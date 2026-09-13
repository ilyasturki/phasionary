<script lang="ts">
    import { SvelteSet } from "svelte/reactivity";
    import CategorySheet from "../lib/CategorySheet.svelte";
    import Menu, { type Action } from "../lib/Menu.svelte";
    import Prompt from "../lib/Prompt.svelte";
    import SyncIndicator from "../lib/SyncIndicator.svelte";
    import TaskSheet from "../lib/TaskSheet.svelte";
    import { app } from "../lib/app.svelte";
    import {
        PRIORITY_GLYPH,
        PRIORITY_ORDER,
        STATUS_GLYPH,
        TAG_COLORS,
        type Category,
        type Priority,
        type Project,
        type Status,
        type TagColor,
        type Task,
        categoryStatus,
        findTask,
        formatEstimate,
        isSeparator,
        nextStatus,
    } from "../lib/domain";
    import {
        addCategory,
        addTask,
        deleteCategory,
        deleteTask,
        moveCategory,
        moveTaskToCategory,
        moveTaskWithin,
        setStatus,
        updateCategory,
        updateTask,
    } from "../lib/ops";

    let { project }: { project: Project } = $props();

    type Overlay =
        | { kind: "none" }
        | { kind: "task"; id: string }
        | { kind: "category"; id: string }
        | { kind: "newTask"; categoryID: string; separator: boolean }
        | { kind: "newCategory" }
        | { kind: "menu" }
        | { kind: "moveTo"; id: string }
        | { kind: "confirm"; label: string; run: () => void };

    let overlay = $state<Overlay>({ kind: "none" });
    const collapsed = new SvelteSet<string>();

    // Menu runs the action before onclose; an action that opened another overlay must survive it.
    function close(replaced: Overlay["kind"]) {
        return () => {
            if (overlay.kind === replaced) overlay = { kind: "none" };
        };
    }

    const selectedTask = $derived(app.selectedTask);
    const selectedCategory = $derived(app.selectedCategory);
    const target = $derived(selectedCategory ?? selectedTask?.category ?? project.categories[0] ?? null);

    function tapTask(t: Task) {
        if (app.selected === t.id) overlay = { kind: "task", id: t.id };
        else app.selected = t.id;
    }

    function tapCategory(c: Category) {
        if (app.selected !== c.id) app.selected = c.id;
        else if (collapsed.has(c.id)) collapsed.delete(c.id);
        else collapsed.add(c.id);
    }

    // app.edit hands a snapshot; resolve by id inside it, never the live object.
    function withTask(id: string, fn: (p: Project, c: Category, t: Task) => ReturnType<typeof updateTask>) {
        return app.edit((p) => {
            const found = findTask(p, id);
            return found ? fn(p, found.category, found.task) : [];
        });
    }

    function withCategory(id: string, fn: (p: Project, c: Category) => ReturnType<typeof updateCategory>) {
        return app.edit((p) => {
            const c = p.categories.find((x) => x.id === id);
            return c ? fn(p, c) : [];
        });
    }

    function cycleStatus(t: Task) {
        void withTask(t.id, (p, _c, task) => setStatus(p, task, nextStatus((task.status || "todo") as Status)));
    }

    function cyclePriority(t: Task) {
        const order: Priority[] = ["", ...PRIORITY_ORDER];
        const next = order[(order.indexOf(t.priority ?? "") + 1) % order.length];
        void withTask(t.id, (p, _c, task) => updateTask(p, task, { priority: next }));
    }

    function cycleTag(t: Task) {
        const order: TagColor[] = ["", ...TAG_COLORS];
        const next = order[(order.indexOf(t.tag_color ?? "") + 1) % order.length];
        void withTask(t.id, (p, _c, task) =>
            updateTask(p, task, { tag_color: next, ...(next === "" ? { tag_label: "" } : {}) }),
        );
    }

    async function createTask(title: string) {
        const o = overlay;
        if (o.kind !== "newTask") return;
        overlay = { kind: "none" };
        await withCategory(o.categoryID, (p, c) => {
            const { task, drafts } = addTask(p, c, title, o.separator);
            app.selected = task.id;
            return drafts;
        });
    }

    async function createCategory(name: string) {
        overlay = { kind: "none" };
        await app.edit((p) => {
            const { category, drafts } = addCategory(p, name);
            app.selected = category.id;
            return drafts;
        });
    }

    function removeTask(t: Task) {
        overlay = {
            kind: "confirm",
            label: `Delete “${t.title || "separator"}”?`,
            run: () => {
                if (app.selected === t.id) app.selected = null;
                void withTask(t.id, (p, c, task) => deleteTask(p, c, task));
            },
        };
    }

    function removeCategory(c: Category) {
        const n = c.tasks.length;
        overlay = {
            kind: "confirm",
            label: `Delete “${c.name}” and its ${n} ${n === 1 ? "task" : "tasks"}?`,
            run: () => {
                if (app.selected === c.id) app.selected = null;
                void withCategory(c.id, (p, cat) => deleteCategory(p, cat));
            },
        };
    }

    function taskActions(t: Task, c: Category): Action[] {
        const i = c.tasks.findIndex((x) => x.id === t.id);
        return [
            { label: "Move up", disabled: i <= 0, run: () => withTask(t.id, (p, cat, task) => moveTaskWithin(p, cat, task, -1)) },
            {
                label: "Move down",
                disabled: i < 0 || i >= c.tasks.length - 1,
                run: () => withTask(t.id, (p, cat, task) => moveTaskWithin(p, cat, task, 1)),
            },
            {
                label: "Move to category…",
                disabled: project.categories.length < 2,
                run: () => (overlay = { kind: "moveTo", id: t.id }),
            },
            { label: "New separator here", run: () => (overlay = { kind: "newTask", categoryID: c.id, separator: true }) },
            { label: "Delete task", danger: true, run: () => removeTask(t) },
        ];
    }

    function categoryActions(c: Category): Action[] {
        const i = project.categories.findIndex((x) => x.id === c.id);
        return [
            { label: "New task", run: () => (overlay = { kind: "newTask", categoryID: c.id, separator: false }) },
            { label: "Edit category", run: () => (overlay = { kind: "category", id: c.id }) },
            { label: "Move up", disabled: i <= 0, run: () => withCategory(c.id, (p, cat) => moveCategory(p, cat, -1)) },
            {
                label: "Move down",
                disabled: i >= project.categories.length - 1,
                run: () => withCategory(c.id, (p, cat) => moveCategory(p, cat, 1)),
            },
            { label: "Delete category", danger: true, run: () => removeCategory(c) },
        ];
    }

    const menuActions = $derived.by((): Action[] => {
        if (selectedTask) return taskActions(selectedTask.task, selectedTask.category);
        if (selectedCategory) return categoryActions(selectedCategory);
        return [{ label: "New category", run: () => (overlay = { kind: "newCategory" }) }];
    });

    const menuTitle = $derived(
        selectedTask ? selectedTask.task.title || "Separator" : (selectedCategory?.name ?? project.name),
    );
</script>

<div class="header">
    <button class="hit" aria-label="Back to projects" onclick={() => (app.route = "projects")}>
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 4l-6 6 6 6" />
        </svg>
    </button>
    <span class="bold ellipsis">■ {project.name}</span>
    <span class="spacer"></span>
    <SyncIndicator />
</div>

<div class="list">
    {#each project.categories as category (category.id)}
        {@const status = categoryStatus(category)}
        <button
            class="row cat"
            class:selected={app.selected === category.id}
            onclick={() => tapCategory(category)}
        >
            <span class="pre">{app.selected === category.id ? ">" : " "}</span>
            <span class="title bold">
                {collapsed.has(category.id) ? "▶" : "▼"}
                {category.name}
                {#if status}
                    <span class="pre st-{status}">{STATUS_GLYPH[status]}</span>
                {/if}
                {#if category.estimate_minutes}
                    <span class="muted">~{formatEstimate(category.estimate_minutes)}</span>
                {/if}
            </span>
        </button>

        {#if !collapsed.has(category.id)}
            {#each category.tasks as task (task.id)}
                {#if isSeparator(task)}
                    <button
                        class="row sep"
                        class:selected={app.selected === task.id}
                        onclick={() => tapTask(task)}
                    >
                        <span class="pre">{app.selected === task.id ? ">" : " "}</span>
                        {#if task.title}
                            <span class="bold">{task.title}</span>
                        {/if}
                        <span class="rule"></span>
                    </button>
                {:else}
                    <button
                        class="row"
                        class:selected={app.selected === task.id}
                        class:completed={task.status === "completed" || task.status === "cancelled"}
                        onclick={() => tapTask(task)}
                    >
                        <span class="pre">{app.selected === task.id ? ">" : " "}</span>
                        <span class="pre st-{task.status}">{STATUS_GLYPH[task.status as Status] ?? "[ ]"}</span>
                        {#if task.priority}
                            <span class="pre prio prio-{task.priority}">{PRIORITY_GLYPH[task.priority]}</span>
                        {/if}
                        {#if task.tag_color}
                            <span class="pre dot tag-{task.tag_color}">●</span>
                        {/if}
                        <span class="title">
                            {task.title}{#if task.tag_label}<span class="muted"> {task.tag_label}</span>{/if}<span
                                class="muted meta"
                                >{task.description ? " ¶" : ""}{task.estimate_minutes
                                    ? ` ~${formatEstimate(task.estimate_minutes)}`
                                    : ""}</span
                            >
                        </span>
                    </button>
                {/if}
            {/each}
        {/if}
    {:else}
        <div class="row muted">No categories yet.</div>
    {/each}

    <button class="row muted" onclick={() => (overlay = { kind: "newCategory" })}>+ New category</button>
</div>

<div class="bar">
    <button
        class="chip"
        aria-label={target ? `Add a task to ${target.name}` : "Add a task"}
        disabled={!target}
        onclick={() => target && (overlay = { kind: "newTask", categoryID: target.id, separator: false })}
    >
        +
    </button>
    {#if selectedTask && !isSeparator(selectedTask.task)}
        {@const task = selectedTask.task}
        <button class="chip pre" aria-label="Cycle status" onclick={() => cycleStatus(task)}>
            <span class="st-{task.status}">{STATUS_GLYPH[task.status as Status] ?? "[ ]"}</span>
        </button>
        <button class="chip" aria-label="Cycle priority" onclick={() => cyclePriority(task)}>
            <span class="prio-{task.priority || 'none'}">{PRIORITY_GLYPH[task.priority ?? ""] || "▲"}</span>
        </button>
        <button class="chip" aria-label="Cycle tag" onclick={() => cycleTag(task)}>
            <span class="tag-{task.tag_color || 'none'}">●</span>
        </button>
        <button class="chip" onclick={() => (overlay = { kind: "task", id: task.id })}>edit</button>
    {:else if selectedTask}
        <button class="chip" onclick={() => (overlay = { kind: "task", id: selectedTask!.task.id })}>edit</button>
    {:else if selectedCategory}
        <button class="chip" onclick={() => (overlay = { kind: "category", id: selectedCategory.id })}>edit</button>
    {:else}
        <button class="chip" onclick={() => (overlay = { kind: "newCategory" })}>+ category</button>
    {/if}
    <span class="spacer"></span>
    <button class="chip" aria-label="More actions" onclick={() => (overlay = { kind: "menu" })}>⋯</button>
</div>

{#if overlay.kind === "task"}
    {@const found = findTask(project, overlay.id)}
    {#if found}
        {@const task = found.task}
        <TaskSheet
            {task}
            onsave={(patch) => {
                overlay = { kind: "none" };
                void withTask(task.id, (p, _c, t) => updateTask(p, t, patch));
            }}
            ondelete={() => removeTask(task)}
            onclose={close("task")}
        />
    {/if}
{:else if overlay.kind === "category"}
    {@const category = project.categories.find((c) => c.id === (overlay as { id: string }).id)}
    {#if category}
        <CategorySheet
            {category}
            onsave={(patch) => {
                overlay = { kind: "none" };
                void withCategory(category.id, (p, c) => updateCategory(p, c, patch));
            }}
            ondelete={() => removeCategory(category)}
            onclose={close("category")}
        />
    {/if}
{:else if overlay.kind === "newTask"}
    <Prompt
        title={overlay.separator ? "New separator" : "New task"}
        label={overlay.separator ? "Label (optional)" : "Title"}
        confirm="Add"
        onsubmit={createTask}
        onclose={close("newTask")}
    />
{:else if overlay.kind === "newCategory"}
    <Prompt
        title="New category"
        label="Name"
        confirm="Add"
        onsubmit={createCategory}
        onclose={close("newCategory")}
    />
{:else if overlay.kind === "menu"}
    <Menu title={menuTitle} actions={menuActions} onclose={close("menu")} />
{:else if overlay.kind === "moveTo"}
    {@const id = overlay.id}
    <Menu
        title="Move to category"
        actions={project.categories
            .filter((c) => c.id !== selectedTask?.category.id)
            .map((c) => ({
                label: c.name,
                run: () =>
                    withTask(id, (p, from, t) => {
                        const to = p.categories.find((x) => x.id === c.id);
                        return to ? moveTaskToCategory(p, from, t, to) : [];
                    }),
            }))}
        onclose={close("moveTo")}
    />
{:else if overlay.kind === "confirm"}
    {@const confirm = overlay}
    <Menu
        title={confirm.label}
        actions={[{ label: "Delete", danger: true, run: confirm.run }]}
        onclose={close("confirm")}
    />
{/if}

<style>
    .cat {
        align-items: center;
        margin-top: 10px;
    }

    .sep {
        align-items: baseline;
        gap: 8px;
    }

    .meta {
        white-space: nowrap;
    }

    .prio-none,
    .tag-none {
        color: var(--muted);
    }
</style>
