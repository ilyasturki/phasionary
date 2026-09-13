<script lang="ts" generics="T extends string | number">
    let {
        label,
        choices,
        value = $bindable(),
    }: { label: string; choices: { value: T; label: string; glyph?: string; color?: string }[]; value: T } = $props();
</script>

<div class="group" role="radiogroup" aria-label={label}>
    <span class="label">{label}</span>
    <div class="wrap">
        {#each choices as c (c.value)}
            {@const glyph = c.glyph !== undefined}
            <button
                class="chip"
                class:icon={glyph}
                class:pre={glyph}
                class:on={value === c.value}
                role="radio"
                aria-checked={value === c.value}
                aria-label={glyph ? c.label : undefined}
                title={glyph ? c.label : undefined}
                onclick={() => (value = c.value)}
            >
                {#if glyph}<span class={value === c.value ? "" : c.color}>{c.glyph}</span>{:else}{c.label}{/if}
            </button>
        {/each}
    </div>
</div>
