import { describe, expect, it } from "vitest";
import {
  inferBaseFromPathname,
  normalizeViteBase,
  resolveAppBase,
  stripBase,
  withBase,
} from "./router-base";

describe("router-base", () => {
  it("treats absolute Vite base as the mount path", () => {
    expect(normalizeViteBase("/")).toBe("");
    expect(normalizeViteBase("/demo/")).toBe("/demo");
    expect(normalizeViteBase("/reader/demo/")).toBe("/reader/demo");
    expect(normalizeViteBase("./")).toBeNull();
  });

  it("infers /demo as the library mount for relative demo builds", () => {
    expect(inferBaseFromPathname("/demo")).toBe("/demo");
    expect(inferBaseFromPathname("/demo/")).toBe("/demo");
    expect(inferBaseFromPathname("/reader/demo/")).toBe("/reader/demo");
  });

  it("strips known routes when inferring base", () => {
    expect(inferBaseFromPathname("/demo/book/3")).toBe("/demo");
    expect(inferBaseFromPathname("/reader/demo/settings/library")).toBe("/reader/demo");
    expect(inferBaseFromPathname("/book/3")).toBe("");
    expect(inferBaseFromPathname("/")).toBe("");
  });

  it("resolves production root installs without inference", () => {
    expect(resolveAppBase("/demo", "/")).toBe("");
    expect(resolveAppBase("/demo/", "./")).toBe("/demo");
  });

  it("strips and prefixes the mount base", () => {
    expect(stripBase("/demo", "/demo")).toBe("/");
    expect(stripBase("/demo/", "/demo")).toBe("/");
    expect(stripBase("/demo/book/1", "/demo")).toBe("/book/1");
    expect(withBase("/", "/demo")).toBe("/demo/");
    expect(withBase("/book/1", "/demo")).toBe("/demo/book/1");
    expect(withBase("/login?reason=required", "/demo")).toBe("/demo/login?reason=required");
    expect(withBase("/", "")).toBe("/");
  });
});
