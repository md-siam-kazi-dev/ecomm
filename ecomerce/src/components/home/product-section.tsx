import type { ReactNode } from "react";
import { ArrowRight } from "lucide-react";
import Link from "next/link";

import { FadeUp } from "@/components/animations";
import { cn } from "@/lib/utils";

interface ProductSectionProps {
  title: string;
  description?: string;
  /** Optional "view all" target, e.g. /products. */
  href?: string;
  linkLabel?: string;
  children: ReactNode;
  className?: string;
}

/** Shared section shell for product rows: heading, copy and a view-all link. */
export function ProductSection({
  title,
  description,
  href,
  linkLabel = "View all",
  children,
  className,
}: ProductSectionProps) {
  return (
    <section
      className={cn(
        "mx-auto w-full max-w-7xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8",
        className
      )}
    >
      <FadeUp inView className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            {title}
          </h2>
          {description && (
            <p className="mt-2 max-w-xl text-sm text-muted-foreground sm:text-base">
              {description}
            </p>
          )}
        </div>
        {href && (
          <Link
            href={href}
            className="inline-flex items-center gap-1.5 rounded-md text-sm font-medium text-primary transition-colors hover:text-primary-dark focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            {linkLabel}
            <ArrowRight className="size-4" aria-hidden="true" />
          </Link>
        )}
      </FadeUp>
      {children}
    </section>
  );
}
