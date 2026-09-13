import { type Category, type Project, type Status, type Task, SEPARATOR, newID, nowTimestamp } from "./domain";

export interface Draft {
    kind: string;
    project_id: string;
    entity_id?: string;
    fields?: Record<string, unknown>;
}

export interface Op extends Draft {
    op_id: string;
    device_id: string;
    seq: number;
    ts: string;
}

function categoryReorder(p: Project, c: Category): Draft {
    return { kind: "category.reorder", project_id: p.id, entity_id: c.id, fields: { task_ids: c.tasks.map((t) => t.id) } };
}

function projectReorder(p: Project): Draft {
    return { kind: "project.reorder", project_id: p.id, fields: { category_ids: p.categories.map((c) => c.id) } };
}

export function createProject(name: string): { project: Project; drafts: Draft[] } {
    const now = nowTimestamp();
    const project: Project = { id: newID(), name, created_at: now, updated_at: now, categories: [] };
    return {
        project,
        drafts: [{ kind: "project.create", project_id: project.id, fields: { name, created_at: now } }],
    };
}

export function renameProject(p: Project, name: string): Draft[] {
    if (p.name === name) return [];
    p.name = name;
    return [{ kind: "project.update", project_id: p.id, fields: { name } }];
}

export function deleteProject(p: Project): Draft[] {
    return [{ kind: "project.delete", project_id: p.id }];
}

export function addCategory(p: Project, name: string): { category: Category; drafts: Draft[] } {
    const now = nowTimestamp();
    const category: Category = { id: newID(), name, created_at: now, tasks: [] };
    p.categories.push(category);
    return {
        category,
        drafts: [
            { kind: "category.create", project_id: p.id, entity_id: category.id, fields: { name, created_at: now } },
            projectReorder(p),
        ],
    };
}

export function updateCategory(p: Project, c: Category, patch: Partial<Category>): Draft[] {
    const fields: Record<string, unknown> = {};
    if (patch.name !== undefined && patch.name !== c.name) fields.name = patch.name;
    if (patch.estimate_minutes !== undefined && patch.estimate_minutes !== (c.estimate_minutes ?? 0)) {
        fields.estimate_minutes = patch.estimate_minutes;
    }
    if (Object.keys(fields).length === 0) return [];
    Object.assign(c, patch);
    return [{ kind: "category.update", project_id: p.id, entity_id: c.id, fields }];
}

export function deleteCategory(p: Project, c: Category): Draft[] {
    p.categories = p.categories.filter((x) => x.id !== c.id);
    // Its tasks go with it on the server, so they need no tombstone of their own.
    return [{ kind: "category.delete", project_id: p.id, entity_id: c.id }, projectReorder(p)];
}

export function moveCategory(p: Project, c: Category, delta: number): Draft[] {
    const i = p.categories.findIndex((x) => x.id === c.id);
    const dst = i + delta;
    if (i < 0 || dst < 0 || dst >= p.categories.length) return [];
    [p.categories[i], p.categories[dst]] = [p.categories[dst], p.categories[i]];
    return [projectReorder(p)];
}

function newTaskFields(t: Task, categoryID: string): Record<string, unknown> {
    const fields: Record<string, unknown> = {
        category_id: categoryID,
        created_at: t.created_at,
        updated_at: t.updated_at,
    };
    for (const key of ["title", "status", "priority", "estimate_minutes", "description", "kind", "tag_color", "tag_label"] as const) {
        const v = t[key];
        if (v !== undefined && v !== "" && v !== 0) fields[key] = v;
    }
    return fields;
}

export function addTask(p: Project, c: Category, title: string, separator = false): { task: Task; drafts: Draft[] } {
    const now = nowTimestamp();
    const task: Task = separator
        ? { id: newID(), title, status: "", kind: SEPARATOR, created_at: now, updated_at: now }
        : { id: newID(), title, status: "todo", created_at: now, updated_at: now };
    c.tasks.push(task);
    return {
        task,
        drafts: [
            { kind: "task.create", project_id: p.id, entity_id: task.id, fields: newTaskFields(task, c.id) },
            categoryReorder(p, c),
        ],
    };
}

const TASK_FIELDS = [
    "title",
    "status",
    "priority",
    "completion_date",
    "estimate_minutes",
    "description",
    "tag_color",
    "tag_label",
] as const;

export function updateTask(p: Project, t: Task, patch: Partial<Task>): Draft[] {
    const fields: Record<string, unknown> = {};
    for (const key of TASK_FIELDS) {
        if (patch[key] === undefined) continue;
        const before = t[key] ?? (key === "estimate_minutes" ? 0 : "");
        const after = patch[key];
        if (before !== after) fields[key] = after;
    }
    if (Object.keys(fields).length === 0) return [];
    Object.assign(t, patch);
    t.updated_at = nowTimestamp();
    fields.updated_at = t.updated_at;
    return [{ kind: "task.update", project_id: p.id, entity_id: t.id, fields }];
}

export function setStatus(p: Project, t: Task, status: Status): Draft[] {
    return updateTask(p, t, {
        status,
        completion_date: status === "completed" ? nowTimestamp() : "",
    });
}

export function deleteTask(p: Project, c: Category, t: Task): Draft[] {
    c.tasks = c.tasks.filter((x) => x.id !== t.id);
    return [{ kind: "task.delete", project_id: p.id, entity_id: t.id }, categoryReorder(p, c)];
}

export function moveTaskWithin(p: Project, c: Category, t: Task, delta: number): Draft[] {
    const i = c.tasks.findIndex((x) => x.id === t.id);
    const dst = i + delta;
    if (i < 0 || dst < 0 || dst >= c.tasks.length) return [];
    [c.tasks[i], c.tasks[dst]] = [c.tasks[dst], c.tasks[i]];
    return [categoryReorder(p, c)];
}

export function moveTaskToCategory(p: Project, from: Category, t: Task, to: Category): Draft[] {
    if (from.id === to.id) return [];
    from.tasks = from.tasks.filter((x) => x.id !== t.id);
    to.tasks.push(t);
    return [
        { kind: "task.move", project_id: p.id, entity_id: t.id, fields: { category_id: to.id } },
        categoryReorder(p, from),
        categoryReorder(p, to),
    ];
}
