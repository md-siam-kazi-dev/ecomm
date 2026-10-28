import Link from "next/link";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/** Category-level not-found (PRD §36) with a recovery action. */
export default function CategoryNotFound() {
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col items-center gap-4 px-4 py-24 text-center">
      <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
        Category not found
      </h1>
      <p className="max-w-md text-sm text-muted-foreground sm:text-base">
        The category you are looking for does not exist or may have moved.
      </p>
      <Link href="/products" className={cn(buttonVariants())}>
        Browse all products
      </Link>
    </div>
  );
}
