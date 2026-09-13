export function shortcutsSuspended(e: KeyboardEvent): boolean {
    if (e.metaKey || e.ctrlKey || e.altKey) return true;
    const t = e.target instanceof HTMLElement ? e.target : null;
    if (t && (t.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(t.tagName))) return true;
    if (document.querySelector('[role="dialog"]')) return true;
    // A focused button or link already activates on these; the shortcut would fire it twice.
    return !!t && (e.key === " " || e.key === "Enter") && !!t.closest("button, a");
}

export function scrollRowIntoView(id: string): void {
    document.getElementById(`row-${id}`)?.scrollIntoView({ block: "nearest" });
}

// A resizable textarea only grows, so a height the user dragged survives typing.
export function autogrow(node: HTMLTextAreaElement): { destroy(): void } {
    const fit = () => {
        if (node.classList.contains("resizable")) {
            if (node.scrollHeight > node.clientHeight) node.style.height = `${node.scrollHeight}px`;
            return;
        }
        node.style.height = "auto";
        node.style.height = `${node.scrollHeight}px`;
    };
    node.addEventListener("input", fit);
    // The sheet is still laying out on mount; a frame later scrollHeight is real.
    const frame = requestAnimationFrame(fit);
    return {
        destroy() {
            cancelAnimationFrame(frame);
            node.removeEventListener("input", fit);
        },
    };
}
