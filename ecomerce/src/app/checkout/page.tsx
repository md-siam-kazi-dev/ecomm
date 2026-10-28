import type { Metadata } from "next";

import { CheckoutView } from "@/components/checkout/checkout-view";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Checkout",
  description:
    "Complete your ShopKart order — contact, shipping and delivery details.",
  path: "/checkout",
  noIndex: true,
});

export default async function CheckoutPage() {
  // Ordering requires an account (orders are tied to the signed-in user).
  await requireCurrentUser();
  return <CheckoutView />;
}
