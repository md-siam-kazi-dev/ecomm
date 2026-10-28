import type { Metadata } from "next";

import { OrderHistory } from "@/components/account/order-history";
import { getMyOrders } from "@/lib/account-orders";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Order History",
  description:
    "Track and review every order you have placed at ShopKart, with status and totals.",
  path: "/account/orders",
  noIndex: true,
});

export default async function AccountOrdersPage() {
  await requireCurrentUser();
  const orders = await getMyOrders();

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold">Order history</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Track and review everything you have ordered.
        </p>
      </div>
      <OrderHistory orders={orders} />
    </div>
  );
}
