<script lang="ts">
    import { untrack } from "svelte";
    import Sheet from "./Sheet.svelte";
    import {
        ESTIMATE_PRESETS,
        PRIORITY_CHIP,
        PRIORITY_ORDER,
        STATUSES,
        STATUS_GLYPH,
        TAG_COLORS,
        type Priority,
        type Status,
        type TagColor,
        type Task,
        estimateLabel,
        isSeparator,
    } from "./domain";

    let {
        task,
        onsave,
        ondelete,
        onclose,
    }: {
        task: Task;
        onsave: (patch: Partial<Task>) => void;
        ondelete: () => void;
        onclose: () => void;
    } = $props();

    const seed = untrack(() => $state.snapshot(task)) as Task;
    const separator = isSeparator(seed);

    let title = $state(seed.title ?? "");
    let description = $state(seed.description ?? "");
    let status = $state<Status | "">(seed.status ?? "todo");
    let priority = $state<Priority>(seed.priority ?? "");
    let estimate = $state(seed.estimate_minutes ?? 0);
    let tagColor = $state<TagColor>(seed.tag_color ?? "");
    let tagLabel = $state(seed.tag_label ?? "");

    function save() {
        onsave({
            title,
            description,
            status: status as Status,
            priority,
            estimate_minutes: estimate,
            tag_color: tagColor,
            tag_label: tagColor === "" ? "" : tagLabel,
        });
    }
</script>

<Sheet title={separator ? "Edit Separator" : "Edit Task"} {onclose}>
    <div class="group">
        <span class="muted">{separator ? "Label" : "Title"}</span>
        <input class="field" bind:value={title} autocomplete="off" />
    </div>

    {#if !separator}
        <div class="group">
            <span class="muted">Description</span>
            <textarea class="field" rows="3" bind:value={description}></textarea>
        </div>

        <div class="group">
            <span class="muted">Status</span>
            <div class="wrap">
                {#each STATUSES as value (value)}
                    <button class="chip pre" class:on={status === value} onclick={() => (status = value)}>
                        <span class={status === value ? "" : `st-${value}`}>{STATUS_GLYPH[value]}</span>
                    </button>
                {/each}
            </div>
        </div>

        <div class="group">
            <span class="muted">Priority</span>
            <div class="wrap">
                <button class="chip" class:on={priority === ""} onclick={() => (priority = "")}>none</button>
                {#each PRIORITY_ORDER as value (value)}
                    <button
                        class="chip"
                        aria-label={value}
                        class:on={priority === value}
                        onclick={() => (priority = value)}
                    >
                        <span class={priority === value ? "" : `prio-${value}`}>{PRIORITY_CHIP[value]}</span>
                    </button>
                {/each}
            </div>
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

        <div class="group">
            <span class="muted">Tag</span>
            <div class="wrap">
                <button class="chip" class:on={tagColor === ""} onclick={() => (tagColor = "")}>none</button>
                {#each TAG_COLORS as value (value)}
                    <button class="chip" class:on={tagColor === value} onclick={() => (tagColor = value)}>
                        <span class={tagColor === value ? "" : `tag-${value}`}>●</span>
                    </button>
                {/each}
            </div>
            {#if tagColor}
                <input class="field" bind:value={tagLabel} placeholder="Label" autocomplete="off" />
            {/if}
        </div>
    {/if}

    <div class="actions">
        <button class="chip danger" aria-label="Delete" onclick={ondelete}>✕</button>
        <button class="chip" onclick={onclose}>Cancel</button>
        <button class="chip on" onclick={save}>Save</button>
    </div>
</Sheet>
