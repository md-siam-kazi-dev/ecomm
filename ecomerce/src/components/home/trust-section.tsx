import { Headset, RotateCcw, ShieldCheck, Truck } from "lucide-react";

import { FadeUp } from "@/components/animations";
import { FREE_SHIPPING_THRESHOLD } from "@/lib/constants";
import { formatPrice } from "@/lib/formatters";

const TRUST_ITEMS = [
  {
    icon: Truck,
    title: "Free Shipping",
    description: `Free standard shipping on orders over ${formatPrice(FREE_SHIPPING_THRESHOLD)}.`,
  },
  {
    icon: RotateCcw,
    title: "30-Day Returns",
    description: "Changed your mind? Send it back within 30 days, no questions asked.",
  },
  {
    icon: ShieldCheck,
    title: "Secure Checkout",
    description: "Encrypted payments and protected personal data, always.",
  },
  {
    icon: Headset,
    title: "24/7 Support",
    description: "Real humans ready to help — day, night and weekends.",
  },
];

/** "Why shop with us" trust signals (PRD §9). */
export function TrustSection() {
  return (
    <section className="border-y bg-secondary">
      <div className="mx-auto w-full max-w-7xl px-4 py-12 sm:px-6 sm:py-14 lg:px-8">
        <h2 className="sr-only">Why shop with us</h2>
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {TRUST_ITEMS.map((item, index) => (
            <FadeUp
              key={item.title}
              inView
              delay={index * 0.08}
              className="flex items-start gap-3.5"
            >
              <span className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <item.icon className="size-5" aria-hidden="true" />
              </span>
              <div>
                <h3 className="text-sm font-semibold">{item.title}</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  {item.description}
                </p>
              </div>
            </FadeUp>
          ))}
        </div>
      </div>
    </section>
  );
}
