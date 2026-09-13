<script lang="ts">
    import type { Snippet } from "svelte";

    let { title, onclose, children }: { title: string; onclose: () => void; children: Snippet } = $props();

    function focusFirst(node: HTMLElement) {
        (node.querySelector<HTMLElement>("input, textarea") ?? node).focus();
    }
</script>

<svelte:window onkeydown={(e) => e.key === "Escape" && (e.preventDefault(), onclose())} />

<div
    class="scrim"
    role="button"
    tabindex="-1"
    aria-label="Close"
    onclick={onclose}
    onkeydown={(e) => e.key === "Enter" && onclose()}
></div>
<div class="sheet" role="dialog" aria-modal="true" aria-label={title} tabindex="-1" use:focusFirst>
    <h2>{title}</h2>
    {@render children()}
</div>
