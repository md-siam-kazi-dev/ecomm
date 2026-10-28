"use client";

import { Suspense } from "react";
import Link from "next/link";
import { Heart } from "lucide-react";

import { ProductGrid } from "@/components/products/product-grid";
import { ProductGridSkeleton } from "@/components/products/product-grid-skeleton";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useHasMounted } from "@/hooks/use-has-mounted";
import { useProductsByIds } from "@/hooks/use-products-by-ids";
import { useWishlistStore } from "@/store/wishlist-store";

function WishlistSkeleton() {
  return (
    <div aria-hidden="true" className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <Skeleton className="h-9 w-52" />
      <div className="mt-8">
        <ProductGridSkeleton count={4} />
      </div>
    </div>
  );
}

function WishlistProducts({ ids }: { ids: string[] }) {
  const products = useProductsByIds(ids);
  return <ProductGrid products={products} />;
}

function EmptyWishlist() {
  return (
    <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed px-6 py-20 text-center">
      <span className="flex size-16 items-center justify-center rounded-full bg-primary-light text-primary">
        <Heart className="size-8 fill-current" aria-hidden="true" />
      </span>
      <div>
        <h2 className="text-xl font-semibold">Your wishlist is empty</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Tap the heart on any product to save it for later.
        </p>
      </div>
      <Link href="/products" className={buttonVariants()}>
        Continue Shopping
      </Link>
    </div>
  );
}

interface WishlistViewProps {
  /**
   * Renders just the list without the page shell — used inside the account
   * layout (PRD §23), which already provides the page heading.
   */
  embedded?: boolean;
}

/** Wishlist content (PRD §22) — client, reads the persisted store. */
export function WishlistView({ embedded = false }: WishlistViewProps) {
  const hasMounted = useHasMounted();
  const ids = useWishlistStore((state) => state.items);

  if (!hasMounted) {
    return embedded ? <ProductGridSkeleton count={4} /> : <WishlistSkeleton />;
  }

  const content =
    ids.length === 0 ? (
      <EmptyWishlist />
    ) : (
      <Suspense
        fallback={<ProductGridSkeleton count={Math.min(4, ids.length)} />}
      >
        <WishlistProducts ids={ids} />
      </Suspense>
    );

  if (embedded) {
    return content;
  }

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          Your Wishlist
        </h1>
        {ids.length > 0 && (
          <p className="mt-2 text-sm text-muted-foreground">
            {ids.length} saved {ids.length === 1 ? "item" : "items"} — tap the
            heart on a product to remove it.
          </p>
        )}
      </header>

      {content}
    </div>
  );
}
