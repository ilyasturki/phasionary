<script lang="ts">
    import { app } from "../lib/app.svelte";
    import { SyncError } from "../lib/sync";

    // Clear the fragment so a screenshot or back-navigation cannot replay the code.
    function scannedCode(): string {
        const match = /[#&]code=([A-Za-z0-9-]+)/.exec(location.hash);
        if (!match) return "";
        history.replaceState(null, "", location.pathname + location.search);
        return match[1].toUpperCase();
    }

    const scanned = scannedCode();
    let code = $state(scanned);
    let name = $state(defaultName());
    let busy = $state(false);
    let error = $state<string | null>(null);

    const ready = $derived(code.trim().length > 0 && name.trim().length > 0);

    if (scanned) void enroll();

    function defaultName(): string {
        const ua = navigator.userAgent;
        if (/android/i.test(ua)) return "Android phone";
        if (/iphone|ipad/i.test(ua)) return "iPhone";
        return "Browser";
    }

    async function enroll() {
        if (!ready || busy) return;
        busy = true;
        error = null;
        try {
            await app.login(code.trim().toUpperCase(), name.trim());
        } catch (err) {
            error = err instanceof SyncError ? err.message : String(err);
        } finally {
            busy = false;
        }
    }
</script>

<form class="login" onsubmit={(e) => { e.preventDefault(); void enroll(); }}>
    <div class="brand">
        <div class="wordmark">■ phasionary</div>
        <div class="muted">{location.host}</div>
    </div>

    <div class="fields">
        <div class="group">
            <label class="muted" for="code">Enrollment code</label>
            <input
                id="code"
                class="field code"
                bind:value={code}
                autocapitalize="characters"
                autocomplete="off"
                spellcheck="false"
                placeholder="XXXX-XXXX"
            />
            <div class="muted">
                Printed by <span class="fg">phasionary-server pair</span> on the server. Valid for ten minutes.
            </div>
        </div>

        <div class="group">
            <label class="muted" for="name">Device name</label>
            <input id="name" class="field" bind:value={name} autocomplete="off" />
        </div>

        {#if error}
            <div class="error">{error}</div>
        {/if}

        <button class="primary" type="submit" disabled={!ready || busy}>
            {busy ? "Enrolling…" : "Enroll this device"}
        </button>
    </div>

    <div class="spacer"></div>

    <p class="muted note">
        Your projects are downloaded to this device and stay usable offline. Changes sync when the server is
        reachable.
    </p>
</form>

<style>
    .login {
        display: flex;
        flex: 1;
        flex-direction: column;
        gap: 32px;
        padding: 80px 24px 32px;
        overflow-y: auto;
    }

    .brand {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .wordmark {
        font-size: 24px;
        line-height: 32px;
        font-weight: 700;
    }

    .fields {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .code {
        height: 52px;
        border-color: var(--fg);
        font-size: 20px;
        letter-spacing: 3px;
        font-weight: 700;
    }

    .fg {
        color: var(--fg);
    }

    .error {
        color: var(--red);
        text-wrap: pretty;
    }
</style>
