import { api, ApiError, restoreSession, ensureCsrf, clearCsrfCache } from "$lib/api/client";
import { onUnauthorized, onForbidden, type AuthRedirectReason } from "$lib/api/session";
import { unauthorizedRedirect, isAuthPagePathname } from "$lib/auth-redirect";
import { router } from "$lib/router.svelte";
import { toast } from "$lib/stores/toast.svelte";
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
    if (isAuthPagePathname(window.location.pathname)) return;
    router.navigate("/error/forbidden", true);
  }

  handleUnauthorized(reason: AuthRedirectReason = "required") {
    this.user = null;
    if (typeof window === "undefined") return;
    const target = unauthorizedRedirect(window.location.pathname, reason);
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
        this.error = e instanceof Error ? e.message : "Failed to load session";
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
    toast.success("Admin account created");
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
      toast.success("Signed in");
      return result;
    } catch (e) {
      this.error = e instanceof ApiError ? e.message : "Login failed";
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
      toast.success("Signed in");
    } catch (e) {
      this.error = e instanceof ApiError ? e.message : "Invalid code";
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
    toast.success("Account created");
  }

  async logout() {
    try {
      await api.logout();
    } catch {
      // session may already be gone
    }
    clearCsrfCache();
    this.user = null;
    toast.info("Signed out");
    this.handleUnauthorized("logged_out");
  }

  async register(username: string, password: string) {
    try {
      const u = await api.register(username, password);
      this.authEnabled = true;
      toast.success("Account created");
      return u;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Registration failed");
      throw e;
    }
  }

  async updateProfile(username: string) {
    this.user = await api.updateProfile(username);
    toast.success("Profile updated");
  }

  async changePassword(currentPassword: string, newPassword: string) {
    await api.changePassword(currentPassword, newPassword);
    toast.success("Password changed");
  }

  get needsLogin(): boolean {
    return this.authEnabled && !this.setupNeeded && !this.user;
  }

  get canAccessApp(): boolean {
    return !this.loading && !this.setupNeeded && !this.needsLogin;
  }
}

export const auth = new AuthStore();
