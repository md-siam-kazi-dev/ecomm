import type { Metadata } from "next";

import { ContentPage } from "@/components/content/content-page";
import { FREE_SHIPPING_THRESHOLD } from "@/lib/constants";
import { formatPrice } from "@/lib/formatters";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Shipping & Returns",
  description:
    "ShopKart shipping options, delivery times, free-shipping threshold and our 30-day return policy.",
  path: "/shipping-returns",
});

export default function ShippingReturnsPage() {
  return (
    <ContentPage
      title="Shipping & Returns"
      description="Everything you need to know about how we deliver your order and how returns work."
      sections={[
        {
          heading: "Shipping options",
          paragraphs: [
            `Standard delivery arrives in 5–7 business days and is free on orders over ${formatPrice(
              FREE_SHIPPING_THRESHOLD
            )}. Below that threshold a flat standard rate applies.`,
            "Express delivery arrives in 2–3 business days for a flat fee, regardless of order value.",
            "Delivery estimates are shown at checkout before you place your order.",
          ],
        },
        {
          heading: "Order tracking",
          paragraphs: [
            "Once your order ships you can follow its progress from your account. Open Account → Orders and select the order to see its current status and estimated arrival.",
          ],
        },
        {
          heading: "Returns and exchanges",
          paragraphs: [
            "You have 30 days from delivery to return most items in their original condition for a full refund.",
            "To start a return, contact support with your order number. We will share return instructions and a prepaid label where eligible.",
            "Refunds are issued to the original payment method once the returned item is received and inspected.",
          ],
        },
        {
          heading: "Items that cannot be returned",
          paragraphs: [
            "For hygiene and safety reasons, opened personal-care and perishable items cannot be returned. This is noted on the product page where it applies.",
          ],
        },
      ]}
    />
  );
}
