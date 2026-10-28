import { CURRENCY } from "./constants";

const priceFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: CURRENCY,
});

/** Formats a numeric price (e.g. `39.99`) as a localized currency string (`$39.99`). */
export function formatPrice(value: number): string {
  return priceFormatter.format(value);
}

/**
 * Returns the discount percentage between the current price and the
 * compare-at price, or `null` when the product is not discounted.
 * Percentages are rounded to the nearest whole number (33.3 → 33% OFF).
 */
export function calculateDiscountPercentage(
  price: number,
  compareAtPrice?: number | null
): number | null {
  if (!compareAtPrice || compareAtPrice <= price) return null;
  return Math.round(((compareAtPrice - price) / compareAtPrice) * 100);
}

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  year: "numeric",
  month: "short",
  day: "numeric",
});

/** Formats an ISO date string (e.g. `"2026-01-05"`) as `"Jan 5, 2026"`. */
export function formatDate(value: string | Date): string {
  return dateFormatter.format(
    typeof value === "string" ? new Date(value) : value
  );
}

const dateTimeFormatter = new Intl.DateTimeFormat("en-US", {
  year: "numeric",
  month: "short",
  day: "numeric",
  hour: "numeric",
  minute: "2-digit",
});

/** Formats an ISO timestamp as `"Jan 5, 2026, 3:04 PM"`. */
export function formatDateTime(value: string | Date): string {
  return dateTimeFormatter.format(
    typeof value === "string" ? new Date(value) : value
  );
}
