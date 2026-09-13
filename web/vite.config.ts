import { readFileSync } from "node:fs";
import { createHash } from "node:crypto";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import type { Plugin } from "vite";
import { defineConfig } from "vitest/config";

function serviceWorker(): Plugin {
    return {
        name: "phasionary-sw",
        apply: "build",
        generateBundle(_options, bundle) {
            const assets = Object.keys(bundle).map((name) => "/" + name);
            // public/ files are copied outside the bundle, so they are named here.
            const shell = ["/", "/manifest.webmanifest", "/icon.svg", "/icon-192.png", "/icon-512.png", ...assets];
            const version = createHash("sha256").update(shell.join("\n")).digest("hex").slice(0, 12);
            const source = readFileSync(new URL("./src/sw.js", import.meta.url), "utf8")
                .replaceAll("__SHELL__", JSON.stringify(shell))
                .replaceAll("__VERSION__", JSON.stringify(version));
            // A stale placeholder fails worker evaluation silently in the browser.
            const leftover = source.match(/__[A-Z]+__/);
            if (leftover) this.error(`sw.js still holds ${leftover[0]}`);
            this.emitFile({ type: "asset", fileName: "sw.js", source });
        },
    };
}

export default defineConfig({
    plugins: [svelte(), serviceWorker()],
    build: { target: "es2022", assetsInlineLimit: 0 },
    server: { proxy: { "/v1": "http://127.0.0.1:7777" } },
    test: { environment: "node", include: ["src/**/*.test.ts"], unstubGlobals: true },
});
