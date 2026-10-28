import type { Metadata } from "next";
import { Suspense } from "react";

import { AuditLogsManager } from "@/components/admin/audit-logs-manager";
import { PageHeader } from "@/components/dashboard/page-header";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Audit Logs",
  description: "Review the administrative audit trail.",
  path: "/admin/audit-logs",
  noIndex: true,
});

export default function AuditLogsPage() {
  return (
    <>
      <PageHeader
        title="Audit logs"
        description="An append-only record of admin, vendor and security activity. Read-only: events cannot be edited or deleted."
      />

      <Suspense
        fallback={
          <div role="status" aria-busy="true">
            <span className="sr-only">Loading audit logs</span>
            <TableSkeleton rows={8} columns={4} />
          </div>
        }
      >
        <AuditLogsManager />
      </Suspense>
    </>
  );
}
