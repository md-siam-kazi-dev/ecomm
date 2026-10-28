import type { Metadata } from "next";

import { ContentPage } from "@/components/content/content-page";
import { SITE_CONFIG } from "@/lib/constants";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Privacy Policy",
  description:
    "How ShopKart collects, uses and protects your information in this demo storefront.",
  path: "/privacy-policy",
});

export default function PrivacyPolicyPage() {
  return (
    <ContentPage
      title="Privacy Policy"
      description={`How ${SITE_CONFIG.name} handles information in this demo storefront.`}
      lastUpdated="September 2026"
      sections={[
        {
          heading: "Overview",
          paragraphs: [
            "This is a demonstration storefront. It does not run a production backend and does not transmit the information you enter to any server.",
          ],
        },
        {
          heading: "Information we store",
          paragraphs: [
            "Your cart and wishlist are saved in your browser's local storage so they survive a page refresh. Orders you place are kept in your browser session only.",
            "The address book and profile settings are held in memory for the duration of your visit and are not persisted.",
          ],
        },
        {
          heading: "Checkout details",
          paragraphs: [
            "The checkout collects contact and shipping details to demonstrate the flow. No card numbers, bank details or other financial information are ever collected or stored.",
          ],
        },
        {
          heading: "Analytics and cookies",
          paragraphs: [
            "This demo does not include third-party analytics or advertising trackers.",
          ],
        },
        {
          heading: "Your choices",
          paragraphs: [
            "You can clear the cart, wishlist and session orders at any time by clearing your browser storage for this site.",
          ],
        },
      ]}
    />
  );
}
