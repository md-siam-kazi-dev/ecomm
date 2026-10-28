import type { Metadata } from "next";

import { PageHeader } from "@/components/dashboard/page-header";
import { ProductForm } from "@/components/vendor/product-form";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Edit Product",
  description: "Edit a product and resubmit it for review.",
  path: "/vendor/products",
  noIndex: true,
});

export default async function EditVendorProductPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <>
      <PageHeader
        title="Edit product"
        description="Fix anything the admin flagged, then resubmit for review."
      />

      <ProductForm productId={id} />
    </>
  );
}
