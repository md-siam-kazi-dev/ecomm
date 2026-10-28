import type { Metadata } from "next";

import { PageHeader } from "@/components/dashboard/page-header";
import { VendorReviews } from "@/components/vendor/vendor-reviews";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Reviews",
  description: "Monitor ratings and reply to reviews on your products.",
  path: "/vendor/reviews",
  noIndex: true,
});

export default function VendorReviewsPage() {
  return (
    <>
      <PageHeader
        title="Reviews"
        description="Star feedback on your products. Reply to build trust."
      />

      <VendorReviews />
    </>
  );
}
