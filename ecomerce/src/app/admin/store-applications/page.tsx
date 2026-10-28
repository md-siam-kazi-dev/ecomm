import type { Metadata } from "next";
import { Suspense } from "react";

import { StoreApplicationsManager } from "@/components/admin/store-applications-manager";
import { PageHeader } from "@/components/dashboard/page-header";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Store Applications",
  description: "Review and process store applications.",
  path: "/admin/store-applications",
  noIndex: true,
});

export default function StoreApplicationsPage() {
  return (
    <>
      <PageHeader
        title="Store applications"
        description="Review seller applications and approve or reject them. Every filter is applied on the server."
      />

      <Suspense
        fallback={
          <div role="status" aria-busy="true">
            <span className="sr-only">Loading store applications</span>
            <TableSkeleton rows={8} columns={5} />
          </div>
        }
      >
        <StoreApplicationsManager />
      </Suspense>
    </>
  );
}
