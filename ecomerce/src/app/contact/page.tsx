import type { Metadata } from "next";

import { ContactForm } from "@/components/content/contact-form";
import { ContentPage } from "@/components/content/content-page";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Contact Us",
  description:
    "Get in touch with the ShopKart team — questions about orders, returns, products or partnerships.",
  path: "/contact",
});

export default function ContactPage() {
  return (
    <ContentPage
      title="Contact Us"
      description="Questions about an order, a return or a product? Our support team is here to help."
      sections={[
        {
          heading: "Customer support",
          paragraphs: [
            "Email support@shopkart.example and we will get back to you within one business day. This demo store does not monitor real messages.",
            "For order-specific questions, please include your order number so we can help you faster.",
          ],
        },
        {
          heading: "Business hours",
          paragraphs: [
            "Monday to Friday, 9:00–18:00. Weekend and public holiday enquiries are answered on the next business day.",
          ],
        },
        {
          heading: "Returns",
          paragraphs: [
            "Prefer to start a return? Visit our Shipping & Returns page for step-by-step instructions and the 30-day window.",
          ],
        },
      ]}
    >
      <ContactForm />
    </ContentPage>
  );
}
