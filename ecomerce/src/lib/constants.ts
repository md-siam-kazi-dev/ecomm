/**
 * Global site constants. Safe to import from both server and client
 * components — keep this file free of browser-only APIs.
 */

export const SITE_CONFIG = {
  name: "ShopKart",
  tagline: "Modern essentials for everyday living",
  description:
    "Shop everyday essentials, electronics, home goods and more at ShopKart. Free shipping, easy returns and secure checkout.",
  /** Absolute origin used for canonical URLs, sitemap and Open Graph. */
  url: process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000",
} as const;

export const CURRENCY = "USD";

/** Orders at or above this amount ship free (drives cart shipping estimate). */
export const FREE_SHIPPING_THRESHOLD = 75;

/** Flat-rate shipping applied below the free-shipping threshold. */
export const STANDARD_SHIPPING_RATE = 5.99;

/** Flat rate for express delivery, regardless of order value. */
export const EXPRESS_SHIPPING_RATE = 12.99;

/** Products with stock at or below this count are flagged "low stock". */
export const LOW_STOCK_THRESHOLD = 10;

/** Products shown per listing page (load-more increments multiply this). */
export const CATALOG_PAGE_SIZE = 12;

export const ANNOUNCEMENT_MESSAGE = `Free shipping on orders over $${FREE_SHIPPING_THRESHOLD} · 30-day easy returns`;

export interface NavLink {
  label: string;
  href: string;
}

/**
 * Primary header navigation. A "Categories" dropdown will be added once the
 * data layer provides categories (Phase 3).
 */
export const NAV_LINKS: NavLink[] = [
  { label: "Shop", href: "/products" },
  { label: "Deals", href: "/products?sort=discount" },
  { label: "New Arrivals", href: "/products?sort=newest" },
];
