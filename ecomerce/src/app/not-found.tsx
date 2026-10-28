import Link from "next/link";
import { Compass } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";

/** Global 404 (PRD §36/§54) — explains the problem and offers a way forward. */
export default function NotFound() {
  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col items-center gap-4 px-4 py-24 text-center sm:px-6">
      <span className="flex size-16 items-center justify-center rounded-full bg-primary-light text-primary">
        <Compass className="size-8" aria-hidden="true" />
      </span>
      <p className="text-sm font-medium text-primary-dark">404</p>
      <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
        Page not found
      </h1>
      <p className="max-w-md text-sm text-muted-foreground sm:text-base">
        The page you are looking for does not exist or may have moved.
      </p>
      <div className="mt-2 flex flex-wrap justify-center gap-3">
        <Link href="/" className={buttonVariants()}>
          Back to home
        </Link>
        <Link
          href="/products"
          className={buttonVariants({ variant: "outline" })}
        >
          Browse products
        </Link>
      </div>
    </div>
  );
}
