import type { Metadata } from "next";

import { OrderSuccessView } from "@/components/checkout/order-success-view";
import { getMyOrder } from "@/lib/account-orders";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Order Confirmation",
  description: "Your ShopKart order has been placed successfully.",
  path: "/order/success",
  noIndex: true,
});

export default async function OrderSuccessPage({
  searchParams,
}: {
  searchParams: Promise<{ order?: string }>;
}) {
  await requireCurrentUser();
  const { order: orderId } = await searchParams;
  const order = orderId ? await getMyOrder(orderId) : null;

  return <OrderSuccessView order={order} />;
}
