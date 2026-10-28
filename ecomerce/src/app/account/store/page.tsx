import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, BadgeCheck, LineChart, Store, Wallet } from "lucide-react";

import { StoreStatusView } from "@/components/store/store-status-view";
import { Button } from "@/components/ui/button";
import {
  getApplicationDetail,
  getMyApplications,
  pickPrimaryApplication,
} from "@/lib/store-applications";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Your Store",
  description: "Apply for a store, track its review and manage activation.",
  path: "/account/store",
  noIndex: true,
});

const PERKS = [
  { icon: Store, title: "Open a storefront", text: "Apply with your existing customer account — no separate sign-up." },
  { icon: BadgeCheck, title: "Reviewed by admins", text: "Every application is reviewed before your store goes live." },
  { icon: LineChart, title: "Sell to customers", text: "List products, manage orders and grow once you're approved." },
  { icon: Wallet, title: "Simple activation", text: "Activate your store after approval and the application fee." },
];

export default async function AccountStorePage() {
  await requireCurrentUser();

  const applications = (await getMyApplications()) ?? [];
  const primary = pickPrimaryApplication(applications);
  const detail = primary ? await getApplicationDetail(primary.id) : null;

  if (!detail) {
    return (
      <div className="flex flex-col gap-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Start selling</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Apply for a store and get it in front of customers.
          </p>
        </div>

        <section className="rounded-2xl bg-background p-6 ring-1 ring-foreground/10 sm:p-8">
          <div className="grid gap-5 sm:grid-cols-2">
            {PERKS.map((perk) => (
              <div key={perk.title} className="flex gap-3">
                <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary-dark">
                  <perk.icon className="size-4" aria-hidden="true" />
                </span>
                <div>
                  <p className="font-medium">{perk.title}</p>
                  <p className="mt-0.5 text-sm text-muted-foreground">{perk.text}</p>
                </div>
              </div>
            ))}
          </div>
          <div className="mt-8">
            <Button asChild size="lg" className="h-11 text-base">
              <Link href="/account/store/apply">
                Create a store
                <ArrowRight className="size-4" aria-hidden="true" />
              </Link>
            </Button>
          </div>
        </section>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold tracking-tight">Your store</h1>
      <StoreStatusView application={detail} />
    </div>
  );
}
