import { describe, expect, it } from "vitest";
import {
  chordFromEvent,
  chordsEqual,
  eventMatchesChord,
  formatChord,
  isMacPlatform,
  isTypingTarget,
  normalizeKey,
  parseChord,
  serializeChord,
} from "./chords";

function keyEvent(init: KeyboardEventInit): KeyboardEvent {
  return new KeyboardEvent("keydown", init);
}

describe("normalizeKey", () => {
  it("maps the space key to a stable name", () => {
    expect(normalizeKey(" ")).toBe("Space");
  });

  it("lowercases single character keys", () => {
    expect(normalizeKey("K")).toBe("k");
    expect(normalizeKey("?")).toBe("?");
  });

  it("keeps named keys verbatim", () => {
    expect(normalizeKey("ArrowDown")).toBe("ArrowDown");
    expect(normalizeKey("Escape")).toBe("Escape");
  });
});

describe("chordFromEvent", () => {
  it("treats meta and control as the same modifier", () => {
    expect(chordFromEvent(keyEvent({ key: "k", metaKey: true }))).toEqual({
      key: "k",
      mod: true,
      shift: undefined,
      alt: undefined,
    });
    expect(chordFromEvent(keyEvent({ key: "k", ctrlKey: true })).mod).toBe(true);
  });

  it("omits modifiers that are not held", () => {
    const chord = chordFromEvent(keyEvent({ key: "b" }));
    expect(chord.mod).toBe(false);
    expect(chord.shift).toBeUndefined();
    expect(chord.alt).toBeUndefined();
  });

  it("captures shift and alt", () => {
    const chord = chordFromEvent(keyEvent({ key: "L", shiftKey: true, altKey: true }));
    expect(chord).toMatchObject({ key: "l", shift: true, alt: true });
  });
});

describe("chordsEqual", () => {
  it("treats missing modifiers as false", () => {
    expect(chordsEqual({ key: "k", mod: true }, { key: "k", mod: true, shift: undefined })).toBe(
      true,
    );
  });

  it("separates chords that differ by a modifier", () => {
    expect(chordsEqual({ key: "l", mod: true }, { key: "l", mod: true, shift: true })).toBe(false);
  });

  it("separates chords that differ by key", () => {
    expect(chordsEqual({ key: "1", mod: true }, { key: "2", mod: true })).toBe(false);
  });
});

describe("eventMatchesChord", () => {
  it("matches a mod chord from either meta or control", () => {
    const chord = { key: "k", mod: true };
    expect(eventMatchesChord(keyEvent({ key: "k", metaKey: true }), chord)).toBe(true);
    expect(eventMatchesChord(keyEvent({ key: "k", ctrlKey: true }), chord)).toBe(true);
  });

  it("rejects the bare key when the chord requires a modifier", () => {
    expect(eventMatchesChord(keyEvent({ key: "k" }), { key: "k", mod: true })).toBe(false);
  });

  it("rejects extra modifiers the chord does not declare", () => {
    const chord = { key: "b", mod: true };
    expect(eventMatchesChord(keyEvent({ key: "b", ctrlKey: true, shiftKey: true }), chord)).toBe(
      false,
    );
  });

  it("matches a shifted question mark", () => {
    const chord = { key: "?", shift: true };
    expect(eventMatchesChord(keyEvent({ key: "?", shiftKey: true }), chord)).toBe(true);
  });
});

describe("formatChord", () => {
  it("uses Ctrl and plus separators off Mac", () => {
    expect(formatChord({ key: "k", mod: true })).toBe("Ctrl+K");
    expect(formatChord({ key: "l", mod: true, shift: true })).toBe("Ctrl+Shift+L");
  });

  it("uses glyphs without separators on Mac", () => {
    expect(formatChord({ key: "k", mod: true }, true)).toBe("⌘K");
    expect(formatChord({ key: "l", mod: true, shift: true }, true)).toBe("⌘⇧L");
  });

  it("skips the shift label for question mark and slash", () => {
    expect(formatChord({ key: "?", shift: true })).toBe("?");
    expect(formatChord({ key: "/", shift: true })).toBe("/");
  });

  it("keeps comma and space readable", () => {
    expect(formatChord({ key: ",", mod: true })).toBe("Ctrl+,");
    expect(formatChord({ key: "Space" })).toBe("Space");
  });

  it("labels alt per platform", () => {
    expect(formatChord({ key: "b", alt: true })).toBe("Alt+B");
    expect(formatChord({ key: "b", alt: true }, true)).toBe("⌥B");
  });
});

describe("isMacPlatform", () => {
  function withPlatform(value: string, fn: () => void) {
    const original = Object.getOwnPropertyDescriptor(Navigator.prototype, "platform");
    Object.defineProperty(navigator, "platform", { value, configurable: true });
    try {
      fn();
    } finally {
      if (original) Object.defineProperty(Navigator.prototype, "platform", original);
      delete (navigator as unknown as Record<string, unknown>).platform;
    }
  }

  it("detects Apple platforms", () => {
    withPlatform("MacIntel", () => expect(isMacPlatform()).toBe(true));
    withPlatform("iPhone", () => expect(isMacPlatform()).toBe(true));
  });

  it("returns false elsewhere", () => {
    withPlatform("Win32", () => expect(isMacPlatform()).toBe(false));
    withPlatform("Linux x86_64", () => expect(isMacPlatform()).toBe(false));
  });
});

describe("isTypingTarget", () => {
  it("detects form fields", () => {
    for (const tag of ["input", "textarea", "select"]) {
      expect(isTypingTarget(document.createElement(tag))).toBe(true);
    }
  });

  it("detects contenteditable hosts and their descendants", () => {
    const host = document.createElement("div");
    host.setAttribute("contenteditable", "true");
    const child = document.createElement("span");
    host.appendChild(child);
    document.body.appendChild(host);

    expect(isTypingTarget(host)).toBe(true);
    expect(isTypingTarget(child)).toBe(true);

    host.remove();
  });

  it("ignores plain elements and non elements", () => {
    expect(isTypingTarget(document.createElement("button"))).toBe(false);
    expect(isTypingTarget(null)).toBe(false);
  });
});

describe("serializeChord and parseChord", () => {
  it("round trips a chord", () => {
    const chord = { key: "l", mod: true, shift: true };
    const parsed = parseChord(JSON.parse(serializeChord(chord)));
    expect(parsed).not.toBeNull();
    expect(chordsEqual(parsed!, chord)).toBe(true);
  });

  it("normalizes false modifiers back to undefined", () => {
    expect(parseChord({ key: "k", mod: true, shift: false, alt: false })).toEqual({
      key: "k",
      mod: true,
      shift: undefined,
      alt: undefined,
    });
  });

  it("normalizes the stored key", () => {
    expect(parseChord({ key: "K" })?.key).toBe("k");
    expect(parseChord({ key: " " })?.key).toBe("Space");
  });

  it("rejects malformed input", () => {
    expect(parseChord(null)).toBeNull();
    expect(parseChord("Ctrl+K")).toBeNull();
    expect(parseChord({})).toBeNull();
    expect(parseChord({ key: "" })).toBeNull();
    expect(parseChord({ key: 3 })).toBeNull();
  });
});
