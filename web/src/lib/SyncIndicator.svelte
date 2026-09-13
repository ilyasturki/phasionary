<script lang="ts">
    import { app } from "./app.svelte";

    const label = $derived(
        app.syncState === "error" ? "sync error"
        : app.syncState !== "idle" ? app.syncState
        : app.pending > 0 ? `${app.pending} changes pending` : "synced",
    );
</script>

<button class="hit" aria-label={label} title={label} onclick={() => (app.route = "sync")}>
    {#if app.syncState === "syncing"}
        <svg class="spin" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--blue)" stroke-width="1.5" stroke-linecap="round">
            <path d="M16 10a6 6 0 0 1-10.4 4.1M4 10a6 6 0 0 1 10.4-4.1" />
            <path d="M14 3v3h-3M6 17v-3h3" />
        </svg>
    {:else if app.syncState === "error"}
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--red)" stroke-width="1.5" stroke-linecap="round">
            <circle cx="10" cy="10" r="7" />
            <path d="M10 6v5M10 14h.01" />
        </svg>
    {:else if app.syncState === "offline"}
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--muted)" stroke-width="1.5" stroke-linecap="round">
            <path d="M3 3l14 14M6 13a5 5 0 0 1 8-4M3.5 9.5a9 9 0 0 1 4-2.5M10 16h.01" />
        </svg>
        {#if app.pending > 0}
            <span class="count">{app.pending}</span>
        {/if}
    {:else if app.pending > 0}
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--yellow)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M6 15V5m0 0L3 8m3-3l3 3" />
            <path d="M14 5v10m0 0l3-3m-3 3l-3-3" />
        </svg>
        <span class="count">{app.pending}</span>
    {:else}
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="var(--muted)" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 10.5l4 4 8-9" />
        </svg>
    {/if}
</button>

<style>
    button {
        gap: 4px;
    }

    .count {
        color: var(--yellow);
        font-weight: 700;
    }

    .spin {
        animation: spin 1.1s linear infinite;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }

    @media (prefers-reduced-motion: reduce) {
        .spin {
            animation: none;
        }
    }
</style>
