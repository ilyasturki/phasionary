package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"
)

// No 0/O/1/I: the code is read off one screen and typed on another.
const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const (
	codeLen = 8
	CodeTTL = 10 * time.Minute
)

func nowTimestamp() string { return time.Now().UTC().Format(time.RFC3339) }

func (db *DB) NewEnrollCode(ctx context.Context) (string, error) {
	raw := make([]byte, codeLen)
	rand.Read(raw)
	var b strings.Builder
	for i, r := range raw {
		if i == codeLen/2 {
			b.WriteByte('-')
		}
		b.WriteByte(codeAlphabet[int(r)%len(codeAlphabet)])
	}
	code := b.String()
	expires := time.Now().UTC().Add(CodeTTL).Format(time.RFC3339)
	tx, err := db.begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM enroll_codes WHERE expires_at <= ?`, nowTimestamp()); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`INSERT INTO enroll_codes (code, expires_at) VALUES (?, ?)`, code, expires); err != nil {
		return "", err
	}
	return code, tx.Commit()
}

// Accepts the code however it was typed: any case, with or without the dash.
func normalizeCode(input string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(input) {
		if r >= 'A' && r <= 'Z' || r >= '2' && r <= '9' {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if len(s) != codeLen {
		return ""
	}
	return s[:codeLen/2] + "-" + s[codeLen/2:]
}

// Single-use by construction: the delete is the check.
func consumeCode(tx *sql.Tx, input string) (bool, error) {
	code := normalizeCode(input)
	if code == "" {
		return false, nil
	}
	res, err := tx.Exec(`DELETE FROM enroll_codes WHERE code = ? AND expires_at > ?`, code, nowTimestamp())
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func newToken() string { return rand.Text() }

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
