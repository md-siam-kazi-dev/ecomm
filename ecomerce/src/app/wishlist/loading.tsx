import { ProductGridSkeleton } from "@/components/products/product-grid-skeleton";
import { Skeleton } from "@/components/ui/skeleton";

/** Route-level loading skeleton for /wishlist (PRD §35). */
export default function WishlistLoading() {
  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <Skeleton className="h-9 w-52" />
      <div className="mt-8">
        <ProductGridSkeleton count={4} />
      </div>
    </div>
  );
}
