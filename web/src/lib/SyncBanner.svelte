<script lang="ts">
    import { app, href } from "./app.svelte";

    const offline = $derived(app.offline);

    const text = $derived.by(() => {
        if (app.syncing) return "retrying…";
        const queued = app.pending === 1 ? "1 change waiting" : `${app.pending} changes waiting`;
        if (offline) return app.pending > 0 ? `offline · ${queued}` : "offline";
        return `sync failed · ${app.error}`;
    });
</script>

{#if app.error !== null}
    <div class="banner" class:offline role="alert">
        <span class="pre bold">{offline ? "◌" : "!"}</span>
        <span class="text">{text}</span>
        <button class="chip small" disabled={app.syncing} onclick={() => app.sync()}>retry</button>
        <a class="chip small" href={href("sync")}>details</a>
    </div>
{/if}

<style>
    .banner {
        display: flex;
        align-items: center;
        gap: 10px;
        min-height: 40px;
        padding: 6px 16px;
        color: var(--red);
        background: color-mix(in srgb, var(--red) 12%, transparent);
        border-bottom: 1px solid color-mix(in srgb, var(--red) 40%, transparent);
    }

    .banner.offline {
        color: var(--yellow);
        background: color-mix(in srgb, var(--yellow) 14%, var(--bg));
        border-bottom-color: color-mix(in srgb, var(--yellow) 45%, transparent);
    }

    .text {
        flex: 1;
        min-width: 0;
        overflow-wrap: anywhere;
    }

    .small {
        height: 28px;
        padding: 0 10px;
        color: inherit;
        border-color: currentColor;
        font-weight: 400;
    }
</style>
