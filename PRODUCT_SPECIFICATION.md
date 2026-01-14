# Project Planner - Product Specification

## 1. Overview / Product Vision

Project Planner is a task-based project management web application designed for individuals and teams who need a straightforward way to organize, track, and complete work items.

### Core Value Proposition

- **Simplicity First**: A focused to-do list experience without unnecessary complexity
- **Flexible Organization**: User-defined categories that adapt to any workflow
- **Clear Progress Tracking**: Visual status management from inception to completion

### Platform

Web application accessible via modern browsers, built with responsive design for desktop, tablet, and mobile use.

---

## 2. Target Users & Use Cases

### Primary Users

| User Type | Description |
|-----------|-------------|
| **Product Managers** | Track feature requests, prioritize backlogs, monitor progress |
| **Developers** | Manage bug fixes, technical improvements, and feature work |
| **Freelancers** | Organize client projects with deadlines and time estimates |
| **Small Teams** | Coordinate work items without enterprise-level overhead |

### Core Use Cases

1. **Personal Task Management**: Individual users maintaining their own task lists across multiple project categories
2. **Work Planning**: Breaking down larger initiatives into trackable tasks with time estimates
3. **Deadline Tracking**: Managing time-sensitive deliverables with due dates
4. **Progress Visibility**: Understanding what's in progress, what's pending, and what's complete

---

## 3. Core Concepts & Entities

### 3.1 Tasks

A **Task** is the fundamental unit of work in Project Planner.

| Attribute | Required | Description |
|-----------|----------|-------------|
| **Title** | Yes | Brief name or summary of the task (max 200 characters) |
| **Description** | No | Extended details, notes, or context for the task |
| **Deadline** | No | Target completion date; optionally includes time |
| **Time Estimate** | No | Expected effort required (value + unit) |
| **Category** | Yes | Classification from user's category list |
| **Status** | Yes | Current state of the task |
| **Section** | Yes | Which temporal phase: Current, Future, or Past |
| **Position** | Yes | Order within section/category (determines display order) |
| **Priority** | No | Explicit priority level: High, Medium, or Low |
| **Parent Task** | No | Reference to parent task (for subtasks only) |
| **Notes** | No | Free-text annotations, links, or references |
| **Completion Date** | No | Auto-set when task is marked Completed |

#### Time Estimate Units

Users select from the following units when estimating task duration:

- Minutes
- Hours
- Days

Example: "2 hours", "30 minutes", "3 days"

#### Subtasks

Tasks support one level of nesting via subtasks:

| Aspect | Behavior |
|--------|----------|
| **Depth Limit** | One level only (parent → children, no deeper nesting) |
| **Inheritance** | Subtasks inherit parent's section and category |
| **Independence** | Subtasks have their own status, deadline, and estimate |
| **Progress Display** | Parent task shows aggregate progress (e.g., "3/5 subtasks done") |
| **Completion** | Parent can only be marked Completed when all subtasks are Completed/Cancelled |

#### Priority Levels

| Level | Description |
|-------|-------------|
| **High** | Urgent or critical tasks |
| **Medium** | Standard priority |
| **Low** | Can be deferred if needed |
| *(None)* | No explicit priority assigned (default) |

Priority works alongside position ordering: users can both drag-and-drop to reorder and assign explicit priority labels.

### 3.2 Categories

A **Category** is a user-defined label used to classify and organize tasks.

| Aspect | Behavior |
|--------|----------|
| **User-Defined** | Users create, rename, and delete categories freely |
| **Single Assignment** | Each task belongs to exactly one category |
| **No Limit** | No restriction on number of categories a user can create |

#### Default Categories

New users receive the following pre-populated categories (all editable/removable):

- **Feature** - New functionality or capabilities
- **Fix** - Bug fixes and corrections
- **Ergonomy** - User experience improvements
- **Documentation** - Documentation and guides
- **Research** - Investigation and exploration tasks

#### System Category: Uncategorized

- A protected system category named **"Uncategorized"** exists for all users
- Cannot be renamed or deleted
- Receives tasks when their original category is deleted

### 3.3 Status

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
| **Auto-Archive Option** | Completed and Cancelled tasks can auto-move to Past section (user setting) |
| **Visual Distinction** | Cancelled tasks display with strikethrough styling |
| **Completion Date** | Set automatically when task moves to Completed status |

### 3.4 Sections

A **Section** organizes tasks by temporal phase, representing where work sits in its lifecycle.

| Section | Description |
|---------|-------------|
| **Current** | Active work - tasks being worked on now or in the immediate focus |
| **Future** | Planned work - backlog items for later |
| **Past** | Archived work - completed or cancelled tasks |

#### Section Behavior

| Aspect | Behavior |
|--------|----------|
| **Fixed Sections** | Current, Future, and Past are system-defined (cannot be renamed or deleted) |
| **Manual Movement** | Users can move tasks between sections freely |
| **Auto-Archive** | Optionally, Completed/Cancelled tasks auto-move to Past |
| **Category Grouping** | Within each section, tasks are displayed grouped by category |
| **Default View** | Current section is shown on app load |

---

## 4. Key Features

### 4.1 User Authentication

| Feature | Description |
|---------|-------------|
| **Account Creation** | Users register with email and password |
| **Login/Logout** | Secure session management |
| **Personal Workspace** | Each user has private access to their own tasks and categories |

### 4.2 Category Management

| Action | Behavior |
|--------|----------|
| **Create Category** | Add a new category with a unique name |
| **Rename Category** | Change an existing category's display name |
| **Delete Category** | Remove category; all associated tasks move to "Uncategorized" |
| **View Categories** | List all user categories including task counts |

#### Constraints

- Category names must be unique per user (case-insensitive)
- "Uncategorized" is reserved and cannot be created, renamed, or deleted
- Empty categories can be deleted without warning
- Categories with tasks prompt confirmation before deletion

### 4.3 Task Management

| Action | Behavior |
|--------|----------|
| **Create Task** | Add new task with required fields |
| **Edit Task** | Modify any task attribute |
| **Delete Task** | Remove task permanently (with confirmation) |
| **Change Category** | Reassign task to a different category |
| **Update Status** | Change task status freely between any values |
| **Move to Section** | Move task between Current, Future, and Past |
| **Reorder Tasks** | Drag-and-drop to change position within section/category |
| **Set Priority** | Assign High, Medium, or Low priority label |
| **Add Subtask** | Create a child task under an existing task |

#### Task Creation Requirements

When creating a task, users must provide:
- Title
- Deadline (date, optionally time)
- Time estimate (value and unit)
- Category (selected from existing categories)

Defaults for new tasks:
- Status: **"To Do"**
- Section: **"Current"**
- Priority: *(None)*
- Position: End of category list

#### Subtask Creation

When creating a subtask:
- User selects a parent task
- Subtask inherits parent's section and category
- Subtask requires its own title, deadline, and time estimate
- Status defaults to "To Do"

### 4.4 Task Views & Filtering

#### Primary Navigation: Sections

| Section Tab | Description |
|-------------|-------------|
| **Current** | Active work and immediate focus (default view) |
| **Future** | Planned backlog items |
| **Past** | Archived completed and cancelled tasks |

#### Within-Section Display

| Aspect | Behavior |
|--------|----------|
| **Category Grouping** | Tasks are grouped under category headers |
| **Position Ordering** | Tasks appear in user-defined order (drag-and-drop) |
| **Subtask Display** | Subtasks appear indented under their parent |

#### Filters (Applied Within Any Section)

| Filter | Description |
|--------|-------------|
| **By Category** | Show only tasks from a specific category |
| **By Status** | Show only tasks with a specific status |
| **By Priority** | Show only tasks with a specific priority level |
| **Overdue** | Tasks past their deadline (not Completed/Cancelled) |

#### Default View

- Section: Current
- Grouping: By category
- Sort: By position within category

---

## 5. User Flows

### 5.1 Account Setup

```
1. User navigates to application
2. User selects "Create Account"
3. User enters email and password
4. Account is created with default categories
5. User is logged in and sees empty task list
```

### 5.2 Managing Categories

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
   a. System displays confirmation with task count
   b. User confirms deletion
   c. Tasks are moved to "Uncategorized"
4. If category is empty:
   a. Category is deleted immediately
5. Category is removed from list
```

### 5.3 Managing Tasks

#### Creating a Task

```
1. User selects "New Task"
2. User enters title
3. User optionally enters description
4. User sets deadline (date picker, optional time)
5. User enters time estimate (value + unit dropdown)
6. User selects category from dropdown
7. User saves task
8. Task is created with status "To Do"
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

### 5.4 Daily Workflow Example

```
1. User logs in
2. User reviews "To Do" tasks
3. User moves priority task to "In Progress"
4. User works on task
5. User marks task "Completed"
6. User creates new tasks as needed
7. User logs out
```

### 5.5 Working with Sections

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

### 6.2 Accessibility

| Requirement | Description |
|-------------|-------------|
| **Keyboard Navigation** | All features accessible via keyboard |
| **Screen Reader Support** | Proper ARIA labels and semantic HTML |
| **Color Contrast** | WCAG AA compliance for text legibility |
| **Focus Indicators** | Visible focus states for interactive elements |

### 6.3 Performance Expectations

| Interaction | Target |
|-------------|--------|
| **Page Load** | Under 3 seconds on standard connection |
| **Task Operations** | Under 1 second response time |
| **Category Operations** | Under 1 second response time |
| **Search/Filter** | Results displayed within 500ms |

### 6.4 Data & Reliability

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
| **Task** | A unit of work with title, deadline, estimate, category, and status |
| **Category** | A user-defined label for organizing tasks |
| **Status** | The current workflow state of a task (To Do, In Progress, Completed) |
| **Time Estimate** | Expected duration to complete a task |
| **Deadline** | Target date (and optionally time) for task completion |
| **Uncategorized** | System category for tasks whose original category was deleted |
