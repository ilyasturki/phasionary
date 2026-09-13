<script lang="ts">
    import { untrack } from "svelte";
    import Sheet from "./Sheet.svelte";
    import { ESTIMATE_PRESETS, type Category, estimateLabel } from "./domain";

    let {
        category,
        onsave,
        ondelete,
        onclose,
    }: {
        category: Category;
        onsave: (patch: Partial<Category>) => void;
        ondelete: () => void;
        onclose: () => void;
    } = $props();

    const seed = untrack(() => $state.snapshot(category)) as Category;
    let name = $state(seed.name);
    let estimate = $state(seed.estimate_minutes ?? 0);
</script>

<Sheet title="Edit Category" {onclose}>
    <div class="group">
        <span class="muted">Name</span>
        <input class="field" bind:value={name} autocomplete="off" />
    </div>

    <div class="group">
        <span class="muted">Estimate</span>
        <div class="wrap">
            {#each ESTIMATE_PRESETS as value (value)}
                <button class="chip" class:on={estimate === value} onclick={() => (estimate = value)}>
                    {estimateLabel(value)}
                </button>
            {/each}
        </div>
    </div>

    <div class="actions">
        <button class="chip danger" aria-label="Delete" onclick={ondelete}>✕</button>
        <button class="chip" onclick={onclose}>Cancel</button>
        <button class="chip on" disabled={!name.trim()} onclick={() => onsave({ name: name.trim(), estimate_minutes: estimate })}>
            Save
        </button>
    </div>
</Sheet>
