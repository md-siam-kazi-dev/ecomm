"use client";

import { useEffect, useState } from "react";

import { searchCategories, searchProducts } from "@/lib/products";
import type { CategoryWithCount } from "@/lib/products";
import type { Product } from "@/types/product";

import { useDebouncedValue } from "./use-debounced-value";

interface SearchSuggestions {
  /** The trimmed query these suggestions were resolved for. */
  key: string;
  products: Product[];
  categories: CategoryWithCount[];
}

/**
 * Debounced product + category suggestions for the search combobox.
 * Returns `null` while loading or when the resolved key no longer matches
 * the query (stale results are derived away, never rendered).
 */
export function useSearchSuggestions(
  query: string,
  delay = 300
): SearchSuggestions | null {
  const debouncedQuery = useDebouncedValue(query.trim(), delay);
  const [result, setResult] = useState<SearchSuggestions | null>(null);

  useEffect(() => {
    if (!debouncedQuery) return;

    let cancelled = false;

    void Promise.all([
      searchProducts(debouncedQuery, 6),
      searchCategories(debouncedQuery, 3),
    ])
      .then(([products, categories]) => {
        if (!cancelled) {
          setResult({ key: debouncedQuery, products, categories });
        }
      })
      .catch(() => {
        // Suggestions are best-effort: a backend failure must not surface as an
        // unhandled rejection — fall back to an empty (unavailable) state.
        if (!cancelled) {
          setResult({ key: debouncedQuery, products: [], categories: [] });
        }
      });

    return () => {
      cancelled = true;
    };
  }, [debouncedQuery]);

  return result && result.key === debouncedQuery ? result : null;
}
