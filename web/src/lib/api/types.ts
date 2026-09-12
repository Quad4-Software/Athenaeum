import type * as gen from "./generated/models";

// Interfaces mirroring Go API models live in generated/models.ts, emitted by
// genapi from the structs in internal/models (run `task generate`). They are
// re-exported here so existing imports keep working.
export * from "./generated/models";

export type { PasswordStrength, PasswordPolicy } from "$lib/utils/password-strength";

export type BookFormat =
  | "epub"
  | "pdf"
  | "mobi"
  | "azw3"
  | "azw"
  | "kfx"
  | "cbz"
  | "cbr"
  | "audiobook"
  | "mp3"
  | "m4b"
  | "m4a"
  | "ogg"
  | "flac"
  | "audio"
  | "comic"
  | "kindle"
  | "papers";

export const AUDIO_FORMATS = ["mp3", "m4b", "m4a", "ogg", "flac", "audiobook"] as const;

export function isAudioFormat(format: string): boolean {
  return (AUDIO_FORMATS as readonly string[]).includes(format);
}

export function isComicFormat(format: string): boolean {
  return format === "cbz" || format === "cbr" || format === "comic";
}

export function isMobiFormat(format: string): boolean {
  return format === "mobi" || format === "azw3" || format === "azw";
}

export function isPaperBook(book: { doi?: string; arxivId?: string; pubmedId?: string }): boolean {
  return Boolean(book.doi || book.arxivId || book.pubmedId);
}

// The following aliases keep generated shapes but narrow fields where the API
// contract is tighter than the Go type (string enums) or where the client
// sends partial payloads the Go struct marks as required.

export type Book = Omit<gen.Book, "format"> & { format: BookFormat };

export type BookPage = Omit<gen.BookPage, "items"> & { items: Book[] };

export type MetadataSearchQuery = Partial<gen.MetadataSearchQuery>;

export type Permission =
  "read" | "edit_metadata" | "delete_books" | "manage_library" | "manage_users";

export type User = Omit<gen.User, "permissions" | "totpEnabled"> & {
  permissions?: Permission[];
  totpEnabled?: boolean;
};

export type Invite = Omit<gen.Invite, "kind"> & { kind: "permanent" | "guest" };

export type OIDCMatchBy = "username" | "email" | "sub";

export type OIDCConfig = Omit<gen.OIDCConfig, "matchBy"> & { matchBy: OIDCMatchBy };

export type CollectionKind = "manual" | "smart" | "auto" | "reading";

export type SmartQuery = Omit<gen.SmartQuery, "format"> & { format?: BookFormat | "" };

export type Collection = Omit<gen.Collection, "kind" | "query"> & {
  kind: CollectionKind;
  query?: SmartQuery;
};

export type LibraryCreateInput = Omit<gen.LibraryCreateInput, "mountPath" | "backend"> & {
  mountPath?: string;
  backend?: "local" | "s3";
};

export type I18nLocaleInfo = Omit<gen.I18nLocaleInfo, "source"> & {
  source: "bundled" | "custom";
};

export type I18nCatalog = Omit<gen.I18nCatalog, "locales"> & {
  locales: I18nLocaleInfo[];
};

export type GuestCredentials = Omit<gen.GuestCredentials, "user"> & { user: User };

export type ServerConfig = gen.ServerConfig & { metricsPassword?: string };

export type WebhookEvent =
  | "user.create"
  | "user.delete"
  | "invite.created"
  | "invite.accepted"
  | "book.upload"
  | "library.scan.complete";

export type LoginResult = User | gen.LoginChallenge;

export function isLoginChallenge(r: LoginResult): r is gen.LoginChallenge {
  return "needsTotp" in r && r.needsTotp === true;
}

// Hand-written shapes without a single Go struct behind them (map literals in
// handlers, unexported request bodies, or UI-only client types).

export interface HealthResponse {
  status: string;
  database?: string;
  version?: string;
  webVersion?: string;
  scanning?: boolean;
  lastScan?: string;
  diskFreeBytes?: number;
  telemetry?: gen.TelemetryConfig;
}

export interface TOTPSetup {
  secret: string;
  otpauthUrl: string;
}

export interface SandboxComponentStatus {
  state: string;
  detail?: string;
  error?: string;
  reason?: string;
}

export interface SandboxStatus {
  mode: string;
  landlock: SandboxComponentStatus;
  seccomp: SandboxComponentStatus;
}

export type SystemStats = Omit<gen.SystemStats, "sandbox"> & { sandbox?: SandboxStatus };

export interface UserLibraries {
  libraryIds: number[];
}

export type SortKey = "recent" | "oldest" | "title" | "author" | "progress";

export interface BookQueryParams {
  search?: string;
  sort?: SortKey;
  format?: BookFormat | "";
  series?: string;
  author?: string;
  library?: number;
  collection?: number;
  favorites?: boolean;
  inProgress?: boolean;
  tag?: string;
  limit?: number;
  offset?: number;
}

export type SidebarSectionId =
  "libraries" | "formats" | "series" | "favorites" | "continue" | "reading" | "shelves";

export interface SidebarPrefs {
  order: SidebarSectionId[];
  hidden: SidebarSectionId[];
}
