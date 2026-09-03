// Indonesian formatting.

const LOCALE = "id-ID";

/** 2 September 2026 */
export function formatTanggal(value: string | Date | null | undefined): string {
  const d = toDate(value);
  if (!d) return "";
  return d.toLocaleDateString(LOCALE, { day: "numeric", month: "long", year: "numeric" });
}

/** 2 Sep 2026 */
export function formatTanggalPendek(value: string | Date | null | undefined): string {
  const d = toDate(value);
  if (!d) return "";
  return d.toLocaleDateString(LOCALE, { day: "numeric", month: "short", year: "numeric" });
}

/** 1.234.567 */
export function formatAngka(value: number | null | undefined): string {
  if (value === null || value === undefined || Number.isNaN(value)) return "0";
  return value.toLocaleString(LOCALE);
}

function toDate(value: string | Date | null | undefined): Date | null {
  if (!value) return null;
  const d = value instanceof Date ? value : new Date(value);
  // An invalid date renders as "Invalid Date" otherwise, which is worse than
  // rendering nothing.
  return Number.isNaN(d.getTime()) ? null : d;
}
