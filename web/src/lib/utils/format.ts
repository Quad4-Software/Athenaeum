/**
 * formatBytes renders a byte count as a human-readable size using binary
 * units (KiB-based, labelled KB/MB/GB for familiarity).
 */
export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return "-";
  if (bytes < 1024) return `${bytes} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let size = bytes / 1024;
  let i = 0;
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024;
    i++;
  }
  return `${size.toFixed(1)} ${units[i]}`;
}

/** seriesLabel joins a series name and index into a single label. */
export function seriesLabel(series?: string, index?: number): string {
  if (!series) return "";
  return index ? `${series} #${index}` : series;
}
