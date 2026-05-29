// Package web serves the htmx-based mobile companion UI for Phasionary.
//
// It binds to localhost by default. Set Config.Token to require token auth
// for any non-static request; the token can be supplied via ?token=… (which
// the middleware exchanges for a session cookie) or Authorization: Bearer.
package web

import (
	"context"
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"phasionary/internal/data"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

type Config struct {
	Addr  string
	Token string
}

type Server struct {
	store data.ProjectRepository
	cfg   Config

	pages    map[string]*template.Template
	partials map[string]*template.Template

	staticHandler http.Handler
}

func New(store data.ProjectRepository, cfg Config) (*Server, error) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}
	s := &Server{
		store:         store,
		cfg:           cfg,
		staticHandler: http.StripPrefix("/static/", http.FileServerFS(sub)),
	}
	if err := s.parseTemplates(); err != nil {
		return nil, err
	}
	return s, nil
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

	log.Printf("phasionary serve listening on http://%s", displayAddr(s.cfg.Addr))
	if s.cfg.Token != "" {
		log.Printf("auth enabled; first visit must supply ?token=… (e.g. http://%s/?token=%s)", displayAddr(s.cfg.Addr), redact(s.cfg.Token))
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
