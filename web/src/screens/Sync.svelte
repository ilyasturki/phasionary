<script lang="ts">
    import Menu from "../lib/Menu.svelte";
    import { app, href } from "../lib/app.svelte";

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

    const health = $derived.by(() => {
        if (app.syncing) return { text: "syncing…", cls: "" };
        if (app.error !== null) return { text: app.offline ? "offline" : "failed", cls: app.offline ? "warn" : "error" };
        return { text: "ok", cls: "ok" };
    });

    const back = $derived(app.projectID ? href("tasks", app.projectID) : href("projects"));
</script>

<svelte:window onkeydown={(e) => e.key === "Escape" && !confirming && (location.hash = back)} />

<div class="screen">
    <div class="top">
        <div class="header">
            <a class="hit back" href={back} aria-label="Back">
                <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M12 4l-6 6 6 6" />
                </svg>
            </a>
            <h1 class="bold">Sync</h1>
        </div>
    </div>

    <div class="body">
        <dl class="kv">
            <dt class="muted">State</dt>
            <dd class={health.cls}>{health.text}</dd>
            <dt class="muted">Device</dt>
            <dd>{app.device?.name}</dd>
            <dt class="muted">Server</dt>
            <dd>{location.host}</dd>
            <dt class="muted">Pending</dt>
            <dd class:pending={app.pending > 0}>
                {app.pending === 0 ? "nothing" : `${app.pending} change${app.pending === 1 ? "" : "s"}`}
            </dd>
            <dt class="muted">Last sync</dt>
            <dd>{relative}</dd>
            {#if result}
                <dt class="muted">Result</dt>
                <dd class="muted">{result}</dd>
            {/if}
            {#if app.error}
                <dt class="muted">Error</dt>
                <dd class="error">{app.error}</dd>
            {/if}
        </dl>

        <button class="primary" disabled={app.syncing} onclick={() => app.sync()}>
            {app.syncing ? "Syncing…" : "Sync now"}
        </button>

        <p class="muted note">
            Syncs on its own when the app opens, after each edit, and when it comes back to the foreground. A
            banner stays up while a sync is failing.
        </p>
    </div>

    <div class="spacer"></div>

    <div class="foot">
        <button class="chip danger logout" onclick={() => (confirming = true)}>Log out of this device</button>
    </div>
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
    .body {
        display: flex;
        flex-direction: column;
        gap: 20px;
        padding: 8px 16px 0;
        max-width: 480px;
    }

    .kv {
        display: grid;
        grid-template-columns: 110px minmax(0, 1fr);
        gap: 4px 12px;
    }

    .ok {
        color: var(--green);
    }

    .warn {
        color: var(--yellow);
    }

    .pending {
        color: var(--yellow);
        font-weight: 700;
    }

    .error {
        color: var(--red);
        overflow-wrap: anywhere;
    }

    .foot {
        display: flex;
        padding: 16px;
    }

    .logout {
        font-weight: 400;
    }
</style>
