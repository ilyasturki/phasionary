# Phasionary - Tech Stack Document

## Overview

This document defines the technology stack for Phasionary, a task-based project planning web application. The stack prioritizes developer experience, type safety, and performance while maintaining simplicity for a single-user V1 scope.

---

## Runtime & Package Management

- **Node.js** (24.x) - JavaScript runtime
- **Bun** (Latest) - Package manager, script runner

---

## Frontend

### Core Framework

- **Nuxt 4** - Full-stack Vue framework with SSR/SSG capabilities
- **Vue 3** - Composition API, reactive UI components
- **TypeScript** - Type safety across the application

### Styling

- **Tailwind CSS 4** - Utility-first CSS framework

Configuration notes for terminal UI aesthetic:
- Border radius: `0px` globally
- No gradients, solid flat colors
- Monospace typography where appropriate

### State Management

- **Pinia** - Vue's official state management (included with Nuxt)
- **VueUse** - Composition utilities (localStorage, keyboard, etc.)

### UI Components & Interactions

- **shadcn-vue** - Accessible, customizable component library (modify components/styles as needed)
- **@vueuse/core** - Composables for keyboard navigation, focus management

### Forms & Validation

- **Valibot** - Lightweight schema validation (~1KB vs Zod's ~12KB)
- **@vee-validate/nuxt** - Form handling with Valibot integration

### Date & Time

- **Temporal API** - Modern date/time handling (ES2024+)
- **temporal-polyfill** - Polyfill for broader browser support

### Icons

- **@nuxt/icon** - Icon component with Iconify integration
- **Lucide icons** - Clean, terminal-friendly icon set

---

## Backend

### API Layer

- **Nuxt Server Routes** - Built-in API endpoints (`/server/api/`)
- **H3** - Nuxt's underlying HTTP framework
- **Valibot** - Request/response validation (shared with frontend)

### Authentication

- **better-auth** - Modern authentication library
- **bcrypt** - Password hashing
- **Secure cookies** - HTTP-only session tokens

### Database

- **SQLite** - Lightweight relational database, single file
- **better-sqlite3** - Synchronous SQLite driver for Node.js
- **Drizzle ORM** - Type-safe SQL queries, lightweight, excellent DX
- **drizzle-kit** - Database migrations

SQLite is ideal for V1 single-user scope: zero configuration, easy backups (single file), and simple Railway deployment.

---

## Testing

- **Vitest** - Unit and component testing
- **@vue/test-utils** - Vue component testing utilities
- **Playwright** - End-to-end testing
- **@nuxt/test-utils** - Nuxt-specific testing helpers

---

## Code Quality

- **Prettier** - Code formatting
- **TypeScript strict mode** - Type checking
- **Vue TSC** - Vue-specific type checking

> Note: ESLint to be added separately by user.

---

## Infrastructure & Deployment

### Development

- **Bun** - Dev server, script execution, package management
- **SQLite file** - Local database (`data/app.db`)

### Production: Railway

- **Platform** - Railway (PaaS)
- **Build** - `bun run build`
- **Start** - `bun run start` or `node .output/server/index.mjs`
- **Database** - SQLite file with Railway volume for persistence
- **Volume Mount** - `/app/data` for SQLite file persistence

Railway deployment notes:
- Use a Railway volume to persist the SQLite database file across deployments
- Configure `NUXT_DATABASE_PATH` to point to the volume mount
- Single service deployment (Nuxt handles both frontend and API)

### Environment Variables

```env
# Database (SQLite path)
NUXT_DATABASE_PATH=./data/app.db

# Auth
SESSION_SECRET=your-secret-key
COOKIE_SECURE=true

# App
NUXT_PUBLIC_APP_URL=https://your-app.railway.app
```

---

## Accessibility Requirements

- **Keyboard Navigation** - VueUse keyboard composables, focus-trap
- **Screen Reader** - Semantic HTML, ARIA labels via shadcn-vue
- **Color Contrast** - WCAG AA via Tailwind config
- **Focus Indicators** - Custom focus-visible styles

---

## Performance Targets

- **Page Load** (< 3s) - Nuxt SSR, code splitting
- **Task Operations** (< 1s) - Optimistic updates, Pinia
- **Filter Response** (< 500ms) - Client-side filtering

---

## Summary

- **Runtime** - Node.js 24, Bun
- **Framework** - Nuxt 4 + Vue 3
- **Language** - TypeScript
- **Styling** - Tailwind CSS 4
- **UI Components** - shadcn-vue
- **State** - Pinia
- **Validation** - Valibot
- **Database** - SQLite + Drizzle ORM
- **Auth** - better-auth + bcrypt
- **Testing** - Vitest + Playwright
- **Formatting** - Prettier
- **Deployment** - Railway (PaaS)
