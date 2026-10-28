"use client";

import Image from "next/image";
import Link from "next/link";
import { Heart, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { QuantitySelector } from "@/components/products/quantity-selector";
import { Button } from "@/components/ui/button";
import { useCartLines } from "@/hooks/use-cart-lines";
import { formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import { useCartStore } from "@/store/cart-store";
import { useWishlistStore } from "@/store/wishlist-store";

interface CartLineItemsProps {
  /** Compact rows for the cart drawer. */
  compact?: boolean;
  /** Closes the drawer on product navigation. */
  onNavigate?: () => void;
  className?: string;
}

/**
 * Cart line items (PRD §17): image, name, variant, unit price, quantity
 * stepper clamped to stock, line total, remove and move-to-wishlist actions.
 * Lines come from the server cart; every action round-trips to the server and
 * the store updates optimistically in the meantime (PRD §50).
 */
export function CartLineItems({
  compact = false,
  onNavigate,
  className,
}: CartLineItemsProps) {
  const lines = useCartLines();
  const updateQuantity = useCartStore((state) => state.updateQuantity);
  const removeItem = useCartStore((state) => state.removeItem);
  const addToWishlist = useWishlistStore((state) => state.addToWishlist);

  return (
    <ul className={cn("flex flex-col", compact ? "gap-5" : "gap-6", className)}>
      {lines.map((line) => (
        <li key={line.id} className="flex gap-3 sm:gap-4">
          <Link
            href={`/products/${line.slug}`}
            onClick={onNavigate}
            className="relative size-20 shrink-0 overflow-hidden rounded-lg border focus-visible:outline-2 focus-visible:outline-offset-2 sm:size-24"
          >
            <Image
              src={line.imageUrl}
              alt={line.imageAlt}
              fill
              sizes="96px"
              className="object-cover"
            />
          </Link>

          <div className="flex min-w-0 flex-1 flex-col gap-1.5">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <Link
                  href={`/products/${line.slug}`}
                  onClick={onNavigate}
                  className="line-clamp-2 text-sm font-medium underline-offset-2 transition-colors hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
                >
                  {line.name}
                </Link>
                {line.variantLabel && (
                  <p className="text-xs text-muted-foreground">
                    {line.variantLabel}
                  </p>
                )}
                <p className="text-xs text-muted-foreground">
                  {formatPrice(line.unitPrice)} each
                </p>
                {!line.inStock && (
                  <p className="text-xs font-medium text-destructive">
                    Out of stock
                  </p>
                )}
              </div>
              <span className="text-sm font-semibold tabular-nums">
                {formatPrice(line.lineTotal)}
              </span>
            </div>

            <div className="mt-auto flex flex-wrap items-center gap-2">
              <QuantitySelector
                value={line.quantity}
                onChange={(value) => void updateQuantity(line.id, value)}
                max={line.maxQuantity}
                label={`Quantity for ${line.name}`}
              />
              {!compact && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-8 gap-1.5 px-2 text-xs text-muted-foreground hover:text-foreground"
                  onClick={() => {
                    addToWishlist(line.productId);
                    void removeItem(line.id);
                    toast.success("Moved to wishlist", {
                      description: line.name,
                    });
                  }}
                >
                  <Heart className="size-3.5" aria-hidden="true" />
                  Move to wishlist
                </Button>
              )}
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="ml-auto size-8 text-muted-foreground hover:text-destructive"
                aria-label={`Remove ${line.name} from cart`}
                onClick={() => {
                  void removeItem(line.id);
                  toast.success("Removed from cart", {
                    description: line.name,
                  });
                }}
              >
                <Trash2 className="size-4" aria-hidden="true" />
              </Button>
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}
