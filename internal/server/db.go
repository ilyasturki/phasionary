package server

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	dbFileName    = "sync.db"
	schemaVersion = 1
)

type DB struct {
	sql *sql.DB
}

// WAL + busy_timeout let the enroll subcommand share the file with a running
// server; one connection serializes this process's writers.
func OpenDB(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}
	dsn := "file:" + filepath.Join(dataDir, dbFileName) +
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.sql.Close() }

// version is set on project rows only: the server cursor at the last request
// that touched the project; pulls filter on it.
const schema = `
CREATE TABLE IF NOT EXISTS meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS devices (
	id           TEXT PRIMARY KEY,
	name         TEXT NOT NULL DEFAULT '',
	token_hash   TEXT NOT NULL UNIQUE,
	created_at   TEXT NOT NULL,
	last_seen_at TEXT NOT NULL DEFAULT '',
	last_seq     INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS enroll_codes (
	code       TEXT PRIMARY KEY,
	expires_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS entities (
	project_id TEXT NOT NULL,
	id         TEXT NOT NULL,
	kind       TEXT NOT NULL,
	fields     TEXT NOT NULL DEFAULT '{}',
	field_ts   TEXT NOT NULL DEFAULT '{}',
	deleted_at TEXT NOT NULL DEFAULT '',
	version    INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (project_id, id)
);
CREATE INDEX IF NOT EXISTS entities_project_version ON entities (kind, version);
`

func (db *DB) migrate() error {
	if _, err := db.sql.Exec(schema); err != nil {
		return err
	}
	var n int
	err := db.sql.QueryRow(`SELECT value FROM meta WHERE key = 'schema'`).Scan(&n)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = db.sql.Exec(`INSERT INTO meta (key, value) VALUES ('schema', ?), ('cursor', '0')`, schemaVersion)
		return err
	case err != nil:
		return err
	}
	if n > schemaVersion {
		return errors.New("database written by a newer phasionary-server; upgrade to open it")
	}
	return nil
}

func (db *DB) begin(ctx context.Context) (*sql.Tx, error) {
	return db.sql.BeginTx(ctx, nil)
}

func readCursor(tx *sql.Tx) (uint64, error) {
	var v uint64
	err := tx.QueryRow(`SELECT value FROM meta WHERE key = 'cursor'`).Scan(&v)
	return v, err
}

func writeCursor(tx *sql.Tx, cursor uint64) error {
	_, err := tx.Exec(`UPDATE meta SET value = ? WHERE key = 'cursor'`, int64(cursor))
	return err
}
