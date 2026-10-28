"use client";

import Link from "next/link";

import { ErrorState } from "@/components/errors/error-state";
import { buttonVariants } from "@/components/ui/button";

/** Root error boundary (PRD §54) — keeps the storefront chrome and offers
 * a retry plus a safe way back. */
export default function RouteError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="mx-auto w-full max-w-2xl px-4 py-20 sm:px-6">
      <ErrorState onRetry={reset} />
      <div className="mt-6 flex flex-wrap justify-center gap-3">
        <Link href="/" className={buttonVariants({ variant: "outline" })}>
          Back to home
        </Link>
        <Link href="/products" className={buttonVariants({ variant: "ghost" })}>
          Browse products
        </Link>
      </div>
    </div>
  );
}
