import type { Metadata } from "next";
import type { ReactNode } from "react";

import { AccountNav } from "@/components/account/account-nav";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Your Account",
  description:
    "Manage your ShopKart orders, addresses, wishlist and account settings.",
  path: "/account",
  noIndex: true,
});

/**
 * Account section shell (PRD §23): real Better Auth session header plus a
 * persistent navigation rail. Unauthenticated visitors are redirected to sign
 * in before the page renders.
 */
export default async function AccountLayout({
  children,
}: {
  children: ReactNode;
}) {
  const user = await requireCurrentUser();

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <div className="grid gap-6 lg:grid-cols-[220px_1fr] lg:gap-8">
        <aside className="flex flex-col gap-6">
          <header className="flex flex-col gap-1">
            <p className="text-sm text-muted-foreground">Signed in as</p>
            <h1 className="truncate text-lg font-semibold tracking-tight">
              {user.name}
            </h1>
            <p className="truncate text-sm text-muted-foreground">{user.email}</p>
          </header>
          <AccountNav />
        </aside>
        <div className="min-w-0">{children}</div>
      </div>
    </div>
  );
}
