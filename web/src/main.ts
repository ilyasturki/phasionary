import { mount } from "svelte";
import App from "./App.svelte";
import "./app.css";

mount(App, { target: document.getElementById("app")! });

if ("serviceWorker" in navigator && import.meta.env.PROD) {
    window.addEventListener("load", () => {
        void navigator.serviceWorker.register("/sw.js");
    });
}
