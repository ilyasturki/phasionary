# Project Overview

Phasionary is a single-user task-based project planning web application built with Nuxt 4. Users organize tasks across projects using categories, sections (Current/Future/Past), and automatic priority-based sorting.

# Commands

```bash
bun run dev           # Start dev server (localhost:3000)
bun run build         # Production build
bun run test          # Run tests once
bun run test:watch    # Watch mode testing
bun run db:generate   # Generate Drizzle migration files
bun run db:push       # Apply schema changes to SQLite
bun run db:studio     # Launch web-based DB browser
```

# Architecture

## Tech Stack
- **Frontend**: Nuxt 4, Vue 3 (Composition API), Tailwind CSS 4, Pinia
- **Backend**: Nuxt Server Routes (H3), Drizzle ORM, SQLite
- **Auth**: better-auth with email/password credentials, session-based HTTP-only cookies
- **Validation**: Valibot schemas (frontend and backend)
- **Testing**: Vitest with @vue/test-utils and happy-dom

## Directory Structure
```
app/
  components/     # Vue components (kebab-case)
  composables/    # Composition functions (use-*.ts, auto-imported)
  layouts/        # Page layouts (auth.vue, default.vue)
  middleware/     # Route guards (auth.ts)
  pages/          # File-based routing
  stores/         # Pinia stores (tasks.ts, projects.ts, categories.ts)
server/
  api/            # RESTful endpoints (method-suffixed: [id].get.ts, [id].post.ts)
  db/             # Schema (schema.ts) and database singleton (index.ts)
docs/             # Detailed documentation (read PRODUCT_SPEC.md for features)
```

## Key Patterns

**API Endpoints** (`server/api/`):
- Naming: `index.get.ts`, `index.post.ts`, `[id].put.ts`, `[id].delete.ts`
- Response format: `{ data: {...} }` for success, `{ error: { code, message } }` for errors
- Auth: Check `event.context.session` in handlers
- Resources nested under projects: `/api/projects/[projectId]/tasks/`, `/api/projects/[projectId]/categories/`

**Database Access**:
```typescript
import { getDb, schema } from '../../db'
const db = getDb()
const result = await db.select().from(schema.tasks).where(...)
```

**Pinia Stores**: Auto-imported as `useTasksStore()`, `useProjectsStore()`, `useCategoriesStore()`

**Task Sorting**: Automatic within categories by priority → deadline → time estimate → title

## Database Schema

Main entities: `users`, `projects`, `categories`, `tasks` (plus better-auth tables: `sessions`, `accounts`, `verifications`)

- Tasks have: title, description, deadline, timeEstimateValue/Unit, status (todo/in_progress/completed/cancelled), section (current/future/past), priority (high/medium/low/null), categoryId, projectId
- Cascade deletes: user → projects → categories/tasks

## Environment Variables

```env
NUXT_DATABASE_PATH=./data/app.db
NUXT_SESSION_SECRET=<secret>
NUXT_PUBLIC_APP_URL=http://localhost:3000
```
