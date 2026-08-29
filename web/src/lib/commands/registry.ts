import { router } from "$lib/router.svelte";
import { library } from "$lib/stores/library.svelte";
import { ui } from "$lib/stores/ui.svelte";
import { theme } from "$lib/stores/theme.svelte";
import { density } from "$lib/stores/density.svelte";
import { commandPalette } from "$lib/stores/commandPalette.svelte";
import { can } from "$lib/permissions";
import type { BookFormat } from "$lib/api/types";
import type { CommandDef, CommandId } from "./types";

function goLibraryHome(clear = true) {
  if (clear) library.clearFilters();
  router.navigate("/");
  ui.closeMobileNav();
}

function goFormat(format: BookFormat | "") {
  library.setFormat(format);
  router.navigate("/");
  ui.closeMobileNav();
}

const COMMANDS: CommandDef[] = [
  {
    id: "palette.open",
    titleKey: "commands.paletteOpen",
    section: "actions",
    scope: "shell",
    defaultBinding: { key: "k", mod: true },
    run: () => commandPalette.toggle(),
  },
  {
    id: "nav.library",
    titleKey: "commands.navLibrary",
    section: "navigate",
    scope: "shell",
    defaultBinding: { key: "1", mod: true },
    run: () => goLibraryHome(true),
  },
  {
    id: "nav.continue",
    titleKey: "commands.navContinue",
    section: "navigate",
    scope: "shell",
    defaultBinding: { key: "2", mod: true },
    run: () => {
      library.setInProgress(true);
      router.navigate("/");
      ui.closeMobileNav();
    },
  },
  {
    id: "nav.favorites",
    titleKey: "commands.navFavorites",
    section: "navigate",
    scope: "shell",
    run: () => {
      library.setFavorites(true);
      router.navigate("/");
      ui.closeMobileNav();
    },
  },
  {
    id: "nav.collections",
    titleKey: "commands.navCollections",
    section: "navigate",
    scope: "shell",
    defaultBinding: { key: "3", mod: true },
    run: () => {
      router.navigate("/collections");
      ui.closeMobileNav();
    },
  },
  {
    id: "nav.settings",
    titleKey: "commands.navSettings",
    section: "settings",
    scope: "shell",
    defaultBinding: { key: ",", mod: true },
    run: () => {
      router.navigate("/settings/library");
      ui.closeMobileNav();
    },
  },
  {
    id: "nav.format.epub",
    titleKey: "commands.formatEpub",
    section: "navigate",
    scope: "shell",
    run: () => goFormat("epub"),
  },
  {
    id: "nav.format.pdf",
    titleKey: "commands.formatPdf",
    section: "navigate",
    scope: "shell",
    run: () => goFormat("pdf"),
  },
  {
    id: "nav.format.comic",
    titleKey: "commands.formatComic",
    section: "navigate",
    scope: "shell",
    run: () => goFormat("comic"),
  },
  {
    id: "nav.format.kindle",
    titleKey: "commands.formatKindle",
    section: "navigate",
    scope: "shell",
    run: () => goFormat("kindle"),
  },
  {
    id: "nav.format.audio",
    titleKey: "commands.formatAudio",
    section: "navigate",
    scope: "shell",
    run: () => goFormat("audio"),
  },
  {
    id: "ui.toggleSidebar",
    titleKey: "commands.toggleSidebar",
    section: "actions",
    scope: "shell",
    defaultBinding: { key: "b", mod: true },
    run: () => ui.toggleSidebar(),
  },
  {
    id: "ui.toggleTheme",
    titleKey: "commands.toggleTheme",
    section: "actions",
    scope: "shell",
    defaultBinding: { key: "l", mod: true, shift: true },
    run: () => theme.toggle(),
  },
  {
    id: "ui.toggleDensity",
    titleKey: "commands.toggleDensity",
    section: "actions",
    scope: "shell",
    run: () => density.toggle(),
  },
  {
    id: "library.scan",
    titleKey: "commands.scanLibrary",
    section: "actions",
    scope: "shell",
    when: () => can("manage_library"),
    run: () => void library.triggerScan(),
  },
  {
    id: "help.shortcuts",
    titleKey: "commands.helpShortcuts",
    section: "settings",
    scope: "shell",
    defaultBinding: { key: "?", shift: true },
    run: () => {
      router.navigate("/settings/library");
      ui.closeMobileNav();
      queueMicrotask(() => {
        document.getElementById("settings-keyboard-shortcuts")?.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      });
    },
  },
];

const BY_ID = new Map(COMMANDS.map((c) => [c.id, c]));

export function listCommands(): CommandDef[] {
  return COMMANDS;
}

export function getCommand(id: CommandId): CommandDef | undefined {
  return BY_ID.get(id);
}

export async function runCommand(id: CommandId): Promise<boolean> {
  const cmd = BY_ID.get(id);
  if (!cmd) return false;
  if (cmd.when && !cmd.when()) return false;
  await cmd.run();
  return true;
}

export function visibleCommands(): CommandDef[] {
  return COMMANDS.filter((c) => !c.when || c.when());
}
