import * as db from "./db";
import type { Device } from "./db";
import { type Project, newID } from "./domain";

export class SyncError extends Error {
    constructor(
        message: string,
        readonly status = 0,
    ) {
        super(message);
        this.name = "SyncError";
    }

    // 0: fetch threw; 502–504: the proxy's upstream is down.
    get offline(): boolean {
        return this.status === 0 || this.status === 502 || this.status === 503 || this.status === 504;
    }
}

export interface Summary {
    pushed: number;
    written: number;
    unchanged: number;
    skipped: number;
    removed: number;
}

interface SyncResponse {
    acked_through_seq: number;
    server_cursor: number;
    projects: { id: string; deleted?: boolean; project?: Project }[];
}

async function post<T>(path: string, token: string, body: unknown): Promise<T> {
    const resp = await fetch(path, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(body),
    }).catch(() => {
        throw new SyncError("the server is unreachable");
    });
    const raw = await resp.text();
    if (!resp.ok) {
        let message = `the server answered ${resp.status}`;
        try {
            const parsed = JSON.parse(raw) as { error?: string };
            if (parsed.error) message = parsed.error;
        } catch {}
        throw new SyncError(message, resp.status);
    }
    return JSON.parse(raw) as T;
}

export async function enroll(code: string, name: string): Promise<Device> {
    const device_id = newID();
    const { token } = await post<{ token: string }>("/v1/enroll", "", { code, device_id, name });
    const device: Device = { device_id, name, token, cursor: 0 };
    await db.saveDevice(device);
    return device;
}

export async function run(): Promise<Summary> {
    const device = await db.getDevice();
    if (!device) throw new SyncError("this device is not enrolled");

    const ops = await db.pendingOps();
    const resp = await post<SyncResponse>("/v1/sync", device.token, {
        device_id: device.device_id,
        cursor: device.cursor,
        ops,
    });

    const summary: Summary = { pushed: ops.length, written: 0, unchanged: 0, skipped: 0, removed: 0 };
    await db.pruneThrough(resp.acked_through_seq);

    for (const snap of resp.projects) {
        if (snap.deleted) {
            await db.removeProject(snap.id);
            summary.removed++;
            continue;
        }
        if (!snap.project) continue;
        summary[await db.applySnapshot(snap.project)]++;
    }

    device.cursor = resp.server_cursor;
    device.last_sync = new Date().toISOString();
    await db.saveDevice(device);
    return summary;
}
