import type { Project } from "./domain";
import { newID, nowTimestamp, sameProject } from "./domain";
import type { Draft, Op } from "./ops";

const DB_NAME = "phasionary";
const DB_VERSION = 1;

export interface Device {
    device_id: string;
    name: string;
    token: string;
    cursor: number;
    last_sync?: string;
}

export function request<T>(req: IDBRequest<T>): Promise<T> {
    return new Promise((resolve, reject) => {
        req.onsuccess = () => resolve(req.result);
        req.onerror = () => reject(req.error);
    });
}

function done(tx: IDBTransaction): Promise<void> {
    return new Promise((resolve, reject) => {
        tx.oncomplete = () => resolve();
        tx.onerror = () => reject(tx.error);
        tx.onabort = () => reject(tx.error);
    });
}

let handle: Promise<IDBDatabase> | null = null;

export function open(): Promise<IDBDatabase> {
    if (handle) return handle;
    handle = new Promise((resolve, reject) => {
        const req = indexedDB.open(DB_NAME, DB_VERSION);
        req.onupgradeneeded = () => {
            const db = req.result;
            db.createObjectStore("meta");
            db.createObjectStore("projects", { keyPath: "id" });
            const outbox = db.createObjectStore("outbox", { keyPath: "seq", autoIncrement: true });
            outbox.createIndex("project_id", "project_id");
        };
        req.onsuccess = () => resolve(req.result);
        req.onerror = () => reject(req.error);
    });
    return handle;
}

export async function close(): Promise<void> {
    const open = handle;
    handle = null;
    if (open) (await open).close();
}

async function read<T>(store: string, fn: (s: IDBObjectStore) => IDBRequest<T>): Promise<T> {
    return request(fn((await open()).transaction(store).objectStore(store)));
}

async function write(store: string, fn: (s: IDBObjectStore) => void): Promise<void> {
    const tx = (await open()).transaction(store, "readwrite");
    fn(tx.objectStore(store));
    await done(tx);
}

export async function getDevice(): Promise<Device | null> {
    return (await read<Device | undefined>("meta", (s) => s.get("device"))) ?? null;
}

export function saveDevice(device: Device): Promise<void> {
    return write("meta", (s) => s.put(device, "device"));
}

const byName = new Intl.Collator(undefined, { sensitivity: "base" });

export async function listProjects(): Promise<Project[]> {
    const projects = await read<Project[]>("projects", (s) => s.getAll());
    return projects.sort((a, b) => byName.compare(a.name, b.name) || a.created_at.localeCompare(b.created_at));
}

// Queued ops postdate the push, so this snapshot is stale for the project.
export async function applySnapshot(project: Project): Promise<"written" | "skipped" | "unchanged"> {
    const tx = (await open()).transaction(["projects", "outbox"], "readwrite");
    const queued = await request(tx.objectStore("outbox").index("project_id").count(project.id));
    if (queued > 0) {
        tx.abort();
        return "skipped";
    }
    const current = await request<Project | undefined>(tx.objectStore("projects").get(project.id));
    if (current && sameProject(current, project)) {
        tx.abort();
        return "unchanged";
    }
    tx.objectStore("projects").put(project);
    await done(tx);
    return "written";
}

export function removeProject(id: string): Promise<void> {
    return write("projects", (s) => s.delete(id));
}

export async function commit(project: Project | null, deviceID: string, drafts: Draft[]): Promise<void> {
    const tx = (await open()).transaction(["projects", "outbox"], "readwrite");
    if (project) tx.objectStore("projects").put(project);
    const ts = nowTimestamp();
    const outbox = tx.objectStore("outbox");
    for (const draft of drafts) {
        outbox.put({ op_id: newID(), device_id: deviceID, ts, ...draft });
    }
    await done(tx);
}

export function pendingOps(): Promise<Op[]> {
    return read<Op[]>("outbox", (s) => s.getAll());
}

export function pendingCount(): Promise<number> {
    return read("outbox", (s) => s.count());
}

export function pruneThrough(seq: number): Promise<void> {
    return write("outbox", (s) => s.delete(IDBKeyRange.upperBound(seq)));
}

export async function clearAll(): Promise<void> {
    const tx = (await open()).transaction(["meta", "projects", "outbox"], "readwrite");
    for (const name of ["meta", "projects", "outbox"]) tx.objectStore(name).clear();
    await done(tx);
}
