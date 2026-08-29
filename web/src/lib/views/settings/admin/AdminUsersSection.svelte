<script lang="ts">
  import { Copy, Users } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import PasswordStrength from "$lib/components/PasswordStrength.svelte";
  import GuestAdmin from "$lib/components/GuestAdmin.svelte";
  import InviteAdmin from "$lib/components/InviteAdmin.svelte";
  import { libraries } from "$lib/stores/libraries.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { api } from "$lib/api/client";
  import { scorePassword } from "$lib/utils/password-strength";
  import { apiAction, apiErrorMessage } from "$lib/utils/api-action";
  import type { GuestCredentials, Permission, User } from "$lib/api/types";
  import { EDITABLE_PERMISSIONS, PERMISSION_LABELS } from "$lib/permissions";
  import { untrack } from "svelte";

  let regUser = $state("");
  let regPass = $state("");
  let regMsg = $state<string | null>(null);

  let users = $state<User[]>([]);
  let usersLoading = $state(false);
  let resetUserId = $state<number | null>(null);
  let resetPass = $state("");
  let resetSaving = $state(false);

  let accessUserId = $state<number | null>(null);
  let accessLibraryIds = $state<number[]>([]);
  let accessSaving = $state(false);
  let accessMsg = $state<string | null>(null);

  let guestUser = $state("");
  let guestHours = $state(24);
  let guestSaving = $state(false);
  let newGuestCreds = $state<GuestCredentials | null>(null);
  let permUserId = $state<number | null>(null);
  let userPermissions = $state<Permission[]>([]);
  let permSaving = $state(false);

  let allowRegistration = $state(false);
  let authSettingsSaving = $state(false);

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void loadUsers();
      void loadAuthSettings();
    });
  });

  async function loadAuthSettings() {
    try {
      const s = await api.getAuthSettings();
      allowRegistration = s.allowRegistration;
    } catch {
      // ignore
    }
  }

  async function saveAllowRegistration() {
    authSettingsSaving = true;
    const s = await apiAction(
      () =>
        api.saveAuthSettings({
          allowRegistration,
          requireTotp: false,
        }),
      {
        success: "Registration setting saved",
        errorFallback: "Failed to save setting",
      },
    );
    if (s) allowRegistration = s.allowRegistration;
    authSettingsSaving = false;
  }

  async function loadUsers() {
    usersLoading = true;
    const list = await apiAction(() => api.listUsers(), {
      errorFallback: "Failed to load users",
    });
    if (list) users = list;
    usersLoading = false;
  }

  async function createGuest(event: Event) {
    event.preventDefault();
    guestSaving = true;
    newGuestCreds = null;
    const creds = await apiAction(() => api.createGuest(guestUser.trim(), guestHours), {
      success: "Guest account created",
      errorFallback: "Failed to create guest",
    });
    if (creds) {
      newGuestCreds = creds;
      guestUser = "";
      void loadUsers();
    }
    guestSaving = false;
  }

  async function copyGuestCreds() {
    if (!newGuestCreds) return;
    const text = `Username: ${newGuestCreds.user.username}\nPassword: ${newGuestCreds.password}`;
    try {
      await navigator.clipboard.writeText(text);
      toast.success(i18n.t("admin.guests.credsCopied"));
    } catch {
      toast.error(i18n.t("admin.guests.copyFailed"));
    }
  }

  function openPermissions(u: User) {
    permUserId = permUserId === u.id ? null : u.id;
    accessUserId = null;
    resetUserId = null;
    userPermissions = [...(u.permissions ?? [])];
  }

  function toggleUserPermission(perm: Permission) {
    if (userPermissions.includes(perm)) {
      userPermissions = userPermissions.filter((p) => p !== perm);
    } else {
      userPermissions = [...userPermissions, perm];
    }
  }

  async function saveUserPermissions(userId: number) {
    permSaving = true;
    const updated = await apiAction(() => api.setUserPermissions(userId, userPermissions), {
      success: "Permissions updated",
      errorFallback: "Failed to update permissions",
    });
    if (updated) {
      users = users.map((u) => (u.id === updated.id ? updated : u));
      permUserId = null;
    }
    permSaving = false;
  }

  async function deleteUser(userId: number) {
    const ok = await confirmDialog.ask({
      title: i18n.t("settings.deleteUserTitle"),
      message: i18n.t("settings.deleteUser"),
      confirmLabel: i18n.t("confirm.delete"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    const deleted = await apiAction(() => api.deleteUser(userId), {
      success: "User deleted",
      errorFallback: "Delete failed",
    });
    if (deleted) void loadUsers();
  }

  async function toggleAdmin(u: User) {
    const updated = await apiAction(() => api.setUserAdmin(u.id, !u.isAdmin), {
      success: u.isAdmin ? "Admin revoked" : "Admin granted",
      errorFallback: "Update failed",
    });
    if (updated) void loadUsers();
  }

  async function openLibraryAccess(u: User) {
    if (accessUserId === u.id) {
      accessUserId = null;
      return;
    }
    permUserId = null;
    accessUserId = u.id;
    accessMsg = null;
    accessLibraryIds = [];
    try {
      const res = await api.getUserLibraries(u.id);
      accessLibraryIds = [...res.libraryIds];
    } catch (e) {
      accessMsg = apiErrorMessage(e, "Failed to load library access");
    }
  }

  function toggleAccessLibrary(libId: number) {
    if (accessLibraryIds.includes(libId)) {
      accessLibraryIds = accessLibraryIds.filter((id) => id !== libId);
    } else {
      accessLibraryIds = [...accessLibraryIds, libId];
    }
  }

  async function saveLibraryAccess(event: Event) {
    event.preventDefault();
    if (accessUserId == null) return;
    accessSaving = true;
    accessMsg = null;
    try {
      await api.setUserLibraries(accessUserId, accessLibraryIds);
      toast.success("Library access updated");
    } catch (e) {
      accessMsg = apiErrorMessage(e, "Failed to save");
      toast.error(accessMsg);
    } finally {
      accessSaving = false;
    }
  }

  async function resetPassword(event: Event) {
    event.preventDefault();
    if (resetUserId == null || !scorePassword(resetPass, auth.passwordPolicy).valid) return;
    resetSaving = true;
    const ok = await apiAction(() => api.resetUserPassword(resetUserId!, resetPass), {
      success: "Password reset",
      errorFallback: "Reset failed",
    });
    if (ok) {
      resetUserId = null;
      resetPass = "";
    }
    resetSaving = false;
  }

  async function register(event: Event) {
    event.preventDefault();
    regMsg = null;
    if (!scorePassword(regPass, auth.passwordPolicy).valid) {
      regMsg = "Password does not meet strength requirements";
      return;
    }
    try {
      await auth.register(regUser, regPass);
      regUser = "";
      regPass = "";
      void loadUsers();
    } catch (e) {
      regMsg = apiErrorMessage(e, "Registration failed");
    }
  }
</script>

<div
  id="admin-users"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">Users</h2>
  <div class="mt-4 space-y-6">
    <div class="space-y-2 border-b border-border pb-4">
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={allowRegistration} />
        Allow public self-registration
      </label>
      <Button type="button" size="sm" loading={authSettingsSaving} onclick={saveAllowRegistration}>
        Save registration setting
      </Button>
    </div>
    <form class="space-y-3" onsubmit={register}>
      <p class="text-sm font-medium text-fg">Create user</p>
      <input type="text" placeholder="Username" bind:value={regUser} class="field-input" />
      <input type="password" placeholder="Password" bind:value={regPass} class="field-input" />
      <PasswordStrength password={regPass} policy={auth.passwordPolicy} />
      <div class="pt-1">
        <button
          type="submit"
          class="btn btn-primary"
          disabled={!scorePassword(regPass, auth.passwordPolicy).valid}
        >
          Create account
        </button>
      </div>
      {#if regMsg}<p class="text-xs text-muted">{regMsg}</p>{/if}
    </form>

    <form class="space-y-3 border-t border-border pt-6" onsubmit={createGuest}>
      <p class="text-sm font-medium text-fg">Create guest account</p>
      <p class="text-xs text-muted">
        Temporary accounts with auto-generated credentials. The password is shown only once.
      </p>
      <input
        type="text"
        placeholder="Username (optional, auto-generated if empty)"
        bind:value={guestUser}
        class="field-input"
      />
      <label class="block text-xs text-muted">
        Expires in hours
        <input type="number" min="1" max="8760" bind:value={guestHours} class="field-input mt-1" />
      </label>
      <div class="pt-1">
        <button type="submit" class="btn btn-primary" disabled={guestSaving}>
          {guestSaving ? "Creating..." : "Create guest"}
        </button>
      </div>
      {#if newGuestCreds}
        <div class="rounded-lg border border-border bg-elevated p-3 text-sm">
          <p class="font-medium text-fg">{newGuestCreds.user.username}</p>
          <p class="mt-1 font-mono text-xs text-fg">{newGuestCreds.password}</p>
          {#if newGuestCreds.user.expiresAt}
            <p class="mt-1 text-xs text-muted">
              Expires {new Date(newGuestCreds.user.expiresAt).toLocaleString()}
            </p>
          {/if}
          <button
            type="button"
            class="btn btn-ghost mt-2 text-xs ring-1 ring-border"
            onclick={copyGuestCreds}
          >
            <Copy size={14} />
            Copy credentials
          </button>
        </div>
      {/if}
    </form>

    <GuestAdmin />
    <InviteAdmin />

    <div class="space-y-3 border-t border-border pt-6">
      <p class="text-sm font-medium text-fg">{i18n.t("admin.users.title")}</p>
      {#if usersLoading}
        <Skeleton height="3rem" rounded="lg" />
      {:else if users.length === 0}
        <EmptyState
          size="sm"
          title={i18n.t("admin.users.emptyTitle")}
          body={i18n.t("admin.users.emptyBody")}
        >
          {#snippet icon(size)}
            <Users {size} />
          {/snippet}
        </EmptyState>
      {:else}
        <ul class="divide-y divide-border rounded-lg border border-border">
          {#each users as u (u.id)}
            <li class="px-3 py-2">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p class="text-sm font-medium text-fg">
                    {u.username}
                    {#if u.isAdmin}<span class="text-xs text-muted">(admin)</span>{/if}
                    {#if u.isGuest}<span class="text-xs text-muted">(guest)</span>{/if}
                  </p>
                  <p class="text-xs text-subtle">
                    Joined {new Date(u.createdAt).toLocaleDateString()}
                    {#if u.expiresAt}
                      · Expires {new Date(u.expiresAt).toLocaleString()}
                    {/if}
                  </p>
                </div>
                <div class="flex shrink-0 flex-wrap gap-1">
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={() => toggleAdmin(u)}
                  >
                    {u.isAdmin ? "Revoke admin" : "Make admin"}
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={() => openPermissions(u)}
                  >
                    Permissions
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={() => openLibraryAccess(u)}
                  >
                    Libraries
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={() => {
                      resetUserId = resetUserId === u.id ? null : u.id;
                      resetPass = "";
                    }}
                  >
                    Reset password
                  </button>
                  {#if u.id !== auth.user?.id}
                    <button
                      type="button"
                      class="btn btn-ghost text-xs text-danger"
                      onclick={() => deleteUser(u.id)}
                    >
                      Delete
                    </button>
                  {/if}
                </div>
              </div>
              {#if resetUserId === u.id}
                <form class="mt-2 space-y-2" onsubmit={resetPassword}>
                  <input
                    type="password"
                    bind:value={resetPass}
                    required
                    class="field-input"
                    placeholder="New password for {u.username}"
                  />
                  <PasswordStrength password={resetPass} policy={auth.passwordPolicy} />
                  <Button
                    type="submit"
                    size="sm"
                    loading={resetSaving}
                    disabled={!scorePassword(resetPass, auth.passwordPolicy).valid}
                  >
                    Confirm reset
                  </Button>
                </form>
              {/if}
              {#if permUserId === u.id && !u.isAdmin}
                <form
                  class="mt-2 space-y-2 rounded-lg border border-border p-3"
                  onsubmit={(e) => {
                    e.preventDefault();
                    void saveUserPermissions(u.id);
                  }}
                >
                  <p class="text-xs text-muted">Choose what this user can do.</p>
                  <ul class="space-y-1">
                    {#each EDITABLE_PERMISSIONS as perm (perm)}
                      <li>
                        <label class="flex items-center gap-2 text-sm text-fg">
                          <input
                            type="checkbox"
                            checked={userPermissions.includes(perm)}
                            onchange={() => toggleUserPermission(perm)}
                          />
                          {PERMISSION_LABELS[perm]}
                        </label>
                      </li>
                    {/each}
                  </ul>
                  <Button type="submit" size="sm" loading={permSaving}>Save permissions</Button>
                </form>
              {/if}
              {#if accessUserId === u.id}
                <form
                  class="mt-2 space-y-2 rounded-lg border border-border p-3"
                  onsubmit={saveLibraryAccess}
                >
                  <p class="text-xs text-muted">
                    Restrict to selected libraries. Leave all unchecked for access to every mount.
                  </p>
                  <ul class="space-y-1">
                    {#each libraries.items as lib (lib.id)}
                      <li>
                        <label class="flex items-center gap-2 text-sm text-fg">
                          <input
                            type="checkbox"
                            checked={accessLibraryIds.includes(lib.id)}
                            onchange={() => toggleAccessLibrary(lib.id)}
                          />
                          {lib.name}
                        </label>
                      </li>
                    {/each}
                  </ul>
                  <Button type="submit" size="sm" loading={accessSaving}>Save access</Button>
                  {#if accessMsg}<p class="text-xs text-danger">{accessMsg}</p>{/if}
                </form>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>
</div>
