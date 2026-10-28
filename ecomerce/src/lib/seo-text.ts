/**
 * Pure, dependency-free text helpers for SEO metadata.
 *
 * Kept free of imports (no React, no Next, no config) so the behaviour is
 * deterministic and unit-testable in isolation. Used by `src/lib/seo.ts` to
 * never leak HTML, control characters or unbounded author text into `<title>`,
 * meta descriptions or JSON-LD.
 */

/** Anything that looks like an HTML tag, including multi-line tags. */
const HTML_TAG = /<[^>]*>/g;

/** C0/C1 control characters (NUL, newlines, tabs, ...). */
const CONTROL_CHARACTERS = /[\u0000-\u001F\u007F-\u009F]/g;

/** Any run of whitespace (includes U+2028/U+2029 line separators). */
const WHITESPACE = /\s+/g;

/** Punctuation/space left dangling after cutting back to a word boundary. */
const DANGLING_PUNCTUATION = /[\s.,;:!?—–-]+$/u;

/** Appended by {@link truncate} when text had to be cut. */
const ELLIPSIS = "…";

/**
 * Normalises arbitrary (possibly author-supplied) text for use in metadata:
 * strips HTML tags, removes control characters, collapses every run of
 * whitespace — newlines included — to a single space, then trims.
 *
 * Returns an empty string for `undefined`/`null`/empty input.
 */
export function sanitizeText(value: string | undefined | null): string {
  if (!value) return "";
  return value
    .replace(HTML_TAG, " ")
    .replace(CONTROL_CHARACTERS, " ")
    .replace(WHITESPACE, " ")
    .trim();
}

/**
 * Truncates `value` to at most `max` characters — the trailing "…" counts
 * towards the limit — cutting on a word boundary rather than mid-word. Returns
 * `value` untouched when it already fits, so no ellipsis is added to text that
 * was not actually cut.
 */
export function truncate(value: string, max: number): string {
  if (max <= 0) return "";
  if (value.length <= max) return value;

  const room = max - ELLIPSIS.length;
  if (room <= 0) return value.slice(0, max);

  const clipped = value.slice(0, room);

  // If the character we are dropping is whitespace we already cut on a word
  // boundary; otherwise back up to the start of the previous word.
  let base = clipped;
  if (!/\s/.test(value.charAt(room))) {
    const lastSpace = clipped.lastIndexOf(" ");
    if (lastSpace > 0) base = clipped.slice(0, lastSpace);
  }

  const cleaned = base.replace(DANGLING_PUNCTUATION, "");
  return `${cleaned || clipped}${ELLIPSIS}`;
}

/**
 * Resolves a site-relative path (e.g. `/products/blender`) or an already
 * absolute URL against `base`, returning an absolute URL string.
 *
 * Returns `""` for empty input (or when `base` cannot resolve the value) so
 * callers can skip the field instead of emitting a broken URL.
 */
export function absoluteUrl(pathOrUrl: string, base: string): string {
  if (!pathOrUrl) return "";

  const value = pathOrUrl.trim();
  if (!value) return "";

  try {
    return new URL(value, base).toString();
  } catch {
    return "";
  }
}
