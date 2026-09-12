import type {
  AuditPage,
  SystemStats,
  ServerConfig,
  IntegrityReport,
  MaintenanceStatus,
  SMTPSettingsPublic,
  PocketIDSettingsPublic,
  Invite,
  InviteCreateResult,
  Webhook,
  WebhookDelivery,
} from "./types";
import { request } from "./core";
import { opURL } from "./op";

export const adminApi = {
  listAudit: (limit = 50, offset = 0, action = "", q = "") => {
    const params = new URLSearchParams({
      limit: String(limit),
      offset: String(offset),
    });
    if (action) params.set("action", action);
    if (q.trim()) params.set("q", q.trim());
    return request<AuditPage>(`${opURL("GET__api_auth_audit")}?${params}`);
  },

  getSystemStats: () => request<SystemStats>(opURL("GET__api_system_stats")),

  getServerConfig: () => request<ServerConfig>(opURL("GET__api_admin_server")),

  saveServerConfig: (config: ServerConfig) =>
    request<ServerConfig>(opURL("PUT__api_admin_server"), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    }),

  getSMTP: () => request<SMTPSettingsPublic>("/api/admin/smtp"),

  saveSMTP: (config: {
    enabled: boolean;
    host: string;
    port: number;
    username: string;
    password?: string;
    fromAddr: string;
    useTls: boolean;
  }) =>
    request<SMTPSettingsPublic>("/api/admin/smtp", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    }),

  getPocketID: () => request<PocketIDSettingsPublic>(opURL("GET__api_admin_pocketid")),

  savePocketID: (config: {
    enabled: boolean;
    baseUrl: string;
    apiKey?: string;
    defaultGroupIds: string[];
  }) =>
    request<PocketIDSettingsPublic>(opURL("PUT__api_admin_pocketid"), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    }),

  testPocketID: () =>
    request<{ ok: string }>(opURL("POST__api_admin_pocketid_test"), { method: "POST" }),

  applyPocketIDOIDC: () =>
    request<Record<string, unknown>>(opURL("POST__api_admin_pocketid_apply_oidc"), {
      method: "POST",
    }),

  listInvites: (status?: string) =>
    request<Invite[]>(
      `${opURL("GET__api_invites")}${status ? `?status=${encodeURIComponent(status)}` : ""}`,
    ),

  createInvite: (body: {
    kind: string;
    email?: string;
    username?: string;
    permissions?: string[];
    expiresInHours?: number;
    guestExpiresInHours?: number;
    provisionPocketId?: boolean;
  }) =>
    request<InviteCreateResult>(opURL("POST__api_invites"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),

  revokeInvite: (id: number) =>
    request<{ ok: string }>(opURL("DELETE__api_invites__id", { id }), { method: "DELETE" }),

  listWebhooks: () => request<Webhook[]>(opURL("GET__api_admin_webhooks")),

  createWebhook: (body: { url: string; secret?: string; events: string[]; enabled?: boolean }) =>
    request<Webhook>(opURL("POST__api_admin_webhooks"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),

  updateWebhook: (
    id: number,
    body: { url?: string; secret?: string; events?: string[]; enabled?: boolean },
  ) =>
    request<Webhook>(opURL("PUT__api_admin_webhooks__id", { id }), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),

  deleteWebhook: (id: number) =>
    request<{ ok: string }>(opURL("DELETE__api_admin_webhooks__id", { id }), { method: "DELETE" }),

  listWebhookDeliveries: (id: number, limit = 50, offset = 0) =>
    request<WebhookDelivery[]>(
      `${opURL("GET__api_admin_webhooks__id__deliveries", { id })}?limit=${limit}&offset=${offset}`,
    ),

  testWebhook: (id: number) =>
    request<{ ok: string }>(opURL("POST__api_admin_webhooks__id__test", { id }), {
      method: "POST",
    }),

  adminContentIndex: () =>
    request<{ status: string }>("/api/admin/content-index", { method: "POST" }),

  adminTaskStatus: () => request<MaintenanceStatus>("/api/admin/tasks/status"),

  adminVerifyIntegrity: () =>
    request<IntegrityReport>("/api/admin/tasks/verify", { method: "POST" }),

  adminPruneMissing: () =>
    request<{ removed: number }>("/api/admin/tasks/prune-missing", { method: "POST" }),

  adminCleanupCovers: () =>
    request<{ removed: number }>("/api/admin/tasks/cleanup-covers", { method: "POST" }),

  adminRegenerateCovers: (libraryId?: number) =>
    request<{ ok: boolean }>("/api/admin/tasks/regenerate-covers", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(libraryId != null ? { libraryId } : {}),
    }),

  adminCleanupSeries: () =>
    request<{ updated: number }>("/api/admin/tasks/cleanup-series", { method: "POST" }),
  adminCleanupText: () =>
    request<{ updated: number }>("/api/admin/tasks/cleanup-text", { method: "POST" }),

  exportConfig: () => request<Record<string, unknown>>("/api/admin/config/export"),

  importConfig: (body: Record<string, unknown>) =>
    request<{ ok: boolean }>("/api/admin/config/import", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),

  restoreBackup: (file: File) => {
    const form = new FormData();
    form.append("file", file);
    // FormData body: request() adds Accept + CSRF but no Content-Type, so the
    // multipart boundary is still set by fetch.
    return request<{ status: string; message: string }>("/api/admin/restore", {
      method: "POST",
      body: form,
    });
  },
};
