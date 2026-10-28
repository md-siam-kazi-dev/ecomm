import type { Metadata } from "next";

import { AccountOverview } from "@/components/account/account-overview";
import { getMyOrders } from "@/lib/account-orders";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Account Overview",
  description:
    "Your ShopKart account at a glance — recent orders, spend, wishlist and default address.",
  path: "/account",
  noIndex: true,
});

export default async function AccountPage() {
  const user = await requireCurrentUser();
  const orders = await getMyOrders();

  return <AccountOverview user={user} orders={orders} />;
}
