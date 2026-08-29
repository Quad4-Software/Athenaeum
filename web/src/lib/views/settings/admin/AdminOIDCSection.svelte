<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { api } from "$lib/api/client";
  import type { OIDCConfig } from "$lib/api/types";
  import { apiAction, apiErrorMessage } from "$lib/utils/api-action";
  import { untrack } from "svelte";

  let oidcConfig = $state<OIDCConfig | null>(null);
  let oidcLoading = $state(false);
  let oidcSaving = $state(false);
  let oidcDiscovering = $state(false);
  let oidcMsg = $state<string | null>(null);
  let oidcSecret = $state("");

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void loadOIDCConfig();
    });
  });

  async function loadOIDCConfig() {
    oidcLoading = true;
    oidcMsg = null;
    try {
      oidcConfig = await api.getOIDCConfig();
      oidcSecret = "";
    } catch (e) {
      oidcMsg = apiErrorMessage(e, "Failed to load OIDC settings");
    } finally {
      oidcLoading = false;
    }
  }

  async function discoverOIDC(event: Event) {
    event.preventDefault();
    if (!oidcConfig?.issuerUrl.trim()) return;
    oidcDiscovering = true;
    oidcMsg = null;
    const discovered = await apiAction(() => api.discoverOIDC(oidcConfig!.issuerUrl.trim()), {
      success: "OIDC endpoints populated",
      errorFallback: "Discovery failed",
    });
    if (discovered) {
      oidcConfig = {
        ...oidcConfig!,
        issuerUrl: discovered.issuerUrl,
        authorizeUrl: discovered.authorizeUrl,
        tokenUrl: discovered.tokenUrl,
        userinfoUrl: discovered.userinfoUrl,
        jwksUrl: discovered.jwksUrl,
        logoutUrl: discovered.logoutUrl ?? "",
      };
    } else {
      oidcMsg = "Discovery failed";
    }
    oidcDiscovering = false;
  }

  async function saveOIDCConfig(event: Event) {
    event.preventDefault();
    if (!oidcConfig) return;
    oidcSaving = true;
    oidcMsg = null;
    const payload: OIDCConfig = {
      ...oidcConfig,
      clientSecret: oidcSecret || oidcConfig.clientSecret || "",
    };
    const saved = await apiAction(() => api.saveOIDCConfig(payload), {
      success: "Authentication settings saved",
      errorFallback: "Failed to save OIDC settings",
    });
    if (saved) {
      oidcConfig = saved;
      oidcSecret = "";
    } else {
      oidcMsg = "Failed to save OIDC settings";
    }
    oidcSaving = false;
  }
</script>

<div
  id="admin-oidc"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">OpenID Connect</h2>
  <p class="mt-1 text-xs text-muted">
    Configure Keycloak, Auth0, Authentik, Authelia, or any OIDC provider. Callback URL:
    <code class="text-fg">/auth/oidc/callback</code>
  </p>
  {#if oidcLoading}
    <Skeleton height="8rem" rounded="lg" class="mt-4" />
  {:else if oidcConfig}
    <form class="mt-4 space-y-3" onsubmit={saveOIDCConfig}>
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={oidcConfig.enabled} />
        Enable OpenID Connect
      </label>
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={oidcConfig.loginLocal} />
        Allow local username/password login
      </label>
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={oidcConfig.autoRegister} />
        Auto-register new SSO users
      </label>
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={oidcConfig.autoLaunch} />
        Auto-redirect to SSO when local login is disabled
      </label>
      <div class="flex gap-2">
        <input
          type="url"
          bind:value={oidcConfig.issuerUrl}
          placeholder="Issuer URL"
          class="field-input flex-1"
        />
        <Button type="button" size="sm" loading={oidcDiscovering} onclick={discoverOIDC}>
          Auto-populate
        </Button>
      </div>
      <input
        type="url"
        bind:value={oidcConfig.authorizeUrl}
        placeholder="Authorize URL"
        class="field-input"
      />
      <input
        type="url"
        bind:value={oidcConfig.tokenUrl}
        placeholder="Token URL"
        class="field-input"
      />
      <input
        type="url"
        bind:value={oidcConfig.userinfoUrl}
        placeholder="Userinfo URL"
        class="field-input"
      />
      <input
        type="url"
        bind:value={oidcConfig.jwksUrl}
        placeholder="JWKS URL"
        class="field-input"
      />
      <input
        type="text"
        bind:value={oidcConfig.clientId}
        placeholder="Client ID"
        class="field-input"
      />
      <input
        type="password"
        bind:value={oidcSecret}
        placeholder={oidcConfig.clientSecretSet ? "Client secret (unchanged)" : "Client secret"}
        class="field-input"
      />
      <input
        type="text"
        bind:value={oidcConfig.buttonText}
        placeholder="Sign-in button text"
        class="field-input"
      />
      <select bind:value={oidcConfig.matchBy} class="field-input">
        <option value="username">Match existing users by username</option>
        <option value="email">Match existing users by email</option>
        <option value="sub">Match only by OIDC subject</option>
      </select>
      <input
        type="text"
        bind:value={oidcConfig.groupClaim}
        placeholder="Group claim (default: groups)"
        class="field-input"
      />
      <input
        type="text"
        bind:value={oidcConfig.adminGroups}
        placeholder="Admin groups (comma-separated)"
        class="field-input"
      />
      <div class="pt-1">
        <Button type="submit" size="sm" loading={oidcSaving}>Save SSO settings</Button>
      </div>
      {#if oidcMsg}<p class="text-xs text-danger">{oidcMsg}</p>{/if}
    </form>
  {/if}
</div>
