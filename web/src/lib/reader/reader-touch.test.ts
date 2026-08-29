import { describe, expect, it, vi } from "vitest";
import { readerGestures } from "./reader-touch";

function touch(node: HTMLElement, type: string, clientX: number, clientY = 0) {
  const touchObj = { clientX, clientY, identifier: 0, target: node } as unknown as Touch;
  const event = new Event(type, { bubbles: true, cancelable: true }) as TouchEvent;
  Object.defineProperty(event, "touches", {
    value: type === "touchend" ? [] : [touchObj],
    configurable: true,
  });
  Object.defineProperty(event, "changedTouches", {
    value: [touchObj],
    configurable: true,
  });
  node.dispatchEvent(event);
}

describe("readerGestures", () => {
  it("fires swipe left for next page", () => {
    const node = document.createElement("div");
    const onSwipeLeft = vi.fn();
    const action = readerGestures(node, { onSwipeLeft });
    touch(node, "touchstart", 200, 100);
    touch(node, "touchend", 120, 100);
    expect(onSwipeLeft).toHaveBeenCalledOnce();
    action.destroy();
  });

  it("fires swipe right for previous page", () => {
    const node = document.createElement("div");
    const onSwipeRight = vi.fn();
    const action = readerGestures(node, { onSwipeRight });
    touch(node, "touchstart", 120, 100);
    touch(node, "touchend", 200, 100);
    expect(onSwipeRight).toHaveBeenCalledOnce();
    action.destroy();
  });

  it("ignores short swipes", () => {
    const node = document.createElement("div");
    const onSwipeLeft = vi.fn();
    const action = readerGestures(node, { onSwipeLeft });
    touch(node, "touchstart", 200, 100);
    touch(node, "touchend", 180, 100);
    expect(onSwipeLeft).not.toHaveBeenCalled();
    action.destroy();
  });
});
