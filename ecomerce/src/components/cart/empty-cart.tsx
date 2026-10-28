import Link from "next/link";
import { ShoppingBag } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface EmptyCartProps {
  /** Compact layout for the cart drawer. */
  compact?: boolean;
  /** Closes the drawer on navigation. */
  onNavigate?: () => void;
}

/** Empty cart state (PRD §17): explains what happened + recovery action. */
export function EmptyCart({ compact = false, onNavigate }: EmptyCartProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center gap-4 text-center",
        compact ? "py-10" : "py-20"
      )}
    >
      <span className="flex size-16 items-center justify-center rounded-full bg-muted">
        <ShoppingBag className="size-8 text-muted-foreground" aria-hidden="true" />
      </span>
      <div>
        <h2 className={cn("font-semibold", compact ? "text-lg" : "text-xl")}>
          Your cart is empty
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Looks like you have not added anything yet.
        </p>
      </div>
      <Link
        href="/products"
        onClick={onNavigate}
        className={cn(buttonVariants())}
      >
        Continue Shopping
      </Link>
    </div>
  );
}
