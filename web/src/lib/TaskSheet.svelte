<script lang="ts">
    import { untrack } from "svelte";
    import Chips from "./Chips.svelte";
    import LineField from "./LineField.svelte";
    import Sheet from "./Sheet.svelte";
    import {
        ESTIMATE_CHOICES,
        PRIORITY_CHIP,
        PRIORITY_ORDER,
        STATUSES,
        STATUS_GLYPH,
        STATUS_NAME,
        TAG_COLORS,
        type Priority,
        type Status,
        type TagColor,
        type Task,
        isSeparator,
    } from "./domain";
    import { autogrow } from "./keys";

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

    const statusChoices = STATUSES.map((s) => ({ value: s, label: STATUS_NAME[s], glyph: STATUS_GLYPH[s], color: `st-${s}` }));
    const priorityChoices = [
        { value: "" as Priority, label: "none" },
        ...PRIORITY_ORDER.map((p) => ({ value: p, label: p, glyph: PRIORITY_CHIP[p], color: `prio-${p}` })),
    ];

    const valid = $derived(separator || title.trim() !== "");

    function save() {
        if (!valid) return;
        onsave({
            title: title.trim(),
            description: description.trim(),
            status: status as Status,
            priority,
            estimate_minutes: estimate,
            tag_color: tagColor,
            tag_label: tagColor === "" ? "" : tagLabel.trim(),
        });
    }

    function onSheetKey(e: KeyboardEvent) {
        if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
            e.preventDefault();
            save();
        }
    }
</script>

<Sheet title={separator ? "Edit separator" : "Edit task"} {onclose}>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="form" onkeydown={onSheetKey}>
        <div class="group">
            <label for="task-title">{separator ? "Label" : "Title"}</label>
            <LineField
                id="task-title"
                class="field bold"
                bind:value={title}
                onenter={save}
                placeholder={separator ? "optional" : ""}
            />
        </div>

        {#if !separator}
            <div class="group">
                <label for="task-description">Description</label>
                <textarea
                    id="task-description"
                    class="field resizable"
                    rows="3"
                    bind:value={description}
                    use:autogrow
                    placeholder="notes, links, context…"
                ></textarea>
            </div>

            <div class="columns">
                <Chips label="Status" choices={statusChoices} bind:value={status} />
                <Chips label="Priority" choices={priorityChoices} bind:value={priority} />
            </div>

            <Chips label="Estimate" choices={ESTIMATE_CHOICES} bind:value={estimate} />

            <div class="group">
                <span class="label" id="tag-label">Tag</span>
                <div class="tag" role="radiogroup" aria-labelledby="tag-label">
                    <div class="wrap">
                        <button
                            class="chip"
                            class:on={tagColor === ""}
                            role="radio"
                            aria-checked={tagColor === ""}
                            onclick={() => (tagColor = "")}
                        >
                            none
                        </button>
                        {#each TAG_COLORS as value (value)}
                            <button
                                class="chip icon"
                                class:on={tagColor === value}
                                role="radio"
                                aria-checked={tagColor === value}
                                aria-label={value}
                                title={value}
                                onclick={() => (tagColor = value)}
                            >
                                <span class={tagColor === value ? "" : `tag-${value}`}>●</span>
                            </button>
                        {/each}
                    </div>
                    {#if tagColor}
                        <input
                            class="field"
                            bind:value={tagLabel}
                            placeholder="label (optional)"
                            aria-label="Tag label"
                            autocomplete="off"
                        />
                    {/if}
                </div>
            </div>
        {/if}

        <div class="footer">
            <button class="chip danger delete" onclick={ondelete}>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M2.5 4.5h11M6 4.5V3h4v1.5M4 4.5l.7 8.5h6.6l.7-8.5M6.5 7v4M9.5 7v4" />
                </svg>
                Delete {separator ? "separator" : "task"}
            </button>
            <span class="spacer"></span>
            <button class="chip" onclick={onclose}>Cancel</button>
            <button class="chip on" disabled={!valid} onclick={save}>Save</button>
        </div>
    </div>
</Sheet>

<style>
    .form {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .columns {
        display: grid;
        gap: 16px;
    }

    @media (min-width: 640px) {
        .columns {
            grid-template-columns: auto 1fr;
            column-gap: 32px;
        }
    }

    .tag {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    @media (max-width: 400px) {
        .delete {
            padding: 0 10px;
        }
    }
</style>
