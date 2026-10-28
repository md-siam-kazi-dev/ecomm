import { Suspense } from "react";

import { ActiveFilters } from "@/components/products/active-filters";
import { CatalogResults } from "@/components/products/catalog-results";
import { FiltersSheet } from "@/components/products/filters-sheet";
import { ProductFilters } from "@/components/products/product-filters";
import { ProductGridSkeleton } from "@/components/products/product-grid-skeleton";
import { ProductSort } from "@/components/products/product-sort";
import type { CategoryWithCount } from "@/lib/products";
import type { CatalogFilters } from "@/lib/query-params";

interface CatalogBrowserProps {
  filters: CatalogFilters;
  categories: CategoryWithCount[];
  basePath: string;
  showCategoryFilter?: boolean;
  /** Set on category pages — results always stay in this category. */
  fixedCategorySlug?: string;
  fixedCategoryName?: string;
}

/**
 * Shared listing layout for /products and /category/[slug]: desktop sidebar
 * or mobile sheet, sort toolbar, active-filter chips and a Suspense-wrapped
 * result grid (PRD §12/§21).
 */
export function CatalogBrowser({
  filters,
  categories,
  basePath,
  showCategoryFilter = true,
  fixedCategorySlug,
  fixedCategoryName,
}: CatalogBrowserProps) {
  // Stable signature excluding page — load-more updates results in place
  // without re-triggering the skeleton fallback.
  const signature = JSON.stringify({ ...filters, page: undefined });
  const hasFilters = Boolean(
    (showCategoryFilter && filters.category) ||
      filters.query ||
      filters.minPrice !== undefined ||
      filters.maxPrice !== undefined ||
      filters.rating !== undefined ||
      filters.inStock ||
      filters.onSale
  );

  return (
    <div className="grid gap-8 lg:grid-cols-[240px_1fr]">
      <aside className="hidden lg:block">
        <ProductFilters
          key={signature}
          categories={categories}
          current={filters}
          showCategory={showCategoryFilter}
        />
      </aside>

      <div className="flex min-w-0 flex-col gap-4">
        <div className="flex items-center justify-between gap-3">
          <FiltersSheet
            categories={categories}
            current={filters}
            showCategory={showCategoryFilter}
          />
          <ProductSort sort={filters.sort} />
        </div>

        <ActiveFilters
          filters={filters}
          basePath={basePath}
          showCategory={showCategoryFilter}
          categoryName={fixedCategoryName}
        />

        <Suspense key={signature} fallback={<ProductGridSkeleton count={8} />}>
          <CatalogResults
            filters={filters}
            basePath={basePath}
            fixedCategorySlug={fixedCategorySlug}
            hasFilters={hasFilters}
          />
        </Suspense>
      </div>
    </div>
  );
}
