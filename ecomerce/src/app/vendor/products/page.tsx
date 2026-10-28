import type { Metadata } from "next";
import Link from "next/link";
import { Plus } from "lucide-react";

import { PageHeader } from "@/components/dashboard/page-header";
import { VendorProductsManager } from "@/components/vendor/vendor-products-manager";
import { Button } from "@/components/ui/button";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Products",
  description: "Manage your products and submit them for approval.",
  path: "/vendor/products",
  noIndex: true,
});

export default function VendorProductsPage() {
  return (
    <>
      <PageHeader
        title="Your products"
        description="Add, edit and submit products for admin approval."
        actions={
          <Button asChild>
            <Link href="/vendor/products/new">
              <Plus data-icon="inline-start" aria-hidden="true" />
              Add product
            </Link>
          </Button>
        }
      />

      <VendorProductsManager />
    </>
  );
}
