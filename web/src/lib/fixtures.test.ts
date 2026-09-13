import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { beforeEach, expect, test, vi } from "vitest";

import { nowTimestamp } from "./domain";
import {
    type Draft,
    type Op,
    addCategory,
    addTask,
    createProject,
    deleteCategory,
    deleteProject,
    deleteTask,
    moveCategory,
    moveTaskToCategory,
    moveTaskWithin,
    renameProject,
    setStatus,
    updateCategory,
    updateTask,
} from "./ops";

const FIXTURE = join(import.meta.dirname, "../../../internal/server/testdata/web-ops.json");
const DEVICE = "fixture-device";

let ids = 0;
beforeEach(() => {
    vi.stubGlobal("crypto", {
        getRandomValues(a: Uint8Array) {
            for (let i = 0; i < a.length; i++) a[i] = (ids * 37 + i * 11) % 256;
            ids++;
            return a;
        },
    });
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2030-01-01T00:00:00Z"));
    return () => vi.useRealTimers();
});

let seq = 0;
const ops: Op[] = [];

// Timestamps must advance per step or last-writer-wins is never exercised.
function record(drafts: Draft[]): void {
    for (const draft of drafts) {
        seq++;
        ops.push({ op_id: `op-${seq}`, device_id: DEVICE, seq, ts: nowTimestamp(), ...draft });
    }
    vi.advanceTimersByTime(1000);
}

function apply<T extends { drafts: Draft[] }>(result: T): T {
    record(result.drafts);
    return result;
}

test("the op vocabulary round-trips through the server", () => {
    const { project: life } = apply(createProject("Life"));

    const { category: home } = apply(addCategory(life, "Home"));
    const { category: health } = apply(addCategory(life, "Health"));
    const { category: scratch } = apply(addCategory(life, "Scratch"));

    const { task: water } = apply(addTask(life, home, "Water the plants"));
    const { task: plumber } = apply(addTask(life, home, "Call the plumber"));
    apply(addTask(life, home, "Reading", true));
    const { task: run } = apply(addTask(life, health, "Run 5 km three times this week"));
    apply(addTask(life, scratch, "Throwaway"));

    record(
        updateTask(life, plumber, {
            status: "in_progress",
            priority: "high",
            estimate_minutes: 120,
            description: "Ask about the boiler service too.",
            tag_color: "blue",
            tag_label: "house",
        }),
    );
    record(setStatus(life, water, "completed"));

    record(moveTaskWithin(life, home, plumber, -1));
    record(moveTaskToCategory(life, health, run, home));

    record(updateCategory(life, health, { name: "Health & sport", estimate_minutes: 480 }));
    record(moveCategory(life, health, -1));

    record(deleteTask(life, home, water));
    record(deleteCategory(life, scratch));
    record(renameProject(life, "Life 2026"));

    const { project: gone } = apply(createProject("Old plans"));
    record(deleteProject(gone));

    const encoded = JSON.stringify({ ops, projects: [life], deleted: [gone.id] }, null, 2) + "\n";

    if (process.env.UPDATE_FIXTURES) {
        mkdirSync(dirname(FIXTURE), { recursive: true });
        writeFileSync(FIXTURE, encoded);
        return;
    }
    expect(readFileSync(FIXTURE, "utf8")).toBe(encoded);
});
