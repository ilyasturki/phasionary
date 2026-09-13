<script lang="ts">
    import Sheet from "./Sheet.svelte";

    let {
        sections,
        onclose,
    }: { sections: { title: string; keys: { key: string; label: string }[] }[]; onclose: () => void } = $props();
</script>

<Sheet title="Keyboard shortcuts" {onclose}>
    <div class="sections">
        {#each sections as section (section.title)}
            <div class="section">
                <span class="bold">{section.title}</span>
                <dl>
                    {#each section.keys as k (k.key)}
                        <dt class="pre bold">{k.key}</dt>
                        <dd class="muted">{k.label}</dd>
                    {/each}
                </dl>
            </div>
        {/each}
    </div>
    <div class="actions">
        <span class="spacer"></span>
        <button class="chip" onclick={onclose}>Close</button>
    </div>
</Sheet>

<style>
    .sections {
        display: grid;
        gap: 16px;
    }

    @media (min-width: 640px) {
        .sections {
            grid-template-columns: 1fr 1fr;
            column-gap: 32px;
        }
    }

    .section {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    dl {
        display: grid;
        grid-template-columns: max-content 1fr;
        gap: 2px 14px;
    }

    .actions .chip {
        flex: 0 0 auto;
    }
</style>
