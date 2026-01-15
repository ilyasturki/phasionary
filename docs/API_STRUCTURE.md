# Phasionary - API Structure

## Overview

All API endpoints are served under `/api/` and return JSON responses. Authentication is required for all endpoints except auth routes.

### Conventions

- **Base path**: `/api`
- **Format**: JSON request/response bodies
- **Auth**: Session-based via cookies (managed by better-auth)
- **IDs**: UUID strings

---

## Authentication

Authentication is handled by [better-auth](https://www.better-auth.com/). Standard endpoints are available under `/api/auth/*`.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/sign-up` | Create new account |
| POST | `/api/auth/sign-in` | Log in |
| POST | `/api/auth/sign-out` | Log out |
| GET | `/api/auth/session` | Get current session |

Refer to better-auth documentation for complete endpoint reference.

---

## Projects

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/projects` | List all projects for current user |
| POST | `/api/projects` | Create a new project |
| GET | `/api/projects/[id]` | Get single project |
| PUT | `/api/projects/[id]` | Update project |
| DELETE | `/api/projects/[id]` | Delete project and all contents |

---

## Categories

Categories are scoped to a project.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/projects/[projectId]/categories` | List categories in project |
| POST | `/api/projects/[projectId]/categories` | Create category |
| PUT | `/api/projects/[projectId]/categories/[id]` | Update category |
| DELETE | `/api/projects/[projectId]/categories/[id]` | Delete category (requires task reassignment) |

---

## Tasks

Tasks are scoped to a project.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/projects/[projectId]/tasks` | List tasks (supports filtering via query params) |
| POST | `/api/projects/[projectId]/tasks` | Create task |
| GET | `/api/projects/[projectId]/tasks/[id]` | Get single task |
| PUT | `/api/projects/[projectId]/tasks/[id]` | Update task |
| DELETE | `/api/projects/[projectId]/tasks/[id]` | Delete task |
| PATCH | `/api/projects/[projectId]/tasks/[id]/status` | Update task status |
| PATCH | `/api/projects/[projectId]/tasks/[id]/section` | Move task to section |

### Task Filters

The `GET /api/projects/[projectId]/tasks` endpoint supports query parameters:

| Parameter | Values | Description |
|-----------|--------|-------------|
| `section` | `current`, `future`, `past` | Filter by section |
| `status` | `todo`, `in_progress`, `completed`, `cancelled` | Filter by status |
| `category` | `<category-id>` | Filter by category |
| `priority` | `high`, `medium`, `low` | Filter by priority |

---

## Response Patterns

### Success

```json
{
  "data": { ... }
}
```

### Error

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Project not found"
  }
}
```

### HTTP Status Codes

| Code | Usage |
|------|-------|
| 200 | Success |
| 201 | Created |
| 400 | Validation error |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not found |
| 500 | Server error |
