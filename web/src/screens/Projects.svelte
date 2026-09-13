<script lang="ts">
    import Menu, { type Action } from "../lib/Menu.svelte";
    import Prompt from "../lib/Prompt.svelte";
    import SyncIndicator from "../lib/SyncIndicator.svelte";
    import { app } from "../lib/app.svelte";
    import { type Project, statusCounts } from "../lib/domain";
    import { createProject, deleteProject, renameProject } from "../lib/ops";

    let creating = $state(false);
    let renaming = $state<Project | null>(null);
    let menuFor = $state<Project | null>(null);

    function summary(p: Project): string {
        const counts = statusCounts(p);
        const open = counts.todo + counts.in_progress;
        const total = open + counts.completed + counts.cancelled;
        if (total === 0) return "empty";
        if (open === 0) return "all done";
        return `${open} open · ${counts.in_progress} in progress`;
    }

    async function create(name: string) {
        creating = false;
        const { project, drafts } = createProject(name);
        await app.commit(project, drafts);
        app.open(project.id);
    }

    async function rename(p: Project, name: string) {
        renaming = null;
        const copy = $state.snapshot(p) as Project;
        await app.commit(copy, renameProject(copy, name));
    }

    function actions(p: Project): Action[] {
        return [
            { label: "Rename project", run: () => (renaming = p) },
            { label: "Delete project", danger: true, run: () => app.removeProject(p, deleteProject(p)) },
        ];
    }
</script>

<div class="header">
    <span class="bold">Projects</span>
    <span class="spacer"></span>
    <SyncIndicator />
</div>

<div class="list">
    {#each app.projects as project (project.id)}
        <div class="line">
            <button class="row entry" onclick={() => app.open(project.id)}>
                <span class="title">
                    <span class="bold">■ {project.name}</span>
                    <span class="counts muted">{summary(project)}</span>
                </span>
            </button>
            <button class="hit" aria-label="Project actions" onclick={() => (menuFor = project)}>⋯</button>
        </div>
    {:else}
        <div class="row muted">No projects yet.</div>
    {/each}

    <button class="row muted" onclick={() => (creating = true)}>+ New project</button>
</div>

<div class="bar">
    <span class="muted ellipsis">{app.device?.name ?? ""} · {location.host}</span>
    <span class="spacer"></span>
    <button class="chip" onclick={() => (app.route = "sync")}>Sync</button>
</div>

{#if creating}
    <Prompt
        title="New project"
        label="Name"
        confirm="Create"
        onsubmit={create}
        onclose={() => (creating = false)}
    />
{/if}

{#if renaming}
    <Prompt
        title="Rename project"
        label="Name"
        value={renaming.name}
        onsubmit={(name) => rename(renaming!, name)}
        onclose={() => (renaming = null)}
    />
{/if}

{#if menuFor}
    <Menu title={menuFor.name} actions={actions(menuFor)} onclose={() => (menuFor = null)} />
{/if}

<style>
    .line {
        display: flex;
        align-items: flex-start;
    }

    .entry {
        min-height: 48px;
    }

    .entry .title {
        display: flex;
        flex-direction: column;
    }

    .counts {
        font-weight: 400;
    }
</style>
