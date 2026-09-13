import "fake-indexeddb/auto";
import { beforeEach, expect, test, vi } from "vitest";

import * as db from "./db";
import type { Project } from "./domain";
import { addCategory, addTask, createProject } from "./ops";
import { enroll, run } from "./sync";

const DEVICE = { device_id: "dev-1", name: "Test", token: "t", cursor: 0 };

function project(name: string): Project {
    const now = "2030-01-01T00:00:00Z";
    return { id: `p-${name}`, name, created_at: now, updated_at: now, categories: [] };
}

function respond(body: unknown, status = 200) {
    return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

async function stored(id: string): Promise<Project | undefined> {
    return (await db.listProjects()).find((p) => p.id === id);
}

beforeEach(async () => {
    await db.close();
    await db.request(indexedDB.deleteDatabase("phasionary"));
});

test("enrolling stores the device the server minted a token for", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => respond({ token: "secret" })));

    const device = await enroll("K7RD-3PMQ", "Pixel 8");

    expect(device.token).toBe("secret");
    expect(device.name).toBe("Pixel 8");
    expect(await db.getDevice()).toEqual(device);
});

test("a round pushes the outbox, prunes what was acked and writes snapshots", async () => {
    await db.saveDevice(DEVICE);
    const { project: local, drafts } = createProject("Life");
    await db.commit(local, DEVICE.device_id, drafts);
    expect(await db.pendingCount()).toBe(1);

    const pulled = project("Work");
    const fetchMock = vi.fn(async (_path: string, _init: RequestInit) =>
        respond({ acked_through_seq: 1, server_cursor: 7, projects: [{ id: pulled.id, project: pulled }] }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const summary = await run();

    const sent = JSON.parse(fetchMock.mock.calls[0][1].body as string);
    expect(sent.ops).toHaveLength(1);
    expect(sent.ops[0]).toMatchObject({ kind: "project.create", device_id: DEVICE.device_id, seq: 1 });
    expect(summary).toEqual({ pushed: 1, written: 1, unchanged: 0, skipped: 0, removed: 0 });
    expect(await db.pendingCount()).toBe(0);
    expect((await db.getDevice())?.cursor).toBe(7);
    expect((await stored(pulled.id))?.name).toBe("Work");
});

test("a project edited while the round was in flight keeps the local version", async () => {
    await db.saveDevice(DEVICE);
    const { project: local, drafts } = createProject("Life");
    await db.commit(local, DEVICE.device_id, drafts);

    const stale = { ...structuredClone(local), name: "Stale name" };
    vi.stubGlobal(
        "fetch",
        vi.fn(async () => {
            const edited = { ...structuredClone(local), name: "Renamed" };
            await db.commit(edited, DEVICE.device_id, [
                { kind: "project.update", project_id: local.id, fields: { name: "Renamed" } },
            ]);
            return respond({ acked_through_seq: 1, server_cursor: 2, projects: [{ id: local.id, project: stale }] });
        }),
    );

    const summary = await run();

    expect(summary.skipped).toBe(1);
    expect(summary.written).toBe(0);
    expect((await stored(local.id))?.name).toBe("Renamed");
    expect(await db.pendingCount()).toBe(1);
});

test("a snapshot equal to the local copy is not rewritten", async () => {
    await db.saveDevice(DEVICE);
    const { project: local } = createProject("Life");
    const { category } = addCategory(local, "Home");
    addTask(local, category, "Water the plants");
    await db.commit(local, DEVICE.device_id, []);
    vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
            respond({ acked_through_seq: 0, server_cursor: 1, projects: [{ id: local.id, project: local }] }),
        ),
    );

    expect(await run()).toMatchObject({ unchanged: 1, written: 0 });
});

test("a tombstoned project is removed locally", async () => {
    await db.saveDevice(DEVICE);
    const gone = project("Old");
    await db.commit(gone, DEVICE.device_id, []);
    vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
            respond({ acked_through_seq: 0, server_cursor: 3, projects: [{ id: gone.id, deleted: true }] }),
        ),
    );

    expect((await run()).removed).toBe(1);
    expect(await stored(gone.id)).toBeUndefined();
});

test("an unreachable server reads as offline, a rejected token as an error", async () => {
    await db.saveDevice(DEVICE);

    vi.stubGlobal("fetch", vi.fn(async () => {
        throw new TypeError("network");
    }));
    await expect(run()).rejects.toMatchObject({ offline: true });

    vi.stubGlobal("fetch", vi.fn(async () => respond({ error: "unauthorized" }, 401)));
    await expect(run()).rejects.toMatchObject({ name: "SyncError", status: 401, message: "unauthorized" });
});
