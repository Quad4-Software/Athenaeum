import { describe, expect, it } from "vitest";
import { formatBytes, seriesLabel } from "./format";

describe("formatBytes", () => {
  it("formats bytes under a kilobyte", () => {
    expect(formatBytes(512)).toBe("512 B");
  });

  it("formats kilobytes and megabytes", () => {
    expect(formatBytes(2048)).toBe("2.0 KB");
    expect(formatBytes(5 * 1024 * 1024)).toBe("5.0 MB");
  });

  it("handles invalid input", () => {
    expect(formatBytes(-1)).toBe("-");
    expect(formatBytes(NaN)).toBe("-");
  });
});

describe("seriesLabel", () => {
  it("returns empty when no series", () => {
    expect(seriesLabel(undefined)).toBe("");
  });

  it("appends index when present", () => {
    expect(seriesLabel("Dune", 2)).toBe("Dune #2");
    expect(seriesLabel("Dune")).toBe("Dune");
  });
});
