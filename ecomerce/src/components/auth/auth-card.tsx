import Link from "next/link";
import type { ReactNode } from "react";

import { SITE_CONFIG } from "@/lib/constants";

interface AuthCardProps {
  title: string;
  description: string;
  children: ReactNode;
  /** Prompt below the card, e.g. a link to the other auth screen. */
  footer?: ReactNode;
}

/** Centered card shell shared by the sign-in and sign-up pages. */
export function AuthCard({ title, description, children, footer }: AuthCardProps) {
  return (
    <div className="mx-auto flex w-full max-w-md flex-col px-4 py-14 sm:py-20">
      <Link
        href="/"
        className="mx-auto mb-8 rounded-md text-lg font-semibold tracking-tight focus-visible:outline-2 focus-visible:outline-offset-4"
      >
        {SITE_CONFIG.name}
      </Link>

      <div className="rounded-2xl border bg-card p-6 shadow-sm sm:p-8">
        <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
        <p className="mt-1.5 text-sm text-muted-foreground">{description}</p>
        <div className="mt-6">{children}</div>
      </div>

      {footer && (
        <p className="mt-6 text-center text-sm text-muted-foreground">
          {footer}
        </p>
      )}
    </div>
  );
}
