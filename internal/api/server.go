// Package api serves the JSON HTTP API that the Phasionary mobile app talks to.
//
// It binds to localhost by default. Set Config.Token to require token auth on
// every request, supplied as an "Authorization: Bearer <token>" header.
package api

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"phasionary/internal/data"
)

type Config struct {
	Addr  string
	Token string
}

type Server struct {
	store data.ProjectRepository
	state data.StateRepository
	cfg   Config
}

// New builds the API server. state backs the fold endpoints and is the same
// state.json the TUI reads, so collapsing a category on the phone collapses it
// in the TUI too; it must not be nil.
func New(store data.ProjectRepository, state data.StateRepository, cfg Config) *Server {
	return &Server{store: store, state: state, cfg: cfg}
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	log.Printf("phasionary serve listening on http://%s (JSON API)", displayAddr(s.cfg.Addr))
	if s.cfg.Token != "" {
		log.Printf("auth enabled; send requests with Authorization: Bearer <token> (token %s)", redact(s.cfg.Token))
	} else {
		log.Printf("auth disabled (no --token); only safe on localhost")
	}

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// IsLoopbackAddr returns true if addr binds to a loopback interface only.
// Accepts forms like "127.0.0.1:7777", "localhost:7777", "[::1]:7777", ":7777"
// (the bare-port form, which binds to all interfaces, is reported as non-loopback).
func IsLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "" {
		return false
	}
	switch strings.ToLower(host) {
	case "localhost":
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}

func displayAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "0.0.0.0" + addr
	}
	return addr
}

// redact masks a secret for logs. Short tokens (under 16 chars) are masked in
// full so the log line doesn't reveal most of the secret; longer tokens show
// the first 2 and last 2 chars as a fingerprint.
func redact(s string) string {
	if len(s) < 16 {
		return strings.Repeat("*", len(s))
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}
