"use client";

import Link from "next/link";

import { CartLineItems } from "@/components/cart/cart-line-items";
import { CartSummary } from "@/components/cart/cart-summary";
import { EmptyCart } from "@/components/cart/empty-cart";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import {
  selectCartItemCount,
  selectCartSubtotal,
  useCartStore,
} from "@/store/cart-store";

/** Shown until the server cart has loaded, so an empty cart never flashes. */
function CartViewSkeleton() {
  return (
    <div aria-hidden="true" className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <Skeleton className="h-9 w-40" />
      <div className="mt-8 grid gap-10 lg:grid-cols-[1fr_360px]">
        <div className="flex flex-col gap-6">
          {Array.from({ length: 3 }).map((_, index) => (
            <div key={index} className="flex gap-4">
              <Skeleton className="size-24 rounded-lg" />
              <div className="flex flex-1 flex-col gap-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/3" />
                <Skeleton className="h-8 w-44" />
              </div>
            </div>
          ))}
        </div>
        <Skeleton className="h-72 rounded-2xl" />
      </div>
    </div>
  );
}

/** Cart page content (PRD §17) — client, reads the server-backed cart store. */
export function CartView() {
  const itemCount = useCartStore(selectCartItemCount);
  const subtotal = useCartStore(selectCartSubtotal);
  const status = useCartStore((state) => state.status);

  if (itemCount === 0 && (status === "idle" || status === "loading")) {
    return <CartViewSkeleton />;
  }

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          Your Cart
        </h1>
        {itemCount > 0 && (
          <p className="mt-2 text-sm text-muted-foreground">
            {itemCount} {itemCount === 1 ? "item" : "items"} in your cart
          </p>
        )}
      </header>

      {itemCount === 0 ? (
        <EmptyCart />
      ) : (
        <div className="grid gap-10 lg:grid-cols-[1fr_360px]">
          <section aria-label="Cart items">
            <CartLineItems />
          </section>

          <aside>
            <div className="rounded-2xl border bg-secondary p-6 lg:sticky lg:top-24">
              <h2 className="text-lg font-semibold">Order Summary</h2>
              <div className="mt-4">
                <CartSummary subtotal={subtotal} />
              </div>
              <Link
                href="/checkout"
                className={cn(
                  buttonVariants({ variant: "default", size: "lg" }),
                  "mt-5 h-11 w-full text-base"
                )}
              >
                Proceed to Checkout
              </Link>
              <Link
                href="/products"
                className="mt-3 block rounded-md text-center text-sm text-muted-foreground underline-offset-2 transition-colors hover:text-foreground hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
              >
                Continue shopping
              </Link>
            </div>
          </aside>
        </div>
      )}
    </div>
  );
}
