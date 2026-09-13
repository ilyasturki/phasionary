// Substituted by the phasionary-sw vite plugin at build time.
const SHELL = __SHELL__;
const CACHE = "phasionary-" + __VERSION__;

self.addEventListener("install", (event) => {
    event.waitUntil(caches.open(CACHE).then((cache) => cache.addAll(SHELL)).then(() => self.skipWaiting()));
});

self.addEventListener("activate", (event) => {
    event.waitUntil(
        caches
            .keys()
            .then((keys) => Promise.all(keys.filter((key) => key !== CACHE).map((key) => caches.delete(key))))
            .then(() => self.clients.claim()),
    );
});

self.addEventListener("fetch", (event) => {
    const request = event.request;
    if (request.method !== "GET") return;
    const url = new URL(request.url);
    if (url.origin !== self.location.origin || url.pathname.startsWith("/v1/")) return;

    if (request.mode === "navigate") {
        event.respondWith(fetch(request).catch(() => caches.match("/", { cacheName: CACHE })));
        return;
    }
    event.respondWith(
        caches.match(request, { cacheName: CACHE }).then((hit) => hit || fetch(request)),
    );
});
