<script lang="ts">
  import { LogOut, Monitor } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import PasswordStrength from "$lib/components/PasswordStrength.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { ApiError, api } from "$lib/api/client";
  import { scorePassword } from "$lib/utils/password-strength";
  import { totpQrDataUrl } from "$lib/utils/totp-qr";
  import type { UserSession } from "$lib/api/types";
  import { untrack } from "svelte";

  let profileName = $state("");
  let profileMsg = $state<string | null>(null);
  let profileSaving = $state(false);

  let currentPass = $state("");
  let newPass = $state("");
  let passMsg = $state<string | null>(null);
  let passSaving = $state(false);

  let totpSecret = $state("");
  let totpUrl = $state("");
  let totpQr = $state("");
  let totpCode = $state("");
  let totpPass = $state("");
  let totpMsg = $state<string | null>(null);
  let totpBusy = $state(false);

  let sessions = $state<UserSession[]>([]);
  let sessionsLoading = $state(false);

  $effect(() => {
    if (!auth.user) return;
    untrack(() => {
      profileName = auth.user!.username;
      void loadSessions();
    });
  });

  async function loadSessions() {
    sessionsLoading = true;
    try {
      sessions = await api.listSessions();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to load sessions");
    } finally {
      sessionsLoading = false;
    }
  }

  async function revokeSession(sess: UserSession) {
    const ok = await confirmDialog.ask({
      title: i18n.t("settings.revokeSessionTitle"),
      message: i18n.t("settings.revokeSession", {
        device: sess.device || i18n.t("settings.revokeSessionDevice"),
      }),
      confirmLabel: i18n.t("settings.revoke"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    try {
      await api.revokeSession(sess.id);
      if (sess.current) {
        await auth.logout();
        return;
      }
      toast.success("Session revoked");
      void loadSessions();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to revoke session");
    }
  }

  async function revokeOtherSessions() {
    const ok = await confirmDialog.ask({
      title: i18n.t("settings.revokeOtherTitle"),
      message: i18n.t("settings.revokeOther"),
      confirmLabel: i18n.t("settings.signOutOthers"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    try {
      const res = await api.revokeOtherSessions();
      toast.success(`Revoked ${res.revoked} session(s)`);
      void loadSessions();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to revoke sessions");
    }
  }

  async function saveProfile(event: Event) {
    event.preventDefault();
    profileMsg = null;
    if (!profileName.trim() || profileName === auth.user?.username) return;
    profileSaving = true;
    try {
      await auth.updateProfile(profileName.trim());
    } catch (e) {
      profileMsg = e instanceof ApiError ? e.message : "Failed to update profile";
    } finally {
      profileSaving = false;
    }
  }

  async function changePassword(event: Event) {
    event.preventDefault();
    passMsg = null;
    if (!scorePassword(newPass, auth.passwordPolicy).valid) {
      passMsg = "Password does not meet strength requirements";
      return;
    }
    passSaving = true;
    try {
      await auth.changePassword(currentPass, newPass);
      currentPass = "";
      newPass = "";
    } catch (e) {
      passMsg = e instanceof ApiError ? e.message : "Failed to change password";
    } finally {
      passSaving = false;
    }
  }

  async function setupTotp() {
    totpBusy = true;
    totpMsg = null;
    try {
      const res = await api.totpSetup();
      totpSecret = res.secret;
      totpUrl = res.otpauthUrl;
      totpQr = res.otpauthUrl ? await totpQrDataUrl(res.otpauthUrl) : "";
      totpCode = "";
    } catch (e) {
      totpMsg = e instanceof ApiError ? e.message : "Failed to start 2FA setup";
    } finally {
      totpBusy = false;
    }
  }

  async function enableTotp(event: Event) {
    event.preventDefault();
    totpBusy = true;
    totpMsg = null;
    try {
      await api.totpEnable(totpCode);
      totpSecret = "";
      totpUrl = "";
      totpQr = "";
      totpCode = "";
      if (auth.user) auth.user = { ...auth.user, totpEnabled: true };
      toast.success("Two-factor authentication enabled");
    } catch (e) {
      totpMsg = e instanceof ApiError ? e.message : "Could not enable 2FA";
    } finally {
      totpBusy = false;
    }
  }

  async function disableTotp(event: Event) {
    event.preventDefault();
    totpBusy = true;
    totpMsg = null;
    try {
      await api.totpDisable(totpPass, totpCode);
      totpPass = "";
      totpCode = "";
      if (auth.user) auth.user = { ...auth.user, totpEnabled: false };
      toast.success("Two-factor authentication disabled");
    } catch (e) {
      totpMsg = e instanceof ApiError ? e.message : "Could not disable 2FA";
    } finally {
      totpBusy = false;
    }
  }

  async function logout() {
    await auth.logout();
  }
</script>

<div class="space-y-6">
  <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h2 class="text-sm font-semibold text-fg">Account</h2>
        {#if auth.user}
          <p class="mt-2 text-sm text-muted">
            Signed in as <span class="text-fg">{auth.user.username}</span>
            {#if auth.user.isAdmin}<span class="text-muted">(admin)</span>{/if}
          </p>
        {:else}
          <p class="mt-2 text-sm text-muted">Not signed in.</p>
        {/if}
      </div>
      {#if auth.user}
        <button class="btn btn-ghost shrink-0 ring-1 ring-border" onclick={logout}>
          <LogOut size={16} />
          {i18n.t("auth.signOut")}
        </button>
      {/if}
    </div>
  </div>

  {#if auth.user}
    <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
      <form class="space-y-3" onsubmit={saveProfile}>
        <p class="text-sm font-medium text-fg">Profile</p>
        <input
          type="text"
          bind:value={profileName}
          required
          minlength="2"
          class="field-input"
          placeholder="Username"
        />
        <div class="pt-1">
          <Button
            type="submit"
            size="sm"
            loading={profileSaving}
            disabled={profileName === auth.user.username}
          >
            Save username
          </Button>
        </div>
        {#if profileMsg}<p class="text-xs text-danger">{profileMsg}</p>{/if}
      </form>
    </div>

    <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
      <form class="space-y-3" onsubmit={changePassword}>
        <p class="text-sm font-medium text-fg">Change password</p>
        {#if auth.user.localAuth === false}
          <p class="text-xs text-muted">This account uses single sign-on only.</p>
        {:else}
          <input
            type="password"
            bind:value={currentPass}
            required
            autocomplete="current-password"
            class="field-input"
            placeholder="Current password"
          />
          <input
            type="password"
            bind:value={newPass}
            required
            autocomplete="new-password"
            class="field-input"
            placeholder="New password"
          />
          <PasswordStrength password={newPass} policy={auth.passwordPolicy} />
          <div class="pt-1">
            <Button
              type="submit"
              size="sm"
              loading={passSaving}
              disabled={!scorePassword(newPass, auth.passwordPolicy).valid}
            >
              Update password
            </Button>
          </div>
        {/if}
        {#if passMsg}<p class="text-xs text-danger">{passMsg}</p>{/if}
      </form>
    </div>

    <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
      <p class="text-sm font-medium text-fg">Two-factor authentication</p>
      <p class="mt-1 text-xs text-muted">Protect your account with an authenticator app (TOTP).</p>
      {#if auth.user.totpEnabled}
        <form class="mt-3 space-y-3" onsubmit={disableTotp}>
          <input
            type="password"
            bind:value={totpPass}
            class="field-input"
            placeholder="Current password"
            required={auth.user.localAuth !== false}
          />
          <input
            type="text"
            inputmode="numeric"
            bind:value={totpCode}
            class="field-input"
            placeholder="Authenticator code"
            required
          />
          <Button type="submit" size="sm" loading={totpBusy}>Disable 2FA</Button>
        </form>
      {:else if totpSecret}
        <form class="mt-3 space-y-3" onsubmit={enableTotp}>
          {#if totpQr}
            <img
              src={totpQr}
              alt="Authenticator QR code"
              width="192"
              height="192"
              class="rounded-[var(--radius-control)] border border-border bg-white p-2"
            />
          {/if}
          <p class="text-xs text-muted">Scan with your authenticator app, or enter the secret manually.</p>
          <p class="break-all text-xs text-muted">Secret: {totpSecret}</p>
          {#if totpUrl}
            <a class="block break-all text-xs text-primary underline" href={totpUrl}>{totpUrl}</a>
          {/if}
          <input
            type="text"
            inputmode="numeric"
            bind:value={totpCode}
            class="field-input"
            placeholder="Enter code to confirm"
            required
          />
          <Button type="submit" size="sm" loading={totpBusy}>Enable 2FA</Button>
        </form>
      {:else}
        <div class="mt-3">
          <Button type="button" size="sm" loading={totpBusy} onclick={setupTotp}>
            Set up authenticator
          </Button>
        </div>
      {/if}
      {#if totpMsg}<p class="mt-2 text-xs text-danger">{totpMsg}</p>{/if}
    </div>

    <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
      <div class="space-y-2">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <p class="text-sm font-medium text-fg">Active sessions</p>
          {#if sessions.length > 1}
            <button
              type="button"
              class="btn btn-ghost text-xs ring-1 ring-border"
              onclick={revokeOtherSessions}
            >
              Sign out other devices
            </button>
          {/if}
        </div>
        {#if sessionsLoading}
          <Skeleton height="4rem" rounded="lg" />
        {:else if sessions.length === 0}
          <EmptyState
            size="sm"
            title={i18n.t("settings.sessionsEmptyTitle")}
            body={i18n.t("settings.sessionsEmptyBody")}
          >
            {#snippet icon(size)}
              <Monitor {size} />
            {/snippet}
          </EmptyState>
        {:else}
          <ul class="divide-y divide-border rounded-lg border border-border">
            {#each sessions as sess (sess.id)}
              <li class="px-3 py-2">
                <div class="flex flex-wrap items-start justify-between gap-2">
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-fg">
                      {sess.device || "Unknown device"}
                      {#if sess.current}
                        <span class="ml-1 text-xs text-primary">(this device)</span>
                      {/if}
                    </p>
                    <p class="text-xs text-muted">
                      {sess.authMethod === "oidc" ? "SSO" : "Local"}
                      {#if sess.ip}
                        · {sess.ip}
                      {/if}
                    </p>
                    <p class="text-xs text-subtle">
                      Last active {new Date(sess.lastSeenAt).toLocaleString()}
                      · expires {new Date(sess.expiresAt).toLocaleDateString()}
                    </p>
                  </div>
                  {#if !sess.current}
                    <button
                      type="button"
                      class="btn btn-ghost text-xs text-danger"
                      onclick={() => revokeSession(sess)}
                    >
                      Revoke
                    </button>
                  {/if}
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>
  {/if}
</div>
