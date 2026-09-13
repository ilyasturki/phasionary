<script lang="ts">
    import type { Snippet } from "svelte";

    let { title, onclose, children }: { title: string; onclose: () => void; children: Snippet } = $props();
</script>

<svelte:window on:keydown={(e) => e.key === "Escape" && onclose()} />

<div
    class="scrim"
    role="button"
    tabindex="-1"
    aria-label="Close"
    onclick={onclose}
    onkeydown={(e) => e.key === "Enter" && onclose()}
></div>
<div class="sheet" role="dialog" aria-modal="true" aria-label={title}>
    <h2>{title}</h2>
    {@render children()}
</div>
