import type { Metadata } from "next";

import { PageHeader } from "@/components/dashboard/page-header";
import { VendorOrdersTable } from "@/components/vendor/vendor-orders-table";
import { buildMetadata } from "@/lib/seo";
import { getVendorOrders } from "@/lib/vendor";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Orders",
  description: "Fulfil the orders that contain your products.",
  path: "/vendor/orders",
  noIndex: true,
});

export default async function VendorOrdersPage() {
  const orders = await getVendorOrders();

  return (
    <>
      <PageHeader
        title="Orders"
        description={`${orders.length} orders contain your products. Confirm, ship and deliver them.`}
      />

      <VendorOrdersTable orders={orders} />
    </>
  );
}
