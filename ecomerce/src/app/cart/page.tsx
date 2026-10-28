import type { Metadata } from "next";

import { CartView } from "@/components/cart/cart-view";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Your Cart",
  description:
    "Review the items in your ShopKart shopping cart, adjust quantities and proceed to checkout.",
  path: "/cart",
  noIndex: true,
});

export default function CartPage() {
  return <CartView />;
}
