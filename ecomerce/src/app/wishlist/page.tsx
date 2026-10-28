import type { Metadata } from "next";

import { WishlistView } from "@/components/wishlist/wishlist-view";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Your Wishlist",
  description:
    "Products you have saved for later at ShopKart — move them to your cart whenever you are ready.",
  path: "/wishlist",
  noIndex: true,
});

export default function WishlistPage() {
  return <WishlistView />;
}
