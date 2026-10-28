import type { Metadata } from "next";

import { OrderDetailView } from "@/components/account/order-detail";
import { getMyOrder } from "@/lib/account-orders";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

interface AccountOrderPageProps {
  params: Promise<{ id: string }>;
}

export async function generateMetadata({
  params,
}: AccountOrderPageProps): Promise<Metadata> {
  const { id } = await params;
  const order = await getMyOrder(id);

  return buildMetadata({
    title: order ? `Order ${order.orderNumber}` : "Order details",
    description:
      "Review the items, totals, shipping and delivery details for your ShopKart order.",
    path: `/account/orders/${id}`,
    noIndex: true,
  });
}

export default async function AccountOrderPage({
  params,
}: AccountOrderPageProps) {
  const { id } = await params;
  const [user, order] = await Promise.all([
    requireCurrentUser(),
    getMyOrder(id),
  ]);
  const defaultAddress =
    user.addresses.find((address) => address.isDefault) ??
    user.addresses[0] ??
    null;

  return (
    <OrderDetailView
      orderId={id}
      order={order}
      defaultAddress={defaultAddress}
    />
  );
}
