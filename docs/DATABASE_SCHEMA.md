# Phasionary - Data Schema (JSON)

## Overview
- Storage: Multiple JSON files (one per project)
- Single-user, no auth tables
- Store timestamps in UTC ISO 8601 strings
- Directory auto-created on startup if missing

## Storage Locations
- Linux: `~/.local/share/phasionary/`
- Override with `PHASIONARY_DATA_PATH`

## Directory Structure
```
~/.local/share/phasionary/
├── {project-uuid-1}.json
├── {project-uuid-2}.json
└── ...
```

## File Schema (per project)
```json
{
  "id": "uuid",
  "name": "Project name",
  "created_at": "2026-01-22T10:00:00Z",
  "updated_at": "2026-01-22T10:00:00Z",
  "categories": [
    {
      "id": "uuid",
      "name": "Category name",
      "created_at": "2026-01-22T10:00:00Z",
      "tasks": [
        {
          "id": "uuid",
          "title": "Task title",
          "status": "todo",
          "section": "current",
          "created_at": "2026-01-22T10:00:00Z",
          "updated_at": "2026-01-22T10:00:00Z"
        }
      ]
    }
  ]
}
```

## Optional Fields
- Projects: `description`
- Categories: none
- Tasks: `description`, `deadline`, `time_estimate_value`, `time_estimate_unit`, `priority`, `notes`, `completion_date`

## Constraints
- Project names are unique (case-insensitive)
- Category names are unique within a project (case-insensitive)
- `status` is `todo`, `in_progress`, `completed`, or `cancelled`
- `section` is `current`, `future`, or `past`
- `time_estimate_unit` is `minutes`, `hours`, or `days`

## Relationships (Nested Hierarchy)
- Projects contain categories
- Categories contain tasks
- No foreign key IDs needed (implied by nesting)

## Cascade Rules
- Delete project → deletes entire file (including all categories and tasks)
- Delete category → deletes all nested tasks
