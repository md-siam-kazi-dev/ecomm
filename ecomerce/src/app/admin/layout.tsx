import type { Metadata } from "next";
import { redirect } from "next/navigation";
import type { ReactNode } from "react";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Admin",
  description: "ShopKart admin dashboard.",
  path: "/admin",
  noIndex: true,
});

/**
 * Admin area shell (PRD §24/§7): signed-in admins only. Unauthenticated
 * visitors are sent to sign-in; authenticated non-admins are sent to their
 * account. The backend remains the authority for admin API access.
 */
export default async function AdminLayout({ children }: { children: ReactNode }) {
  const user = await requireCurrentUser();
  if (user.role !== "admin") redirect("/account");

  return (
    <DashboardShell
      role="admin"
      name={user.name}
      email={user.email}
      avatarUrl={user.avatarUrl}
    >
      {children}
    </DashboardShell>
  );
}
