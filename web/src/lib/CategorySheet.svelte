<script lang="ts">
    import { untrack } from "svelte";
    import Chips from "./Chips.svelte";
    import LineField from "./LineField.svelte";
    import Sheet from "./Sheet.svelte";
    import { ESTIMATE_CHOICES, type Category } from "./domain";

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

    const tasks = seed.tasks.length;
    const valid = $derived(name.trim() !== "");

    function save() {
        if (valid) onsave({ name: name.trim(), estimate_minutes: estimate });
    }
</script>

<Sheet title="Edit category" {onclose}>
    <div class="group">
        <label for="category-name">Name</label>
        <LineField id="category-name" class="field bold" bind:value={name} onenter={save} />
    </div>

    <Chips label="Estimate" choices={ESTIMATE_CHOICES} bind:value={estimate} />

    <div class="footer">
        <button
            class="chip danger delete"
            title={tasks ? `Deletes its ${tasks} ${tasks === 1 ? "task" : "tasks"} too` : undefined}
            onclick={ondelete}
        >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M2.5 4.5h11M6 4.5V3h4v1.5M4 4.5l.7 8.5h6.6l.7-8.5M6.5 7v4M9.5 7v4" />
            </svg>
            Delete category
        </button>
        <span class="spacer"></span>
        <button class="chip" onclick={onclose}>Cancel</button>
        <button class="chip on" disabled={!valid} onclick={save}>Save</button>
    </div>
</Sheet>
