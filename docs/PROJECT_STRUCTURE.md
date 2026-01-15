# Phasionary - Project Structure

## Overview

This document describes the directory structure for Phasionary, a Nuxt 4 full-stack application. The structure follows Nuxt 4 conventions with the `app/` directory for client-side code and `server/` directory for backend logic.

---

## Directory Tree

```
project-planner/
├── app/                          # Client-side application code
│   ├── assets/css/               # Global styles and Tailwind @theme config
│   ├── components/               # Reusable Vue components
│   ├── composables/              # Reusable composition functions
│   ├── layouts/                  # Page layout templates
│   ├── middleware/               # Route middleware for navigation guards
│   ├── pages/                    # File-based routing
│   ├── plugins/                  # Vue plugins initialized at startup
│   ├── stores/                   # Pinia state stores
│   └── app.vue                   # Root Vue component
├── server/                       # Server-side API and database
│   ├── api/                      # RESTful API endpoints (H3)
│   ├── db/                       # Database (Drizzle ORM + SQLite)
│   ├── middleware/               # Server middleware
│   └── utils/                    # Server-side utilities
├── public/                       # Static assets served at root URL
├── docs/                         # Project documentation
├── nuxt.config.ts                # Nuxt configuration
├── package.json                  # Dependencies and scripts
├── tsconfig.json                 # TypeScript configuration
└── .env                          # Environment variables
```

---

## File Naming Conventions

### General

- **kebab-case** for all file names: `task-card.vue`, `use-auth.ts`, `category-manager.vue`

### API Routes

Nuxt server routes use method suffixes:

- `index.get.ts` — GET
- `index.post.ts` — POST
- `[id].get.ts` — GET with dynamic param
- `[id].put.ts` — PUT with dynamic param
- `[id].delete.ts` — DELETE with dynamic param

### Pages

- `index.vue` — Index route for directory
- `[id].vue` — Dynamic route parameter
- `[...slug].vue` — Catch-all route

### Components

- kebab-case naming: `task-card.vue`, `project-selector.vue`
- Colocate related components in subdirectories if needed

### Composables

- kebab-case with `use-` prefix: `use-auth.ts`, `use-tasks.ts`
- Auto-imported by Nuxt from `composables/` directory
