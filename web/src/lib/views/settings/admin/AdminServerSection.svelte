<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { api } from "$lib/api/client";
  import type { ServerConfig } from "$lib/api/types";
  import { apiAction } from "$lib/utils/api-action";
  import { untrack } from "svelte";

  let serverConfig = $state<ServerConfig | null>(null);
  let serverConfigLoading = $state(false);
  let serverConfigSaving = $state(false);
  let metricsPassword = $state("");

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void loadServerConfig();
    });
  });

  async function loadServerConfig() {
    serverConfigLoading = true;
    const cfg = await apiAction(() => api.getServerConfig(), {
      errorFallback: "Failed to load server settings",
    });
    if (cfg) {
      serverConfig = {
        ...cfg,
        autoScanEnabled: cfg.autoScanEnabled ?? false,
        autoScanIntervalSec: cfg.autoScanIntervalSec ?? 300,
        scanWorkers: cfg.scanWorkers ?? 2,
      };
    }
    serverConfigLoading = false;
  }

  async function saveServerConfig(event: Event) {
    event.preventDefault();
    if (!serverConfig) return;
    serverConfigSaving = true;
    const payload: ServerConfig = {
      ...serverConfig,
      metricsPassword: metricsPassword || undefined,
    };
    const saved = await apiAction(() => api.saveServerConfig(payload), {
      success: "Server settings saved",
      errorFallback: "Failed to save server settings",
    });
    if (saved) {
      serverConfig = saved;
      metricsPassword = "";
    }
    serverConfigSaving = false;
  }
</script>

<div
  id="admin-server"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">Server</h2>
  <p class="mt-1 text-xs text-muted">
    Prometheus metrics, reverse proxy trust, CORS, and Content-Security-Policy.
  </p>
  {#if serverConfigLoading && !serverConfig}
    <div class="mt-3">
      <Skeleton height="10rem" rounded="lg" />
    </div>
  {:else if serverConfig}
    <form class="mt-4 space-y-4" onsubmit={saveServerConfig}>
      <div class="rounded-lg border border-border p-3 space-y-2">
        <p class="text-sm font-medium text-fg">Prometheus /metrics</p>
        <label class="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" bind:checked={serverConfig.metricsEnabled} />
          Enable metrics endpoint
        </label>
        <label class="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" bind:checked={serverConfig.metricsAuth} />
          Require HTTP Basic Auth
        </label>
        <input
          type="text"
          bind:value={serverConfig.metricsUsername}
          placeholder="Metrics username"
          class="field-input"
          disabled={!serverConfig.metricsAuth}
        />
        <input
          type="password"
          bind:value={metricsPassword}
          placeholder={serverConfig.metricsPasswordSet
            ? "Metrics password (unchanged)"
            : "Metrics password"}
          class="field-input"
          disabled={!serverConfig.metricsAuth}
        />
      </div>
      <div class="rounded-lg border border-border p-3 space-y-2">
        <p class="text-sm font-medium text-fg">Reverse proxy</p>
        <p class="text-xs text-muted">
          Comma-separated trusted proxy IPs or CIDRs (e.g. 127.0.0.1,::1,10.0.0.0/8). When set,
          X-Forwarded-For, X-Forwarded-Proto, and X-Forwarded-Host are honored from those addresses.
        </p>
        <input
          type="text"
          bind:value={serverConfig.trustedProxies}
          placeholder="127.0.0.1, ::1"
          class="field-input"
        />
      </div>
      <div class="rounded-lg border border-border p-3 space-y-2">
        <p class="text-sm font-medium text-fg">CORS</p>
        <label class="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" bind:checked={serverConfig.corsEnabled} />
          Enable CORS for API requests
        </label>
        <input
          type="text"
          bind:value={serverConfig.corsOrigins}
          placeholder="https://app.example.com, https://other.example.com"
          class="field-input"
          disabled={!serverConfig.corsEnabled}
        />
      </div>
      <div class="rounded-lg border border-border p-3 space-y-2">
        <p class="text-sm font-medium text-fg">Content-Security-Policy</p>
        <label class="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" bind:checked={serverConfig.cspEnabled} />
          Send CSP header
        </label>
        <textarea
          bind:value={serverConfig.cspPolicy}
          placeholder="Leave empty for the default self-hosted policy"
          class="field-input font-mono text-xs"
          disabled={!serverConfig.cspEnabled}></textarea>
      </div>
      <div class="rounded-lg border border-border p-3 space-y-2">
        <p class="text-sm font-medium text-fg">
          {i18n.t("admin.server.autoScanTitle")}
        </p>
        <label class="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" bind:checked={serverConfig.autoScanEnabled} />
          {i18n.t("admin.server.autoScanEnable")}
        </label>
        <label class="block text-xs text-muted">
          {i18n.t("admin.server.autoScanInterval")}
          <input
            type="number"
            min="60"
            max="86400"
            bind:value={serverConfig.autoScanIntervalSec}
            class="field-input mt-1"
          />
        </label>
      </div>
      <div class="rounded-lg border border-border p-3 space-y-2">
        <p class="text-sm font-medium text-fg">
          {i18n.t("admin.server.scanWorkersTitle")}
        </p>
        <p class="text-xs text-muted">{i18n.t("admin.server.scanWorkersHelp")}</p>
        <label class="block text-xs text-muted">
          {i18n.t("admin.server.scanWorkersLabel")}
          <input
            type="number"
            min="1"
            max="32"
            bind:value={serverConfig.scanWorkers}
            class="field-input mt-1"
          />
        </label>
      </div>
      <div class="pt-1">
        <Button type="submit" size="sm" loading={serverConfigSaving}
          >{i18n.t("admin.server.save")}</Button
        >
      </div>
    </form>
  {/if}
</div>
