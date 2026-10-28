import type { Metadata } from "next";

import { CatalogBrowser } from "@/components/products/catalog-browser";
import { getCategories } from "@/lib/products";
import {
  hasActiveFilters,
  parseCatalogFilters,
  type CatalogSearchParams,
} from "@/lib/query-params";
import { buildMetadata } from "@/lib/seo";

const BASE_PATH = "/products";

interface ProductsPageProps {
  searchParams: Promise<CatalogSearchParams>;
}

/**
 * Indexation matrix (SEO plan, Part 4): content filters (`category`, `q`, `min`,
 * `max`, `rating`, `stock`, `sale`) → `noindex, follow` with a canonical to the
 * base path; sort-only and the base view are indexable at the base path; deeper
 * pages with no content filters are indexable and self-canonical (`?page=N`).
 */
export async function generateMetadata({
  searchParams,
}: ProductsPageProps): Promise<Metadata> {
  const params = await searchParams;
  const filters = parseCatalogFilters(params);
  const hasContentFilters = hasActiveFilters(filters, true);
  const isPaged = !hasContentFilters && filters.page > 1;

  return buildMetadata({
    title: isPaged ? `All Products — Page ${filters.page}` : "All Products",
    description:
      "Browse the full ShopKart catalog — filter by category, price, rating and availability, and sort to find exactly what you need.",
    path: isPaged ? `${BASE_PATH}?page=${filters.page}` : BASE_PATH,
    noIndex: hasContentFilters,
    follow: true,
  });
}

/** Product listing page (PRD §12) — filters, sort and page live in the URL. */
export default async function ProductsPage({
  searchParams,
}: ProductsPageProps) {
  const params = await searchParams;
  const filters = parseCatalogFilters(params);
  const categories = await getCategories();

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          All Products
        </h1>
        <p className="mt-2 max-w-2xl text-sm text-muted-foreground sm:text-base">
          Browse everything in the ShopKart catalog — filter by category,
          price, rating and availability.
        </p>
      </header>

      <CatalogBrowser filters={filters} categories={categories} basePath="/products" />
    </div>
  );
}
