import type { Metadata } from "next";

import { OrdersTable } from "@/components/admin/orders-table";
import { PageHeader } from "@/components/dashboard/page-header";
import { getOrders } from "@/lib/orders";
import { buildMetadata } from "@/lib/seo";
import { getCustomers } from "@/lib/users";

export const metadata: Metadata = buildMetadata({
  title: "Admin Orders",
  description: "Review and manage ShopKart orders.",
  path: "/admin/orders",
  noIndex: true,
});

export default async function AdminOrdersPage() {
  const [orders, customers] = await Promise.all([getOrders(), getCustomers()]);

  return (
    <>
      <PageHeader
        title="Orders"
        description={`${orders.length} orders across all statuses.`}
      />

      <OrdersTable orders={orders} customers={customers} />
    </>
  );
}
