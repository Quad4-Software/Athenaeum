import { describe, expect, it, vi } from "vitest";
import {
  applyEpubFontPct,
  applyEpubSurfaceBackground,
  applyEpubThemeOverrides,
  buildEpubPrefKeys,
  canSelectEpubFont,
  clampEpubFontPct,
  createEpubPrefsSaver,
  decideEpubNarration,
  epubFontIdFromSelect,
  epubFontUploadErrorKey,
  epubHighlightStyles,
  epubjsSpread,
  epubLoadErrorMessage,
  epubNarrationParagraphs,
  epubPageKeyHandlers,
  epubPercentFromCfi,
  epubSpreadWidthClass,
  EPUB_MAX_FONT,
  EPUB_MIN_FONT,
  errorMessage,
  gatherEpubNarrateContents,
  loadAndPaintEpubHighlights,
  loadEpubSpreadMode,
  loadInitialEpubDisplayPrefs,
  loadReaderTheme,
  loadStoredNumber,
  mergeEpubReaderPrefs,
  mergeRemoteEpubPrefs,
  nextEpubFontPct,
  normalizeEpubContents,
  paintEpubHighlight,
  paintEpubHighlightList,
  persistEpubDisplayPrefs,
  persistEpubSpreadMode,
  persistEpubThemePrefs,
  preloadEpubSpineSections,
  prevEpubFontPct,
  readEpubSelectionText,
  resolveEpubFontId,
  resolveLoadedCustomFont,
  scheduleEpubLocationsGenerate,
  takeFileInput,
} from "./epub-reader";

describe("epubjsSpread", () => {
  it("maps UI modes to epubjs spread values", () => {
    expect(epubjsSpread("single")).toBe("none");
    expect(epubjsSpread("always")).toBe("always");
    expect(epubjsSpread("auto")).toBe("auto");
  });
});

describe("epubSpreadWidthClass", () => {
  it("widens the surface for double-page spreads", () => {
    expect(epubSpreadWidthClass("single")).toBe("max-w-3xl");
    expect(epubSpreadWidthClass("auto")).toBe("max-w-5xl");
    expect(epubSpreadWidthClass("always")).toBe("max-w-6xl");
  });
});

describe("font size helpers", () => {
  it("clamps and steps font percentage", () => {
    expect(clampEpubFontPct(50)).toBe(EPUB_MIN_FONT);
    expect(clampEpubFontPct(200)).toBe(EPUB_MAX_FONT);
    expect(nextEpubFontPct(100)).toBe(110);
    expect(prevEpubFontPct(100)).toBe(90);
    expect(nextEpubFontPct(EPUB_MAX_FONT)).toBe(EPUB_MAX_FONT);
    expect(prevEpubFontPct(EPUB_MIN_FONT)).toBe(EPUB_MIN_FONT);
  });
});

describe("storage loaders", () => {
  it("loads spread, theme, and numeric prefs with fallbacks", () => {
    expect(loadEpubSpreadMode("always")).toBe("always");
    expect(loadEpubSpreadMode("nope")).toBe("auto");
    expect(loadReaderTheme("sepia")).toBe("sepia");
    expect(loadReaderTheme("neon")).toBe("light");
    expect(loadStoredNumber("120", 100)).toBe(120);
    expect(loadStoredNumber(null, 100)).toBe(100);
  });
});

describe("epubHighlightStyles", () => {
  it("returns known colors and falls back to yellow", () => {
    expect(epubHighlightStyles("green").fill).toBe("#86efac");
    expect(epubHighlightStyles("unknown").fill).toBe("#fde047");
  });
});

describe("epubPercentFromCfi", () => {
  it("returns percentage or zero on failure", () => {
    expect(epubPercentFromCfi(() => 0.42, "cfi")).toBe(0.42);
    expect(epubPercentFromCfi(() => 0, "cfi")).toBe(0);
    expect(
      epubPercentFromCfi(() => {
        throw new Error("missing");
      }, "cfi"),
    ).toBe(0);
  });
});

describe("mergeEpubReaderPrefs", () => {
  it("applies only valid preference fields", () => {
    const base = {
      fontPct: 100,
      theme: "light" as const,
      lineHeight: 1.6,
      marginPx: 24,
      spread: "auto" as const,
    };
    expect(
      mergeEpubReaderPrefs(base, {
        fontPct: 120,
        theme: "night",
        lineHeight: 1.8,
        marginPx: 32,
        spread: "single",
        junk: true,
      }),
    ).toEqual({
      fontPct: 120,
      theme: "night",
      lineHeight: 1.8,
      marginPx: 32,
      spread: "single",
    });
    expect(mergeEpubReaderPrefs(base, { theme: "neon", spread: "wide" })).toEqual(base);
  });
});

describe("normalizeEpubContents", () => {
  it("normalizes single, array, and missing contents", () => {
    const frame = { document: undefined };
    expect(normalizeEpubContents(frame)).toEqual([frame]);
    expect(normalizeEpubContents([frame, frame])).toEqual([frame, frame]);
    expect(normalizeEpubContents(null)).toEqual([]);
    expect(normalizeEpubContents(undefined)).toEqual([]);
  });
});

describe("epubFontUploadErrorKey", () => {
  it("detects size errors", () => {
    expect(epubFontUploadErrorKey("file too large")).toBe("reader.fontTooLarge");
    expect(epubFontUploadErrorKey("network")).toBe("reader.fontUploadFailed");
  });
});

describe("epubPageKeyHandlers", () => {
  it("maps reader actions onto page key handlers", () => {
    const prev = vi.fn();
    const next = vi.fn();
    const largerFont = vi.fn();
    const smallerFont = vi.fn();
    const openShortcuts = vi.fn();
    const handlers = epubPageKeyHandlers({
      prev,
      next,
      largerFont,
      smallerFont,
      openShortcuts,
    });
    handlers.prev();
    handlers.next();
    handlers.zoomIn?.();
    handlers.zoomOut?.();
    handlers.shortcuts?.();
    expect(prev).toHaveBeenCalledOnce();
    expect(next).toHaveBeenCalledOnce();
    expect(largerFont).toHaveBeenCalledOnce();
    expect(smallerFont).toHaveBeenCalledOnce();
    expect(openShortcuts).toHaveBeenCalledOnce();
  });
});

describe("decideEpubNarration", () => {
  it("toggles off when already active", () => {
    expect(
      decideEpubNarration({
        narratorActive: true,
        provider: "browser",
        kokoroEnabled: true,
        browserAvailable: true,
      }),
    ).toEqual({ kind: "toggle-off" });
  });

  it("errors when browser TTS is unavailable", () => {
    expect(
      decideEpubNarration({
        narratorActive: false,
        provider: "browser",
        kokoroEnabled: false,
        browserAvailable: false,
      }),
    ).toEqual({ kind: "error", key: "narrator.errUnavailable" });
  });

  it("falls back from kokoro to browser when needed", () => {
    expect(
      decideEpubNarration({
        narratorActive: false,
        provider: "kokoro",
        kokoroEnabled: false,
        browserAvailable: true,
      }),
    ).toEqual({ kind: "start", switchToBrowser: true });
    expect(
      decideEpubNarration({
        narratorActive: false,
        provider: "kokoro",
        kokoroEnabled: false,
        browserAvailable: false,
      }),
    ).toEqual({ kind: "error", key: "narrator.errKokoro" });
  });

  it("starts with the current provider when ready", () => {
    expect(
      decideEpubNarration({
        narratorActive: false,
        provider: "kokoro",
        kokoroEnabled: true,
        browserAvailable: false,
      }),
    ).toEqual({ kind: "start", switchToBrowser: false });
  });
});

describe("resolveEpubFontId", () => {
  it("keeps custom only when a font is stored", () => {
    expect(resolveEpubFontId({ preference: "custom", hasCustomFont: true })).toBe("custom");
    expect(resolveEpubFontId({ preference: "custom", hasCustomFont: false })).toBe("book");
    expect(resolveEpubFontId({ preference: "serif", hasCustomFont: false })).toBe("serif");
  });
});

describe("scheduleEpubLocationsGenerate", () => {
  it("uses idle callback when available", () => {
    const run = vi.fn();
    const idle = vi.fn((cb: () => void) => {
      cb();
      return 1;
    });
    vi.stubGlobal("requestIdleCallback", idle);
    scheduleEpubLocationsGenerate(run);
    expect(idle).toHaveBeenCalled();
    expect(run).toHaveBeenCalledOnce();
    vi.unstubAllGlobals();
  });
});

describe("epubLoadErrorMessage", () => {
  it("prefers trimmed messages", () => {
    expect(epubLoadErrorMessage("  boom  ", "fallback")).toBe("boom");
    expect(epubLoadErrorMessage("   ", "fallback")).toBe("fallback");
    expect(epubLoadErrorMessage(undefined, "fallback")).toBe("fallback");
  });
});

describe("canSelectEpubFont", () => {
  it("blocks custom when no uploaded font exists", () => {
    expect(canSelectEpubFont("custom", false)).toBe(false);
    expect(canSelectEpubFont("custom", true)).toBe(true);
    expect(canSelectEpubFont("book", false)).toBe(true);
  });
});

describe("persistEpubDisplayPrefs", () => {
  it("writes all display prefs to localStorage", () => {
    const store = new Map<string, string>();
    vi.stubGlobal("localStorage", {
      setItem(key: string, value: string) {
        store.set(key, value);
      },
      getItem(key: string) {
        return store.get(key) ?? null;
      },
    });
    persistEpubDisplayPrefs(
      { font: "f", theme: "t", line: "l", margin: "m", spread: "s" },
      {
        fontPct: 110,
        theme: "sepia",
        lineHeight: 1.7,
        marginPx: 28,
        spread: "always",
      },
    );
    expect(store.get("f")).toBe("110");
    expect(store.get("t")).toBe("sepia");
    expect(store.get("l")).toBe("1.7");
    expect(store.get("m")).toBe("28");
    expect(store.get("s")).toBe("always");
    vi.unstubAllGlobals();
  });
});

describe("loadInitialEpubDisplayPrefs", () => {
  it("returns defaults without storage", () => {
    expect(
      loadInitialEpubDisplayPrefs(null, {
        font: "f",
        theme: "t",
        line: "l",
        margin: "m",
        spread: "s",
      }),
    ).toEqual({
      fontPct: 100,
      theme: "light",
      lineHeight: 1.6,
      marginPx: 24,
      spread: "auto",
    });
  });

  it("reads stored values", () => {
    const store = new Map([
      ["f", "120"],
      ["t", "night"],
      ["l", "1.8"],
      ["m", "32"],
      ["s", "single"],
    ]);
    expect(
      loadInitialEpubDisplayPrefs(
        { getItem: (k) => store.get(k) ?? null },
        { font: "f", theme: "t", line: "l", margin: "m", spread: "s" },
      ),
    ).toEqual({
      fontPct: 120,
      theme: "night",
      lineHeight: 1.8,
      marginPx: 32,
      spread: "single",
    });
  });
});

describe("applyEpubThemeOverrides", () => {
  it("sets color background line-height and margins", () => {
    const override = vi.fn();
    applyEpubThemeOverrides(
      { override },
      {
        fg: "#111",
        bg: "#eee",
        lineHeight: 1.7,
        marginPx: 20,
      },
    );
    expect(override).toHaveBeenCalledWith("color", "#111");
    expect(override).toHaveBeenCalledWith("background", "#eee");
    expect(override).toHaveBeenCalledWith("line-height", "1.7");
    expect(override).toHaveBeenCalledWith("margin-left", "20px");
    expect(override).toHaveBeenCalledWith("margin-right", "20px");
  });
});

describe("paintEpubHighlight", () => {
  it("adds a highlight with styled class", () => {
    const add = vi.fn();
    paintEpubHighlight({ add }, "epubcfi(/6/4)", 7, "blue");
    expect(add).toHaveBeenCalledWith(
      "highlight",
      "epubcfi(/6/4)",
      {},
      expect.any(Function),
      "hl-7",
      expect.objectContaining({ fill: "#93c5fd" }),
    );
  });
});

describe("preloadEpubSpineSections", () => {
  it("loads adjacent spine sections", () => {
    const load = vi.fn(() => Promise.resolve());
    const get = vi.fn((index: number) => (index === 2 || index === 4 ? { load } : undefined));
    preloadEpubSpineSections({ spine: { length: 10, get } }, 3);
    expect(get).toHaveBeenCalledWith(2);
    expect(get).toHaveBeenCalledWith(4);
    expect(load).toHaveBeenCalledTimes(2);
  });
});

describe("applyEpubSurfaceBackground", () => {
  it("sets container background color", () => {
    const container = { style: { backgroundColor: "" } } as HTMLElement;
    applyEpubSurfaceBackground(container, () => undefined, "#abc");
    expect(container.style.backgroundColor).toBe("#abc");
  });
});

describe("persistEpubThemePrefs", () => {
  it("writes theme line and margin only", () => {
    const store = new Map<string, string>();
    vi.stubGlobal("localStorage", {
      setItem(key: string, value: string) {
        store.set(key, value);
      },
    });
    persistEpubThemePrefs(
      { theme: "t", line: "l", margin: "m" },
      { theme: "sepia", lineHeight: 1.5, marginPx: 18 },
    );
    expect(store.get("t")).toBe("sepia");
    expect(store.get("l")).toBe("1.5");
    expect(store.get("m")).toBe("18");
    vi.unstubAllGlobals();
  });
});

describe("gatherEpubNarrateContents", () => {
  it("normalizes contents and swallows errors", () => {
    const frame = { document: undefined };
    expect(gatherEpubNarrateContents(() => frame)).toEqual([frame]);
    expect(
      gatherEpubNarrateContents(() => {
        throw new Error("gone");
      }),
    ).toEqual([]);
  });
});

describe("readEpubSelectionText", () => {
  it("returns trimmed selection or empty on failure", () => {
    expect(readEpubSelectionText(() => ({ toString: () => "  hi  " }) as Selection)).toBe("hi");
    expect(readEpubSelectionText(() => null)).toBe("");
    expect(
      readEpubSelectionText(() => {
        throw new Error("no window");
      }),
    ).toBe("");
  });
});

describe("resolveLoadedCustomFont", () => {
  it("flags preference reset when custom is missing", () => {
    expect(resolveLoadedCustomFont({ preference: "custom", hasCustomFont: true })).toEqual({
      fontId: "custom",
      shouldResetPreference: false,
    });
    expect(resolveLoadedCustomFont({ preference: "custom", hasCustomFont: false })).toEqual({
      fontId: "book",
      shouldResetPreference: true,
    });
  });
});

describe("createEpubPrefsSaver", () => {
  it("debounces save when ready", async () => {
    vi.useFakeTimers();
    const save = vi.fn(() => Promise.resolve());
    const saver = createEpubPrefsSaver({
      delayMs: 100,
      isReady: () => true,
      getPrefs: () => ({
        fontPct: 100,
        theme: "light",
        lineHeight: 1.6,
        marginPx: 24,
        spread: "auto",
      }),
      save,
    });
    saver.queue();
    saver.queue();
    expect(save).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(100);
    expect(save).toHaveBeenCalledOnce();
    vi.useRealTimers();
  });

  it("skips queue when not ready", async () => {
    vi.useFakeTimers();
    const save = vi.fn(() => Promise.resolve());
    const saver = createEpubPrefsSaver({
      delayMs: 50,
      isReady: () => false,
      getPrefs: () => ({
        fontPct: 100,
        theme: "light",
        lineHeight: 1.6,
        marginPx: 24,
        spread: "auto",
      }),
      save,
    });
    saver.queue();
    await vi.advanceTimersByTimeAsync(50);
    expect(save).not.toHaveBeenCalled();
    vi.useRealTimers();
  });
});

describe("buildEpubPrefKeys", () => {
  it("maps storage key names", () => {
    expect(buildEpubPrefKeys((n) => `x:${n}`)).toEqual({
      font: "x:epub-font",
      theme: "x:epub-theme",
      line: "x:epub-line",
      margin: "x:epub-margin",
      spread: "x:epub-spread",
    });
  });
});

describe("applyEpubFontPct", () => {
  it("sets font size and persists", () => {
    const fontSize = vi.fn();
    const store = new Map<string, string>();
    vi.stubGlobal("localStorage", {
      setItem(key: string, value: string) {
        store.set(key, value);
      },
    });
    applyEpubFontPct({ fontSize }, 110, "font-key");
    expect(fontSize).toHaveBeenCalledWith("110%");
    expect(store.get("font-key")).toBe("110");
    vi.unstubAllGlobals();
  });
});

describe("persistEpubSpreadMode", () => {
  it("writes spread mode", () => {
    const store = new Map<string, string>();
    vi.stubGlobal("localStorage", {
      setItem(key: string, value: string) {
        store.set(key, value);
      },
    });
    persistEpubSpreadMode("spread", "always");
    expect(store.get("spread")).toBe("always");
    vi.unstubAllGlobals();
  });
});

describe("paintEpubHighlightList", () => {
  it("paints each highlight", () => {
    const add = vi.fn();
    paintEpubHighlightList({ add }, [
      { location: "a", id: 1, color: "green" },
      { location: "b", id: 2 },
    ]);
    expect(add).toHaveBeenCalledTimes(2);
  });
});

describe("loadAndPaintEpubHighlights", () => {
  it("loads then paints", async () => {
    const add = vi.fn();
    await loadAndPaintEpubHighlights({ add }, async () => [
      { location: "c", id: 3, color: "blue" },
    ]);
    expect(add).toHaveBeenCalledOnce();
  });

  it("no-ops without annotations and swallows load errors", async () => {
    await expect(loadAndPaintEpubHighlights(undefined, async () => [])).resolves.toBeUndefined();
    await expect(
      loadAndPaintEpubHighlights({ add: vi.fn() }, async () => {
        throw new Error("fail");
      }),
    ).resolves.toBeUndefined();
  });
});

describe("takeFileInput", () => {
  it("returns the first file and clears the input", () => {
    const file = new File(["x"], "f.ttf");
    const input = {
      files: [file],
      value: "f.ttf",
    };
    const event = { currentTarget: input } as unknown as Event;
    expect(takeFileInput(event)).toBe(file);
    expect(input.value).toBe("");
  });
});

describe("errorMessage", () => {
  it("reads Error messages", () => {
    expect(errorMessage(new Error("boom"))).toBe("boom");
    expect(errorMessage("nope")).toBe("");
  });
});

describe("epubFontIdFromSelect", () => {
  it("returns null when custom is blocked", () => {
    const event = {
      currentTarget: { value: "custom" },
    } as unknown as Event;
    expect(epubFontIdFromSelect(event, false)).toBeNull();
    expect(epubFontIdFromSelect(event, true)).toBe("custom");
  });
});

describe("epubNarrationParagraphs", () => {
  it("gathers contents then maps paragraphs", () => {
    const frame = { document: undefined };
    const paragraphsFrom = vi.fn(() => ["one", "two"]);
    expect(epubNarrationParagraphs(() => frame, paragraphsFrom)).toEqual(["one", "two"]);
    expect(paragraphsFrom).toHaveBeenCalledWith([frame]);
  });
});

describe("mergeRemoteEpubPrefs", () => {
  it("merges remote prefs object", () => {
    expect(
      mergeRemoteEpubPrefs(
        {
          fontPct: 100,
          theme: "light",
          lineHeight: 1.6,
          marginPx: 24,
          spread: "auto",
        },
        { fontPct: 130 },
      ).fontPct,
    ).toBe(130);
  });
});
