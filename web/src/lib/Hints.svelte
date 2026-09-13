<script lang="ts">
    let { hints }: { hints: { key: string; label: string }[] } = $props();
</script>

<div class="hints" aria-hidden="true">
    {#each hints as hint, i (hint.key)}
        {#if i > 0}<span class="sep muted">·</span>{/if}
        <span class="bold">{hint.key}</span>
        <span class="muted">{hint.label}</span>
    {/each}
</div>

<style>
    .hints {
        display: none;
        position: sticky;
        bottom: 0;
        z-index: 2;
        align-items: center;
        gap: 8px;
        min-height: 36px;
        padding: 4px 16px;
        background: var(--bg);
        border-top: 1px solid var(--rule);
        white-space: nowrap;
        overflow: hidden;
    }

    @media (min-width: 640px) {
        .hints {
            display: flex;
        }
    }

    .sep {
        margin: 0 2px;
    }
</style>
