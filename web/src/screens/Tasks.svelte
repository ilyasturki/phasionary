<script lang="ts">
    import { SvelteSet } from "svelte/reactivity";
    import CategorySheet from "../lib/CategorySheet.svelte";
    import HelpSheet from "../lib/HelpSheet.svelte";
    import Highlight from "../lib/Highlight.svelte";
    import Hints from "../lib/Hints.svelte";
    import Menu, { type Action } from "../lib/Menu.svelte";
    import Prompt from "../lib/Prompt.svelte";
    import SearchField from "../lib/SearchField.svelte";
    import SyncBanner from "../lib/SyncBanner.svelte";
    import TaskSheet from "../lib/TaskSheet.svelte";
    import { app, href } from "../lib/app.svelte";
    import {
        PRIORITY_GLYPH,
        PRIORITY_ORDER,
        STATUS_GLYPH,
        STATUS_NAME,
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
    import { scrollRowIntoView, shortcutsSuspended } from "../lib/keys";
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
        | { kind: "menu"; id: string | null }
        | { kind: "moveTo"; id: string }
        | { kind: "help" }
        | { kind: "confirm"; label: string; action: string; run: () => void };

    let overlay = $state<Overlay>({ kind: "none" });
    let query = $state("");
    let search = $state<SearchField | null>(null);
    const collapsed = new SvelteSet<string>();

    function close() {
        overlay = { kind: "none" };
    }

    const selectedTask = $derived(app.selectedTask);
    const selectedCategory = $derived(app.selectedCategory);
    const target = $derived(selectedCategory ?? selectedTask?.category ?? project.categories[0] ?? null);

    function matches(t: Task, q: string): boolean {
        return [t.title, t.description, t.tag_label].some((s) => s?.toLowerCase().includes(q));
    }

    const shown = $derived.by(() => {
        const q = query.trim().toLowerCase();
        return project.categories.flatMap((category) => {
            if (!q || category.name.toLowerCase().includes(q)) return [{ category, tasks: category.tasks }];
            const tasks = category.tasks.filter((t) => matches(t, q));
            return tasks.length ? [{ category, tasks }] : [];
        });
    });

    const matchCount = $derived(shown.reduce((n, s) => n + s.tasks.filter((t) => !isSeparator(t)).length, 0));

    function folded(c: Category): boolean {
        return !query && collapsed.has(c.id);
    }

    const rows = $derived(shown.flatMap((s) => [s.category.id, ...(folded(s.category) ? [] : s.tasks.map((t) => t.id))]));

    function move(delta: number) {
        if (rows.length === 0) return;
        const at = app.selected ? rows.indexOf(app.selected) : -1;
        const i = at < 0 ? (delta > 0 ? 0 : rows.length - 1) : Math.min(Math.max(at + delta, 0), rows.length - 1);
        app.selected = rows[i];
        scrollRowIntoView(rows[i]);
    }

    function tapTask(t: Task) {
        if (app.selected === t.id) overlay = { kind: "task", id: t.id };
        else app.selected = t.id;
    }

    function tapCategory(c: Category) {
        if (app.selected === c.id) overlay = { kind: "category", id: c.id };
        else app.selected = c.id;
    }

    function toggleFold(c: Category) {
        if (collapsed.has(c.id)) collapsed.delete(c.id);
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

    function cycleStatus(t: Task, delta = 1) {
        void withTask(t.id, (p, _c, task) => setStatus(p, task, nextStatus((task.status || "todo") as Status, delta)));
    }

    function cyclePriority(t: Task) {
        const order: Priority[] = ["", ...PRIORITY_ORDER];
        const next = order[(order.indexOf(t.priority ?? "") + 1) % order.length];
        void withTask(t.id, (p, _c, task) => updateTask(p, task, { priority: next }));
    }

    function stepPriority(t: Task, delta: number) {
        const order: Priority[] = ["", ...[...PRIORITY_ORDER].reverse()];
        const i = Math.min(Math.max(order.indexOf(t.priority ?? "") + delta, 0), order.length - 1);
        void withTask(t.id, (p, _c, task) => updateTask(p, task, { priority: order[i] }));
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
        const separator = isSeparator(t);
        overlay = {
            kind: "confirm",
            label: separator ? `Delete separator${t.title ? ` “${t.title}”` : ""}?` : `Delete task “${t.title}”?`,
            action: separator ? "Delete separator" : "Delete task",
            run: () => {
                if (app.selected === t.id) app.selected = null;
                void withTask(t.id, (p, c, task) => deleteTask(p, c, task));
            },
        };
    }

    function removeCategory(c: Category) {
        const n = c.tasks.filter((t) => !isSeparator(t)).length;
        overlay = {
            kind: "confirm",
            label: n ? `Delete “${c.name}” and its ${n} ${n === 1 ? "task" : "tasks"}?` : `Delete “${c.name}”?`,
            action: "Delete category",
            run: () => {
                if (app.selected === c.id) app.selected = null;
                void withCategory(c.id, (p, cat) => deleteCategory(p, cat));
            },
        };
    }

    function taskActions(t: Task, c: Category): Action[] {
        const i = c.tasks.findIndex((x) => x.id === t.id);
        const separator = isSeparator(t);
        return [
            { label: "Edit", run: () => (overlay = { kind: "task", id: t.id }), key: "⏎" },
            ...(separator ? [] : [{ label: "Cycle status", run: () => cycleStatus(t), key: "space" }]),
            { label: "Move up", disabled: i <= 0, run: () => withTask(t.id, (p, cat, task) => moveTaskWithin(p, cat, task, -1)), key: "K" },
            {
                label: "Move down",
                disabled: i < 0 || i >= c.tasks.length - 1,
                run: () => withTask(t.id, (p, cat, task) => moveTaskWithin(p, cat, task, 1)),
                key: "J",
            },
            {
                label: "Move to category…",
                disabled: project.categories.length < 2,
                run: () => (overlay = { kind: "moveTo", id: t.id }),
            },
            { label: "New separator here", run: () => (overlay = { kind: "newTask", categoryID: c.id, separator: true }), key: "-" },
            { label: separator ? "Delete separator" : "Delete task", danger: true, run: () => removeTask(t), key: "d" },
        ];
    }

    function categoryActions(c: Category): Action[] {
        const i = project.categories.findIndex((x) => x.id === c.id);
        return [
            { label: "New task", run: () => (overlay = { kind: "newTask", categoryID: c.id, separator: false }), key: "a" },
            { label: "Edit category", run: () => (overlay = { kind: "category", id: c.id }), key: "⏎" },
            { label: collapsed.has(c.id) ? "Unfold" : "Fold", run: () => toggleFold(c), key: "space" },
            { label: "Move up", disabled: i <= 0, run: () => withCategory(c.id, (p, cat) => moveCategory(p, cat, -1)), key: "K" },
            {
                label: "Move down",
                disabled: i >= project.categories.length - 1,
                run: () => withCategory(c.id, (p, cat) => moveCategory(p, cat, 1)),
                key: "J",
            },
            { label: "Delete category", danger: true, run: () => removeCategory(c), key: "d" },
        ];
    }

    const projectActions: Action[] = [
        { label: "New category", run: () => (overlay = { kind: "newCategory" }), key: "A" },
        { label: "Keyboard shortcuts", run: () => (overlay = { kind: "help" }), key: "?" },
        { label: "Sync…", run: () => app.go("sync") },
    ];

    function menuFor(id: string | null): { title: string; actions: Action[] } {
        const category = id ? project.categories.find((c) => c.id === id) : null;
        if (category) return { title: category.name, actions: categoryActions(category) };
        const found = id ? findTask(project, id) : null;
        if (found) return { title: found.task.title || "Separator", actions: taskActions(found.task, found.category) };
        return { title: project.name, actions: projectActions };
    }

    function moveActions(id: string): Action[] {
        const from = findTask(project, id)?.category;
        return project.categories
            .filter((c) => c.id !== from?.id)
            .map((c) => ({
                label: c.name,
                run: () =>
                    withTask(id, (p, src, t) => {
                        const to = p.categories.find((x) => x.id === c.id);
                        return to ? moveTaskToCategory(p, src, t, to) : [];
                    }),
            }));
    }

    function onkeydown(e: KeyboardEvent) {
        if (shortcutsSuspended(e)) return;
        const task = selectedTask && !isSeparator(selectedTask.task) ? selectedTask.task : null;
        switch (e.key) {
            case "j":
            case "ArrowDown":
                move(1);
                break;
            case "k":
            case "ArrowUp":
                move(-1);
                break;
            case "g":
                move(-rows.length);
                break;
            case "G":
                move(rows.length);
                break;
            case "Enter":
                if (selectedTask) overlay = { kind: "task", id: selectedTask.task.id };
                else if (selectedCategory) overlay = { kind: "category", id: selectedCategory.id };
                else return;
                break;
            case " ":
                if (task) cycleStatus(task, e.shiftKey ? -1 : 1);
                else if (selectedCategory) toggleFold(selectedCategory);
                else return;
                break;
            case "a":
            case "-":
                if (!target) return;
                overlay = { kind: "newTask", categoryID: target.id, separator: e.key === "-" };
                break;
            case "A":
                overlay = { kind: "newCategory" };
                break;
            case "d":
                if (selectedTask) removeTask(selectedTask.task);
                else if (selectedCategory) removeCategory(selectedCategory);
                else return;
                break;
            case "h":
            case "l":
                if (!task) return;
                stepPriority(task, e.key === "l" ? 1 : -1);
                break;
            case "t":
                if (!task) return;
                cycleTag(task);
                break;
            case "J":
            case "K": {
                const delta = e.key === "J" ? 1 : -1;
                if (selectedTask) void withTask(selectedTask.task.id, (p, c, t) => moveTaskWithin(p, c, t, delta));
                else if (selectedCategory) void withCategory(selectedCategory.id, (p, c) => moveCategory(p, c, delta));
                else return;
                break;
            }
            case "/":
                search?.focus();
                break;
            case "?":
                overlay = { kind: "help" };
                break;
            case "Escape":
                if (query) query = "";
                else if (app.selected) app.selected = null;
                else return;
                break;
            default:
                return;
        }
        e.preventDefault();
    }

    const hints = [
        { key: "?", label: "help" },
        { key: "a", label: "add" },
        { key: "A", label: "category" },
        { key: "space", label: "status" },
        { key: "⏎", label: "edit" },
        { key: "/", label: "search" },
        { key: "d", label: "delete" },
    ];

    const help = [
        {
            title: "Navigation",
            keys: [
                { key: "j / k", label: "move down / up" },
                { key: "g / G", label: "first / last row" },
                { key: "space", label: "fold / unfold a category" },
                { key: "/", label: "search" },
                { key: "esc", label: "clear search, then selection" },
            ],
        },
        {
            title: "Tasks",
            keys: [
                { key: "a", label: "add task" },
                { key: "A", label: "add category" },
                { key: "-", label: "add separator" },
                { key: "⏎", label: "edit" },
                { key: "space / ⇧space", label: "cycle status forward / back" },
                { key: "h / l", label: "priority down / up" },
                { key: "t", label: "cycle tag" },
                { key: "d", label: "delete" },
            ],
        },
        {
            title: "Organize",
            keys: [{ key: "J / K", label: "move item down / up" }],
        },
        {
            title: "Editing",
            keys: [
                { key: "⏎", label: "save (title)" },
                { key: "ctrl+⏎", label: "save (anywhere)" },
                { key: "esc", label: "cancel" },
            ],
        },
    ];
</script>

<svelte:window {onkeydown} />

<div class="screen">
    <div class="top">
        <SyncBanner />
        <div class="header">
            <a class="hit back" href={href("projects")} aria-label="Back to projects">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M12 4l-6 6 6 6" />
                </svg>
            </a>
            <h1 class="bold ellipsis">■ {project.name}</h1>
            <span class="spacer"></span>
            <div class="search-slot">
                <SearchField bind:this={search} bind:value={query} placeholder="search tasks" count={matchCount} />
            </div>
            <button class="hit" aria-label="More actions" onclick={() => (overlay = { kind: "menu", id: app.selected })}>⋯</button>
        </div>
        <div class="subheader">
            <SearchField bind:value={query} placeholder="search tasks" count={matchCount} />
        </div>
    </div>

    <ul class="list" aria-label="Tasks">
        {#each shown as { category, tasks } (category.id)}
            {@const status = categoryStatus(category)}
            {@const isFolded = folded(category)}
            {@const current = app.selected === category.id}
            <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
            <li
                id="row-{category.id}"
                class="row cat clickable"
                class:selected={current}
                onclick={(e) => !(e.target as Element).closest("button") && (app.selected = category.id)}
            >
                <span class="gutter pre" aria-hidden="true">{current ? ">" : " "}</span>
                <button
                    class="glyph fold muted"
                    aria-label={isFolded ? "Unfold" : "Fold"}
                    aria-expanded={!isFolded}
                    disabled={query !== ""}
                    onclick={() => toggleFold(category)}
                >
                    {isFolded ? "▶" : "▼"}
                </button>
                <button class="title bold" onclick={() => tapCategory(category)}>
                    <Highlight text={category.name} {query} />
                    {#if status}
                        <span class="pre st-{status}"> {STATUS_GLYPH[status]}</span>
                    {/if}
                    {#if category.estimate_minutes}
                        <span class="muted"> ~{formatEstimate(category.estimate_minutes)}</span>
                    {/if}
                    {#if isFolded}
                        <span class="muted"> {category.tasks.filter((t) => !isSeparator(t)).length}</span>
                    {/if}
                </button>
                <button class="more" aria-label="Actions for {category.name}" onclick={() => (overlay = { kind: "menu", id: category.id })}>
                    ⋯
                </button>
            </li>

            {#if !isFolded}
                <li class="gap" aria-hidden="true"></li>
                {#each tasks as task (task.id)}
                    {@const sel = app.selected === task.id}
                    {#if isSeparator(task)}
                        <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
                        <li
                            id="row-{task.id}"
                            class="row sep clickable"
                            class:selected={sel}
                            onclick={(e) => !(e.target as Element).closest("button") && tapTask(task)}
                        >
                            <span class="gutter pre" aria-hidden="true">{sel ? ">" : " "}</span>
                            {#if task.title}
                                <button class="title label bold" aria-label="Separator {task.title}" onclick={() => tapTask(task)}>
                                    <Highlight text={task.title} {query} />
                                </button>
                            {/if}
                            <span class="rule"></span>
                            <button class="more" aria-label="Separator actions" onclick={() => (overlay = { kind: "menu", id: task.id })}>⋯</button>
                        </li>
                    {:else}
                        <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
                        <li
                            id="row-{task.id}"
                            class="row task clickable"
                            class:selected={sel}
                            class:completed={task.status === "completed" || task.status === "cancelled"}
                            onclick={(e) => !(e.target as Element).closest("button") && (app.selected = task.id)}
                        >
                            <span class="gutter pre" aria-hidden="true">{sel ? ">" : " "}</span>
                            <button
                                class="glyph st-{task.status}"
                                aria-label="{STATUS_NAME[task.status as Status] ?? 'to do'} — cycle status"
                                title="Cycle status"
                                onclick={() => cycleStatus(task)}
                            >
                                {STATUS_GLYPH[task.status as Status] ?? "[ ]"}
                            </button>
                            {#if task.priority && PRIORITY_GLYPH[task.priority]}
                                <span class="pre prio-{task.priority}" title="{task.priority} priority">{PRIORITY_GLYPH[task.priority]}</span>
                            {/if}
                            {#if task.tag_color}
                                <span class="pre tag-{task.tag_color}">●</span>
                            {/if}
                            <button class="title prio-{task.priority ?? ''}" onclick={() => tapTask(task)}>
                                <Highlight text={task.title} {query} />{#if task.tag_label}<span class="muted"> <Highlight text={task.tag_label} {query} /></span>{/if}<span
                                    class="muted meta"
                                    >{task.description ? " ¶" : ""}{task.estimate_minutes ? ` ~${formatEstimate(task.estimate_minutes)}` : ""}</span
                                >
                            </button>
                            <button class="more" aria-label="Actions for {task.title}" onclick={() => (overlay = { kind: "menu", id: task.id })}>⋯</button>
                        </li>
                    {/if}
                {:else}
                    <li class="row empty muted">empty</li>
                {/each}
            {/if}
        {:else}
            <li class="row muted">{query ? "no matches" : "No categories yet."}</li>
        {/each}

        {#if !query}
            <li class="gap" aria-hidden="true"></li>
            <li>
                <button class="row new" onclick={() => (overlay = { kind: "newCategory" })}>
                    <span class="gutter pre"> </span>
                    <span class="tag-green bold">+</span>
                    <span class="title">New category</span>
                </button>
            </li>
        {/if}
    </ul>

    <div class="bar">
        <button
            class="chip icon"
            aria-label={target ? `Add a task to ${target.name}` : "Add a task"}
            disabled={!target}
            onclick={() => target && (overlay = { kind: "newTask", categoryID: target.id, separator: false })}
        >
            +
        </button>
        {#if selectedTask && !isSeparator(selectedTask.task)}
            {@const task = selectedTask.task}
            <button class="chip icon pre" aria-label="Cycle status" onclick={() => cycleStatus(task)}>
                <span class="st-{task.status}">{STATUS_GLYPH[task.status as Status] ?? "[ ]"}</span>
            </button>
            <button class="chip icon" aria-label="Cycle priority" onclick={() => cyclePriority(task)}>
                <span class="prio-{task.priority || 'none'}">{PRIORITY_GLYPH[task.priority ?? ""] || "▲"}</span>
            </button>
            <button class="chip icon" aria-label="Cycle tag" onclick={() => cycleTag(task)}>
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
        <button class="chip icon" aria-label="More actions" onclick={() => (overlay = { kind: "menu", id: app.selected })}>⋯</button>
    </div>

    <Hints {hints} />
</div>

{#if overlay.kind === "task"}
    {@const task = findTask(project, overlay.id)?.task}
    {#if task}
        <TaskSheet
            {task}
            onsave={(patch) => {
                // Dispatch before closing: the const goes null with the overlay.
                void withTask(task.id, (p, _c, t) => updateTask(p, t, patch));
                overlay = { kind: "none" };
            }}
            ondelete={() => removeTask(task)}
            onclose={close}
        />
    {/if}
{:else if overlay.kind === "category"}
    {@const category = project.categories.find((c) => c.id === (overlay as { id: string }).id)}
    {#if category}
        <CategorySheet
            {category}
            onsave={(patch) => {
                void withCategory(category.id, (p, c) => updateCategory(p, c, patch));
                overlay = { kind: "none" };
            }}
            ondelete={() => removeCategory(category)}
            onclose={close}
        />
    {/if}
{:else if overlay.kind === "newTask"}
    <Prompt
        title={overlay.separator ? "New separator" : "New task"}
        label={overlay.separator ? "Label (optional)" : "Title"}
        confirm="Add"
        onsubmit={createTask}
        onclose={close}
    />
{:else if overlay.kind === "newCategory"}
    <Prompt title="New category" label="Name" confirm="Add" onsubmit={createCategory} onclose={close} />
{:else if overlay.kind === "menu"}
    {@const menu = menuFor(overlay.id)}
    <Menu title={menu.title} actions={menu.actions} onclose={close} />
{:else if overlay.kind === "moveTo"}
    <Menu title="Move to category" actions={moveActions(overlay.id)} onclose={close} />
{:else if overlay.kind === "help"}
    <HelpSheet sections={help} onclose={close} />
{:else if overlay.kind === "confirm"}
    {@const confirm = overlay}
    <Menu title={confirm.label} actions={[{ label: confirm.action, danger: true, run: confirm.run }]} onclose={close} />
{/if}

<style>
    .cat {
        align-items: center;
        margin-top: 18px;
    }

    .fold:disabled {
        cursor: default;
    }

    .fold:hover {
        color: inherit;
    }

    .gap {
        height: 12px;
    }

    .sep {
        align-items: baseline;
    }

    .sep .label {
        flex: 0 1 auto;
    }

    .meta {
        white-space: nowrap;
    }

    .empty {
        padding-left: calc(16px + 1ch + 8px);
        min-height: 28px;
    }

    .prio-none,
    .tag-none {
        color: var(--muted);
    }
</style>
