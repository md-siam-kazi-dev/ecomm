import type { Metadata } from "next";

import { ContentPage } from "@/components/content/content-page";
import { SITE_CONFIG } from "@/lib/constants";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Terms of Service",
  description:
    "The terms that govern your use of the ShopKart demo storefront.",
  path: "/terms-of-service",
});

export default function TermsOfServicePage() {
  return (
    <ContentPage
      title="Terms of Service"
      description={`The terms that govern your use of ${SITE_CONFIG.name}.`}
      lastUpdated="September 2026"
      sections={[
        {
          heading: "Demonstration store",
          paragraphs: [
            "This site is a demonstration e-commerce storefront. Products, prices, availability and reviews are illustrative only and no real transactions take place.",
          ],
        },
        {
          heading: "Orders and payment",
          paragraphs: [
            "Placing an order does not create a binding contract. No payment is collected and no physical goods are shipped.",
          ],
        },
        {
          heading: "Acceptable use",
          paragraphs: [
            "You agree not to misuse the site, attempt to disrupt it, or use it in a way that violates applicable law.",
          ],
        },
        {
          heading: "Intellectual property",
          paragraphs: [
            "Brand names, copy and imagery are used for demonstration purposes. Third-party marks belong to their respective owners.",
          ],
        },
        {
          heading: "Changes to these terms",
          paragraphs: [
            "These terms may be updated as the demonstration evolves. Continued use of the site constitutes acceptance of the current version.",
          ],
        },
      ]}
    />
  );
}
