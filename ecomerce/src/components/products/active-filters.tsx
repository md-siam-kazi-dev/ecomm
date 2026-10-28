import Link from "next/link";
import { X } from "lucide-react";

import { formatPrice } from "@/lib/formatters";
import {
  buildCatalogPath,
  serializeCatalogFilters,
  type CatalogFilters,
} from "@/lib/query-params";

interface ActiveFiltersProps {
  filters: CatalogFilters;
  basePath: string;
  showCategory?: boolean;
  /** Friendly label for a fixed category (category pages). */
  categoryName?: string;
}

interface FilterChip {
  key: string;
  label: string;
  remove: Record<string, string | number | null>;
}

/** Removable chips for every active filter + a clear-all action (PRD §12). */
export function ActiveFilters({
  filters,
  basePath,
  showCategory = true,
  categoryName,
}: ActiveFiltersProps) {
  const chips: FilterChip[] = [];

  if (showCategory && filters.category) {
    chips.push({
      key: "category",
      label: categoryName ?? filters.category,
      remove: { category: null },
    });
  }
  if (filters.query) {
    chips.push({ key: "q", label: `“${filters.query}”`, remove: { q: null } });
  }
  if (filters.minPrice !== undefined) {
    chips.push({
      key: "min",
      label: `Min ${formatPrice(filters.minPrice)}`,
      remove: { min: null },
    });
  }
  if (filters.maxPrice !== undefined) {
    chips.push({
      key: "max",
      label: `Max ${formatPrice(filters.maxPrice)}`,
      remove: { max: null },
    });
  }
  if (filters.rating !== undefined) {
    chips.push({
      key: "rating",
      label: `${filters.rating}★ & up`,
      remove: { rating: null },
    });
  }
  if (filters.inStock) {
    chips.push({ key: "stock", label: "In stock", remove: { stock: null } });
  }
  if (filters.onSale) {
    chips.push({ key: "sale", label: "On sale", remove: { sale: null } });
  }

  if (chips.length === 0) return null;

  return (
    <div className="flex flex-wrap items-center gap-2">
      {chips.map((chip) => (
        <Link
          key={chip.key}
          href={buildCatalogPath(basePath, serializeCatalogFilters(filters), {
            ...chip.remove,
            page: null,
          })}
          className="inline-flex items-center gap-1 rounded-full border bg-secondary px-2.5 py-1 text-xs font-medium transition-colors hover:bg-muted focus-visible:outline-2 focus-visible:outline-offset-2"
        >
          {chip.label}
          <X className="size-3" aria-hidden="true" />
          <span className="sr-only">(remove filter)</span>
        </Link>
      ))}
      <Link
        href={basePath}
        className="rounded-md px-1.5 py-1 text-xs font-medium text-primary transition-colors hover:text-primary-dark focus-visible:outline-2 focus-visible:outline-offset-2"
      >
        Clear all
      </Link>
    </div>
  );
}
