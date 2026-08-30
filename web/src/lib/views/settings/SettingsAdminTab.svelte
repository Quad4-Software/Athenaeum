<script lang="ts">
  import { i18n } from "$lib/stores/i18n.svelte";
  import AdminSystemSection from "./admin/AdminSystemSection.svelte";
  import AdminTasksSection from "./admin/AdminTasksSection.svelte";
  import AdminServerSection from "./admin/AdminServerSection.svelte";
  import AdminSMTPSection from "./admin/AdminSMTPSection.svelte";
  import AdminTTSSection from "./admin/AdminTTSSection.svelte";
  import AdminBackupSection from "./admin/AdminBackupSection.svelte";
  import AdminOIDCSection from "./admin/AdminOIDCSection.svelte";
  import AdminPocketIDSection from "./admin/AdminPocketIDSection.svelte";
  import AdminWebhooksSection from "./admin/AdminWebhooksSection.svelte";
  import AdminUsersSection from "./admin/AdminUsersSection.svelte";
  import AdminAuditSection from "./admin/AdminAuditSection.svelte";

  const adminSections = $derived([
    { id: "system", label: i18n.t("settings.adminSections.system") },
    { id: "tasks", label: i18n.t("settings.adminSections.tasks") },
    { id: "server", label: i18n.t("settings.adminSections.server") },
    { id: "smtp", label: i18n.t("settings.adminSections.smtp") },
    { id: "tts", label: i18n.t("settings.adminSections.tts") },
    { id: "backup", label: i18n.t("settings.adminSections.backup") },
    { id: "oidc", label: i18n.t("settings.adminSections.oidc") },
    { id: "pocketid", label: i18n.t("settings.adminSections.pocketid") },
    { id: "webhooks", label: i18n.t("settings.adminSections.webhooks") },
    { id: "users", label: i18n.t("settings.adminSections.users") },
    { id: "audit", label: i18n.t("settings.adminSections.audit") },
  ] as { id: string; label: string }[]);

  function scrollToAdminSection(id: string) {
    document.getElementById(`admin-${id}`)?.scrollIntoView({ behavior: "smooth", block: "start" });
  }
</script>

<div class="flex flex-col gap-6 xl:flex-row xl:items-start">
  <nav
    class="sticky top-0 z-10 flex shrink-0 gap-1 overflow-x-auto border-b border-border bg-bg/95 pb-3 backdrop-blur xl:top-0 xl:max-h-[calc(100dvh-8rem)] xl:w-40 xl:flex-col xl:items-stretch xl:gap-0.5 xl:overflow-x-visible xl:overflow-y-auto xl:border-b-0 xl:border-r xl:bg-transparent xl:pb-0 xl:pr-4 xl:backdrop-blur-none"
    aria-label="Administration sections"
  >
    {#each adminSections as section (section.id)}
      <button
        type="button"
        class="shrink-0 rounded-lg px-3 py-2 text-left text-sm text-muted transition-colors hover:bg-surface-hover hover:text-fg"
        onclick={() => scrollToAdminSection(section.id)}
      >
        {section.label}
      </button>
    {/each}
  </nav>

  <div class="min-w-0 flex-1 space-y-6">
    <AdminSystemSection />
    <AdminTasksSection />
    <AdminServerSection />
    <AdminSMTPSection />
    <AdminTTSSection />
    <AdminBackupSection />
    <AdminOIDCSection />
    <AdminPocketIDSection />
    <AdminWebhooksSection />
    <AdminUsersSection />
    <AdminAuditSection />
  </div>
</div>
