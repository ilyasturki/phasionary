<script lang="ts">
    import type { HTMLTextareaAttributes } from "svelte/elements";
    import { autogrow } from "./keys";

    let {
        value = $bindable(""),
        onenter,
        select = false,
        ...rest
    }: HTMLTextareaAttributes & { value?: string; onenter: () => void; select?: boolean } = $props();

    function onkeydown(e: KeyboardEvent) {
        if (e.key !== "Enter" || e.shiftKey) return;
        e.preventDefault();
        onenter();
    }

    function focus(node: HTMLTextAreaElement) {
        node.focus();
        if (select) node.select();
        else node.setSelectionRange(node.value.length, node.value.length);
    }
</script>

<textarea
    {...rest}
    rows="1"
    bind:value
    oninput={() => (value = value.replace(/[\r\n]+/g, " "))}
    {onkeydown}
    use:autogrow
    use:focus
    autocomplete="off"
    enterkeyhint="done"
></textarea>
