import type { Metadata } from "next";

import { PageHeader } from "@/components/dashboard/page-header";
import { ProductForm } from "@/components/vendor/product-form";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Add Product",
  description: "Create a product and submit it for admin review.",
  path: "/vendor/products/new",
  noIndex: true,
});

export default function NewVendorProductPage() {
  return (
    <>
      <PageHeader
        title="Add product"
        description="Save a draft, or submit for review — approved products go live."
      />

      <ProductForm />
    </>
  );
}
