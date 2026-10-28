import { Suspense } from "react";
import type { Metadata } from "next";

import { ProductGridSkeleton } from "@/components/products/product-grid-skeleton";
import { SearchInput } from "@/components/search/search-input";
import { SearchResults } from "@/components/search/search-results";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Search",
  description:
    "Search the full ShopKart catalog — find products and categories in seconds.",
  path: "/search",
  noIndex: true,
  follow: true,
});

interface SearchPageProps {
  searchParams: Promise<{ q?: string | string[] }>;
}

/** Search page (PRD §20) — the query lives in the URL: /search?q=blender. */
export default async function SearchPage({ searchParams }: SearchPageProps) {
  const params = await searchParams;
  const query = typeof params.q === "string" ? params.q : "";

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <header className="mx-auto max-w-2xl text-center">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          Search
        </h1>
        <p className="mt-2 text-sm text-muted-foreground sm:text-base">
          Find products and categories across the entire store.
        </p>
      </header>

      <div className="mx-auto mt-8 max-w-2xl">
        <SearchInput initialQuery={query} />
      </div>

      {query.trim() && (
        <div className="mx-auto mt-12 max-w-5xl">
          <Suspense key={query} fallback={<ProductGridSkeleton count={4} />}>
            <SearchResults query={query} />
          </Suspense>
        </div>
      )}
    </div>
  );
}
