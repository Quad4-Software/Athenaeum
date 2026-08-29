import type {
  User,
  GuestCredentials,
  PasswordStrength,
  UserLibraries,
  UserSession,
  AuthMethods,
  OIDCConfig,
  OIDCDiscovery,
  APIKey,
  APIKeyCreated,
  APIDoc,
  ReaderPrefs,
  LoginResult,
  TOTPSetup,
  AuthSettings,
  InviteMeta,
} from "./types";
import { request } from "./core";

export const authApi = {
  login: (username: string, password: string, altcha?: string) =>
    request<LoginResult>("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password, altcha: altcha || undefined }),
    }),

  verifyTotp: (totpToken: string, code: string) =>
    request<User>("/api/auth/totp/verify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ totpToken, code }),
    }),

  totpSetup: () => request<TOTPSetup>("/api/auth/totp/setup", { method: "POST" }),

  totpEnable: (code: string) =>
    request<{ ok: boolean }>("/api/auth/totp/enable", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
    }),

  totpDisable: (password: string, code: string) =>
    request<{ ok: boolean }>("/api/auth/totp/disable", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password, code }),
    }),

  registerPublic: (username: string, password: string, altcha?: string) =>
    request<User>("/api/auth/register-public", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password, altcha: altcha || undefined }),
    }),

  getAuthSettings: () => request<AuthSettings>("/api/auth/settings"),

  saveAuthSettings: (settings: AuthSettings) =>
    request<AuthSettings>("/api/auth/settings", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(settings),
    }),

  logout: () => request<void>("/api/auth/logout", { method: "POST" }),

  me: () => request<User>("/api/auth/me"),

  authSetup: () =>
    request<{
      needed: boolean;
      authEnabled: boolean;
      allowRegistration?: boolean;
      passwordPolicy?: import("./types").PasswordPolicy;
      altcha?: import("./types").AltchaPublic;
    }>("/api/auth/setup"),

  setupAdmin: (username: string, password: string, altcha?: string) =>
    request<User>("/api/auth/setup", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password, altcha: altcha || undefined }),
    }),

  register: (username: string, password: string) =>
    request<User>("/api/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    }),

  createGuest: (username: string, expiresInHours: number) =>
    request<GuestCredentials>("/api/auth/users/guest", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: username || undefined, expiresInHours }),
    }),

  updateProfile: (username: string) =>
    request<User>("/api/auth/profile", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username }),
    }),

  changePassword: (currentPassword: string, newPassword: string) =>
    request<{ ok: boolean }>("/api/auth/password", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ currentPassword, newPassword }),
    }),

  listUsers: () => request<User[]>("/api/auth/users"),

  resetUserPassword: (userId: number, password: string) =>
    request<{ ok: boolean }>(`/api/auth/users/${userId}/password`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password }),
    }),

  deleteUser: (userId: number) =>
    request<{ ok: boolean }>(`/api/auth/users/${userId}`, { method: "DELETE" }),

  setUserAdmin: (userId: number, admin: boolean) =>
    request<User>(`/api/auth/users/${userId}/admin`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ admin }),
    }),

  setUserPermissions: (userId: number, permissions: string[]) =>
    request<User>(`/api/auth/users/${userId}/permissions`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ permissions }),
    }),

  getUserLibraries: (userId: number) =>
    request<UserLibraries>(`/api/auth/users/${userId}/libraries`),

  setUserLibraries: (userId: number, libraryIds: number[]) =>
    request<UserLibraries>(`/api/auth/users/${userId}/libraries`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ libraryIds }),
    }),

  checkPasswordStrength: (password: string) =>
    request<PasswordStrength>("/api/auth/password/check", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password }),
    }),

  authMethods: () => request<AuthMethods>("/api/auth/methods"),

  listSessions: () => request<UserSession[]>("/api/auth/sessions"),

  revokeSession: (sessionId: string, userId?: number) =>
    request<void>(
      `/api/auth/sessions/${encodeURIComponent(sessionId)}${userId != null ? `?userId=${userId}` : ""}`,
      { method: "DELETE" },
    ),

  revokeOtherSessions: () =>
    request<{ revoked: number }>("/api/auth/sessions", { method: "DELETE" }),

  listUserSessions: (userId: number) =>
    request<UserSession[]>(`/api/auth/users/${userId}/sessions`),

  getOIDCConfig: () => request<OIDCConfig>("/api/auth/oidc/config"),

  saveOIDCConfig: (config: OIDCConfig) =>
    request<OIDCConfig>("/api/auth/oidc/config", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    }),

  discoverOIDC: (issuerUrl: string) =>
    request<OIDCDiscovery>("/api/auth/oidc/discover", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ issuerUrl }),
    }),

  oidcLoginUrl: () => "/auth/oidc/login",

  listAPIKeys: () => request<APIKey[]>("/api/auth/api-keys"),

  createAPIKey: (name: string) =>
    request<APIKeyCreated>("/api/auth/api-keys", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    }),

  deleteAPIKey: (id: number) => request<void>(`/api/auth/api-keys/${id}`, { method: "DELETE" }),

  getAPIDocs: () => request<APIDoc>("/api/docs"),

  listGuests: (expiringWithinHours?: number) =>
    request<User[]>(
      `/api/auth/users/guests${expiringWithinHours != null ? `?expiringWithinHours=${expiringWithinHours}` : ""}`,
    ),

  bulkDeleteGuests: (ids: number[]) =>
    request<{ deleted: number }>("/api/auth/users/guests/bulk-delete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ids }),
    }),

  extendGuest: (userId: number, expiresInHours: number) =>
    request<User>(`/api/auth/users/guests/${userId}/extend`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ expiresInHours }),
    }),

  getReaderPrefs: () => request<ReaderPrefs>("/api/auth/reader-prefs"),

  saveReaderPrefs: (prefs: Record<string, unknown>) =>
    request<ReaderPrefs>("/api/auth/reader-prefs", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prefs }),
    }),

  getKindleEmail: () => request<{ email: string }>("/api/auth/kindle-email"),

  saveKindleEmail: (email: string) =>
    request<{ email: string }>("/api/auth/kindle-email", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email }),
    }),

  getInviteMeta: (token: string) => request<InviteMeta>(`/api/invite/${encodeURIComponent(token)}`),

  acceptInvite: (token: string, body: { username?: string; password?: string }) =>
    request<Record<string, unknown>>(`/api/invite/${encodeURIComponent(token)}/accept`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
};
