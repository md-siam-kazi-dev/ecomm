"use client";

import Link from "next/link";

import { CartLineItems } from "@/components/cart/cart-line-items";
import { CartSummary } from "@/components/cart/cart-summary";
import { EmptyCart } from "@/components/cart/empty-cart";
import { buttonVariants } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import {
  FREE_SHIPPING_THRESHOLD,
  STANDARD_SHIPPING_RATE,
} from "@/lib/constants";
import { formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import {
  selectCartItemCount,
  selectCartSubtotal,
  useCartStore,
} from "@/store/cart-store";
import { useUiStore } from "@/store/ui-store";

/**
 * Global cart drawer (PRD §11/§16) — opened from the header cart icon.
 * Line data is reconstructed through the repository; all actions update
 * the persisted cart store immediately.
 */
export function CartDrawer() {
  const isOpen = useUiStore((state) => state.isCartOpen);
  const closeCart = useUiStore((state) => state.closeCart);
  const itemCount = useCartStore(selectCartItemCount);
  const subtotal = useCartStore(selectCartSubtotal);
  const status = useCartStore((state) => state.status);
  const isLoading = itemCount === 0 && (status === "idle" || status === "loading");

  const shipping =
    subtotal > 0 && subtotal < FREE_SHIPPING_THRESHOLD
      ? STANDARD_SHIPPING_RATE
      : 0;
  const total = subtotal + shipping;

  return (
    <Sheet
      open={isOpen}
      onOpenChange={(open) => {
        if (!open) closeCart();
      }}
    >
      <SheetContent side="right" className="flex w-full flex-col gap-0 p-0 sm:max-w-md">
        <SheetHeader className="border-b p-4">
          <SheetTitle>
            Your cart
            {itemCount > 0 &&
              ` (${itemCount} ${itemCount === 1 ? "item" : "items"})`}
          </SheetTitle>
          <SheetDescription className="sr-only">
            Review the items in your shopping cart.
          </SheetDescription>
        </SheetHeader>

        {isLoading ? (
          <div className="flex flex-1 flex-col gap-6 p-4" aria-hidden="true">
            {Array.from({ length: 3 }).map((_, index) => (
              <div key={index} className="flex gap-4">
                <Skeleton className="size-20 shrink-0 rounded-lg" />
                <div className="flex flex-1 flex-col gap-2">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/3" />
                  <Skeleton className="h-8 w-40" />
                </div>
              </div>
            ))}
          </div>
        ) : itemCount === 0 ? (
          <EmptyCart compact onNavigate={closeCart} />
        ) : (
          <>
            <div className="flex-1 overflow-y-auto p-4">
              <CartLineItems compact onNavigate={closeCart} />
            </div>

            <div className="flex flex-col gap-4 border-t p-4">
              <CartSummary subtotal={subtotal} />
              <div className="flex flex-col gap-2">
                <Link
                  href="/checkout"
                  onClick={closeCart}
                  className={cn(buttonVariants({ variant: "default", size: "lg" }), "h-11 w-full text-base")}
                >
                  Checkout · {formatPrice(total)}
                </Link>
                <Link
                  href="/cart"
                  onClick={closeCart}
                  className={cn(buttonVariants({ variant: "outline" }), "w-full")}
                >
                  View cart
                </Link>
              </div>
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  );
}
