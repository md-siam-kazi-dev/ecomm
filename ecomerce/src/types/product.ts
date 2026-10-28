/** A product image with meaningful, human-readable alt text (PRD §46). */
export interface ProductImage {
  url: string;
  alt: string;
}

/**
 * A single selectable variant option (PRD §8).
 * Example: `{ name: "Color", value: "Black" }` or `{ name: "Size", value: "Medium" }`.
 */
export interface ProductVariant {
  id: string;
  /** Variant axis, e.g. `"Color"` or `"Size"`. */
  name: string;
  /** Option value, e.g. `"Black"` or `"Medium"`. */
  value: string;
  /** Optional price override; falls back to the product price when omitted. */
  price?: number;
  stock: number;
  sku: string;
}

/** Optional per-product SEO overrides (PRD §27). */
export interface ProductSeo {
  title: string;
  description: string;
  keywords: string[];
  imageAlt: string;
}

export type StockStatus = "in-stock" | "low-stock" | "out-of-stock";

export interface Product {
  id: string;
  slug: string;
  name: string;
  shortDescription: string;
  description: string;
  price: number;
  /** Original price when the product is discounted. */
  compareAtPrice?: number;
  currency: string;
  images: ProductImage[];
  thumbnail: ProductImage;
  category: string;
  categorySlug: string;
  brand: string;
  /** The store that sells this product, when available. */
  storeName?: string;
  storeSlug?: string;
  /** Absent when the product has no reviews yet. */
  rating?: number;
  reviewCount: number;
  stock: number;
  sku: string;
  tags: string[];
  features: string[];
  specifications: { label: string; value: string }[];
  variants: ProductVariant[];
  isFeatured: boolean;
  isNew: boolean;
  isBestSeller: boolean;
  /** ISO date the product was added — drives "New Arrivals" sorting. */
  createdAt: string;
  /** ISO date the product was last changed — drives sitemap `lastmod`. */
  updatedAt?: string;
  seo?: ProductSeo;
}
