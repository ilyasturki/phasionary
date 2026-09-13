import * as db from "./db";
import type { Device } from "./db";
import { type Project, findTask, nowTimestamp } from "./domain";
import type { Draft } from "./ops";
import { SyncError, type Summary, enroll, run } from "./sync";

const SYNC_DEBOUNCE_MS = 800;

export function href(route: "projects" | "sync" | "tasks", projectID?: string): string {
    if (route === "tasks" && projectID) return `#/p/${projectID}`;
    return route === "sync" ? "#/sync" : "#/";
}

class App {
    device = $state<Device | null>(null);
    projects = $state<Project[]>([]);
    route = $state<"login" | "projects" | "tasks" | "sync">("login");
    projectID = $state<string | null>(null);
    selected = $state<string | null>(null);
    booted = $state(false);

    pending = $state(0);
    syncing = $state(false);
    lastResult = $state<Summary | null>(null);
    // Set by a failed sync, cleared by the next one that succeeds; offline says which kind.
    error = $state<string | null>(null);
    offline = $state(false);

    #timer: ReturnType<typeof setTimeout> | null = null;
    #running = false;
    #again = false;

    project = $derived(this.projects.find((p) => p.id === this.projectID) ?? null);
    selectedTask = $derived(this.project && this.selected ? findTask(this.project, this.selected) : null);
    selectedCategory = $derived(this.project?.categories.find((c) => c.id === this.selected) ?? null);

    async boot(): Promise<void> {
        this.device = await db.getDevice();
        if (this.device) {
            [this.projects, this.pending] = await Promise.all([db.listProjects(), db.pendingCount()]);
            this.#applyHash();
            void this.sync();
        }
        window.addEventListener("hashchange", () => this.device && this.#applyHash());
        this.booted = true;
    }

    // On "sync", projectID stays the project it was opened from.
    #applyHash(): void {
        if (location.hash === "#/sync") {
            this.route = "sync";
            return;
        }
        const projectID = /^#\/p\/([A-Za-z0-9_-]{1,64})$/.exec(location.hash)?.[1] ?? null;
        if (projectID !== this.projectID) this.selected = null;
        this.projectID = projectID;
        this.route = projectID ? "tasks" : "projects";
    }

    go(route: "projects" | "sync" | "tasks", projectID?: string): void {
        location.hash = href(route, projectID);
    }

    async login(code: string, name: string): Promise<void> {
        this.device = await enroll(code, name);
        this.#applyHash();
        await this.sync();
    }

    async logout(): Promise<void> {
        await db.clearAll();
        this.device = null;
        this.projects = [];
        this.projectID = null;
        this.selected = null;
        this.pending = 0;
        this.lastResult = null;
        this.error = null;
        this.offline = false;
        this.syncing = false;
        this.route = "login";
        history.replaceState(null, "", location.pathname + location.search);
    }

    async edit(fn: (project: Project) => Draft[]): Promise<void> {
        const project = this.project;
        if (!project || !this.device) return;
        const copy = $state.snapshot(project) as Project;
        await this.commit(copy, fn(copy));
    }

    async removeProject(project: Project, drafts: Draft[]): Promise<void> {
        if (!this.device) return;
        await db.commit(null, this.device.device_id, drafts);
        await db.removeProject(project.id);
        this.projects = this.projects.filter((p) => p.id !== project.id);
        if (this.projectID === project.id) this.go("projects");
        this.pending += drafts.length;
        this.schedule();
    }

    async commit(project: Project, drafts: Draft[]): Promise<void> {
        if (!this.device || drafts.length === 0) return;
        // Provisional; the next pull replaces it with the server's.
        project.updated_at = nowTimestamp();
        await db.commit(project, this.device.device_id, drafts);
        const i = this.projects.findIndex((p) => p.id === project.id);
        if (i < 0) this.projects = [...this.projects, project];
        else this.projects[i] = project;
        this.pending += drafts.length;
        this.schedule();
    }

    schedule(): void {
        if (this.#timer) clearTimeout(this.#timer);
        this.#timer = setTimeout(() => {
            this.#timer = null;
            void this.sync();
        }, SYNC_DEBOUNCE_MS);
    }

    async sync(): Promise<void> {
        if (!this.device) return;
        if (this.#running) {
            this.#again = true;
            return;
        }
        this.#running = true;
        this.syncing = true;
        try {
            const summary = await run();
            this.lastResult = summary;
            this.error = null;
            this.offline = false;
            this.syncing = false;
            if (summary.written > 0 || summary.removed > 0) {
                this.projects = await db.listProjects();
            }
            this.pending = await db.pendingCount();
            this.device = await db.getDevice();
        } catch (err) {
            this.error = err instanceof SyncError ? err.message : String(err);
            this.offline = !(err instanceof SyncError) || err.offline;
            this.syncing = false;
        } finally {
            this.#running = false;
            if (this.#again) {
                this.#again = false;
                void this.sync();
            }
        }
    }
}

export const app = new App();
