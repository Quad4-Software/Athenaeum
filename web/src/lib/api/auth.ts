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
import type { EpubDisplayPrefs } from "$lib/reader/epub-reader";
import { request } from "./core";
import { opURL } from "./op";

export const authApi = {
  login: (username: string, password: string, altcha?: string) =>
    request<LoginResult>(opURL("POST__api_auth_login"), {
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

  logout: () => request<void>(opURL("POST__api_auth_logout"), { method: "POST" }),

  me: () => request<User>(opURL("GET__api_auth_me")),

  authSetup: () =>
    request<{
      needed: boolean;
      authEnabled: boolean;
      allowRegistration?: boolean;
      passwordPolicy?: import("./types").PasswordPolicy;
      altcha?: import("./types").AltchaPublic;
    }>(opURL("GET__api_auth_setup")),

  setupAdmin: (username: string, password: string, altcha?: string) =>
    request<User>(opURL("POST__api_auth_setup"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password, altcha: altcha || undefined }),
    }),

  register: (username: string, password: string) =>
    request<User>(opURL("POST__api_auth_register"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    }),

  createGuest: (username: string, expiresInHours: number) =>
    request<GuestCredentials>(opURL("POST__api_auth_users_guest"), {
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

  listUsers: () => request<User[]>(opURL("GET__api_auth_users")),

  resetUserPassword: (userId: number, password: string) =>
    request<{ ok: boolean }>(opURL("PUT__api_auth_users__id__password", { id: userId }), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ password }),
    }),

  deleteUser: (userId: number) =>
    request<{ ok: boolean }>(opURL("DELETE__api_auth_users__id", { id: userId }), {
      method: "DELETE",
    }),

  setUserAdmin: (userId: number, admin: boolean) =>
    request<User>(opURL("PUT__api_auth_users__id__admin", { id: userId }), {
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
    request<UserLibraries>(opURL("GET__api_auth_users__id__libraries", { id: userId })),

  setUserLibraries: (userId: number, libraryIds: number[]) =>
    request<UserLibraries>(opURL("PUT__api_auth_users__id__libraries", { id: userId }), {
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

  authMethods: () => request<AuthMethods>(opURL("GET__api_auth_methods")),

  listSessions: () => request<UserSession[]>(opURL("GET__api_auth_sessions")),

  revokeSession: (sessionId: string, userId?: number) =>
    request<void>(
      `${opURL("DELETE__api_auth_sessions__id", { id: sessionId })}${userId != null ? `?userId=${userId}` : ""}`,
      { method: "DELETE" },
    ),

  revokeOtherSessions: () =>
    request<{ revoked: number }>(opURL("DELETE__api_auth_sessions"), { method: "DELETE" }),

  revokeAllSessions: () =>
    request<{ revoked: number }>(`${opURL("DELETE__api_auth_sessions")}?all=true`, {
      method: "DELETE",
    }),

  listUserSessions: (userId: number) =>
    request<UserSession[]>(opURL("GET__api_auth_users__id__sessions", { id: userId })),

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

  listAPIKeys: () => request<APIKey[]>(opURL("GET__api_auth_api_keys")),

  createAPIKey: (name: string) =>
    request<APIKeyCreated>(opURL("POST__api_auth_api_keys"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    }),

  deleteAPIKey: (id: number) =>
    request<void>(opURL("DELETE__api_auth_api_keys__id", { id }), { method: "DELETE" }),

  getAPIDocs: () => request<APIDoc>(opURL("GET__api_docs")),

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

  getReaderPrefs: () => request<ReaderPrefs>(opURL("GET__api_auth_reader_prefs")),

  saveReaderPrefs: (prefs: EpubDisplayPrefs) =>
    request<ReaderPrefs>(opURL("PUT__api_auth_reader_prefs"), {
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

  getInviteMeta: (token: string) => request<InviteMeta>(opURL("GET__api_invite__token", { token })),

  acceptInvite: (token: string, body: { username?: string; password?: string }) =>
    request<Record<string, unknown>>(opURL("POST__api_invite__token__accept", { token }), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
};
