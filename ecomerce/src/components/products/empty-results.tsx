import Link from "next/link";
import { SearchX } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface EmptyResultsProps {
  basePath: string;
  hasFilters: boolean;
}

/** Empty catalog state (PRD §37): explains what happened + recovery action. */
export function EmptyResults({ basePath, hasFilters }: EmptyResultsProps) {
  return (
    <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed px-6 py-16 text-center">
      <span className="flex size-14 items-center justify-center rounded-full bg-muted">
        <SearchX className="size-7 text-muted-foreground" aria-hidden="true" />
      </span>
      <div>
        <h2 className="text-lg font-semibold">No products found</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {hasFilters
            ? "Try adjusting or clearing your filters to see more products."
            : "Check back soon — new products are on the way."}
        </p>
      </div>
      {hasFilters && (
        <Link
          href={basePath}
          className={cn(buttonVariants({ variant: "outline" }))}
        >
          Clear all filters
        </Link>
      )}
    </div>
  );
}
