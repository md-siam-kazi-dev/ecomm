import type { Metadata } from "next";

import { ProductReviewQueue } from "@/components/admin/product-review-queue";
import { PageHeader } from "@/components/dashboard/page-header";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Admin Product Review",
  description: "Review vendor-submitted products and approve or reject them.",
  path: "/admin/products/review",
  noIndex: true,
});

export default function AdminProductReviewPage() {
  return (
    <>
      <PageHeader
        title="Product review"
        description="Approve, reject or request changes on vendor submissions. Only approved products are published."
      />

      <ProductReviewQueue />
    </>
  );
}
