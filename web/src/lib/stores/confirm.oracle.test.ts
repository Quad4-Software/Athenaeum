import { describe, expect, it } from "vitest";
import { confirmDialog } from "./confirm.svelte";

/**
 * PROVED_CONFIRM_SUPERSEDE
 * Guarantee: a second ask() cancels the first pending promise as false
 * and only the latest dialog can accept.
 */
describe("confirmDialog adversarial", () => {
  it("supersedes an in-flight ask without accepting it", async () => {
    const first = confirmDialog.ask({ title: "One", message: "first" });
    const second = confirmDialog.ask({ title: "Two", message: "second" });

    expect(confirmDialog.title).toBe("Two");
    expect(confirmDialog.open).toBe(true);

    confirmDialog.accept();

    await expect(first).resolves.toBe(false);
    await expect(second).resolves.toBe(true);
    expect(confirmDialog.open).toBe(false);
    console.log("PROVED_CONFIRM_SUPERSEDE");
  });

  it("double accept does not resurrect a finished dialog", async () => {
    const p = confirmDialog.ask({ title: "X", message: "y" });
    confirmDialog.accept();
    confirmDialog.accept();
    await expect(p).resolves.toBe(true);
    expect(confirmDialog.open).toBe(false);
  });
});
