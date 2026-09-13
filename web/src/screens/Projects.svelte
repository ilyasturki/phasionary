<script lang="ts">
    import HelpSheet from "../lib/HelpSheet.svelte";
    import Highlight from "../lib/Highlight.svelte";
    import Hints from "../lib/Hints.svelte";
    import Menu, { type Action } from "../lib/Menu.svelte";
    import Prompt from "../lib/Prompt.svelte";
    import SearchField from "../lib/SearchField.svelte";
    import SyncBanner from "../lib/SyncBanner.svelte";
    import { app, href } from "../lib/app.svelte";
    import { type Project, formatRelativeShort, projectStats } from "../lib/domain";
    import { scrollRowIntoView, shortcutsSuspended } from "../lib/keys";
    import { createProject, deleteProject, renameProject } from "../lib/ops";

    type Overlay =
        | { kind: "none" }
        | { kind: "create" }
        | { kind: "rename"; project: Project }
        | { kind: "menu"; project: Project }
        | { kind: "app" }
        | { kind: "help" }
        | { kind: "confirm"; label: string; action: string; run: () => void }
        | { kind: "logout" };

    let overlay = $state<Overlay>({ kind: "none" });
    let query = $state("");
    let cursor = $state<string | null>(null);
    let search = $state<SearchField | null>(null);

    let now = $state(Date.now());
    $effect(() => {
        const timer = setInterval(() => (now = Date.now()), 30_000);
        return () => clearInterval(timer);
    });

    const shown = $derived.by(() => {
        const q = query.trim().toLowerCase();
        return q ? app.projects.filter((p) => p.name.toLowerCase().includes(q)) : app.projects;
    });

    const cursorIndex = $derived(shown.findIndex((p) => p.id === cursor));

    function close() {
        overlay = { kind: "none" };
    }

    async function create(name: string) {
        overlay = { kind: "none" };
        const { project, drafts } = createProject(name);
        await app.commit(project, drafts);
        app.go("tasks", project.id);
    }

    async function rename(p: Project, name: string) {
        overlay = { kind: "none" };
        const copy = $state.snapshot(p) as Project;
        await app.commit(copy, renameProject(copy, name));
    }

    function remove(p: Project) {
        const { open } = projectStats(p);
        overlay = {
            kind: "confirm",
            label: open ? `Delete “${p.name}” and its ${open} open ${open === 1 ? "task" : "tasks"}?` : `Delete “${p.name}”?`,
            action: "Delete project",
            run: () => void app.removeProject(p, deleteProject(p)),
        };
    }

    function actions(p: Project): Action[] {
        return [
            { label: "Open", run: () => app.go("tasks", p.id), key: "⏎" },
            { label: "Rename project", run: () => (overlay = { kind: "rename", project: p }) },
            { label: "Delete project", danger: true, run: () => remove(p), key: "d" },
        ];
    }

    const appActions: Action[] = [
        { label: "Sync…", run: () => app.go("sync") },
        { label: "Keyboard shortcuts", run: () => (overlay = { kind: "help" }), key: "?" },
        { label: "Log out of this device", danger: true, run: () => (overlay = { kind: "logout" }) },
    ];

    function move(delta: number) {
        if (shown.length === 0) return;
        const i = cursorIndex < 0 ? (delta > 0 ? 0 : shown.length - 1) : Math.min(Math.max(cursorIndex + delta, 0), shown.length - 1);
        cursor = shown[i].id;
        scrollRowIntoView(cursor);
    }

    function onkeydown(e: KeyboardEvent) {
        if (shortcutsSuspended(e)) return;
        const current = cursorIndex >= 0 ? shown[cursorIndex] : null;
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
                move(-shown.length);
                break;
            case "G":
                move(shown.length);
                break;
            case "Enter":
                if (current) app.go("tasks", current.id);
                else if (shown.length === 1) app.go("tasks", shown[0].id);
                else return;
                break;
            case "a":
                overlay = { kind: "create" };
                break;
            case "r":
                if (!current) return;
                overlay = { kind: "rename", project: current };
                break;
            case "d":
                if (!current) return;
                remove(current);
                break;
            case "/":
                search?.focus();
                break;
            case "?":
                overlay = { kind: "help" };
                break;
            case "Escape":
                if (query) query = "";
                else if (cursor) cursor = null;
                else return;
                break;
            default:
                return;
        }
        e.preventDefault();
    }

    const hints = [
        { key: "⏎", label: "open" },
        { key: "a", label: "new" },
        { key: "/", label: "search" },
        { key: "j/k", label: "move" },
        { key: "d", label: "delete" },
        { key: "?", label: "help" },
    ];

    const help = [
        {
            title: "Projects",
            keys: [
                { key: "j / k", label: "move down / up" },
                { key: "g / G", label: "first / last" },
                { key: "⏎", label: "open" },
                { key: "a", label: "new project" },
                { key: "r", label: "rename" },
                { key: "d", label: "delete" },
            ],
        },
        {
            title: "App",
            keys: [
                { key: "/", label: "search" },
                { key: "esc", label: "clear search" },
                { key: "?", label: "this list" },
            ],
        },
    ];
</script>

<svelte:window {onkeydown} />

<div class="screen">
    <div class="top">
        <SyncBanner />
        <div class="header">
            <h1 class="bold">Projects <span class="muted">({app.projects.length})</span></h1>
            <span class="spacer"></span>
            <div class="search-slot">
                <SearchField bind:this={search} bind:value={query} placeholder="search projects" count={shown.length} />
            </div>
            <button class="hit" aria-label="App menu" onclick={() => (overlay = { kind: "app" })}>⋯</button>
        </div>
        <div class="subheader">
            <SearchField bind:value={query} placeholder="search projects" count={shown.length} />
        </div>
    </div>

    <ul class="list" aria-label="Projects">
        {#if !query}
            <li>
                <button class="row new" onclick={() => (overlay = { kind: "create" })}>
                    <span class="gutter pre"> </span>
                    <span class="tag-green bold">+</span>
                    <span class="title">New project</span>
                </button>
            </li>
        {/if}

        {#each shown as project (project.id)}
            {@const stats = projectStats(project)}
            {@const current = cursor === project.id}
            <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
            <li
                id="row-{project.id}"
                class="row entry clickable"
                class:selected={current}
                onclick={(e) => !(e.target as Element).closest("a, button") && app.go("tasks", project.id)}
            >
                <span class="gutter pre" aria-hidden="true">{current ? ">" : " "}</span>
                <a class="title link bold" href={href("tasks", project.id)} onfocus={() => (cursor = project.id)}>
                    <Highlight text={project.name} {query} />
                </a>
                <span class="meta">
                    <span class="open" class:muted={!current}>{stats.open} open</span>
                    <span class="progress st-in_progress">{stats.inProgress ? `${stats.inProgress} ▸` : ""}</span>
                    <span class="age" class:muted={!current}>{formatRelativeShort(project.updated_at, now)}</span>
                </span>
                <button class="more" aria-label="Actions for {project.name}" onclick={() => (overlay = { kind: "menu", project })}>
                    ⋯
                </button>
            </li>
        {:else}
            <li class="row muted">{query ? "no matches" : "No projects yet."}</li>
        {/each}
    </ul>

    <Hints {hints} />
</div>

{#if overlay.kind === "create"}
    <Prompt title="New project" label="Name" confirm="Create" onsubmit={create} onclose={close} />
{:else if overlay.kind === "rename"}
    {@const project = overlay.project}
    <Prompt title="Rename project" label="Name" value={project.name} onsubmit={(name) => rename(project, name)} onclose={close} />
{:else if overlay.kind === "menu"}
    <Menu title={overlay.project.name} actions={actions(overlay.project)} onclose={close} />
{:else if overlay.kind === "app"}
    <Menu title={app.device?.name ?? "This device"} actions={appActions} onclose={close} />
{:else if overlay.kind === "help"}
    <HelpSheet sections={help} onclose={close} />
{:else if overlay.kind === "confirm"}
    {@const confirm = overlay}
    <Menu title={confirm.label} actions={[{ label: confirm.action, danger: true, run: confirm.run }]} onclose={close} />
{:else if overlay.kind === "logout"}
    <Menu
        title={app.pending > 0
            ? `Log out and discard ${app.pending} unsynced change${app.pending === 1 ? "" : "s"}?`
            : "Log out of this device?"}
        actions={[{ label: "Log out", danger: true, run: () => app.logout() }]}
        onclose={close}
    />
{/if}

<style>
    .entry {
        align-items: baseline;
        min-height: 40px;
    }

    .meta {
        display: flex;
        gap: 0 16px;
        flex-shrink: 0;
        white-space: nowrap;
        font-weight: 400;
        font-variant-numeric: tabular-nums;
    }

    .selected .meta {
        font-weight: 700;
    }

    .open {
        min-width: 7ch;
        text-align: right;
    }

    .progress,
    .age {
        min-width: 4ch;
        text-align: right;
    }

    @media (max-width: 639.98px) {
        .entry {
            display: grid;
            grid-template-columns: 1ch minmax(0, 1fr) 28px;
            grid-template-areas:
                "gutter name more"
                "gutter meta more";
            column-gap: 8px;
            align-items: start;
        }

        .entry .gutter {
            grid-area: gutter;
        }

        .entry .title {
            grid-area: name;
        }

        .meta {
            grid-area: meta;
            gap: 0;
        }

        .meta > * {
            min-width: 0;
            text-align: left;
        }

        .meta > *:not(:first-child):not(:empty)::before {
            content: "·";
            margin: 0 6px;
            color: var(--muted);
        }

        .entry .more {
            grid-area: more;
            display: flex;
            margin: 0;
        }
    }
</style>
