# Phasionary - Product Specification

## 1. Overview / Product Vision

Phasionary is a task-based project planning web application designed for individuals who need a straightforward way to organize, track, and complete work items across projects.

### Core Value Proposition

- **Simplicity First**: A focused to-do list experience without unnecessary complexity
- **Flexible Organization**: User-defined categories that adapt to any workflow
- **Clear Progress Tracking**: Visual status management from inception to completion

### Platform

Web application accessible via modern browsers, built with responsive design for desktop, tablet, and mobile use. V1 is single-user only (no shared workspaces or collaboration).

---

## 2. Target Users & Use Cases

### Primary Users

| User Type | Description |
|-----------|-------------|
| **Product Managers** | Track feature requests, prioritize backlogs, monitor progress |
| **Developers** | Manage bug fixes, technical improvements, and feature work |
| **Freelancers** | Organize client projects with deadlines and time estimates |
| **Independent Operators** | Manage multiple personal or client projects without overhead |

### Core Use Cases

1. **Personal Task Management**: Individual users maintaining their own task lists across multiple projects and categories
2. **Work Planning**: Breaking down larger initiatives into trackable tasks with time estimates
3. **Deadline Tracking**: Managing time-sensitive deliverables with due dates
4. **Progress Visibility**: Understanding what's in progress, what's pending, and what's complete

---

## 3. Core Concepts & Entities

### 3.1 Projects

A **Project** is the top-level container for tasks and categories.

| Attribute | Required | Description |
|-----------|----------|-------------|
| **Name** | Yes | Project name (max 100 characters) |
| **Description** | No | Optional summary or notes for the project |

#### Project Behavior

| Aspect | Behavior |
|--------|----------|
| **Multi-Project** | Each user can create and manage multiple projects |
| **Scoped Data** | Tasks and categories belong to exactly one project |
| **Active Project** | The UI always operates within a single selected project |
| **Default Project** | A starter project is created at account setup |
| **Unique Names** | Project names are unique per user (case-insensitive) |

### 3.2 Tasks

A **Task** is the fundamental unit of work in Phasionary.

| Attribute | Required | Description |
|-----------|----------|-------------|
| **Project** | Yes | Parent project that owns the task |
| **Title** | Yes | Brief name or summary of the task (max 200 characters) |
| **Description** | No | Extended details, notes, or context for the task |
| **Deadline** | No | Target completion date; optionally includes time |
| **Time Estimate** | No | Expected effort required (value + unit) |
| **Category** | Yes | Classification from the project's category list |
| **Status** | Yes | Current state of the task |
| **Section** | Yes | Which temporal phase: Current, Future, or Past |
| **Position** | Yes | Order within section/category (determines display order) |
| **Priority** | No | Explicit priority level: High, Medium, or Low |
| **Notes** | No | Free-text annotations, links, or references |
| **Completion Date** | No | Auto-set when task is marked Completed |

#### Time Estimate Units

Users select from the following units when estimating task duration:

- Minutes
- Hours
- Days

Example: "2 hours", "30 minutes", "3 days"

#### Deadline Rules

- Date-only deadlines are due at the end of that date in the user's local timezone.
- Overdue calculations use the user's local timezone.

#### Priority Levels

| Level | Description |
|-------|-------------|
| **High** | Urgent or critical tasks |
| **Medium** | Standard priority |
| **Low** | Can be deferred if needed |
| *(None)* | No explicit priority assigned (default) |

Priority works alongside position ordering: users can both drag-and-drop to reorder and assign explicit priority labels.

### 3.3 Categories

A **Category** is a project-defined label used to classify and organize tasks.

| Aspect | Behavior |
|--------|----------|
| **Project-Defined** | Categories are created, renamed, and deleted within a project |
| **Single Assignment** | Each task belongs to exactly one category |
| **No Limit** | No restriction on number of categories a project can have |

#### Default Categories

Each new project receives the following pre-populated categories (all editable/removable):

- **Feature** - New functionality or capabilities
- **Fix** - Bug fixes and corrections
- **Ergonomy** - User experience improvements
- **Documentation** - Documentation and guides
- **Research** - Investigation and exploration tasks

### 3.4 Status

A **Status** represents the current state of a task in the workflow.

| Status | Description |
|--------|-------------|
| **To Do** | Task has not been started |
| **In Progress** | Task is actively being worked on |
| **Completed** | Task has been finished |
| **Cancelled** | Task was abandoned or is no longer needed |

#### Status Transition Rules

- **Free Transitions**: Any status can change to any other status
- No enforced linear workflow
- Users may move tasks backward (e.g., Completed → In Progress) as needed

#### Terminal Status Behavior

| Behavior | Description |
|----------|-------------|
| **Auto-Archive Option** | Completed and Cancelled tasks can auto-move to Past section (per-user setting, default: on) |
| **Visual Distinction** | Cancelled tasks display with strikethrough styling |
| **Completion Date** | Set automatically when task moves to Completed status |

#### Completion Date Rules

- Completion Date is cleared if a task leaves Completed status.

### 3.5 Sections

A **Section** organizes tasks by temporal phase, representing where work sits in its lifecycle.

| Section | Description |
|---------|-------------|
| **Current** | Active work - tasks being worked on now or in the immediate focus |
| **Future** | Planned work - backlog items for later |
| **Past** | Archived view - typically completed or cancelled tasks |

#### Section Behavior

| Aspect | Behavior |
|--------|----------|
| **Fixed Sections** | Current, Future, and Past are system-defined (cannot be renamed or deleted) |
| **Manual Movement** | Users can move tasks between sections; moving to Past requires Completed or Cancelled status |
| **Auto-Archive** | Optionally, Completed/Cancelled tasks auto-move to Past |
| **Category Grouping** | Within each section, tasks are displayed grouped by category |
| **Default View** | Current section is shown on app load |
| **Past Behavior** | Past contains only Completed/Cancelled tasks; non-terminal tasks moved to Past must first be marked Completed or Cancelled |

When a task is reopened (status changed from Completed/Cancelled to another status), it moves to Current by default.

---

## 4. Key Features

### 4.1 User Authentication

| Feature | Description |
|---------|-------------|
| **Account Creation** | Users register with email and password |
| **Login/Logout** | Secure session management |
| **Single-User Scope** | Accounts are private; no sharing or collaboration in v1 |

### 4.2 Project Management

| Action | Behavior |
|--------|----------|
| **Create Project** | Add a new project with a unique name |
| **Rename Project** | Change an existing project's name |
| **Delete Project** | Remove project and all its tasks/categories (with confirmation) |
| **Switch Project** | Change the active project context |

### 4.3 Category Management

| Action | Behavior |
|--------|----------|
| **Create Category** | Add a new category with a unique name (per project) |
| **Rename Category** | Change an existing category's display name |
| **Delete Category** | Remove category; tasks must be reassigned to another category |
| **View Categories** | List all categories in the active project including task counts |

#### Constraints

- Category names must be unique per project (case-insensitive)
- Empty categories can be deleted without warning
- Categories with tasks require reassignment before deletion
- Each project must always have at least one category; deleting the last category is not allowed

### 4.4 Task Management

| Action | Behavior |
|--------|----------|
| **Create Task** | Add new task with required fields |
| **Edit Task** | Modify any task attribute |
| **Delete Task** | Remove task permanently (with confirmation) |
| **Change Category** | Reassign task to a different category in the same project |
| **Update Status** | Change task status freely between any values |
| **Move to Section** | Move task between Current, Future, and Past |
| **Reorder Tasks** | Drag-and-drop to change position within section/category |
| **Set Priority** | Assign High, Medium, or Low priority label |

#### Task Creation Requirements

When creating a task, users must provide:
- Title
- Category (within the active project)

Optional fields:
- Deadline (date, optionally time)
- Time estimate (value and unit)

Defaults for new tasks:
- Status: **"To Do"**
- Section: **"Current"**
- Priority: *(None)*
- Position: End of category list

#### Ordering Rules

- Drag-and-drop sets a task's position within its section and category.
- Moving a task to a different section or category places it at the end of the destination list unless the user drops it at a specific position.
- Priority labels are informational only and do not auto-sort tasks.

### 4.5 Task Views & Filtering

#### Primary Navigation: Sections

| Section Tab | Description |
|-------------|-------------|
| **Current** | Active work and immediate focus (default view) |
| **Future** | Planned backlog items |
| **Past** | Archived completed and cancelled tasks |

#### Project Context

| Element | Description |
|---------|-------------|
| **Project Selector** | Switches the active project and reloads its tasks/categories |

#### Within-Section Display

| Aspect | Behavior |
|--------|----------|
| **Category Grouping** | Tasks are grouped under category headers |
| **Position Ordering** | Tasks appear in user-defined order (drag-and-drop) |

#### Filters (Applied Within Any Section)

| Filter | Description |
|--------|-------------|
| **By Category** | Show only tasks from a specific category |
| **By Status** | Show only tasks with a specific status |
| **By Priority** | Show only tasks with a specific priority level |
| **Overdue** | Tasks past their deadline (not Completed/Cancelled); applies only to Current/Future |

Filters apply only to the active project's tasks.

#### Default View

- Section: Current
- Project: Last active project (or default on first login)
- Grouping: By category
- Sort: By position within category

---

## 5. User Flows

### 5.1 Account Setup

```
1. User navigates to application
2. User selects "Create Account"
3. User enters email and password
4. Account is created with a default project and categories
5. User is logged in and sees empty task list in the default project
```

### 5.2 Managing Projects

#### Creating a Project

```
1. User opens project switcher
2. User selects "New Project"
3. User enters project name
4. System validates uniqueness
5. Project is created with default categories
6. Project becomes the active context
```

#### Switching Projects

```
1. User opens project switcher
2. User selects another project
3. UI loads tasks and categories for the selected project
```

#### Deleting a Project

```
1. User opens project settings
2. User selects "Delete Project"
3. System displays confirmation with task count
4. User confirms deletion
5. Project and all associated tasks/categories are removed
```

### 5.3 Managing Categories

#### Creating a Category

```
1. User navigates to category management
2. User selects "Add Category"
3. User enters category name
4. System validates uniqueness
5. Category is created and available for task assignment
```

#### Renaming a Category

```
1. User navigates to category management
2. User selects existing category
3. User chooses "Rename"
4. User enters new name
5. System validates uniqueness
6. Category name is updated (tasks remain associated)
```

#### Deleting a Category

```
1. User navigates to category management
2. User selects category to delete
3. If category has tasks:
   a. System requires selecting a replacement category
   b. User confirms deletion
   c. Tasks are reassigned to the selected category
4. If category is empty:
   a. Category is deleted immediately
5. Category is removed from list
```

### 5.4 Managing Tasks

#### Creating a Task

```
1. User selects "New Task"
2. User confirms active project
3. User enters title
4. User optionally enters description
5. User sets deadline (date picker, optional time)
6. User enters time estimate (value + unit dropdown)
7. User selects category from dropdown
8. User saves task
9. Task is created with status "To Do"
```

#### Editing a Task

```
1. User selects existing task
2. User modifies desired fields
3. User saves changes
4. Task is updated
```

#### Changing Task Status

```
1. User views task (list or detail view)
2. User selects new status from available options
3. Status is updated immediately
```

#### Deleting a Task

```
1. User selects task to delete
2. System prompts for confirmation
3. User confirms
4. Task is permanently removed
```

### 5.5 Daily Workflow Example

```
1. User logs in
2. User reviews "To Do" tasks
3. User moves priority task to "In Progress"
4. User works on task
5. User marks task "Completed"
6. User creates new tasks as needed
7. User logs out
```

### 5.6 Working with Sections

#### Moving Tasks Through Phases

```
1. User opens app (Current section shown by default)
2. User sees tasks grouped by category
3. User drags tasks to reorder priority within category
4. User completes task (auto-moves to Past, or manual move)
5. User switches to Future tab to review backlog
6. User moves a Future task to Current to start work on it
```

#### Planning Session

```
1. User navigates to Future section
2. User creates new tasks for upcoming work
3. User organizes tasks by category
4. User sets priorities (High/Medium/Low)
5. User moves highest priority items to Current section
```

#### Review and Cleanup

```
1. User navigates to Past section
2. User reviews recently completed tasks
3. User identifies any tasks that need to be reopened
4. User moves reopened tasks back to Current
5. User optionally deletes old completed tasks
```

---

## 6. UX & Non-Functional Guidelines

### 6.1 Responsive Design

| Device | Expectation |
|--------|-------------|
| **Desktop** | Full-featured experience, multi-column layouts |
| **Tablet** | Adapted layouts, touch-optimized controls |
| **Mobile** | Single-column, essential features prioritized |

### 6.2 Visual Style

| Aspect | Guideline |
|--------|-----------|
| **Aesthetic** | Terminal UI inspired design |
| **Colors** | Solid, flat colors without gradients |
| **Border Radius** | None (0px) - sharp, square corners throughout |
| **Typography** | Clean, monospace-friendly presentation |

### 6.3 Accessibility

| Requirement | Description |
|-------------|-------------|
| **Keyboard Navigation** | All features accessible via keyboard |
| **Screen Reader Support** | Proper ARIA labels and semantic HTML |
| **Color Contrast** | WCAG AA compliance for text legibility |
| **Focus Indicators** | Visible focus states for interactive elements |

### 6.4 Performance Expectations

| Interaction | Target |
|-------------|--------|
| **Page Load** | Under 3 seconds on standard connection |
| **Task Operations** | Under 1 second response time |
| **Category Operations** | Under 1 second response time |
| **Filter** | Results displayed within 500ms |

### 6.5 Data & Reliability

| Aspect | Expectation |
|--------|-------------|
| **Data Persistence** | All user data saved reliably |
| **Session Management** | Users remain logged in across browser sessions (optional) |
| **Error Handling** | Clear error messages with recovery guidance |
| **Data Validation** | Input validation with helpful feedback |

---

## Appendix: Glossary

| Term | Definition |
|------|------------|
| **Project** | A container that owns tasks and categories |
| **Task** | A unit of work with title, deadline, estimate, project, category, and status |
| **Category** | A project-defined label for organizing tasks |
| **Status** | The current workflow state of a task (To Do, In Progress, Completed, Cancelled) |
| **Time Estimate** | Expected duration to complete a task |
| **Deadline** | Target date (and optionally time) for task completion |
