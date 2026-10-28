"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { ProductSortOption } from "@/lib/products";
import { buildCatalogPath } from "@/lib/query-params";

const SORT_LABELS: Record<ProductSortOption, string> = {
  featured: "Featured",
  newest: "Newest",
  "price-asc": "Price: Low to High",
  "price-desc": "Price: High to Low",
  rating: "Top Rated",
  discount: "Biggest Discount",
};

interface ProductSortProps {
  sort: ProductSortOption;
}

/** Sort control — writes the choice to the URL (PRD §12). */
export function ProductSort({ sort }: ProductSortProps) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  return (
    <div className="flex flex-1 items-center gap-2 sm:flex-none">
      <span className="hidden text-sm text-muted-foreground sm:block">
        Sort by
      </span>
      <Select
        value={sort}
        onValueChange={(value) => {
          router.push(
            buildCatalogPath(pathname, searchParams, {
              sort: value === "featured" ? null : value,
              page: null,
            }),
            { scroll: false }
          );
        }}
      >
        <SelectTrigger
          className="w-full sm:w-[170px]"
          aria-label="Sort products"
        >
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {(Object.keys(SORT_LABELS) as ProductSortOption[]).map((value) => (
            <SelectItem key={value} value={value}>
              {SORT_LABELS[value]}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
