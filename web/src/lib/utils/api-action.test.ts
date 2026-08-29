import { describe, expect, it, vi, beforeEach } from "vitest";
import { ApiError } from "$lib/api/core";
import { apiAction, apiErrorMessage } from "$lib/utils/api-action";

vi.mock("$lib/stores/toast.svelte", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

import { toast } from "$lib/stores/toast.svelte";

beforeEach(() => {
  vi.clearAllMocks();
});

describe("apiAction", () => {
  it("returns result and toasts success", async () => {
    const result = await apiAction(async () => 42, {
      success: "ok",
      errorFallback: "fail",
    });
    expect(result).toBe(42);
    expect(toast.success).toHaveBeenCalledWith("ok");
  });

  it("toasts ApiError message on failure", async () => {
    const result = await apiAction(
      async () => {
        throw new ApiError(400, "bad request");
      },
      { errorFallback: "fail" },
    );
    expect(result).toBeUndefined();
    expect(toast.error).toHaveBeenCalledWith("bad request");
  });

  it("silences configured statuses", async () => {
    const result = await apiAction(
      async () => {
        throw new ApiError(401, "nope");
      },
      { errorFallback: "fail", silentStatuses: [401] },
    );
    expect(result).toBeUndefined();
    expect(toast.error).not.toHaveBeenCalled();
  });
});

describe("apiErrorMessage", () => {
  it("prefers Error.message", () => {
    expect(apiErrorMessage(new Error("x"), "fallback")).toBe("x");
    expect(apiErrorMessage("nope", "fallback")).toBe("fallback");
  });
});
