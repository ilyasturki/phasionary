# Phasionary - Design System

Terminal-inspired aesthetic: sharp edges, flat colors, monospace typography. No gradients, no rounded corners.

Supports both **light** and **dark** themes. Default to system preference via `prefers-color-scheme`, with optional user override stored in localStorage.

---

## Colors

### Backgrounds

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `bg-base` | `#fafafa` | `#0a0a0a` | Page background |
| `bg-surface` | `#f5f5f5` | `#141414` | Cards, panels |
| `bg-elevated` | `#ffffff` | `#1f1f1f` | Modals, dropdowns |
| `bg-muted` | `#e5e5e5` | `#262626` | Hover states, subtle emphasis |

### Text

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `text-primary` | `#0a0a0a` | `#fafafa` | Headings, important content |
| `text-secondary` | `#525252` | `#a1a1a1` | Body text, descriptions |
| `text-muted` | `#a1a1a1` | `#737373` | Placeholders, disabled |

### Accent

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `accent` | `#16a34a` | `#22c55e` | Primary actions, links, focus |
| `accent-hover` | `#15803d` | `#16a34a` | Hover state |

### Status

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `success` | `#16a34a` | `#22c55e` | Completed, positive |
| `warning` | `#ca8a04` | `#eab308` | Deadlines, caution |
| `error` | `#dc2626` | `#ef4444` | Overdue, destructive |
| `info` | `#2563eb` | `#3b82f6` | In progress, informational |

### Priority

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `priority-high` | `#dc2626` | `#ef4444` | High priority indicator |
| `priority-medium` | `#ca8a04` | `#eab308` | Medium priority indicator |
| `priority-low` | `#2563eb` | `#3b82f6` | Low priority indicator |

---

## Typography

### Font Families

| Token | Value | Usage |
|-------|-------|-------|
| `font-mono` | `'JetBrains Mono', 'Fira Code', monospace` | UI elements, code |
| `font-sans` | `system-ui, sans-serif` | Long-form text (optional) |

### Font Sizes

| Token | Size | Line Height |
|-------|------|-------------|
| `text-xs` | 12px | 16px |
| `text-sm` | 14px | 20px |
| `text-base` | 16px | 24px |
| `text-lg` | 18px | 28px |
| `text-xl` | 20px | 28px |
| `text-2xl` | 24px | 32px |

### Font Weights

| Token | Value | Usage |
|-------|-------|-------|
| `font-normal` | 400 | Body text |
| `font-medium` | 500 | Emphasis |
| `font-semibold` | 600 | Headings |

---

## Spacing

Base unit: **4px**

| Token | Value | Usage |
|-------|-------|-------|
| `space-1` | 4px | Tight gaps |
| `space-2` | 8px | Default gap |
| `space-3` | 12px | Related elements |
| `space-4` | 16px | Section padding |
| `space-6` | 24px | Card padding |
| `space-8` | 32px | Section margins |
| `space-12` | 48px | Large separations |

---

## Borders

| Token | Light | Dark |
|-------|-------|------|
| `radius` | `0px` | `0px` |
| `border-width` | `1px` | `1px` |
| `border-color` | `#e5e5e5` | `#262626` |
| `border-focus` | `#16a34a` | `#22c55e` |

---

## Theme Implementation

Use Tailwind CSS 4's `@theme` layer to define CSS custom properties. Apply theme-specific values using `prefers-color-scheme` media queries.

```css
@theme {
  --color-bg-base: #fafafa;
  --color-text-primary: #0a0a0a;
  /* ... */
}

@media (prefers-color-scheme: dark) {
  :root {
    --color-bg-base: #0a0a0a;
    --color-text-primary: #fafafa;
  }
}
```

For user-controlled theme switching, store preference in `localStorage` and apply a `.dark` class to `<html>`.
