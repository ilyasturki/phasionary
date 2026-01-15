# Phasionary - Product Roadmap and Testing Strategy

This roadmap is derived from the existing project documentation and keeps the V1 scope to a single-user product

---

## Product Roadmap (Iterative)

### Iteration 0: Foundations

**Scope**
- Nuxt 4 app shell and base layouts
- Tailwind tokens applied per design system (flat colors, 0 radius, monospace)
- Pinia setup and shared composables
- Drizzle + SQLite config and migrations
- better-auth wiring
- API response conventions

**Exit criteria**
- App boots reliably
- Auth endpoints reachable
- DB migrations run without errors
- UI shell matches terminal aesthetic
- Tests are added and run after this iteration

---

### Iteration 1a: Core MVP - Auth and Bootstrap

**Scope**
- Sign up / sign in / sign out
- Session validation and route protection
- Default project + categories on signup

**Exit criteria**
- Default data seeded on first account creation
- Authenticated routes enforce session checks
- Tests are added and run after this iteration

---

### Iteration 1b: Core MVP - Projects and Categories

**Scope**
- Project CRUD + project switcher
- Category CRUD with constraints

**Exit criteria**
- Project switching reloads scoped data
- Category constraints enforced (unique, reassignment on delete)
- Tests are added and run after this iteration

---

### Iteration 1c: Core MVP - Tasks and Views

**Scope**
- Task CRUD
- Status and section updates
- Current/Future/Past views grouped by category
- Automatic ordering (priority -> deadline -> estimate -> title)

**Exit criteria**
- End-to-end flow: create tasks -> complete -> see in Past
- Task ordering matches spec in every section
- Tests are added and run after this iteration

---

### Iteration 2: Workflow Polish

**Scope**
- Filters (category/status/priority/overdue)
- Deadlines with time and end-of-day rules
- Time estimates (value + unit)
- Priority UI
- Auto-archive toggle (default on)
- Confirmation dialogs and validation messaging
- Responsive layout refinements

**Exit criteria**
- Planning workflow works across Current/Future/Past
- Filters produce correct results
- Validation errors are clear and consistent
- Tests are added and run after this iteration

---

### Iteration 3: Release Hardening

**Scope**
- Accessibility checks and keyboard navigation
- Performance targets (page load <3s, task ops <1s, filters <500ms)
- Empty states and error recovery flows
- UI polish and consistency pass

**Exit criteria**
- E2E suite green
- A11y checks pass for critical flows
- No critical UX gaps for primary personas
- Tests are added and run after this iteration
