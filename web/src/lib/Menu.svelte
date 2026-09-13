<script module lang="ts">
    export interface Action {
        label: string;
        run: () => void;
        danger?: boolean;
        disabled?: boolean;
    }
</script>

<script lang="ts">
    import Sheet from "./Sheet.svelte";

    let { title, actions, onclose }: { title: string; actions: Action[]; onclose: () => void } = $props();
</script>

<Sheet {title} {onclose}>
    <div class="menu">
        {#each actions as action (action.label)}
            <button
                class="item"
                class:danger={action.danger}
                disabled={action.disabled}
                onclick={() => {
                    action.run();
                    onclose();
                }}
            >
                {action.label}
            </button>
        {/each}
    </div>
</Sheet>

<style>
    .menu {
        display: flex;
        flex-direction: column;
    }

    .item {
        display: flex;
        align-items: center;
        min-height: 48px;
        padding: 8px 0;
        text-align: left;
    }

    .item + .item {
        border-top: 1px solid var(--rule);
    }

    .item:disabled {
        color: var(--muted);
        cursor: default;
    }
</style>
