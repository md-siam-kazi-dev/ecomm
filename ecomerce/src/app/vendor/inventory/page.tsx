import type { Metadata } from "next";

import { PageHeader } from "@/components/dashboard/page-header";
import { VendorInventoryManager } from "@/components/vendor/vendor-inventory-manager";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Inventory",
  description: "View and adjust stock for your products.",
  path: "/vendor/inventory",
  noIndex: true,
});

export default function VendorInventoryPage() {
  return (
    <>
      <PageHeader
        title="Inventory"
        description="Stock levels for your products. Adjust quantities as they change."
      />

      <VendorInventoryManager />
    </>
  );
}
