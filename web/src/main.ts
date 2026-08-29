import { mount } from "svelte";
import "./app.css";
import { api } from "$lib/api/client";
import { initBrand } from "$lib/brand";
import { initPwa } from "$lib/pwa/register";
import { i18n } from "$lib/stores/i18n.svelte";
import { registerGlobalErrorHandlers } from "$lib/telemetry/global-errors";
import App from "./App.svelte";

registerGlobalErrorHandlers();
initBrand();

const target = document.getElementById("app");
if (!target) throw new Error("missing #app mount target");
target.replaceChildren();

void i18n.init().then(async () => {
  initPwa(i18n.t("app.offlineReady"));
  try {
    const health = await api.health();
    if (health.telemetry) {
      const { initSentry } = await import("$lib/telemetry/sentry");
      await initSentry(health.telemetry);
    }
  } catch {
    // health may fail before the server is ready
  }
});

const app = mount(App, { target });

export default app;
