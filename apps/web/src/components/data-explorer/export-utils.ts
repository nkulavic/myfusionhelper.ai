/**
 * Utility functions for exporting and downloading data.
 * Supports JSON, CSV, clipboard, and blob downloads with proper formatting.
 */

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type ExportFileType = "json" | "csv";

export interface FlattenedRecord {
  [key: string]: string | number | boolean | null | undefined;
}

// ---------------------------------------------------------------------------
// flattenJson
// ---------------------------------------------------------------------------

/**
 * Flatten a nested JSON object into a single-level object whose keys use
 * dot-notation (e.g. `"user.address.city"`).
 *
 * Arrays are serialised to a JSON string so tabular exports stay in one cell.
 *
 * @param obj      - The value to flatten.
 * @param prefix   - Key prefix used during recursion (default `""`).
 * @param maxDepth - Maximum nesting depth to recurse into. Objects deeper than
 *                   this are serialised as JSON strings. Pass `Infinity` (the
 *                   default) for unlimited depth.
 * @returns A flat record with dot-notation keys.
 */
export function flattenJson(
  obj: unknown,
  prefix: string = "",
  maxDepth: number = Infinity
): FlattenedRecord {
  if (obj === null || obj === undefined) {
    return prefix ? { [prefix]: obj ?? null } : {};
  }

  // Arrays -> JSON string
  if (Array.isArray(obj)) {
    return { [prefix || "array"]: JSON.stringify(obj) };
  }

  // Primitives
  if (typeof obj !== "object") {
    return prefix ? { [prefix]: obj as string | number | boolean } : { value: obj as string | number | boolean };
  }

  // If we have exceeded maxDepth, serialise the remaining object as JSON
  if (maxDepth <= 0) {
    return { [prefix || "object"]: JSON.stringify(obj) };
  }

  const result: FlattenedRecord = {};

  for (const key of Object.keys(obj as Record<string, unknown>)) {
    const value = (obj as Record<string, unknown>)[key];
    const newKey = prefix ? `${prefix}.${key}` : key;

    if (value === null || value === undefined) {
      result[newKey] = value ?? null;
    } else if (Array.isArray(value)) {
      result[newKey] = JSON.stringify(value);
    } else if (typeof value === "object" && value.constructor === Object) {
      Object.assign(result, flattenJson(value, newKey, maxDepth - 1));
    } else {
      result[newKey] = value as string | number | boolean;
    }
  }

  return result;
}

// ---------------------------------------------------------------------------
// CSV helpers
// ---------------------------------------------------------------------------

/**
 * Escape a single value for safe inclusion in a CSV cell.
 * - `null` / `undefined` become empty strings.
 * - Values containing commas, double-quotes, or newlines are quoted.
 */
function escapeCsvValue(value: unknown): string {
  if (value === null || value === undefined) {
    return "";
  }

  let str = String(value);

  const needsEscaping =
    str.includes(",") ||
    str.includes('"') ||
    str.includes("\n") ||
    str.includes("\r");

  if (needsEscaping) {
    str = str.replace(/"/g, '""');
    return `"${str}"`;
  }

  return str;
}

/**
 * Convert an array of objects to a CSV string.
 *
 * Nested objects are flattened via {@link flattenJson} so every value maps to a
 * single column. The header row contains the union of all keys across every
 * record.
 *
 * @param data - Non-empty array of objects.
 * @returns A CSV-formatted string (headers + data rows separated by `\n`).
 * @throws If `data` is not a non-empty array.
 */
export function jsonToCsv(data: Record<string, unknown>[]): string {
  if (!Array.isArray(data) || data.length === 0) {
    throw new Error("Data must be a non-empty array");
  }

  const flattenedData = data.map((item) => flattenJson(item));

  // Collect the union of all keys to use as headers
  const allKeys = new Set<string>();
  for (const item of flattenedData) {
    for (const key of Object.keys(item)) {
      allKeys.add(key);
    }
  }

  const headers = Array.from(allKeys);

  const rows: string[] = [];

  // Header row
  rows.push(headers.map(escapeCsvValue).join(","));

  // Data rows
  for (const item of flattenedData) {
    const row = headers.map((header) => escapeCsvValue(item[header]));
    rows.push(row.join(","));
  }

  return rows.join("\n");
}

// ---------------------------------------------------------------------------
// Download helpers
// ---------------------------------------------------------------------------

/**
 * Trigger a browser file download from a `Blob`.
 *
 * Creates a temporary `<a>` element, clicks it, and cleans up afterwards.
 *
 * @param blob     - The blob to download.
 * @param filename - The suggested filename for the download.
 */
export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);

  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.style.display = "none";

  document.body.appendChild(anchor);
  anchor.click();

  // Clean up after a short delay so the browser can initiate the download
  setTimeout(() => {
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
  }, 100);
}

/**
 * Download arbitrary data as a `.json` file.
 *
 * @param data     - Any JSON-serialisable value.
 * @param filename - Optional filename (auto-generated if omitted).
 * @param pretty   - Pretty-print with 2-space indentation (default `true`).
 */
export function downloadJson(
  data: unknown,
  filename?: string,
  pretty: boolean = true
): void {
  const jsonString = pretty
    ? JSON.stringify(data, null, 2)
    : JSON.stringify(data);

  const blob = new Blob([jsonString], { type: "application/json" });
  downloadBlob(blob, filename ?? generateFilename("json"));
}

/**
 * Download an array of objects as a `.csv` file.
 *
 * @param data     - Non-empty array of objects.
 * @param filename - Optional filename (auto-generated if omitted).
 */
export function downloadCsv(
  data: Record<string, unknown>[],
  filename?: string
): void {
  const csvString = jsonToCsv(data);
  const blob = new Blob([csvString], { type: "text/csv;charset=utf-8;" });
  downloadBlob(blob, filename ?? generateFilename("csv"));
}

// ---------------------------------------------------------------------------
// Clipboard
// ---------------------------------------------------------------------------

/**
 * Copy text to the clipboard.
 *
 * Uses the modern Clipboard API when available, falling back to a hidden
 * `<textarea>` + `document.execCommand("copy")` for older browsers.
 *
 * @param text - The string to copy.
 */
export async function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text);
    return;
  }

  // Fallback for insecure contexts / older browsers
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.style.position = "fixed";
  textarea.style.left = "-999999px";
  textarea.style.top = "-999999px";
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();

  try {
    document.execCommand("copy");
  } finally {
    document.body.removeChild(textarea);
  }
}

// ---------------------------------------------------------------------------
// Formatting helpers
// ---------------------------------------------------------------------------

/**
 * Format a byte count as a human-readable string
 * (e.g. `1024` -> `"1 KB"`, `0` -> `"0 Bytes"`).
 *
 * @param bytes - Size in bytes (must be >= 0).
 * @returns A formatted size string with up to two decimal places.
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 Bytes";

  const k = 1024;
  const units = ["Bytes", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${Math.round((bytes / Math.pow(k, i)) * 100) / 100} ${units[i]}`;
}

// ---------------------------------------------------------------------------
// Filename generation
// ---------------------------------------------------------------------------

/**
 * Sanitise a string so it is safe for use inside a filename.
 * Lowercases, replaces non-alphanumeric characters with hyphens, and collapses
 * consecutive hyphens.
 */
function sanitizeFilename(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9\-_]/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
}

/**
 * Generate a descriptive filename for an export.
 *
 * Format: `export-{platform}-{YYYY-MM-DD}-{HH-MM-SS}.{ext}`
 *
 * @param type      - The file extension / export type.
 * @param platform  - Optional platform or context label included in the name.
 * @param timestamp - Date used for the timestamp portion (defaults to now).
 * @returns A filename string.
 */
export function generateFilename(
  type: ExportFileType,
  platform?: string,
  timestamp: Date = new Date()
): string {
  const dateStr = timestamp.toISOString().split("T")[0]; // YYYY-MM-DD
  const timeStr = timestamp
    .toISOString()
    .split("T")[1]
    .split(".")[0]
    .replace(/:/g, "-"); // HH-MM-SS

  const platformPart = platform ? `${sanitizeFilename(platform)}-` : "";

  return `export-${platformPart}${dateStr}-${timeStr}.${type}`;
}
