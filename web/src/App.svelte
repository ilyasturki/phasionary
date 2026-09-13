<script lang="ts">
    import { app } from "./lib/app.svelte";
    import Login from "./screens/Login.svelte";
    import Projects from "./screens/Projects.svelte";
    import Sync from "./screens/Sync.svelte";
    import Tasks from "./screens/Tasks.svelte";

    void app.boot();
</script>

<svelte:document on:visibilitychange={() => document.visibilityState === "visible" && app.sync()} />
<svelte:window on:online={() => app.sync()} />

{#if !app.booted}
    <div class="header"><span class="muted">loading…</span></div>
{:else if app.route === "login" || !app.device}
    <Login />
{:else if app.route === "sync"}
    <Sync />
{:else if app.route === "tasks" && app.project}
    <Tasks project={app.project} />
{:else}
    <Projects />
{/if}
