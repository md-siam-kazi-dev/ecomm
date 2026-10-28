import type { Metadata } from "next";

import { WishlistView } from "@/components/wishlist/wishlist-view";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Saved Items",
  description: "Products saved to your ShopKart wishlist.",
  path: "/account/wishlist",
  noIndex: true,
});

export default function AccountWishlistPage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold">Saved items</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Products you have saved for later.
        </p>
      </div>
      <WishlistView embedded />
    </div>
  );
}
