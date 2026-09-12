import { api, ApiError, restoreSession, ensureCsrf, clearCsrfCache } from "$lib/api/client";
import { onUnauthorized, onForbidden, type AuthRedirectReason } from "$lib/api/session";
import { unauthorizedRedirect, isAuthPagePathname } from "$lib/auth-redirect";
import { router } from "$lib/router.svelte";
import { routes } from "$lib/routes";
import { toast } from "$lib/stores/toast.svelte";
import { i18n } from "$lib/stores/i18n.svelte";
import type { AuthMethods, AltchaPublic, PasswordPolicy, User } from "$lib/api/types";
import { DEFAULT_PASSWORD_POLICY } from "$lib/utils/password-strength";

class AuthStore {
  user = $state<User | null>(null);
  authEnabled = $state(false);
  setupNeeded = $state(false);
  methods = $state<AuthMethods | null>(null);
  altcha = $state<AltchaPublic | null>(null);
  passwordPolicy = $state<PasswordPolicy>({ ...DEFAULT_PASSWORD_POLICY });
  loading = $state(true);
  error = $state<string | null>(null);

  private initPromise: Promise<void> | null = null;

  constructor() {
    onUnauthorized((reason) => this.handleUnauthorized(reason));
    onForbidden(() => this.handleForbidden());
  }

  handleForbidden() {
    if (typeof window === "undefined") return;
    if (isAuthPagePathname(router.appPathname())) return;
    router.navigate(routes.error("forbidden"), true);
  }

  handleUnauthorized(reason: AuthRedirectReason = "required") {
    this.user = null;
    if (typeof window === "undefined") return;
    const target = unauthorizedRedirect(router.appPathname(), reason);
    if (target) router.navigate(target, true);
  }

  async init() {
    if (this.initPromise) return this.initPromise;
    this.initPromise = this.bootstrap();
    return this.initPromise;
  }

  private applyPasswordPolicy(policy?: PasswordPolicy | null) {
    if (policy) this.passwordPolicy = { ...DEFAULT_PASSWORD_POLICY, ...policy };
  }

  private async bootstrap() {
    this.loading = true;
    this.error = null;
    try {
      // Plant CSRF cookie early so setup/login POSTs are ready.
      void ensureCsrf().catch(() => undefined);
      const setup = await api.authSetup();
      this.setupNeeded = setup.needed;
      this.authEnabled = setup.authEnabled;
      this.altcha = setup.altcha?.enabled ? setup.altcha : null;
      this.applyPasswordPolicy(setup.passwordPolicy);
      if (setup.needed) {
        this.user = null;
        this.methods = null;
        return;
      }
      if (setup.authEnabled) {
        this.methods = await api.authMethods();
        this.applyPasswordPolicy(this.methods.passwordPolicy);
        if (this.methods.altcha?.enabled) {
          this.altcha = this.methods.altcha;
        }
        try {
          this.user = await api.me();
        } catch (e) {
          if (e instanceof ApiError && e.status === 401) {
            const restored = await restoreSession();
            if (restored) {
              clearCsrfCache();
              try {
                this.user = await api.me();
              } catch (retryErr) {
                if (retryErr instanceof ApiError && retryErr.status === 401) {
                  this.user = null;
                } else {
                  throw retryErr;
                }
              }
            } else {
              this.user = null;
            }
          } else {
            throw e;
          }
        }
      }
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        this.user = null;
      } else {
        this.error = e instanceof Error ? e.message : i18n.t("auth.loadFailed");
        toast.error(this.error);
      }
    } finally {
      this.loading = false;
    }
  }

  async setup(username: string, password: string, altchaPayload?: string) {
    this.user = await api.setupAdmin(username, password, altchaPayload);
    clearCsrfCache();
    this.setupNeeded = false;
    this.authEnabled = true;
    toast.success(i18n.t("auth.adminCreated"));
  }

  async login(username: string, password: string, altchaPayload?: string) {
    this.error = null;
    try {
      const result = await api.login(username, password, altchaPayload);
      clearCsrfCache();
      if ("needsTotp" in result && result.needsTotp) {
        return result;
      }
      this.user = result as import("$lib/api/types").User;
      this.authEnabled = true;
      this.methods = await api.authMethods();
      this.applyPasswordPolicy(this.methods.passwordPolicy);
      if (this.methods.altcha?.enabled) {
        this.altcha = this.methods.altcha;
      }
      toast.success(i18n.t("auth.signedIn"));
      return result;
    } catch (e) {
      this.error = e instanceof ApiError ? e.message : i18n.t("auth.loginFailed");
      toast.error(this.error);
      throw e;
    }
  }

  async verifyTotp(totpToken: string, code: string) {
    this.error = null;
    try {
      this.user = await api.verifyTotp(totpToken, code);
      clearCsrfCache();
      this.authEnabled = true;
      this.methods = await api.authMethods();
      this.applyPasswordPolicy(this.methods.passwordPolicy);
      toast.success(i18n.t("auth.signedIn"));
    } catch (e) {
      this.error = e instanceof ApiError ? e.message : i18n.t("auth.totpInvalid");
      toast.error(this.error);
      throw e;
    }
  }

  async registerPublic(username: string, password: string, altchaPayload?: string) {
    this.user = await api.registerPublic(username, password, altchaPayload);
    clearCsrfCache();
    this.authEnabled = true;
    this.methods = await api.authMethods();
    this.applyPasswordPolicy(this.methods.passwordPolicy);
    toast.success(i18n.t("auth.accountCreated"));
  }

  async logout() {
    try {
      await api.logout();
    } catch {
      // session may already be gone
    }
    clearCsrfCache();
    this.user = null;
    toast.info(i18n.t("auth.signedOut"));
    this.handleUnauthorized("logged_out");
  }

  async register(username: string, password: string) {
    try {
      const u = await api.register(username, password);
      this.authEnabled = true;
      toast.success(i18n.t("auth.accountCreated"));
      return u;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("auth.registerFailed"));
      throw e;
    }
  }

  async updateProfile(username: string) {
    this.user = await api.updateProfile(username);
    toast.success(i18n.t("auth.profileUpdated"));
  }

  async changePassword(currentPassword: string, newPassword: string) {
    await api.changePassword(currentPassword, newPassword);
    toast.success(i18n.t("auth.passwordChanged"));
  }

  get needsLogin(): boolean {
    return this.authEnabled && !this.setupNeeded && !this.user;
  }

  get canAccessApp(): boolean {
    return !this.loading && !this.setupNeeded && !this.needsLogin;
  }
}

export const auth = new AuthStore();
