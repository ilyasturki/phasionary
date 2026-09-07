package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"phasionary/internal/domain"
	"phasionary/internal/journal"
)

const (
	kindProject  = "project"
	kindCategory = "category"
	kindTask     = "task"
)

var knownKinds = map[string]bool{
	journal.KindProjectCreate: true, journal.KindProjectUpdate: true,
	journal.KindProjectDelete: true, journal.KindProjectReorder: true,
	journal.KindCategoryCreate: true, journal.KindCategoryUpdate: true,
	journal.KindCategoryDelete: true, journal.KindCategoryReorder: true,
	journal.KindTaskCreate: true, journal.KindTaskUpdate: true,
	journal.KindTaskDelete: true, journal.KindTaskMove: true,
}

type entity struct {
	kind      string
	fields    map[string]any
	fieldTS   map[string]string
	deletedAt string
}

func newEntity(kind string) *entity {
	return &entity{kind: kind, fields: map[string]any{}, fieldTS: map[string]string{}}
}

func loadEntity(tx *sql.Tx, projectID, id string) (*entity, error) {
	var kind, fields, fieldTS, deletedAt string
	err := tx.QueryRow(`SELECT kind, fields, field_ts, deleted_at FROM entities WHERE project_id = ? AND id = ?`,
		projectID, id).Scan(&kind, &fields, &fieldTS, &deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e := newEntity(kind)
	e.deletedAt = deletedAt
	if err := json.Unmarshal([]byte(fields), &e.fields); err != nil {
		return nil, fmt.Errorf("entity %s/%s fields: %w", projectID, id, err)
	}
	if err := json.Unmarshal([]byte(fieldTS), &e.fieldTS); err != nil {
		return nil, fmt.Errorf("entity %s/%s field_ts: %w", projectID, id, err)
	}
	return e, nil
}

func saveEntity(tx *sql.Tx, projectID, id string, e *entity) error {
	fields, err := json.Marshal(e.fields)
	if err != nil {
		return err
	}
	fieldTS, err := json.Marshal(e.fieldTS)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO entities (project_id, id, kind, fields, field_ts, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (project_id, id) DO UPDATE SET
			fields = excluded.fields, field_ts = excluded.field_ts, deleted_at = excluded.deleted_at`,
		projectID, id, e.kind, string(fields), string(fieldTS), e.deletedAt)
	return err
}

// validateOp rejects what would corrupt a device: IDs become filenames there.
func validateOp(op journal.Op) error {
	if !knownKinds[op.Kind] {
		return fmt.Errorf("op %s: unknown kind %q", op.OpID, op.Kind)
	}
	if err := domain.ValidateID(op.ProjectID); err != nil {
		return fmt.Errorf("op %s: project id: %w", op.OpID, err)
	}
	if !strings.HasPrefix(op.Kind, kindProject+".") {
		if err := domain.ValidateID(op.EntityID); err != nil {
			return fmt.Errorf("op %s: entity id: %w", op.OpID, err)
		}
	}
	return nil
}

func applyOp(tx *sql.Tx, op journal.Op) error {
	kind, verb, _ := strings.Cut(op.Kind, ".")
	id := op.EntityID
	if kind == kindProject {
		id = op.ProjectID
	}
	e, err := loadEntity(tx, op.ProjectID, id)
	if err != nil {
		return err
	}
	if e == nil {
		e = newEntity(kind)
	}
	switch {
	case verb == "delete":
		if e.deletedAt == "" {
			e.deletedAt = op.Timestamp
		}
	// Tombstoned: drop the op. A stale peer must never resurrect a deletion.
	case e.deletedAt != "":
	default:
		for k, v := range op.Fields {
			// >= not >: RFC3339 seconds tie constantly, so later arrival wins.
			if op.Timestamp >= e.fieldTS[k] {
				e.fields[k] = v
				e.fieldTS[k] = op.Timestamp
			}
		}
	}
	return saveEntity(tx, op.ProjectID, id, e)
}

// Insert fallback: a project whose create op never arrived still gets a row,
// so pulls see it.
func bumpProject(tx *sql.Tx, projectID string, version uint64) error {
	res, err := tx.Exec(`UPDATE entities SET version = ? WHERE project_id = ? AND id = ? AND kind = ?`,
		version, projectID, projectID, kindProject)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n > 0 {
		return err
	}
	_, err = tx.Exec(`INSERT INTO entities (project_id, id, kind, version) VALUES (?, ?, ?, ?)`,
		projectID, projectID, kindProject, version)
	return err
}
