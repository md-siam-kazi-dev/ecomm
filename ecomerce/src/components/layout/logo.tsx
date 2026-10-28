import Link from "next/link";
import { ShoppingBag } from "lucide-react";

import { SITE_CONFIG } from "@/lib/constants";
import { cn } from "@/lib/utils";

interface LogoProps {
  className?: string;
}

/**
 * The brand mark on its own, with no link of its own. Use this inside another
 * link or button — nesting the full `Logo` in one produces invalid nested
 * anchors and a doubled-up brand lockup.
 */
export function LogoMark({ className }: LogoProps) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground",
        className
      )}
    >
      <ShoppingBag className="size-4" />
    </span>
  );
}

export function Logo({ className }: LogoProps) {
  return (
    <Link
      href="/"
      aria-label={`${SITE_CONFIG.name} — home`}
      className={cn(
        "inline-flex items-center gap-2 rounded-md focus-visible:outline-2 focus-visible:outline-offset-4",
        className
      )}
    >
      <LogoMark />
      <span className="text-lg font-semibold tracking-tight text-foreground">
        {SITE_CONFIG.name}
      </span>
    </Link>
  );
}
