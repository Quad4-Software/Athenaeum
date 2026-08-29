export type CommandSection = "navigate" | "actions" | "settings";

export type CommandScope = "shell" | "reader" | "global";

export type CommandId =
  | "palette.open"
  | "nav.library"
  | "nav.continue"
  | "nav.favorites"
  | "nav.collections"
  | "nav.settings"
  | "nav.format.epub"
  | "nav.format.pdf"
  | "nav.format.comic"
  | "nav.format.kindle"
  | "nav.format.audio"
  | "ui.toggleSidebar"
  | "ui.toggleTheme"
  | "ui.toggleDensity"
  | "library.scan"
  | "help.shortcuts";

export type KeyChord = {
  key: string;
  mod?: boolean;
  shift?: boolean;
  alt?: boolean;
};

export type CommandDef = {
  id: CommandId;
  titleKey: string;
  section: CommandSection;
  scope: CommandScope;
  defaultBinding?: KeyChord | null;
  when?: () => boolean;
  run: () => void | Promise<void>;
};
