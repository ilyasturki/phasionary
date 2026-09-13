<script lang="ts">
    let { value = $bindable(""), placeholder, count }: { value?: string; placeholder: string; count: number } = $props();

    let input = $state<HTMLInputElement | null>(null);

    export function focus(): void {
        input?.focus();
        input?.select();
    }

    function onkeydown(e: KeyboardEvent) {
        if (e.key !== "Escape") return;
        e.preventDefault();
        if (value) value = "";
        else input?.blur();
    }
</script>

<div class="search field" class:active={value !== ""}>
    <span class="bold muted pre">/</span>
    <input
        bind:this={input}
        bind:value
        type="search"
        {placeholder}
        aria-label="Search"
        autocomplete="off"
        spellcheck="false"
        enterkeyhint="search"
        {onkeydown}
    />
    {#if value !== ""}
        <span class="muted">{count}</span>
        <button type="button" class="hit clear" aria-label="Clear search" onclick={() => (value = "")}>✕</button>
    {/if}
</div>

<style>
    .search {
        gap: 6px;
        min-height: 36px;
        padding: 0 4px 0 10px;
    }

    .search:focus-within {
        border-color: var(--fg);
    }

    .search.active {
        border-color: var(--yellow);
    }

    input {
        flex: 1;
        min-width: 0;
        height: 34px;
        -webkit-appearance: none;
        appearance: none;
    }

    input::-webkit-search-cancel-button {
        display: none;
    }

    input::placeholder {
        color: var(--muted);
    }

    input:focus-visible {
        outline: none;
    }

    .clear {
        min-width: 28px;
        height: 28px;
    }
</style>
