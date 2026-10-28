import Image from "next/image";
import Link from "next/link";

import { AddToCartButton } from "@/components/products/add-to-cart-button";
import { PriceTag } from "@/components/products/price-tag";
import { StarRating } from "@/components/products/star-rating";
import { WishlistButton } from "@/components/products/wishlist-button";
import { Badge } from "@/components/ui/badge";
import { LOW_STOCK_THRESHOLD } from "@/lib/constants";
import { calculateDiscountPercentage } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { Product } from "@/types/product";

interface ProductCardProps {
  product: Product;
  className?: string;
  /**
   * Load the thumbnail eagerly. Set on the first row of a grid so the
   * above-the-fold product image is not lazy-loaded (PRD §34, LCP).
   */
  priority?: boolean;
}

/**
 * The storefront's core reusable card (PRD §13): image, badges, wishlist,
 * rating, price, discount and quick add-to-cart. Hover lifts the card 2px
 * and scales the image slightly; the title link stretches across the whole
 * card (single accessible link) with action buttons layered above it.
 */
export function ProductCard({ product, className, priority }: ProductCardProps) {
  const outOfStock = product.stock <= 0;
  const lowStock = !outOfStock && product.stock <= LOW_STOCK_THRESHOLD;
  const discount = calculateDiscountPercentage(
    product.price,
    product.compareAtPrice
  );
  const hasThumbnail = product.thumbnail.url.length > 0;

  // Cap at two badges — sale first, then new/best-seller (PRD §13).
  const badges: Array<{
    label: string;
    variant: "default" | "secondary" | "destructive";
  }> = [];
  if (outOfStock) {
    badges.push({ label: "Out of Stock", variant: "secondary" });
  } else {
    // Sale tag uses the warning/sale accent.
    if (discount) badges.push({ label: `-${discount}%`, variant: "destructive" });
    if (product.isNew) badges.push({ label: "New", variant: "secondary" });
    else if (product.isBestSeller)
      badges.push({ label: "Best Seller", variant: "secondary" });
  }

  return (
    <article
      className={cn(
        "group relative flex h-full flex-col overflow-hidden rounded-xl border bg-card transition-all duration-300 hover:-translate-y-0.5 hover:shadow-md has-focus-visible:ring-2 has-focus-visible:ring-ring/50",
        className
      )}
    >
      <div className="relative aspect-square overflow-hidden bg-muted">
        {hasThumbnail ? (
          <Image
            src={product.thumbnail.url}
            alt={product.thumbnail.alt}
            fill
            priority={priority}
            sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
            className={cn(
              "object-cover transition-transform duration-300 group-hover:scale-[1.04]",
              outOfStock && "opacity-70"
            )}
          />
        ) : (
          // Neutral placeholder — `next/image` cannot render an empty src.
          <div aria-hidden="true" className="h-full w-full bg-muted" />
        )}
        {badges.length > 0 && (
          <div className="absolute left-2 top-2 z-[2] flex flex-col items-start gap-1">
            {badges.map((badge) => (
              <Badge key={badge.label} variant={badge.variant} className="shadow-sm">
                {badge.label}
              </Badge>
            ))}
          </div>
        )}
        <WishlistButton
          productId={product.id}
          productName={product.name}
          className="absolute right-2 top-2 z-[2] opacity-0 transition-opacity duration-200 focus-visible:opacity-100 group-hover:opacity-100 max-md:opacity-100"
        />
      </div>

      <div className="flex flex-1 flex-col gap-1.5 p-3 sm:gap-2 sm:p-4">
        <p className="text-xs text-muted-foreground">{product.brand}</p>
        <h3 className="line-clamp-2 text-sm font-medium leading-snug">
          <Link
            href={`/products/${product.slug}`}
            className="after:absolute after:inset-0 after:z-[1] focus-visible:outline-none"
          >
            {product.name}
          </Link>
        </h3>
        <StarRating rating={product.rating} reviewCount={product.reviewCount} />
        <PriceTag price={product.price} compareAtPrice={product.compareAtPrice} />
        {lowStock && (
          <p className="text-xs font-medium text-warning">
            Only {product.stock} left in stock
          </p>
        )}
        <AddToCartButton product={product} size="sm" className="mt-auto w-full" />
      </div>
    </article>
  );
}
