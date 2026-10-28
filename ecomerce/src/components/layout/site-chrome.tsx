"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";

interface SiteChromeProps {
  /** Server-rendered chrome passed as props so it stays server-rendered. */
  announcement: ReactNode;
  header: ReactNode;
  footer: ReactNode;
  children: ReactNode;
}

/**
 * Applies the storefront chrome (announcement, header, footer) to public
 * routes and drops it for the admin and vendor areas, which ship their own
 * shells (PRD §15/§24). Those areas render their own `<main id="main-content">`
 * so the skip link still works.
 */
export function SiteChrome({
  announcement,
  header,
  footer,
  children,
}: SiteChromeProps) {
  const pathname = usePathname();
  const isAppArea =
    pathname.startsWith("/admin") || pathname.startsWith("/vendor");

  if (isAppArea) {
    return <>{children}</>;
  }

  return (
    <>
      {announcement}
      {header}
      <main id="main-content" className="flex-1">
        {children}
      </main>
      {footer}
    </>
  );
}
