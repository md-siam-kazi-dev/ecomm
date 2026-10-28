import type { ProductImage } from "./product";

export interface Category {
  id: string;
  slug: string;
  name: string;
  description: string;
  image: ProductImage;
  /** ISO date the category was last changed — drives sitemap `lastmod`. */
  updatedAt?: string;
}

/** Minimal category option for select inputs (value = slug). */
export interface CategoryOption {
  slug: string;
  name: string;
}
