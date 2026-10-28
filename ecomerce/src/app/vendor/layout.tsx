import type { Metadata } from "next";
import { redirect } from "next/navigation";
import type { ReactNode } from "react";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";
import { getMyStores } from "@/lib/store-applications";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Dashboard",
  description: "Manage your products, orders and inventory.",
  path: "/vendor",
  noIndex: true,
});

/**
 * Vendor area shell (PRD §15): signed-in vendors only. Unauthenticated
 * visitors go to sign-in; authenticated non-vendors go to their account. A
 * vendor whose store is not active is sent to the store status page. The
 * backend remains the authority for vendor API access.
 */
export default async function VendorLayout({ children }: { children: ReactNode }) {
  const user = await requireCurrentUser();
  if (user.role !== "vendor") redirect("/account");

  const stores = await getMyStores();
  const hasActiveStore = (stores ?? []).some((store) => store.status === "active");
  if (!hasActiveStore) redirect("/account/store");

  return (
    <DashboardShell
      role="vendor"
      name={user.name}
      email={user.email}
      avatarUrl={user.avatarUrl}
    >
      {children}
    </DashboardShell>
  );
}
