import type { Metadata } from "next";

import { ContentPage } from "@/components/content/content-page";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Frequently Asked Questions",
  description:
    "Answers to common ShopKart questions about orders, shipping, returns, payments and accounts.",
  path: "/faq",
});

const FAQS = [
  {
    question: "How long does delivery take?",
    answer:
      "Standard delivery takes 5–7 business days and express delivery 2–3 business days. You will see the estimate for your address at checkout.",
  },
  {
    question: "When is shipping free?",
    answer:
      "Standard shipping is free on orders over $75. Express delivery has a flat fee regardless of order value.",
  },
  {
    question: "How do I track my order?",
    answer:
      "Open Account → Orders and select an order to see its status and estimated arrival. Every order also appears on your order confirmation page.",
  },
  {
    question: "What is your return policy?",
    answer:
      "Most items can be returned within 30 days of delivery in their original condition. See the Shipping & Returns page for the full details.",
  },
  {
    question: "Which payment methods do you accept?",
    answer:
      "This demo store shows cash on delivery, bank transfer and digital wallet options. No real payment is processed and no card details are collected.",
  },
  {
    question: "Do I need an account to order?",
    answer:
      "This is a demo storefront, so checkout works without a real account. The Account area shows a sample signed-in customer for demonstration.",
  },
  {
    question: "Are the products and reviews real?",
    answer:
      "No. All products, prices, orders and reviews are realistic demo data used to showcase the storefront experience.",
  },
];

export default function FaqPage() {
  return (
    <ContentPage
      title="Frequently Asked Questions"
      description="Quick answers to the questions we hear most often."
    >
      <div className="mt-8 flex flex-col gap-3">
        {FAQS.map((item) => (
          <details
            key={item.question}
            className="rounded-xl border p-4 open:bg-secondary"
          >
            <summary className="cursor-pointer text-sm font-medium focus-visible:outline-2 focus-visible:outline-offset-2">
              {item.question}
            </summary>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
              {item.answer}
            </p>
          </details>
        ))}
      </div>
    </ContentPage>
  );
}
