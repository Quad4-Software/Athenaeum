export function isTypingTarget(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null;
  if (!el) return false;
  if (el.isContentEditable) return true;
  const tag = el.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return true;
  return !!el.closest('input, textarea, select, [contenteditable="true"]');
}

export interface PageKeyHandlers {
  prev: () => void;
  next: () => void;
  zoomIn?: () => void;
  zoomOut?: () => void;
  shortcuts?: () => void;
}

export function handlePageKeys(event: KeyboardEvent, handlers: PageKeyHandlers): boolean {
  if (event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey) return false;
  if (isTypingTarget(event.target)) return false;

  switch (event.key) {
    case "ArrowLeft":
    case "PageUp":
      event.preventDefault();
      handlers.prev();
      return true;
    case "ArrowRight":
    case "PageDown":
      event.preventDefault();
      handlers.next();
      return true;
    case "+":
    case "=":
      if (handlers.zoomIn) {
        event.preventDefault();
        handlers.zoomIn();
        return true;
      }
      break;
    case "-":
      if (handlers.zoomOut) {
        event.preventDefault();
        handlers.zoomOut();
        return true;
      }
      break;
    case "?":
      if (handlers.shortcuts) {
        event.preventDefault();
        handlers.shortcuts();
        return true;
      }
      break;
    default:
      break;
  }
  return false;
}
