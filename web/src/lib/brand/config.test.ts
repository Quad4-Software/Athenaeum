import { describe, expect, it } from "vitest";
import { brand } from "./config";

/**
 * Wire-critical values must stay aligned with Go
 * internal/auth (CSRFCookie, SessionCookie, RefreshCookie, APIKeyPrefix)
 * and internal/brand (APIKeyPrefix, ConfigExportName, OIDCStateCookie).
 */
describe("brand wire constants", () => {
  it("matches server CSRF cookie name", () => {
    expect(brand.csrfCookie).toBe("athenaeum_csrf");
  });

  it("matches server API key prefix", () => {
    expect(brand.apiKeyPrefix).toBe("ath_");
  });

  it("matches server config export filename", () => {
    expect(brand.configExportName).toBe("athenaeum-config.json");
  });

  it("uses storage prefix consistent with app identity", () => {
    expect(brand.storagePrefix).toBe("athenaeum");
  });
});
