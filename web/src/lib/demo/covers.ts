import { DEMO_CATALOG } from "./catalog";

const PALETTE = [
  ["#8B1E1E", "#B34444"],
  ["#1E3A5F", "#3A6A9A"],
  ["#2F4F3E", "#4F7A62"],
  ["#5C3A21", "#8A5A38"],
  ["#3D2B4F", "#6A4F88"],
  ["#1F4E5F", "#3A7A8E"],
  ["#6B2D3C", "#9A4A5C"],
  ["#2C3E50", "#4A647A"],
];

function hash(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

function escapeXml(s: string): string {
  return s
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&apos;");
}

/** SVG data URL cover for offline demo mode (img src works without a backend). */
export function demoCoverDataUrl(title: string, author: string): string {
  const [bg, accent] = PALETTE[hash(title + "|" + author) % PALETTE.length]!;
  const t = escapeXml(title.length > 28 ? `${title.slice(0, 27)}...` : title);
  const a = escapeXml(author.length > 30 ? `${author.slice(0, 29)}...` : author);
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="600" viewBox="0 0 400 600">
  <rect width="400" height="600" fill="${bg}"/>
  <rect y="80" width="400" height="80" fill="${accent}"/>
  <rect x="36" y="200" width="20" height="320" fill="#f0f0eb"/>
  <rect x="70" y="220" width="290" height="280" fill="rgba(20,18,16,0.7)"/>
  <text x="84" y="280" fill="#faf6f0" font-family="Georgia, serif" font-size="22">${t}</text>
  <text x="84" y="320" fill="#c8beb0" font-family="Georgia, serif" font-size="16">${a}</text>
</svg>`;
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}

export function demoCoverUrlForBook(id: number): string {
  const entry = DEMO_CATALOG.find((e) => e.id === id) ?? DEMO_CATALOG[0]!;
  if (entry.coverUrl) return entry.coverUrl;
  return demoCoverDataUrl(entry.title, entry.author);
}
