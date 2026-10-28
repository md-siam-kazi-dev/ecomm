import Link from "next/link";

import { EmptyResults } from "@/components/products/empty-results";
import { ProductGrid } from "@/components/products/product-grid";
import { buttonVariants } from "@/components/ui/button";
import { CATALOG_PAGE_SIZE } from "@/lib/constants";
import { getProducts } from "@/lib/products";
import {
  buildCatalogPath,
  serializeCatalogFilters,
  type CatalogFilters,
} from "@/lib/query-params";

interface CatalogResultsProps {
  filters: CatalogFilters;
  basePath: string;
  fixedCategorySlug?: string;
  hasFilters: boolean;
}

/**
 * Async server results — streams inside Suspense. Renders the product
 * count, the grid and a load-more link that accumulates pages (PRD §12).
 */
export async function CatalogResults({
  filters,
  basePath,
  fixedCategorySlug,
  hasFilters,
}: CatalogResultsProps) {
  const result = await getProducts({
    categorySlug: fixedCategorySlug ?? filters.category,
    query: filters.query,
    minPrice: filters.minPrice,
    maxPrice: filters.maxPrice,
    minRating: filters.rating,
    inStockOnly: filters.inStock,
    onSaleOnly: filters.onSale,
    sort: filters.sort,
    page: 1,
    // Load-more: show PAGE_SIZE × page items in one deterministic call.
    pageSize: CATALOG_PAGE_SIZE * filters.page,
  });

  if (result.total === 0) {
    return <EmptyResults basePath={basePath} hasFilters={hasFilters} />;
  }

  const hasMore = result.items.length < result.total;

  return (
    <div className="flex flex-col gap-6">
      <p className="text-sm text-muted-foreground" aria-live="polite">
        Showing <span className="font-medium text-foreground">{result.items.length}</span>{" "}
        of <span className="font-medium text-foreground">{result.total}</span>{" "}
        {result.total === 1 ? "product" : "products"}
      </p>

      <ProductGrid products={result.items} priorityCount={4} />

      {hasMore && (
        <div className="flex justify-center">
          <Link
            href={buildCatalogPath(
              basePath,
              serializeCatalogFilters(filters),
              { page: filters.page + 1 }
            )}
            scroll={false}
            className={buttonVariants({ variant: "outline" })}
          >
            Load more ({result.total - result.items.length} remaining)
          </Link>
        </div>
      )}
    </div>
  );
}
