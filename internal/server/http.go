package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"phasionary/internal/domain"
	"phasionary/internal/syncproto"
)

// A first upload carries every project; 64 MiB is generous for that.
const maxBodyBytes = 64 << 20

// 1s per failure: an 8-char code with a ten-minute life is not brute-forceable
// at one guess per second.
var enrollFailureDelay = time.Second

type Server struct {
	db   *DB
	addr string
}

func New(db *DB, addr string) *Server {
	return &Server{db: db, addr: addr}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST "+syncproto.EnrollPath, s.handleEnroll)
	mux.HandleFunc("POST "+syncproto.SyncPath, s.handleSync)
	return panicMiddleware(mux)
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Printf("phasionary-server listening on http://%s", s.addr)
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) handleEnroll(w http.ResponseWriter, r *http.Request) {
	var req syncproto.EnrollRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if err := domain.ValidateID(req.DeviceID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}
	if err := domain.ValidateLine(req.Name); err != nil {
		writeError(w, http.StatusBadRequest, "invalid device name")
		return
	}
	token := newToken()

	tx, err := s.db.begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database")
		return
	}
	defer tx.Rollback()
	ok, err := consumeCode(tx, req.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database")
		return
	}
	if !ok {
		// Rollback before sleeping: the pool holds one connection.
		_ = tx.Rollback()
		time.Sleep(enrollFailureDelay)
		writeError(w, http.StatusUnauthorized, "invalid or expired enrollment code")
		return
	}
	_, err = tx.Exec(`INSERT INTO devices (id, name, token_hash, created_at) VALUES (?, ?, ?, ?)`,
		req.DeviceID, req.Name, hashToken(token), nowTimestamp())
	if err != nil {
		if isConstraint(err) {
			writeError(w, http.StatusConflict, "device already enrolled")
			return
		}
		writeError(w, http.StatusInternalServerError, "database")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "database")
		return
	}
	log.Printf("enrolled device %s (%s)", req.DeviceID, strconv.Quote(req.Name))
	writeJSON(w, http.StatusOK, syncproto.EnrollResponse{Token: token})
}

type device struct {
	id      string
	lastSeq uint64
}

func (s *Server) authenticate(r *http.Request) (device, bool) {
	h := r.Header.Get("Authorization")
	if len(h) < 8 || !strings.EqualFold(h[:7], "Bearer ") {
		return device{}, false
	}
	var d device
	err := s.db.sql.QueryRowContext(r.Context(),
		`SELECT id, last_seq FROM devices WHERE token_hash = ?`, hashToken(h[7:])).Scan(&d.id, &d.lastSeq)
	if err != nil {
		return device{}, false
	}
	return d, true
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	dev, ok := s.authenticate(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req syncproto.SyncRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.DeviceID != dev.id {
		writeError(w, http.StatusForbidden, "token belongs to another device")
		return
	}
	var prevSeq uint64
	for i, op := range req.Ops {
		if err := validateOp(op); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if op.DeviceID != dev.id {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("op %s: journaled by another device", op.OpID))
			return
		}
		if i > 0 && op.Seq <= prevSeq {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("op %s: seq not ascending", op.OpID))
			return
		}
		prevSeq = op.Seq
	}

	tx, err := s.db.begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database")
		return
	}
	defer tx.Rollback()
	resp, err := s.syncLocked(tx, dev, req)
	if err != nil {
		log.Printf("sync for device %s: %v", dev.id, err)
		writeError(w, http.StatusInternalServerError, "database")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "database")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) syncLocked(tx *sql.Tx, dev device, req syncproto.SyncRequest) (syncproto.SyncResponse, error) {
	cursor, err := readCursor(tx)
	if err != nil {
		return syncproto.SyncResponse{}, err
	}
	// A lost response makes the device resend merged ops; skipping them is the
	// idempotency.
	acked := dev.lastSeq
	touched := map[string]bool{}
	for _, op := range req.Ops {
		if op.Seq <= acked {
			continue
		}
		if err := applyOp(tx, op); err != nil {
			return syncproto.SyncResponse{}, err
		}
		touched[op.ProjectID] = true
		acked = op.Seq
	}
	if len(touched) > 0 {
		cursor++
		if err := writeCursor(tx, cursor); err != nil {
			return syncproto.SyncResponse{}, err
		}
		for pid := range touched {
			if err := bumpProject(tx, pid, cursor); err != nil {
				return syncproto.SyncResponse{}, err
			}
		}
	}
	if _, err := tx.Exec(`UPDATE devices SET last_seq = ?, last_seen_at = ? WHERE id = ?`,
		acked, nowTimestamp(), dev.id); err != nil {
		return syncproto.SyncResponse{}, err
	}

	rows, err := tx.Query(`SELECT id FROM entities WHERE kind = ? AND version > ? ORDER BY version, id`,
		kindProject, req.Cursor)
	if err != nil {
		return syncproto.SyncResponse{}, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return syncproto.SyncResponse{}, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return syncproto.SyncResponse{}, err
	}
	resp := syncproto.SyncResponse{AckedThroughSeq: acked, ServerCursor: cursor, Projects: []syncproto.ProjectSnapshot{}}
	for _, id := range ids {
		snap, err := snapshot(tx, id)
		if err != nil {
			return syncproto.SyncResponse{}, err
		}
		resp.Projects = append(resp.Projects, snap)
	}
	return resp, nil
}

func decodeBody(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}

func isConstraint(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "constraint")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, syncproto.ErrorResponse{Error: msg})
}

func panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Quote: the path is attacker-controlled.
				log.Printf("panic %s %s: %v\n%s", r.Method, strconv.Quote(r.URL.Path), rec, debug.Stack())
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
