<script lang="ts">
  import { ChevronLeft, ChevronRight, ScrollText } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { api } from "$lib/api/client";
  import type { AuditEntry } from "$lib/api/types";
  import { apiAction } from "$lib/utils/api-action";
  import { untrack } from "svelte";

  const PAGE_SIZE = 30;

  let auditItems = $state<AuditEntry[]>([]);
  let auditTotal = $state(0);
  let auditLoading = $state(false);
  let auditPage = $state(0);
  let auditAction = $state("");
  let auditQuery = $state("");
  let auditQueryDraft = $state("");

  let pageCount = $derived(Math.max(1, Math.ceil(auditTotal / PAGE_SIZE) || 1));

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void loadAudit();
    });
  });

  async function loadAudit() {
    auditLoading = true;
    const page = await apiAction(
      () => api.listAudit(PAGE_SIZE, auditPage * PAGE_SIZE, auditAction, auditQuery),
      { errorFallback: "Failed to load audit log" },
    );
    if (page) {
      auditItems = page.items;
      auditTotal = page.total;
    }
    auditLoading = false;
  }

  function applyFilters() {
    auditQuery = auditQueryDraft.trim();
    auditPage = 0;
    void loadAudit();
  }

  function goPage(next: number) {
    auditPage = Math.max(0, Math.min(next, pageCount - 1));
    void loadAudit();
  }

  function auditLabel(action: string): string {
    switch (action) {
      case "auth.login":
        return "Signed in";
      case "auth.logout":
        return "Signed out";
      case "user.create":
        return "User created";
      case "user.guest":
        return "Guest account created";
      case "user.rename":
        return "Username changed";
      case "password.change":
        return "Password changed";
      case "password.reset":
        return "Password reset";
      case "user.delete":
        return "User deleted";
      case "user.admin":
        return "Admin role changed";
      case "user.libraries":
        return "Library access changed";
      case "book.upload":
        return "Book uploaded";
      case "session.revoke":
        return "Session revoked";
      case "oidc.config":
        return "SSO settings changed";
      case "server.config":
        return "Server settings changed";
      case "apikey.create":
        return "API key created";
      case "apikey.revoke":
        return "API key revoked";
      default:
        return action;
    }
  }
</script>

<div
  id="admin-audit"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <div class="flex flex-wrap items-center justify-between gap-3">
    <h2 class="text-sm font-semibold text-fg">Audit log</h2>
    <button
      type="button"
      class="btn btn-ghost text-xs ring-1 ring-border"
      onclick={() => loadAudit()}
    >
      Refresh
    </button>
  </div>

  <form
    class="mt-3 flex flex-wrap items-center gap-2"
    onsubmit={(e) => {
      e.preventDefault();
      applyFilters();
    }}
  >
    <input
      type="search"
      class="field-input min-w-[12rem] flex-1 text-xs"
      placeholder={i18n.t("admin.audit.searchPlaceholder")}
      bind:value={auditQueryDraft}
    />
    <select
      class="field-input w-auto min-w-[10rem] text-xs"
      bind:value={auditAction}
      onchange={() => {
        auditPage = 0;
        void loadAudit();
      }}
    >
      <option value="">All actions</option>
      <option value="auth.login">Sign in</option>
      <option value="session.revoke">Session revoked</option>
      <option value="oidc.config">SSO settings</option>
      <option value="apikey.create">API key created</option>
      <option value="apikey.revoke">API key revoked</option>
      <option value="user.create">User created</option>
      <option value="user.delete">User deleted</option>
      <option value="password.reset">Password reset</option>
    </select>
    <button type="submit" class="btn btn-ghost text-xs ring-1 ring-border">
      {i18n.t("admin.audit.search")}
    </button>
  </form>

  {#if auditLoading && auditItems.length === 0}
    <Skeleton height="6rem" rounded="lg" class="mt-4" />
  {:else if auditItems.length === 0}
    <EmptyState
      size="sm"
      class="mt-2"
      title={i18n.t("admin.audit.emptyTitle")}
      body={i18n.t("admin.audit.emptyBody")}
    >
      {#snippet icon(size)}
        <ScrollText {size} />
      {/snippet}
    </EmptyState>
  {:else}
    <ul class="mt-4 max-h-80 space-y-2 overflow-y-auto text-xs" class:opacity-60={auditLoading}>
      {#each auditItems as entry (entry.id)}
        <li class="rounded-lg border border-border px-3 py-2">
          <p class="font-medium text-fg">{auditLabel(entry.action)}</p>
          <p class="text-muted">
            {entry.actorName}
            {#if entry.targetName && entry.targetName !== entry.actorName}
              -> {entry.targetName}
            {/if}
            {#if entry.details}
              <span class="text-subtle">({entry.details})</span>
            {/if}
          </p>
          <p class="text-subtle">
            {new Date(entry.createdAt).toLocaleString()}
            {#if entry.ip}
              · {entry.ip}
            {/if}
          </p>
        </li>
      {/each}
    </ul>

    <div class="mt-3 flex flex-wrap items-center justify-between gap-2 text-xs text-muted">
      <p>
        {i18n.t("admin.audit.pageStatus", {
          page: String(auditPage + 1),
          pages: String(pageCount),
          total: String(auditTotal),
        })}
      </p>
      <div class="flex items-center gap-1">
        <button
          type="button"
          class="btn btn-ghost ring-1 ring-border"
          disabled={auditPage <= 0 || auditLoading}
          aria-label={i18n.t("admin.audit.prev")}
          onclick={() => goPage(auditPage - 1)}
        >
          <ChevronLeft size={14} />
        </button>
        <button
          type="button"
          class="btn btn-ghost ring-1 ring-border"
          disabled={auditPage >= pageCount - 1 || auditLoading}
          aria-label={i18n.t("admin.audit.next")}
          onclick={() => goPage(auditPage + 1)}
        >
          <ChevronRight size={14} />
        </button>
      </div>
    </div>
  {/if}
</div>
