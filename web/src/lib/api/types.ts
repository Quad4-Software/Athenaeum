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
  | "kindle";

import type { PasswordPolicy } from "$lib/utils/password-strength";
export type { PasswordStrength, PasswordPolicy } from "$lib/utils/password-strength";

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

export interface MetadataProvider {
  id: string;
  label: string;
  description?: string;
  requiresAsin?: boolean;
}

export interface MetadataMatch {
  source: string;
  sourceId?: string;
  title: string;
  author: string;
  description?: string;
  language?: string;
  series?: string;
  seriesIndex?: number;
  isbn?: string;
  asin?: string;
  coverUrl?: string;
  publishedYear?: number;
}

export interface MetadataSearchQuery {
  title?: string;
  author?: string;
  isbn?: string;
  asin?: string;
  providers?: string[];
}

export interface Chapter {
  index: number;
  title: string;
  startSec: number;
}

export interface Book {
  id: number;
  libraryId?: number;
  title: string;
  author: string;
  series?: string;
  seriesIndex?: number;
  format: BookFormat;
  relPath: string;
  fileSize: number;
  hasCover: boolean;
  language?: string;
  description?: string;
  addedAt: string;
  modifiedAt: string;
  metaEdited?: boolean;
  coverEdited?: boolean;
  contentHash?: string;
  duplicateOf?: number;
  progressPercent?: number;
  tags?: string[];
  userRating?: number;
}

export interface BookUpdate {
  title: string;
  author: string;
  series?: string;
  seriesIndex?: number;
  language?: string;
  description?: string;
}

export interface BookPage {
  items: Book[];
  total: number;
  limit: number;
  offset: number;
}

export interface HealthResponse {
  status: string;
  version?: string;
  webVersion?: string;
  telemetry?: TelemetryConfig;
}

export interface TelemetryConfig {
  sentryDsn?: string;
  environment?: string;
  release?: string;
  tracesSampleRate?: number;
}

export interface LibraryStats {
  totalBooks: number;
  epubCount: number;
  pdfCount: number;
  audioCount: number;
  totalSizeBytes: number;
  authorCount: number;
  seriesCount: number;
  libraryCount: number;
  addedLast7Days: number;
  collectionCount: number;
  readingInProgress?: number;
  readingCompleted?: number;
  favoriteCount?: number;
  userCount?: number;
  lastScanAt?: string;
  scanning: boolean;
  authEnabled: boolean;
}

export interface Progress {
  bookId: number;
  userId?: number;
  location: string;
  percent: number;
  readSeconds?: number;
  updatedAt: string;
}

export interface Bookmark {
  id: number;
  bookId: number;
  location: string;
  label?: string;
  createdAt: string;
}

export interface Tag {
  id: number;
  name: string;
}

export interface BookRating {
  userId?: number;
  bookId: number;
  rating: number;
  updatedAt?: number;
}

export interface ReaderPrefs {
  userId?: number;
  prefs: Record<string, unknown>;
  updatedAt?: number;
}

export interface LoginChallenge {
  needsTotp: boolean;
  totpToken: string;
}

export type LoginResult = User | LoginChallenge;

export function isLoginChallenge(r: LoginResult): r is LoginChallenge {
  return "needsTotp" in r && r.needsTotp === true;
}

export interface TOTPSetup {
  secret: string;
  otpauthUrl: string;
}

export interface AuthSettings {
  allowRegistration: boolean;
  requireTotp: boolean;
}

export interface Highlight {
  id: number;
  bookId: number;
  location: string;
  excerpt?: string;
  note?: string;
  color?: string;
  createdAt: string;
}

export interface ComicPage {
  index: number;
  name: string;
  mimeType: string;
}

export interface ComicManifest {
  total: number;
  pages: ComicPage[];
}

export interface MobiSection {
  index: number;
  title: string;
  html: string;
}

export interface AudiobookTrack {
  index: number;
  title: string;
  relPath: string;
  format: string;
  fileSize: number;
}

export interface ConvertResult {
  targetFormat: string;
  outputPath: string;
  bookId?: number;
  message?: string;
}

export interface ReadingStats {
  totalReadSeconds: number;
  booksInProgress: number;
  booksCompleted: number;
  currentStreakDays: number;
}

export interface AuthorInfo {
  name: string;
  count: number;
}

export interface ScanStatus {
  scanning: boolean;
  indexed: number;
  skipped: number;
  currentPath?: string;
  libraryName?: string;
  startedAt?: string;
  finishedAt?: string;
}

export interface MetadataMatchStatus {
  running: boolean;
  total: number;
  done: number;
  matched: number;
  skipped: number;
  failed: number;
  currentTitle?: string;
  startedAt?: string;
  finishedAt?: string;
}

export interface MetadataAutoMatchRequest {
  bookIds?: number[];
  libraryId?: number;
  applyCover?: boolean;
}

export interface IntegrityReport {
  totalBooks: number;
  missingCount: number;
  missingFiles: { id: number; libraryId: number; title: string; relPath: string }[];
  orphanCovers: number;
}

export interface MaintenanceStatus {
  running: boolean;
  task?: string;
  total: number;
  done: number;
  updated: number;
  skipped: number;
  failed: number;
  currentTitle?: string;
  startedAt?: string;
  finishedAt?: string;
}

export type Permission =
  "read" | "edit_metadata" | "delete_books" | "manage_library" | "manage_users";

export interface User {
  id: number;
  username: string;
  email?: string;
  isAdmin: boolean;
  isGuest?: boolean;
  expiresAt?: string;
  localAuth?: boolean;
  totpEnabled?: boolean;
  permissions?: Permission[];
  createdAt: string;
}

export interface GuestCredentials {
  user: User;
  password: string;
}

export interface ServerConfig {
  metricsEnabled: boolean;
  metricsAuth: boolean;
  metricsUsername: string;
  metricsPassword?: string;
  metricsPasswordSet: boolean;
  trustedProxies: string;
  corsEnabled: boolean;
  corsOrigins: string;
  cspEnabled: boolean;
  cspPolicy: string;
  autoScanEnabled?: boolean;
  autoScanIntervalSec?: number;
  scanWorkers?: number;
}

export interface SMTPSettingsPublic {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  passwordSet: boolean;
  fromAddr: string;
  useTls: boolean;
}

export interface PocketIDSettingsPublic {
  enabled: boolean;
  baseUrl: string;
  apiKeySet: boolean;
  defaultGroupIds: string[];
}

export interface Invite {
  id: number;
  token: string;
  kind: "permanent" | "guest";
  email?: string;
  permissions: string[];
  createdBy: number;
  expiresAt?: string;
  guestExpiresAt?: string;
  pocketIdUserId?: string;
  acceptedAt?: string;
  acceptedUserId?: number;
  revokedAt?: string;
  createdAt: string;
  status: string;
}

export interface InviteCreateResult {
  invite: Invite;
  url: string;
  pocketIdSetupUrl?: string;
  emailSent: boolean;
}

export interface InviteMeta {
  kind: string;
  emailPresent: boolean;
  expiresAt?: string;
  valid: boolean;
  reason?: string;
  pocketIdConfigured: boolean;
}

export type WebhookEvent =
  | "user.create"
  | "user.delete"
  | "invite.created"
  | "invite.accepted"
  | "book.upload"
  | "library.scan.complete";

export interface Webhook {
  id: number;
  url: string;
  secretSet: boolean;
  events: string[];
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface WebhookDelivery {
  id: number;
  webhookId: number;
  event: string;
  payload: string;
  statusCode: number;
  success: boolean;
  attempts: number;
  lastError?: string;
  createdAt: string;
  deliveredAt?: string;
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

export interface SystemStats {
  version: string;
  webVersion: string;
  cpuPercent: number;
  memUsed: number;
  memTotal: number;
  memPercent: number;
  disks: {
    path: string;
    total: number;
    used: number;
    available: number;
    percent: number;
  }[];
  sandbox?: SandboxStatus;
}

export interface UserSession {
  id: string;
  userId?: number;
  ip?: string;
  userAgent?: string;
  device?: string;
  authMethod: string;
  createdAt: string;
  lastSeenAt: string;
  expiresAt: string;
  current: boolean;
}

export interface AltchaWidgetPublic {
  auto?: string;
  display?: string;
  hideFooter?: boolean;
  hideLogo?: boolean;
  language?: string;
  name?: string;
  theme?: string;
  type?: string;
  workers?: number;
}

/** Browser-safe ALTCHA widget configuration from auth methods. */
export interface AltchaPublic {
  enabled: boolean;
  challengeUrl?: string;
  protectLogin: boolean;
  protectSetup: boolean;
  widget: AltchaWidgetPublic;
}

export interface AuthMethods {
  authEnabled: boolean;
  loginLocal: boolean;
  loginOidc: boolean;
  oidcButtonText?: string;
  oidcAutoLaunch: boolean;
  allowRegistration?: boolean;
  passwordPolicy?: PasswordPolicy;
  altcha?: AltchaPublic;
}

export type OIDCMatchBy = "username" | "email" | "sub";

export interface OIDCConfig {
  enabled: boolean;
  loginLocal: boolean;
  issuerUrl: string;
  authorizeUrl: string;
  tokenUrl: string;
  userinfoUrl: string;
  jwksUrl: string;
  logoutUrl?: string;
  clientId: string;
  clientSecret?: string;
  clientSecretSet: boolean;
  signingAlgorithm: string;
  buttonText: string;
  matchBy: OIDCMatchBy;
  autoRegister: boolean;
  autoLaunch: boolean;
  groupClaim?: string;
  adminGroups?: string;
}

export interface OIDCDiscovery {
  issuerUrl: string;
  authorizeUrl: string;
  tokenUrl: string;
  userinfoUrl: string;
  jwksUrl: string;
  logoutUrl?: string;
}

export interface AuditEntry {
  id: number;
  actorId: number;
  actorName: string;
  targetUserId?: number;
  targetName?: string;
  action: string;
  details?: string;
  ip?: string;
  createdAt: string;
}

export interface AuditPage {
  items: AuditEntry[];
  total: number;
  limit: number;
  offset: number;
}

export type CollectionKind = "manual" | "smart" | "auto" | "reading";

export interface SmartQuery {
  format?: BookFormat | "";
  author?: string;
  series?: string;
  search?: string;
  addedDays?: number;
}

export interface Collection {
  id: number;
  userId?: number;
  name: string;
  description?: string;
  kind: CollectionKind;
  query?: SmartQuery;
  bookCount: number;
  createdAt: string;
}

export interface LibraryS3Config {
  endpoint: string;
  region: string;
  bucket: string;
  prefix: string;
  accessKey: string;
  usePathStyle: boolean;
  tls: boolean;
  hasSecretKey: boolean;
}

export interface LibraryS3Input {
  endpoint: string;
  region: string;
  bucket: string;
  prefix: string;
  accessKey: string;
  secretKey: string;
  usePathStyle: boolean;
  tls: boolean;
}

export interface LibraryCreateInput {
  name: string;
  mountPath?: string;
  backend?: "local" | "s3";
  s3?: LibraryS3Input;
}

export interface LibraryMount {
  id: number;
  name: string;
  mountPath: string;
  backend: "local" | "s3" | string;
  s3?: LibraryS3Config;
  sortOrder: number;
  bookCount: number;
  createdAt: string;
}

export interface UploadSession {
  id: string;
  libraryId: number;
  userId: number;
  relPath: string;
  totalSize: number;
  offset: number;
  done: boolean;
  bookId?: number;
  createdAt: string;
  updatedAt: string;
}

export interface UserLibraries {
  libraryIds: number[];
}

export interface SeriesInfo {
  name: string;
  count: number;
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

export interface FSDirEntry {
  name: string;
  path: string;
}

export interface FSBrowseResult {
  path?: string;
  parent?: string;
  entries: FSDirEntry[];
}

export type SidebarSectionId =
  "libraries" | "formats" | "series" | "favorites" | "continue" | "reading" | "shelves";

export interface SidebarPrefs {
  order: SidebarSectionId[];
  hidden: SidebarSectionId[];
}

export interface APIKey {
  id: number;
  userId: number;
  name: string;
  prefix: string;
  createdAt: string;
  lastUsedAt?: string;
}

export interface APIKeyCreated extends APIKey {
  key: string;
}

export interface APIDocEndpoint {
  method: string;
  path: string;
  summary: string;
  auth?: string;
  query?: string;
  body?: string;
  response?: string;
}

export interface APIDocSection {
  title: string;
  endpoints: APIDocEndpoint[];
}

export interface APIDoc {
  title: string;
  version: string;
  auth: string[];
  baseUrl: string;
  contentTypes: string[];
  sections: APIDocSection[];
}

export interface I18nLocaleInfo {
  code: string;
  name: string;
  source: "bundled" | "custom";
}

export interface I18nCatalog {
  locales: I18nLocaleInfo[];
}
