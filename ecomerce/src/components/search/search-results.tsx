import Link from "next/link";
import { SearchX } from "lucide-react";

import { ProductGrid } from "@/components/products/product-grid";
import { buttonVariants } from "@/components/ui/button";
import { searchCategories, searchProducts } from "@/lib/products";
import { cn } from "@/lib/utils";

interface SearchResultsProps {
  query: string;
}

/**
 * Full search results (PRD §20) — server-rendered from the ?q= param with
 * category matches and a product grid. Streams inside Suspense.
 */
export async function SearchResults({ query }: SearchResultsProps) {
  const trimmed = query.trim();

  const [products, categories] = await Promise.all([
    searchProducts(trimmed, 24),
    searchCategories(trimmed, 6),
  ]);

  if (products.length === 0 && categories.length === 0) {
    return (
      <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed px-6 py-16 text-center">
        <span className="flex size-14 items-center justify-center rounded-full bg-muted">
          <SearchX className="size-7 text-muted-foreground" aria-hidden="true" />
        </span>
        <div>
          <h2 className="text-lg font-semibold">No results for “{trimmed}”</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Check the spelling or try a more general term.
          </p>
        </div>
        <Link
          href="/products"
          className={cn(buttonVariants({ variant: "outline" }))}
        >
          Browse all products
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-10">
      {categories.length > 0 && (
        <section aria-labelledby="search-categories-heading">
          <h2
            id="search-categories-heading"
            className="text-sm font-semibold text-muted-foreground"
          >
            Categories
          </h2>
          <div className="mt-3 flex flex-wrap gap-2">
            {categories.map((category) => (
              <Link
                key={category.id}
                href={`/category/${category.slug}`}
                className="inline-flex items-center gap-1.5 rounded-full border bg-secondary px-3.5 py-2 text-sm font-medium transition-colors hover:bg-muted focus-visible:outline-2 focus-visible:outline-offset-2"
              >
                {category.name}
                <span className="text-xs text-muted-foreground">
                  ({category.productCount})
                </span>
              </Link>
            ))}
          </div>
        </section>
      )}

      <section aria-labelledby="search-products-heading">
        <h2
          id="search-products-heading"
          className="text-sm font-semibold text-muted-foreground"
        >
          Products
        </h2>
        <p className="mt-3 text-sm text-muted-foreground" aria-live="polite">
          <span className="font-medium text-foreground">{products.length}</span>{" "}
          {products.length === 1 ? "product" : "products"} matching{" "}
          <span className="font-medium text-foreground">“{trimmed}”</span>
        </p>
        <div className="mt-5">
          <ProductGrid products={products} />
        </div>
      </section>
    </div>
  );
}
