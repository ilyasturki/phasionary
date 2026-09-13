<script lang="ts">
    import Menu from "../lib/Menu.svelte";
    import { app } from "../lib/app.svelte";

    let confirming = $state(false);

    const relative = $derived.by(() => {
        if (!app.device?.last_sync) return "never";
        const seconds = Math.round((Date.now() - Date.parse(app.device.last_sync)) / 1000);
        if (seconds < 60) return "just now";
        const minutes = Math.round(seconds / 60);
        if (minutes < 60) return `${minutes} minute${minutes === 1 ? "" : "s"} ago`;
        const hours = Math.round(minutes / 60);
        return `${hours} hour${hours === 1 ? "" : "s"} ago`;
    });

    const result = $derived.by(() => {
        const r = app.lastResult;
        if (!r) return "";
        const parts = [`pushed ${r.pushed}`];
        if (r.written) parts.push(`pulled ${r.written}`);
        if (r.removed) parts.push(`removed ${r.removed}`);
        if (r.skipped) parts.push(`${r.skipped} deferred`);
        return parts.join(" · ");
    });
</script>

<div class="header">
    <button class="hit back" aria-label="Back" onclick={() => (app.route = app.projectID ? "tasks" : "projects")}>
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 4l-6 6 6 6" />
        </svg>
    </button>
    <span class="bold">Sync</span>
</div>

<div class="body">
    <div class="kv">
        <span class="muted">Device</span><span>{app.device?.name}</span>
        <span class="muted">Server</span><span>{location.host}</span>
        <span class="muted">Pending</span>
        <span class:pending={app.pending > 0}>
            {app.pending === 0 ? "nothing" : `${app.pending} change${app.pending === 1 ? "" : "s"}`}
        </span>
        <span class="muted">Last sync</span><span>{relative}</span>
        {#if result}
            <span class="muted">Result</span><span class="muted">{result}</span>
        {/if}
        {#if app.error}
            <span class="muted">Error</span><span class="error">{app.error}</span>
        {/if}
    </div>

    <button class="primary" disabled={app.syncState === "syncing"} onclick={() => app.sync()}>
        {app.syncState === "syncing" ? "Syncing…" : "Sync now"}
    </button>

    <p class="muted note">
        Syncs on its own when the app opens, after each edit, and when it comes back to the foreground.
    </p>
</div>

<div class="spacer"></div>

<div class="bar">
    <button class="chip logout" onclick={() => (confirming = true)}>Log out of this device</button>
</div>

{#if confirming}
    <Menu
        title={app.pending > 0
            ? `Log out and discard ${app.pending} unsynced change${app.pending === 1 ? "" : "s"}?`
            : "Log out of this device?"}
        actions={[{ label: "Log out", danger: true, run: () => app.logout() }]}
        onclose={() => (confirming = false)}
    />
{/if}

<style>
    .back {
        margin-left: -12px;
    }

    .body {
        display: flex;
        flex-direction: column;
        gap: 20px;
        padding: 8px 16px 0;
    }

    .kv {
        display: grid;
        grid-template-columns: 110px minmax(0, 1fr);
        gap: 4px 12px;
    }

    .pending {
        color: var(--yellow);
        font-weight: 700;
    }

    .error {
        color: var(--red);
        overflow-wrap: anywhere;
    }

    .logout {
        border: none;
        font-weight: 400;
    }
</style>
