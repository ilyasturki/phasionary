package server

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"phasionary/internal/domain"
	"phasionary/internal/syncproto"
)

// A tombstoned or absent project is reported as Deleted.
func snapshot(tx *sql.Tx, projectID string) (syncproto.ProjectSnapshot, error) {
	rows, err := tx.Query(`SELECT id, kind, fields, deleted_at FROM entities WHERE project_id = ?`, projectID)
	if err != nil {
		return syncproto.ProjectSnapshot{}, err
	}
	defer rows.Close()

	var project *entity
	categories := map[string]*entity{}
	tasks := map[string]*entity{}
	for rows.Next() {
		var id, kind, fields, deletedAt string
		if err := rows.Scan(&id, &kind, &fields, &deletedAt); err != nil {
			return syncproto.ProjectSnapshot{}, err
		}
		e := newEntity(kind)
		e.deletedAt = deletedAt
		if err := json.Unmarshal([]byte(fields), &e.fields); err != nil {
			return syncproto.ProjectSnapshot{}, fmt.Errorf("entity %s/%s fields: %w", projectID, id, err)
		}
		switch kind {
		case kindProject:
			project = e
		case kindCategory:
			if deletedAt == "" {
				categories[id] = e
			}
		case kindTask:
			if deletedAt == "" {
				tasks[id] = e
			}
		}
	}
	if err := rows.Err(); err != nil {
		return syncproto.ProjectSnapshot{}, err
	}
	if project == nil || project.deletedAt != "" {
		return syncproto.ProjectSnapshot{ID: projectID, Deleted: true}, nil
	}

	byCategory := map[string]map[string]*entity{}
	for tid, t := range tasks {
		cid, _ := t.fields["category_id"].(string)
		if byCategory[cid] == nil {
			byCategory[cid] = map[string]*entity{}
		}
		byCategory[cid][tid] = t
	}

	p := domain.Project{ID: projectID, Categories: []domain.Category{}}
	if err := decodeFields(project.fields, &p); err != nil {
		return syncproto.ProjectSnapshot{}, err
	}
	p.ID = projectID
	p.UpdatedAt = p.CreatedAt
	for _, cid := range orderedIDs(project.fields["category_ids"], categories) {
		c := domain.Category{Tasks: []domain.Task{}}
		if err := decodeFields(categories[cid].fields, &c); err != nil {
			return syncproto.ProjectSnapshot{}, err
		}
		c.ID = cid
		// category_id wins over any order list: it is the field task.move writes.
		mine := byCategory[cid]
		for _, tid := range orderedIDs(categories[cid].fields["task_ids"], mine) {
			t := domain.Task{}
			if err := decodeFields(mine[tid].fields, &t); err != nil {
				return syncproto.ProjectSnapshot{}, err
			}
			t.ID = tid
			c.Tasks = append(c.Tasks, t)
		}
		p.Categories = append(p.Categories, c)
	}
	domain.StripProjectText(&p)
	return syncproto.ProjectSnapshot{ID: projectID, Project: &p}, nil
}

// Field keys are the domain JSON tags; keys the target type lacks
// (category_id, task_ids) fall away.
func decodeFields(fields map[string]any, target any) error {
	data, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// IDs the winning order list misses are appended by created_at then ID, so
// every replica agrees.
func orderedIDs(order any, live map[string]*entity) []string {
	placed := map[string]bool{}
	var out []string
	if list, ok := order.([]any); ok {
		for _, v := range list {
			id, ok := v.(string)
			if !ok || placed[id] || live[id] == nil {
				continue
			}
			placed[id] = true
			out = append(out, id)
		}
	}
	var rest []string
	for id := range live {
		if !placed[id] {
			rest = append(rest, id)
		}
	}
	slices.SortFunc(rest, func(a, b string) int {
		ca, _ := live[a].fields["created_at"].(string)
		cb, _ := live[b].fields["created_at"].(string)
		return cmp.Or(cmp.Compare(ca, cb), cmp.Compare(a, b))
	})
	return append(out, rest...)
}
