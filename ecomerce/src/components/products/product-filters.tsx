"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { Star } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { CategoryWithCount } from "@/lib/products";
import { buildCatalogPath, type CatalogFilters } from "@/lib/query-params";
import { cn } from "@/lib/utils";

interface ProductFiltersProps {
  categories: CategoryWithCount[];
  current: CatalogFilters;
  showCategory?: boolean;
  /** Called after any navigation — lets the mobile sheet close itself. */
  onNavigate?: () => void;
}

const RATING_OPTIONS = [4, 3, 2];

/**
 * Catalog filter controls (PRD §12): category links with counts, price range
 * form, minimum-rating links and availability toggles. All navigation is
 * URL-based so every filter combination is shareable.
 */
export function ProductFilters({
  categories,
  current,
  showCategory = true,
  onNavigate,
}: ProductFiltersProps) {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const router = useRouter();

  const [minPrice, setMinPrice] = useState(
    current.minPrice !== undefined ? String(current.minPrice) : ""
  );
  const [maxPrice, setMaxPrice] = useState(
    current.maxPrice !== undefined ? String(current.maxPrice) : ""
  );

  function navigate(updates: Record<string, string | number | null>) {
    onNavigate?.();
    return buildCatalogPath(pathname, searchParams, updates);
  }

  function handlePriceApply(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    router.push(
      navigate({ min: minPrice || null, max: maxPrice || null, page: null })
    );
  }

  const linkClassName = (active: boolean) =>
    cn(
      "flex items-center gap-1.5 rounded-md px-2 py-1.5 text-sm transition-colors focus-visible:outline-2 focus-visible:outline-offset-2",
      active
        ? "bg-primary/10 font-medium text-primary"
        : "text-muted-foreground hover:bg-muted hover:text-foreground"
    );

  return (
    <div className="flex flex-col gap-6">
      {showCategory && (
        <section aria-label="Filter by category">
          <h3 className="mb-2 text-sm font-semibold">Category</h3>
          <div className="flex flex-col">
            <Link
              href={navigate({ category: null, page: null })}
              onClick={onNavigate}
              aria-current={!current.category ? "true" : undefined}
              className={linkClassName(!current.category)}
            >
              All products
            </Link>
            {categories.map((category) => (
              <Link
                key={category.id}
                href={navigate({ category: category.slug, page: null })}
                onClick={onNavigate}
                aria-current={
                  current.category === category.slug ? "true" : undefined
                }
                className={linkClassName(current.category === category.slug)}
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

      <section aria-label="Filter by price">
        <h3 className="mb-2 text-sm font-semibold">Price</h3>
        <form onSubmit={handlePriceApply} className="flex flex-col gap-2">
          <div className="flex items-center gap-2">
            <Label htmlFor="filter-price-min" className="sr-only">
              Minimum price
            </Label>
            <Input
              id="filter-price-min"
              type="number"
              inputMode="numeric"
              min={0}
              placeholder="Min"
              value={minPrice}
              onChange={(event) => setMinPrice(event.target.value)}
              className="h-9"
            />
            <span aria-hidden="true" className="text-muted-foreground">
              –
            </span>
            <Label htmlFor="filter-price-max" className="sr-only">
              Maximum price
            </Label>
            <Input
              id="filter-price-max"
              type="number"
              inputMode="numeric"
              min={0}
              placeholder="Max"
              value={maxPrice}
              onChange={(event) => setMaxPrice(event.target.value)}
              className="h-9"
            />
          </div>
          <Button type="submit" variant="outline" size="sm" className="w-full">
            Apply price
          </Button>
        </form>
      </section>

      <section aria-label="Filter by rating">
        <h3 className="mb-2 text-sm font-semibold">Rating</h3>
        <div className="flex flex-col">
          {RATING_OPTIONS.map((rating) => (
            <Link
              key={rating}
              href={navigate({
                rating: current.rating === rating ? null : rating,
                page: null,
              })}
              onClick={onNavigate}
              aria-current={current.rating === rating ? "true" : undefined}
              className={linkClassName(current.rating === rating)}
            >
              <span className="flex" aria-hidden="true">
                {Array.from({ length: rating }).map((_, starIndex) => (
                  <Star
                    key={starIndex}
                    className="size-3.5 fill-amber-400 text-amber-400"
                  />
                ))}
              </span>
              {rating} &amp; up
            </Link>
          ))}
        </div>
      </section>

      <section aria-label="Availability">
        <h3 className="mb-2 text-sm font-semibold">Availability</h3>
        <div className="flex flex-col">
          <Link
            href={navigate({ stock: current.inStock ? null : "1", page: null })}
            onClick={onNavigate}
            aria-current={current.inStock ? "true" : undefined}
            className={linkClassName(current.inStock)}
          >
            In stock only
          </Link>
          <Link
            href={navigate({ sale: current.onSale ? null : "1", page: null })}
            onClick={onNavigate}
            aria-current={current.onSale ? "true" : undefined}
            className={linkClassName(current.onSale)}
          >
            On sale
          </Link>
        </div>
      </section>
    </div>
  );
}
