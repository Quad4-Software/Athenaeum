import type { KeyChord } from "./types";

/** Normalize a key from a KeyboardEvent for storage and matching. */
export function normalizeKey(key: string): string {
  if (key === " ") return "Space";
  if (key.length === 1) return key.toLowerCase();
  return key;
}

export function chordFromEvent(event: KeyboardEvent): KeyChord {
  return {
    key: normalizeKey(event.key),
    mod: event.metaKey || event.ctrlKey,
    shift: event.shiftKey || undefined,
    alt: event.altKey || undefined,
  };
}

export function chordsEqual(a: KeyChord, b: KeyChord): boolean {
  return a.key === b.key && !!a.mod === !!b.mod && !!a.shift === !!b.shift && !!a.alt === !!b.alt;
}

export function eventMatchesChord(event: KeyboardEvent, chord: KeyChord): boolean {
  if (normalizeKey(event.key) !== chord.key) return false;
  if (!!chord.mod !== (event.metaKey || event.ctrlKey)) return false;
  if (!!chord.shift !== event.shiftKey) return false;
  if (!!chord.alt !== event.altKey) return false;
  return true;
}

/** True on Apple platforms, where shortcuts render as glyphs instead of words. */
export function isMacPlatform(): boolean {
  if (typeof navigator === "undefined") return false;
  return /Mac|iPhone|iPad/.test(navigator.platform);
}

/** Human-readable chord for UI (uses Ctrl on non-Mac, Cmd glyph on Mac). */
export function formatChord(chord: KeyChord, isMac = false): string {
  const parts: string[] = [];
  if (chord.mod) parts.push(isMac ? "⌘" : "Ctrl");
  if (chord.alt) parts.push(isMac ? "⌥" : "Alt");
  const skipShiftLabel = chord.key === "?" || chord.key === "/";
  if (chord.shift && !skipShiftLabel) parts.push(isMac ? "⇧" : "Shift");
  const keyLabel =
    chord.key === "Space"
      ? "Space"
      : chord.key === ","
        ? ","
        : chord.key.length === 1
          ? chord.key.toUpperCase()
          : chord.key;
  parts.push(keyLabel);
  return parts.join(isMac ? "" : "+");
}

export function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return true;
  if (target.isContentEditable) return true;
  return target.closest("[contenteditable='true']") != null;
}

export function serializeChord(chord: KeyChord): string {
  return JSON.stringify({
    key: chord.key,
    mod: !!chord.mod,
    shift: !!chord.shift,
    alt: !!chord.alt,
  });
}

export function parseChord(raw: unknown): KeyChord | null {
  if (!raw || typeof raw !== "object") return null;
  const o = raw as Record<string, unknown>;
  if (typeof o.key !== "string" || !o.key) return null;
  return {
    key: normalizeKey(o.key),
    mod: o.mod === true ? true : undefined,
    shift: o.shift === true ? true : undefined,
    alt: o.alt === true ? true : undefined,
  };
}
