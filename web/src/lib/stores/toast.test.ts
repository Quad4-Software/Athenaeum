import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { toast } from "./toast.svelte";

describe("toast store", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    for (const item of [...toast.items]) toast.dismiss(item.id);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("auto-dismisses info toasts", () => {
    const id = toast.info("hello");
    expect(toast.items.some((t) => t.id === id)).toBe(true);
    vi.advanceTimersByTime(5000);
    expect(toast.items.some((t) => t.id === id)).toBe(false);
  });

  it("keeps loading toasts sticky until done", () => {
    const id = toast.loading("working", 0.25);
    expect(toast.items.find((t) => t.id === id)?.sticky).toBe(true);
    expect(toast.items.find((t) => t.id === id)?.progress).toBe(0.25);
    vi.advanceTimersByTime(30_000);
    expect(toast.items.some((t) => t.id === id)).toBe(true);

    toast.update(id, { progress: 0.8, message: "almost" });
    expect(toast.items.find((t) => t.id === id)?.progress).toBe(0.8);

    toast.done(id, "finished", "success");
    expect(toast.items.find((t) => t.id === id)?.kind).toBe("success");
    vi.advanceTimersByTime(5000);
    expect(toast.items.some((t) => t.id === id)).toBe(false);
  });

  it("promise helper resolves with success toast", async () => {
    const p = toast.promise(
      async ({ update }) => {
        update({ progress: 0.5 });
        return 42;
      },
      {
        loading: "go",
        success: (n) => `got ${n}`,
        error: "nope",
      },
    );
    expect(toast.items.some((t) => t.kind === "loading")).toBe(true);
    await expect(p).resolves.toBe(42);
    expect(toast.items.some((t) => t.kind === "success" && t.message === "got 42")).toBe(true);
  });

  it("promise helper surfaces errors", async () => {
    const p = toast.promise(
      async () => {
        throw new Error("boom");
      },
      { loading: "go", success: "ok", error: (e) => (e as Error).message },
    );
    await expect(p).rejects.toThrow("boom");
    expect(toast.items.some((t) => t.kind === "error" && t.message === "boom")).toBe(true);
  });
});
