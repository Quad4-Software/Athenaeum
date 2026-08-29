export interface SwipeNavigationOptions {
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  threshold?: number;
  maxVerticalDrift?: number;
}

export interface TapZoneOptions {
  onTapLeft?: () => void;
  onTapRight?: () => void;
  edgeFraction?: number;
}

export function swipeNavigation(
  node: HTMLElement,
  options: SwipeNavigationOptions,
): { update: (next: SwipeNavigationOptions) => void; destroy: () => void } {
  let opts = options;
  let startX = 0;
  let startY = 0;
  let tracking = false;

  function threshold() {
    return opts.threshold ?? 48;
  }

  function maxVertical() {
    return opts.maxVerticalDrift ?? 72;
  }

  function onTouchStart(event: TouchEvent) {
    if (event.touches.length !== 1) return;
    const touch = event.touches[0];
    startX = touch.clientX;
    startY = touch.clientY;
    tracking = true;
  }

  function onTouchEnd(event: TouchEvent) {
    if (!tracking) return;
    tracking = false;
    const touch = event.changedTouches[0];
    if (!touch) return;
    const dx = touch.clientX - startX;
    const dy = touch.clientY - startY;
    if (Math.abs(dx) < threshold()) return;
    if (Math.abs(dy) > maxVertical()) return;
    if (Math.abs(dx) <= Math.abs(dy)) return;
    if (dx < 0) opts.onSwipeLeft?.();
    else opts.onSwipeRight?.();
  }

  function onTouchCancel() {
    tracking = false;
  }

  node.addEventListener("touchstart", onTouchStart, { passive: true });
  node.addEventListener("touchend", onTouchEnd, { passive: true });
  node.addEventListener("touchcancel", onTouchCancel, { passive: true });

  return {
    update(next: SwipeNavigationOptions) {
      opts = next;
    },
    destroy() {
      node.removeEventListener("touchstart", onTouchStart);
      node.removeEventListener("touchend", onTouchEnd);
      node.removeEventListener("touchcancel", onTouchCancel);
    },
  };
}

export function tapZones(
  node: HTMLElement,
  options: TapZoneOptions,
): { update: (next: TapZoneOptions) => void; destroy: () => void } {
  let opts = options;

  function edgeFraction() {
    return opts.edgeFraction ?? 0.28;
  }

  function onPointerUp(event: PointerEvent) {
    if (event.pointerType === "mouse" && event.button !== 0) return;
    const rect = node.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const fraction = x / rect.width;
    if (fraction < edgeFraction()) opts.onTapLeft?.();
    else if (fraction > 1 - edgeFraction()) opts.onTapRight?.();
  }

  node.addEventListener("pointerup", onPointerUp);

  return {
    update(next: TapZoneOptions) {
      opts = next;
    },
    destroy() {
      node.removeEventListener("pointerup", onPointerUp);
    },
  };
}

export function readerGestures(
  node: HTMLElement,
  options: SwipeNavigationOptions & TapZoneOptions,
): { update: (next: SwipeNavigationOptions & TapZoneOptions) => void; destroy: () => void } {
  const swipe = swipeNavigation(node, options);
  const tap = tapZones(node, options);
  return {
    update(next: SwipeNavigationOptions & TapZoneOptions) {
      swipe.update(next);
      tap.update(next);
    },
    destroy() {
      swipe.destroy();
      tap.destroy();
    },
  };
}

type GestureTarget = Document | HTMLElement;

const boundGestureDocs = new WeakSet<Document>();

function bindGestureTarget(
  target: GestureTarget,
  options: SwipeNavigationOptions & TapZoneOptions,
): () => void {
  const swipe = swipeNavigation(target as HTMLElement, options);
  const tap = tapZones(target as HTMLElement, options);
  return () => {
    swipe.destroy();
    tap.destroy();
  };
}

export function bindDocumentGestures(
  doc: Document,
  options: SwipeNavigationOptions & TapZoneOptions,
): () => void {
  if (boundGestureDocs.has(doc)) return () => undefined;
  boundGestureDocs.add(doc);
  return bindGestureTarget(doc.documentElement, options);
}
