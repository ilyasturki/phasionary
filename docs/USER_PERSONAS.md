# Project Planner - User Personas

## Overview

This document defines the primary user personas for Project Planner. These personas guide product decisions, feature prioritization, and UX design choices. Each persona represents a real user archetype that benefits from a simple, focused task management tool.

### Why These Personas?

Project Planner targets **individuals who work solo** and need project-based task organization without the overhead of team collaboration features. The V1 single-user scope deliberately excludes team-oriented users in favor of those who value simplicity and personal productivity.

---

## Primary Personas

### 1. Maya Chen — Freelance Designer

#### Background

Maya is a 32-year-old freelance UI/UX designer based in Austin, Texas. She left her agency job three years ago to work independently. She typically juggles 3-5 client projects simultaneously, ranging from brand identity work to full product design systems. She works from home and occasionally from coffee shops.

#### Demographics

| Attribute | Detail |
|-----------|--------|
| Role | Freelance UI/UX Designer |
| Experience | 8 years in design, 3 years freelancing |
| Tech Comfort | High — uses Figma, design tools daily |
| Work Style | Project-based, deadline-driven |

#### A Day in Maya's Life

Maya starts her morning reviewing what's due this week. She has a brand guide revision due Wednesday, three app screens due Friday, and client feedback to address on an ongoing project. She needs to quickly see which project needs attention today and track her hours for invoicing. By afternoon, a client emails with an urgent change request — she needs to add it to her list without losing track of her planned work.

#### Goals

- See all active work across clients at a glance
- Track deadlines without complex setup
- Estimate time for accurate client quotes
- Move quickly between planning and doing

#### Frustrations

- **Notion fatigue**: Spent hours building a "perfect system" that became too complex to maintain
- **Context switching**: Most tools require too many clicks to add a simple task
- **Feature bloat**: Doesn't need Gantt charts, team permissions, or integrations she'll never use
- **Visual noise**: Bright colors and busy interfaces feel distracting, not helpful

#### Why Project Planner Works for Maya

The terminal-inspired aesthetic feels focused and professional. She can create a project per client, use categories for task types (Deliverable, Revision, Admin), and quickly see what's due. No team features to configure, no complex views to set up — just tasks organized the way she thinks.

---

### 2. Daniel Kowalski — Solo Developer

#### Background

Daniel is a 28-year-old software developer working remotely for a mid-size company during the day. Evenings and weekends, he builds side projects — currently a CLI tool for developers and a small SaaS app he hopes to launch. He's methodical about his personal projects, treating them with the same rigor as his day job.

#### Demographics

| Attribute | Detail |
|-----------|--------|
| Role | Software Developer (employed) + Side Project Builder |
| Experience | 5 years professional, multiple shipped side projects |
| Tech Comfort | Very high — lives in terminal and IDE |
| Work Style | Systematic, breaks work into small tasks |

#### A Day in Daniel's Life

After work, Daniel has about two hours for his side project. He opens his task manager to see what he committed to this week. Tonight he's fixing a bug a user reported and adding a small feature. He updates the bug to "In Progress," works through it, marks it complete, then moves the feature task to "In Progress." Before closing his laptop, he quickly adds two tasks he thought of during the day to his Future backlog.

#### Goals

- Track bugs, features, and improvements separately
- See what's in progress vs. what's planned
- Maintain velocity on side projects despite limited time
- Keep backlog organized without project management overhead

#### Frustrations

- **Linear/Jira overkill**: Built for teams, too heavy for personal use
- **GitHub Issues limitations**: Okay for public repos, awkward for personal planning
- **Todo apps too simple**: Todoist doesn't understand "projects" the way a developer does
- **Context loss**: Switching between work tools and personal tools is jarring

#### Why Project Planner Works for Daniel

The developer-friendly aesthetic (terminal UI, monospace fonts, sharp corners) feels native. He can use the default categories (Feature, Fix, Research) or customize them. Sections map to his mental model: Current = this week's sprint, Future = backlog, Past = shipped work. No integrations needed — it's a focused space for personal project planning.

---

### 3. Priya Sharma — Independent Consultant

#### Background

Priya is a 41-year-old independent business consultant who left corporate strategy two years ago. She runs her own consulting practice, advising small businesses on operations and growth. She manages client engagements, her own business operations (invoicing, marketing, admin), and occasionally hires subcontractors for specific deliverables.

#### Demographics

| Attribute | Detail |
|-----------|--------|
| Role | Independent Business Consultant |
| Experience | 15 years in strategy, 2 years independent |
| Tech Comfort | Medium-high — comfortable with business tools |
| Work Style | Client-focused, balances delivery with business development |

#### A Day in Priya's Life

Priya's day involves client calls, deliverable work, and business operations. In the morning, she reviews her task list to prepare for a client meeting at 10am. She has a proposal due Friday for a prospective client and needs to follow up on two invoices. After her meeting, she adds three action items that came out of the discussion. She uses priority levels to ensure business development tasks don't get lost under client work.

#### Goals

- Separate client projects from business operations
- Never miss a deliverable deadline
- Track business development alongside client work
- Quick capture during or after meetings

#### Frustrations

- **CRM/PM tool mismatch**: CRMs focus on sales pipeline, not task execution
- **Spreadsheet sprawl**: Started with a "simple" spreadsheet that grew unmanageable
- **Mobile friction**: Needs to add tasks on phone between meetings, most apps are clunky
- **Over-engineering**: Tried Asana, spent more time configuring than using

#### Why Project Planner Works for Priya

She creates a project per client plus one for "Business Operations." Categories help distinguish deliverables from admin work. The responsive design means she can add tasks from her phone during a coffee break. The clean interface shows her what matters without overwhelming her with features she doesn't need.

---

## Competitive Landscape

Project Planner occupies a specific niche: **project-based task management for individuals**. Here's how it compares to alternatives and why users might choose it.

### Positioning Map

| Tool | Target User | Strength | Gap Project Planner Fills |
|------|-------------|----------|---------------------------|
| **Todoist** | General consumers | Simple, fast capture | No project structure, limited organization |
| **Notion** | Power users, teams | Flexible, all-in-one | Overcomplicated for task-focused workflows |
| **Linear** | Dev teams | Beautiful, fast | Team-centric, overkill for individuals |
| **Trello** | Visual thinkers | Intuitive boards | Board fatigue, weak filtering, team-focused |
| **Jira** | Enterprise teams | Comprehensive | Heavy, slow, requires team context |
| **Things 3** | Apple users | Polished, personal | Mac/iOS only, no web access |
| **TickTick** | General productivity | Feature-rich | Busy interface, too many features |

### Why Users Leave These Tools

#### From Todoist
> "I outgrew flat lists. I need to see tasks grouped by project and category, not just by date."

Todoist excels at quick capture but lacks true project hierarchy. Users managing multiple workstreams find themselves creating workarounds (labels, filters) that become unwieldy.

#### From Notion
> "I spent more time building my system than using it. I just want to track tasks, not build a database."

Notion's flexibility is a double-edged sword. Users often over-engineer their setup, then abandon it when maintenance becomes a chore.

#### From Linear/Jira
> "I don't need sprints, story points, or team workflows. I just need to know what to work on today."

Team-oriented tools carry overhead that solo users don't need. The mental model (tickets, epics, sprints) doesn't match personal productivity.

#### From Trello
> "Dragging cards was fun at first, but now I have 50 cards and no clear picture of what's urgent."

Board-based tools work for some workflows but struggle with prioritization and deadline tracking across many items.

### Project Planner's Position

**Simple project-based task management with a developer-friendly aesthetic.**

- More structured than Todoist (projects, categories, sections)
- Simpler than Notion (no databases, no configuration rabbit holes)
- Individual-focused unlike Linear/Jira (no team features)
- List-based unlike Trello (clear priorities, better filtering)
- Web-based unlike Things 3 (accessible anywhere)

---

## Key Takeaways

### Design Principles from Personas

| Principle | Derived From |
|-----------|--------------|
| **Speed over features** | Maya needs quick capture between client calls |
| **Clarity over flexibility** | Daniel wants a clear system, not infinite customization |
| **Focus over comprehensiveness** | Priya left tools that did too much |
| **Terminal aesthetic** | Appeals to Daniel's developer sensibility, signals "focused tool" |

### V1 Priorities

Based on these personas, V1 should emphasize:

1. **Fast task creation** — Minimal friction to add tasks
2. **Clear project separation** — Switch contexts quickly
3. **Deadline visibility** — What's due soon, what's overdue
4. **Section-based workflow** — Current/Future/Past maps to how users think
5. **Responsive design** — Priya needs mobile capture, Maya works from laptop

### What to Avoid

- Team/collaboration features (not our users)
- Complex customization (Notion trap)
- Integrations as a core value prop (keep it focused)
- Gamification or social features (professional users)
