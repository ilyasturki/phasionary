<script lang="ts">
    import { untrack } from "svelte";
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

    function submit(event: SubmitEvent) {
        event.preventDefault();
        if (text.trim()) onsubmit(text.trim());
    }

    function focus(node: HTMLInputElement) {
        node.focus();
        node.select();
    }
</script>

<Sheet {title} {onclose}>
    <form class="group" onsubmit={submit}>
        <span class="muted">{label}</span>
        <input class="field" bind:value={text} use:focus autocomplete="off" />
        <div class="actions">
            <button type="button" class="chip" onclick={onclose}>Cancel</button>
            <button type="submit" class="chip on" disabled={!text.trim()}>{confirm}</button>
        </div>
    </form>
</Sheet>
