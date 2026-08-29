import { describe, expect, it, vi } from "vitest";
import { handlePageKeys, isTypingTarget } from "./reader-keys";

describe("isTypingTarget", () => {
  it("detects form fields and contenteditable elements", () => {
    const input = document.createElement("input");
    expect(isTypingTarget(input)).toBe(true);

    const editable = document.createElement("div");
    editable.setAttribute("contenteditable", "true");
    document.body.appendChild(editable);
    expect(isTypingTarget(editable)).toBe(true);
    editable.remove();

    const button = document.createElement("button");
    expect(isTypingTarget(button)).toBe(false);
  });
});

describe("handlePageKeys", () => {
  it("calls prev and next handlers for arrow keys", () => {
    const prev = vi.fn();
    const next = vi.fn();

    handlePageKeys(
      new KeyboardEvent("keydown", { key: "ArrowLeft", bubbles: true, cancelable: true }),
      { prev, next },
    );
    handlePageKeys(
      new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true, cancelable: true }),
      { prev, next },
    );

    expect(prev).toHaveBeenCalledOnce();
    expect(next).toHaveBeenCalledOnce();
  });

  it("ignores shortcuts while typing", () => {
    const prev = vi.fn();
    const input = document.createElement("input");
    document.body.appendChild(input);

    const event = {
      key: "ArrowLeft",
      defaultPrevented: false,
      altKey: false,
      ctrlKey: false,
      metaKey: false,
      preventDefault: vi.fn(),
      target: input,
    } as unknown as KeyboardEvent;

    const blocked = handlePageKeys(event, { prev, next: vi.fn() });

    expect(blocked).toBe(false);
    expect(prev).not.toHaveBeenCalled();
    input.remove();
  });

  it("supports zoom and shortcut handlers", () => {
    const zoomIn = vi.fn();
    const zoomOut = vi.fn();
    const shortcuts = vi.fn();

    handlePageKeys(new KeyboardEvent("keydown", { key: "=", bubbles: true, cancelable: true }), {
      prev: vi.fn(),
      next: vi.fn(),
      zoomIn,
      zoomOut,
      shortcuts,
    });
    handlePageKeys(new KeyboardEvent("keydown", { key: "-", bubbles: true, cancelable: true }), {
      prev: vi.fn(),
      next: vi.fn(),
      zoomIn,
      zoomOut,
      shortcuts,
    });
    handlePageKeys(new KeyboardEvent("keydown", { key: "?", bubbles: true, cancelable: true }), {
      prev: vi.fn(),
      next: vi.fn(),
      zoomIn,
      zoomOut,
      shortcuts,
    });

    expect(zoomIn).toHaveBeenCalledOnce();
    expect(zoomOut).toHaveBeenCalledOnce();
    expect(shortcuts).toHaveBeenCalledOnce();
  });
});
