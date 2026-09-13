<script lang="ts">
    import { untrack } from "svelte";
    import LineField from "./LineField.svelte";
    import Sheet from "./Sheet.svelte";

    let {
        title,
        label,
        value = "",
        confirm = "Save",
        onsubmit,
        onclose,
    }: {
        title: string;
        label: string;
        value?: string;
        confirm?: string;
        onsubmit: (value: string) => void;
        onclose: () => void;
    } = $props();

    let text = $state(untrack(() => value));

    function submit() {
        if (text.trim()) onsubmit(text.trim());
    }
</script>

<Sheet {title} {onclose}>
    <div class="group">
        <label for="prompt-field">{label}</label>
        <LineField id="prompt-field" class="field" bind:value={text} onenter={submit} select />
        <div class="actions">
            <button class="chip" onclick={onclose}>Cancel</button>
            <button class="chip on" disabled={!text.trim()} onclick={submit}>{confirm}</button>
        </div>
    </div>
</Sheet>
