import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { StoreApplicationForm } from "@/components/store/store-application-form";
import { getMyApplications } from "@/lib/store-applications";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Create a Store",
  description: "Apply to open a store on the marketplace.",
  path: "/account/store/apply",
  noIndex: true,
});

export default async function ApplyPage() {
  await requireCurrentUser();

  // One active application at a time.
  const applications = (await getMyApplications()) ?? [];
  const active = applications.find(
    (a) => a.status === "pending" || a.status === "approved"
  );
  if (active) redirect("/account/store");

  return (
    <div className="flex flex-col gap-6">
      <div>
        <nav aria-label="Breadcrumb" className="text-sm text-muted-foreground">
          <Link href="/account/store" className="hover:text-foreground">
            Your store
          </Link>{" "}
          / Create a store
        </nav>
        <h1 className="mt-2 text-2xl font-bold tracking-tight">Create a store</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Tell us about your business. An administrator will review your
          application before your store goes live.
        </p>
      </div>
      <StoreApplicationForm />
    </div>
  );
}
