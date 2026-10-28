"use client";

import { type ComponentProps } from "react";
import { Heart } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  selectIsWishlisted,
  useWishlistStore,
} from "@/store/wishlist-store";

interface WishlistButtonProps {
  productId: string;
  productName: string;
  className?: string;
  /** Renders a labeled button instead of the compact icon circle. */
  withLabel?: boolean;
  variant?: ComponentProps<typeof Button>["variant"];
}

/**
 * Heart toggle with persistence (Zustand) and toast feedback. Compact
 * circle on product cards, labeled button on the product page.
 */
export function WishlistButton({
  productId,
  productName,
  className,
  withLabel = false,
  variant = "secondary",
}: WishlistButtonProps) {
  const isWishlisted = useWishlistStore(selectIsWishlisted(productId));
  const toggleWishlist = useWishlistStore((state) => state.toggleWishlist);

  function handleToggle() {
    const wishlisted = toggleWishlist(productId);
    toast.success(
      wishlisted ? "Added to wishlist" : "Removed from wishlist",
      { description: productName }
    );
  }

  return (
    <Button
      type="button"
      variant={variant}
      size={withLabel ? "default" : "icon"}
      onClick={handleToggle}
      aria-pressed={isWishlisted}
      aria-label={
        isWishlisted
          ? `Remove ${productName} from wishlist`
          : `Add ${productName} to wishlist`
      }
      className={cn(
        withLabel ? "rounded-lg" : "size-10 rounded-full shadow-sm",
        className
      )}
    >
      <Heart
        className={cn(
          "size-4 transition-all duration-200",
          isWishlisted && "scale-110 fill-primary text-primary"
        )}
        aria-hidden="true"
      />
      {withLabel && (
        <span>{isWishlisted ? "Saved to wishlist" : "Add to wishlist"}</span>
      )}
    </Button>
  );
}
