"use client";

import { useState, type ComponentProps } from "react";
import { Check, Loader2, ShoppingCart } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import type { Product } from "@/types/product";
import { useCartStore } from "@/store/cart-store";

type AddToCartStatus = "idle" | "adding" | "added";

interface AddToCartButtonProps extends ComponentProps<typeof Button> {
  product: Product;
  /** Explicit variant (buy box). Omit for quick-add: first in-stock variant. */
  variantId?: string;
  /** Line quantity. Defaults to 1. */
  quantity?: number;
  /** Stock-availability override — buy box passes variant-level stock. */
  outOfStock?: boolean;
}

const ADDED_FEEDBACK_MS = 1600;

/**
 * Add-to-cart with immediate feedback (PRD §16): idle → spinner → "Added ✓",
 * plus a success toast. The item is persisted to the server cart; the header
 * badge count updates from the returned cart.
 *
 * Quick-add behavior: for variant products the first in-stock variant is
 * added; the product page offers full variant selection. When no variant is
 * given the server resolves the product's default variant.
 */
export function AddToCartButton({
  product,
  variantId,
  quantity = 1,
  outOfStock: outOfStockOverride,
  className,
  ...buttonProps
}: AddToCartButtonProps) {
  const [status, setStatus] = useState<AddToCartStatus>("idle");
  const addItem = useCartStore((state) => state.addItem);

  const variant =
    variantId !== undefined
      ? product.variants.find((candidate) => candidate.id === variantId)
      : product.variants.find((candidate) => candidate.stock > 0);
  const outOfStock = outOfStockOverride ?? product.stock <= 0;

  async function handleAdd() {
    if (outOfStock || status !== "idle") return;

    setStatus("adding");

    const ok = await addItem({
      productId: product.id,
      variantId: variant?.id,
      quantity,
    });

    if (!ok) {
      setStatus("idle");
      toast.error("Could not add to cart", { description: product.name });
      return;
    }

    setStatus("added");
    toast.success("Added to cart", { description: product.name });
    window.setTimeout(() => setStatus("idle"), ADDED_FEEDBACK_MS);
  }

  return (
    <Button
      type="button"
      variant="default"
      onClick={handleAdd}
      disabled={outOfStock || status !== "idle"}
      aria-live="polite"
      className={className}
      {...buttonProps}
    >
      {status === "adding" ? (
        <Loader2 className="size-4 animate-spin" aria-hidden="true" />
      ) : status === "added" ? (
        <Check className="size-4" aria-hidden="true" />
      ) : (
        <ShoppingCart className="size-4" aria-hidden="true" />
      )}
      {outOfStock
        ? "Out of Stock"
        : status === "added"
          ? "Added!"
          : "Add to Cart"}
    </Button>
  );
}
