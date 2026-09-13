<script lang="ts">
    let { text, query = "" }: { text: string; query?: string } = $props();

    const parts = $derived.by(() => {
        const q = query.trim().toLowerCase();
        if (!q) return [{ text, hit: false }];
        const out: { text: string; hit: boolean }[] = [];
        const lower = text.toLowerCase();
        let i = 0;
        for (let at = lower.indexOf(q); at >= 0; at = lower.indexOf(q, i)) {
            if (at > i) out.push({ text: text.slice(i, at), hit: false });
            out.push({ text: text.slice(at, at + q.length), hit: true });
            i = at + q.length;
        }
        if (i < text.length) out.push({ text: text.slice(i), hit: false });
        return out;
    });
</script>

{#each parts as part, i (i)}{#if part.hit}<mark>{part.text}</mark>{:else}{part.text}{/if}{/each}
