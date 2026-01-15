# Project Planner - Database Schema

## Overview

- **Database**: SQLite (single file)
- **ORM**: Drizzle ORM
- **Auth**: better-auth (manages user/session tables)

---

## Tables

### users
Managed by better-auth library.
- `id` - text, primary key
- `email` - text, unique, not null
- `password_hash` - text, not null
- `created_at` - timestamp
- `updated_at` - timestamp

### projects
- `id` - text, primary key (UUID)
- `name` - text, not null, max 100 chars
- `description` - text, nullable
- `user_id` - text, foreign key → users.id
- `created_at` - timestamp
- `updated_at` - timestamp
- **Constraint**: unique(user_id, name) - case-insensitive

### categories
- `id` - text, primary key (UUID)
- `name` - text, not null
- `project_id` - text, foreign key → projects.id
- `created_at` - timestamp
- **Constraint**: unique(project_id, name) - case-insensitive

### tasks
- `id` - text, primary key (UUID)
- `title` - text, not null, max 200 chars
- `description` - text, nullable
- `deadline` - timestamp, nullable
- `time_estimate_value` - integer, nullable
- `time_estimate_unit` - text, nullable (minutes/hours/days)
- `status` - text, not null (todo/in_progress/completed/cancelled)
- `section` - text, not null (current/future/past)
- `priority` - text, nullable (high/medium/low)
- `position` - integer, not null
- `notes` - text, nullable
- `completion_date` - timestamp, nullable
- `project_id` - text, foreign key → projects.id
- `category_id` - text, foreign key → categories.id
- `created_at` - timestamp
- `updated_at` - timestamp

---

## Relationships

- **users → projects**: one-to-many (user owns projects)
- **projects → categories**: one-to-many (project contains categories)
- **projects → tasks**: one-to-many (project contains tasks)
- **categories → tasks**: one-to-many (category groups tasks)

---

## Cascade Rules

- Delete user → delete all projects
- Delete project → delete all categories and tasks
- Delete category → blocked (reassign tasks first)
