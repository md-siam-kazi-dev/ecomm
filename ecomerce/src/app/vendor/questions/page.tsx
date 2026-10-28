import type { Metadata } from "next";

import { PageHeader } from "@/components/dashboard/page-header";
import { VendorQuestions } from "@/components/vendor/vendor-questions";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Questions",
  description: "Answer questions about your products.",
  path: "/vendor/questions",
  noIndex: true,
});

export default function VendorQuestionsPage() {
  return (
    <>
      <PageHeader
        title="Questions"
        description="Product questions from shoppers. Answer them to help buyers decide."
      />

      <VendorQuestions />
    </>
  );
}
