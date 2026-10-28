import Link from "next/link";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/** Store-level not-found (PRD §36) with recovery actions. */
export default function StoreNotFound() {
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col items-center gap-4 px-4 py-24 text-center">
      <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
        Store not found
      </h1>
      <p className="max-w-md text-sm text-muted-foreground sm:text-base">
        The store you are looking for does not exist or may no longer be active.
      </p>
      <div className="flex flex-wrap justify-center gap-3">
        <Link href="/products" className={cn(buttonVariants())}>
          Browse all products
        </Link>
        <Link href="/" className={cn(buttonVariants({ variant: "outline" }))}>
          Back to home
        </Link>
      </div>
    </div>
  );
}
