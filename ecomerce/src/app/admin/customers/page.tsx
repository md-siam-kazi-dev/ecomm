import type { Metadata } from "next";
import { Suspense } from "react";

import { UsersManager } from "@/components/admin/users-manager";
import { PageHeader } from "@/components/dashboard/page-header";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Admin Customers",
  description: "Search, filter and manage ShopKart users.",
  path: "/admin/customers",
  noIndex: true,
});

export default function AdminCustomersPage() {
  return (
    <>
      <PageHeader
        title="User management"
        description="Search, filter and page through accounts. Every filter is applied on the server."
      />

      <Suspense
        fallback={
          <div role="status" aria-busy="true">
            <span className="sr-only">Loading users</span>
            <TableSkeleton rows={8} columns={5} />
          </div>
        }
      >
        <UsersManager />
      </Suspense>
    </>
  );
}
