import type { Metadata } from "next";

import { PageHeader } from "@/components/dashboard/page-header";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty";
import { VendorRevenue } from "@/components/vendor/vendor-revenue";
import { buildMetadata } from "@/lib/seo";
import { getVendorRevenue } from "@/lib/vendor";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Revenue",
  description: "Your store revenue, trends and best-selling products.",
  path: "/vendor/revenue",
  noIndex: true,
});

export default async function VendorRevenuePage() {
  const revenue = await getVendorRevenue();

  return (
    <>
      <PageHeader
        title="Revenue"
        description="Your store's revenue and best-selling products."
      />

      {revenue ? (
        <VendorRevenue revenue={revenue} />
      ) : (
        <Empty className="border p-16">
          <EmptyHeader>
            <EmptyTitle>Revenue isn&apos;t available yet</EmptyTitle>
            <EmptyDescription>
              Once your products start selling, totals and trends appear here.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </>
  );
}
