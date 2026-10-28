import type { Metadata } from "next";

import { AddressBook } from "@/components/account/address-book";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Addresses",
  description: "Manage the shipping addresses saved to your ShopKart account.",
  path: "/account/addresses",
  noIndex: true,
});

export default async function AccountAddressesPage() {
  // Gate the route; the address list itself is loaded from the API.
  await requireCurrentUser();

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold">Addresses</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Save as many delivery addresses as you need — division, district,
          upazila, village or town.
        </p>
      </div>
      <AddressBook />
    </div>
  );
}
