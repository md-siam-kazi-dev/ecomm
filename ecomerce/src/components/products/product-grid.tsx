import { ProductCard } from "@/components/products/product-card";
import { cn } from "@/lib/utils";
import type { Product } from "@/types/product";

interface ProductGridProps {
  products: Product[];
  className?: string;
  /**
   * Number of leading cards to load eagerly (the LCP row). Defaults to 0 so
   * grids that are not the first thing on the page stay lazy.
   */
  priorityCount?: number;
}

/** Responsive product grid — 2 columns on mobile, up to 4 on desktop. */
export function ProductGrid({
  products,
  className,
  priorityCount = 0,
}: ProductGridProps) {
  return (
    <ul
      className={cn(
        "grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-3 lg:grid-cols-4",
        className
      )}
    >
      {products.map((product, index) => (
        <li key={product.id} className="flex">
          <ProductCard
            product={product}
            className="w-full"
            priority={index < priorityCount}
          />
        </li>
      ))}
    </ul>
  );
}
