import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, BadgeCheck, LineChart, ShieldCheck, Store } from "lucide-react";

import { Button } from "@/components/ui/button";
import { buildMetadata } from "@/lib/seo";
import { getCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Become a Vendor",
  description: "Open your own store on ShopKart — apply from your account.",
  path: "/become-a-vendor",
});

const STEPS = [
  { title: "Create a store", text: "Fill in your store details from your existing account." },
  { title: "Admin review", text: "We check your application and share an approval letter." },
  { title: "Application fee", text: "Pay the application fee once your store is approved." },
  { title: "Go live", text: "Activate your store and open the vendor dashboard." },
];

export default async function BecomeAVendorPage() {
  const user = await getCurrentUser();
  const applyHref = user ? "/account/store" : "/auth/sign-up";
  const applyLabel = user ? "Create a store" : "Sign up to get started";

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-12 sm:px-6 lg:px-8">
      <section className="rounded-3xl border bg-gradient-to-b from-primary/5 to-background px-6 py-12 text-center sm:px-12 sm:py-16">
        <span className="inline-flex items-center gap-2 rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary-dark">
          <Store className="size-3.5" aria-hidden="true" />
          Sell on our marketplace
        </span>
        <h1 className="mx-auto mt-5 max-w-2xl text-3xl font-bold tracking-tight sm:text-4xl">
          Open your own store
        </h1>
        <p className="mx-auto mt-4 max-w-xl text-muted-foreground">
          Bring your products to customers across the country. Apply with your
          existing account — no separate seller sign-up.
        </p>
        <div className="mt-8 flex flex-wrap justify-center gap-3">
          <Button asChild size="lg" className="h-11 text-base">
            <Link href={applyHref}>
              {applyLabel}
              <ArrowRight className="size-4" aria-hidden="true" />
            </Link>
          </Button>
          {!user && (
            <Button asChild variant="outline" size="lg" className="h-11 text-base">
              <Link href="/auth/sign-in">Sign in</Link>
            </Button>
          )}
        </div>
      </section>

      <section className="mt-12">
        <h2 className="text-center text-xl font-semibold">How it works</h2>
        <ol className="mt-8 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {STEPS.map((step, index) => (
            <li key={step.title} className="rounded-2xl bg-background p-5 ring-1 ring-foreground/10">
              <span className="flex size-8 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary-dark">
                {index + 1}
              </span>
              <p className="mt-3 font-medium">{step.title}</p>
              <p className="mt-1 text-sm text-muted-foreground">{step.text}</p>
            </li>
          ))}
        </ol>
      </section>

      <section className="mt-12 grid gap-5 sm:grid-cols-3">
        {[
          { icon: ShieldCheck, title: "Trusted review", text: "Admins verify every store before it goes live." },
          { icon: BadgeCheck, title: "Approval letter", text: "Get a clear approval document you can keep." },
          { icon: LineChart, title: "Dashboard tools", text: "Manage products, orders and inventory in one place." },
        ].map((item) => (
          <div key={item.title} className="rounded-2xl bg-background p-5 ring-1 ring-foreground/10">
            <item.icon className="size-5 text-primary" aria-hidden="true" />
            <p className="mt-3 font-medium">{item.title}</p>
            <p className="mt-1 text-sm text-muted-foreground">{item.text}</p>
          </div>
        ))}
      </section>
    </div>
  );
}
