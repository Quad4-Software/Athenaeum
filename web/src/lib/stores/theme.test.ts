import { describe, expect, it } from "vitest";

function systemMode(matchesLight: boolean): "light" | "dark" {
  return matchesLight ? "light" : "dark";
}

function resolve(pref: "light" | "dark" | "system", matchesLight: boolean): "light" | "dark" {
  if (pref === "system") return systemMode(matchesLight);
  return pref;
}

function initial(saved: string | null): "light" | "dark" | "system" {
  if (saved === "light" || saved === "dark" || saved === "system") return saved;
  return "system";
}

describe("theme preference", () => {
  it("defaults to system when nothing saved", () => {
    expect(initial(null)).toBe("system");
  });

  it("restores saved preference", () => {
    expect(initial("dark")).toBe("dark");
    expect(initial("light")).toBe("light");
  });

  it("resolves system from media query", () => {
    expect(resolve("system", true)).toBe("light");
    expect(resolve("system", false)).toBe("dark");
  });

  it("keeps explicit light and dark", () => {
    expect(resolve("light", false)).toBe("light");
    expect(resolve("dark", true)).toBe("dark");
  });
});
