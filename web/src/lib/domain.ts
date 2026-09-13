// Wire format: field names are the Go JSON tags.

export type Status = "todo" | "in_progress" | "completed" | "cancelled";
export type Priority = "critical" | "high" | "medium" | "low" | "trivial" | "";
export type TagColor = "green" | "blue" | "magenta" | "cyan" | "";

export const SEPARATOR = "separator";

export interface Task {
    id: string;
    title: string;
    // A separator carries no status, matching the TUI's bare separator rows.
    status: Status | "";
    created_at: string;
    updated_at: string;
    priority?: Priority;
    completion_date?: string;
    estimate_minutes?: number;
    description?: string;
    kind?: string;
    tag_color?: TagColor;
    tag_label?: string;
}

export interface Category {
    id: string;
    name: string;
    created_at: string;
    estimate_minutes?: number;
    tasks: Task[];
}

export interface Project {
    schema?: number;
    id: string;
    name: string;
    created_at: string;
    updated_at: string;
    categories: Category[];
}

export const STATUSES: Status[] = ["todo", "in_progress", "completed", "cancelled"];
export const PRIORITY_ORDER: Priority[] = ["critical", "high", "medium", "low", "trivial"];
export const TAG_COLORS: TagColor[] = ["green", "blue", "magenta", "cyan"];
export const ESTIMATE_PRESETS = [0, 15, 30, 60, 120, 240, 480, 960, 1440, 2400];

export const STATUS_GLYPH: Record<Status, string> = {
    todo: "[ ]",
    in_progress: "[/]",
    completed: "[x]",
    cancelled: "[-]",
};

export const PRIORITY_GLYPH: Record<string, string> = {
    critical: "⇈",
    high: "▲",
    medium: "",
    low: "▼",
    trivial: "⇊",
};

export const PRIORITY_CHIP: Record<string, string> = { ...PRIORITY_GLYPH, medium: "–" };

export function nowTimestamp(): string {
    return new Date().toISOString().replace(/\.\d{3}Z$/, "Z");
}

// Must pass the server's ValidateID: [A-Za-z0-9_-], at most 64 chars.
export function newID(): string {
    const b = crypto.getRandomValues(new Uint8Array(16));
    const hex = Array.from(b, (v) => v.toString(16).padStart(2, "0")).join("");
    return [hex.slice(0, 8), hex.slice(8, 12), hex.slice(12, 16), hex.slice(16, 20), hex.slice(20)].join("-");
}

export function isSeparator(t: Task): boolean {
    return t.kind === SEPARATOR;
}

export function nextStatus(status: Status): Status {
    return STATUSES[(STATUSES.indexOf(status) + 1) % STATUSES.length];
}

export function formatEstimate(minutes: number | undefined): string {
    if (!minutes) return "";
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    if (hours < 8) return `${hours}h`;
    return `${Math.floor(hours / 8)}d`;
}

export function estimateLabel(minutes: number): string {
    return formatEstimate(minutes) || "None";
}

export function statusCounts(p: Project): Record<Status, number> {
    const counts = { todo: 0, in_progress: 0, completed: 0, cancelled: 0 };
    for (const c of p.categories) {
        for (const t of c.tasks) {
            if (isSeparator(t) || t.status === "") continue;
            counts[t.status]++;
        }
    }
    return counts;
}

export function categoryStatus(c: Category): Status | "" {
    const tasks = c.tasks.filter((t) => !isSeparator(t));
    const finished = (t: Task) => t.status === "completed" || t.status === "cancelled";
    if (tasks.length === 0) return "";
    if (tasks.some((t) => t.status === "in_progress")) return "in_progress";
    if (tasks.every(finished)) return "completed";
    if (tasks.some(finished)) return "in_progress";
    return "todo";
}

export function findTask(p: Project, taskID: string): { category: Category; task: Task } | null {
    for (const c of p.categories) {
        const task = c.tasks.find((t) => t.id === taskID);
        if (task) return { category: c, task };
    }
    return null;
}

// Omits what the protocol never carries: project updated_at and schema.
function comparable(p: Project) {
    return {
        id: p.id,
        name: p.name,
        created_at: p.created_at,
        categories: p.categories.map((c) => ({
            id: c.id,
            name: c.name,
            created_at: c.created_at,
            estimate_minutes: c.estimate_minutes ?? 0,
            tasks: c.tasks.map((t) => ({
                id: t.id,
                title: t.title ?? "",
                status: t.status ?? "",
                created_at: t.created_at,
                updated_at: t.updated_at,
                priority: t.priority ?? "",
                completion_date: t.completion_date ?? "",
                estimate_minutes: t.estimate_minutes ?? 0,
                description: t.description ?? "",
                kind: t.kind ?? "",
                tag_color: t.tag_color ?? "",
                tag_label: t.tag_label ?? "",
            })),
        })),
    };
}

export function sameProject(a: Project, b: Project): boolean {
    return JSON.stringify(comparable(a)) === JSON.stringify(comparable(b));
}
