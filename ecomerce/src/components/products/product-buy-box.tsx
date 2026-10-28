"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { RotateCcw, Truck } from "lucide-react";
import { toast } from "sonner";

import { AddToCartButton } from "@/components/products/add-to-cart-button";
import { ProductVariants } from "@/components/products/product-variants";
import { QuantitySelector } from "@/components/products/quantity-selector";
import { WishlistButton } from "@/components/products/wishlist-button";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  FREE_SHIPPING_THRESHOLD,
  LOW_STOCK_THRESHOLD,
} from "@/lib/constants";
import { calculateDiscountPercentage, formatPrice } from "@/lib/formatters";
import type { Product, ProductVariant } from "@/types/product";
import { useCartStore } from "@/store/cart-store";

interface ProductBuyBoxProps {
  product: Product;
}

const MAX_QUANTITY = 99;

/**
 * The interactive purchase panel (PRD §14/§16): variant-aware price and
 * stock, quantity, add-to-cart, buy-now, wishlist, shipping/return info
 * and SKU. Owns the variant + quantity selection state.
 */
export function ProductBuyBox({ product }: ProductBuyBoxProps) {
  const router = useRouter();
  const addItem = useCartStore((state) => state.addItem);

  const [selectedVariantId, setSelectedVariantId] = useState(() =>
    product.variants.find((variant) => variant.stock > 0)?.id
  );
  const [quantity, setQuantity] = useState(1);

  const activeVariant =
    product.variants.length > 0
      ? product.variants.find(
          (variant) => variant.id === selectedVariantId
        )
      : undefined;

  const price = activeVariant?.price ?? product.price;
  const compareAtPrice = activeVariant?.price
    ? undefined
    : product.compareAtPrice;
  const discount = calculateDiscountPercentage(price, compareAtPrice);
  const stock = activeVariant?.stock ?? product.stock;
  const outOfStock = stock <= 0;
  const lowStock = !outOfStock && stock <= LOW_STOCK_THRESHOLD;
  const sku = activeVariant?.sku ?? product.sku;

  // Group variants by axis — current data uses one axis per product.
  const variantGroups: Array<{ name: string; variants: ProductVariant[] }> = [];
  for (const variant of product.variants) {
    const group = variantGroups.find(
      (candidate) => candidate.name === variant.name
    );
    if (group) group.variants.push(variant);
    else variantGroups.push({ name: variant.name, variants: [variant] });
  }

  function handleVariantSelect(variantId: string) {
    setSelectedVariantId(variantId);
    const variant = product.variants.find(
      (candidate) => candidate.id === variantId
    );
    if (variant && quantity > variant.stock) {
      setQuantity(Math.max(1, Math.min(variant.stock, MAX_QUANTITY)));
    }
  }

  async function handleBuyNow() {
    if (outOfStock) return;
    const ok = await addItem({
      productId: product.id,
      variantId: activeVariant?.id,
      quantity,
    });
    if (!ok) {
      toast.error("Could not add to cart", { description: product.name });
      return;
    }
    router.push("/checkout");
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-center gap-3">
        <span className="text-3xl font-bold tracking-tight">
          {formatPrice(price)}
        </span>
        {compareAtPrice && (
          <span className="text-lg text-muted-foreground line-through">
            {formatPrice(compareAtPrice)}
          </span>
        )}
        {discount && <Badge variant="destructive">-{discount}%</Badge>}
      </div>

      <p aria-live="polite">
        {outOfStock ? (
          <span className="font-medium text-destructive">Out of stock</span>
        ) : lowStock ? (
          <span className="font-medium text-warning">
            Only {stock} left — order soon
          </span>
        ) : (
          <span className="flex items-center gap-1.5 font-medium text-success">
            <span aria-hidden="true" className="size-2 rounded-full bg-success" />
            In stock, ready to ship
          </span>
        )}
      </p>

      {variantGroups.map((group) => (
        <ProductVariants
          key={group.name}
          name={group.name}
          variants={group.variants}
          selectedId={selectedVariantId}
          onSelect={handleVariantSelect}
        />
      ))}

      <div className="flex flex-wrap items-center gap-3">
        <QuantitySelector
          value={quantity}
          onChange={setQuantity}
          max={Math.max(1, Math.min(stock, MAX_QUANTITY))}
        />
        <AddToCartButton
          product={product}
          variantId={activeVariant?.id}
          quantity={quantity}
          outOfStock={outOfStock}
          className="min-w-0 flex-1 sm:min-w-44"
        />
      </div>

      <Button
        type="button"
        variant="default"
        size="lg"
        className="h-11 w-full text-base"
        onClick={handleBuyNow}
        disabled={outOfStock}
      >
        Buy Now
      </Button>

      <WishlistButton
        productId={product.id}
        productName={product.name}
        withLabel
        variant="outline"
        className="w-full"
      />

      <div className="flex flex-col gap-3 rounded-xl border bg-secondary p-4 text-sm">
        <p className="flex items-start gap-2.5">
          <Truck className="mt-0.5 size-4 shrink-0 text-primary" aria-hidden="true" />
          <span>
            <strong className="font-medium">Free shipping</strong> on orders
            over {formatPrice(FREE_SHIPPING_THRESHOLD)} — ships in 1–2
            business days.
          </span>
        </p>
        <p className="flex items-start gap-2.5">
          <RotateCcw className="mt-0.5 size-4 shrink-0 text-primary" aria-hidden="true" />
          <span>
            <strong className="font-medium">30-day returns</strong> — send it
            back, no questions asked.
          </span>
        </p>
      </div>

      <p className="text-xs text-muted-foreground">
        SKU: <span className="font-mono">{sku}</span>
      </p>
    </div>
  );
}
