<script module lang="ts">
    export interface Action {
        label: string;
        run: () => void;
        danger?: boolean;
        disabled?: boolean;
        key?: string;
    }
</script>

<script lang="ts">
    import Sheet from "./Sheet.svelte";

    let { title, actions, onclose }: { title: string; actions: Action[]; onclose: () => void } = $props();
</script>

<Sheet {title} {onclose}>
    <ul class="menu">
        {#each actions as action (action.label)}
            <li>
                <button
                    class="item"
                    class:danger={action.danger}
                    disabled={action.disabled}
                    onclick={() => {
                        onclose();
                        action.run();
                    }}
                >
                    <span class="label">{action.label}</span>
                    {#if action.key}
                        <span class="muted pre">{action.key}</span>
                    {/if}
                </button>
            </li>
        {/each}
    </ul>
</Sheet>

<style>
    .menu {
        display: flex;
        flex-direction: column;
        margin: 0 -16px;
    }

    @media (min-width: 640px) {
        .menu {
            margin: 0 -22px;
        }
    }

    .item {
        display: flex;
        align-items: center;
        gap: 12px;
        width: 100%;
        min-height: 44px;
        padding: 8px 16px;
        border-radius: 0;
        text-align: left;
    }

    @media (min-width: 640px) {
        .item {
            padding: 6px 22px;
        }
    }

    .label {
        flex: 1;
    }

    li + li .item {
        border-top: 1px solid var(--rule);
    }

    .item:hover:not(:disabled) {
        background: var(--hover);
    }

    .item.danger:hover:not(:disabled) {
        background: color-mix(in srgb, var(--red) 12%, transparent);
    }

    .item:disabled {
        color: var(--muted);
        cursor: default;
    }
</style>
