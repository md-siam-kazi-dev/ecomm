import Link from "next/link";
import type { ReactNode } from "react";
import { FileText, Info, Store } from "lucide-react";

import { ApplicationStatusBadge, PaymentStatusBadge } from "@/components/store/application-status-badge";
import { ApplicationTimeline } from "@/components/store/application-timeline";
import { PaymentAction } from "@/components/store/payment-action";
import { Button } from "@/components/ui/button";
import { BUSINESS_TYPE_LABELS } from "@/lib/store-application-labels";
import { formatDate } from "@/lib/formatters";
import type { ApplicationDetail } from "@/types/store-application";

function money(amount: string, currency: string) {
  return `${currency} ${amount}`;
}

/** Status-driven applicant view: letter, fee, dashboard entry, rejection. */
export function StoreStatusView({
  application,
}: {
  application: ApplicationDetail;
}) {
  const { next_action: nextAction, store, fee } = application;
  const letterUrl = `/api/v1/store-applications/${application.id}/approval-letter`;

  return (
    <div className="flex flex-col gap-6">
      <section className="rounded-2xl bg-background p-6 ring-1 ring-foreground/10">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Application {application.application_number}
            </p>
            <h2 className="mt-1 truncate text-xl font-bold tracking-tight">
              {application.store_name}
            </h2>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {application.category} · submitted {formatDate(application.submitted_at)}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <ApplicationStatusBadge status={application.status} />
            {application.payment && <PaymentStatusBadge status={application.payment.status} />}
          </div>
        </div>

        <div className="mt-6">
          <ApplicationTimeline application={application} />
        </div>
      </section>

      {nextAction === "await_review" && (
        <Notice tone="info" title="Your store application is under review">
          An administrator will review your application shortly. You&apos;ll be
          able to pay the application fee and open your dashboard once it&apos;s
          approved.
        </Notice>
      )}

      {application.status === "approved" && (
        <section className="rounded-2xl bg-background p-6 ring-1 ring-foreground/10">
          <div className="flex items-center gap-2">
            <FileText className="size-4 text-primary" aria-hidden="true" />
            <h3 className="font-semibold">Your store application has been approved</h3>
          </div>
          <p className="mt-2 text-sm text-muted-foreground">
            {application.approved_at
              ? `Approved on ${formatDate(application.approved_at)}.`
              : "Your approval letter is ready."}{" "}
            {fee?.enabled
              ? "Complete the application fee to activate your store."
              : "Your store is eligible for activation."}
          </p>
          <div className="mt-4 flex flex-wrap gap-3">
            <Button variant="outline" asChild>
              <Link href={letterUrl} target="_blank" rel="noopener noreferrer">
                <FileText className="size-4" aria-hidden="true" />
                View approval letter
              </Link>
            </Button>
          </div>

          {fee?.enabled && (
            <div className="mt-5 border-t pt-5">
              <dl className="mb-4 grid grid-cols-2 gap-2 text-sm">
                <dt className="text-muted-foreground">Application fee</dt>
                <dd className="text-right font-medium">
                  {money(fee.amount, fee.currency)}
                </dd>
              </dl>
              <PaymentAction application={application} />
            </div>
          )}
        </section>
      )}

      {nextAction === "open_dashboard" && store && (
        <section className="rounded-2xl bg-success/10 p-6 ring-1 ring-success/20">
          <div className="flex items-center gap-2">
            <Store className="size-4 text-success" aria-hidden="true" />
            <h3 className="font-semibold">Your store is active</h3>
          </div>
          <p className="mt-2 text-sm text-muted-foreground">
            Head to your vendor dashboard to add products, manage orders and grow
            your store.
          </p>
          <div className="mt-4">
            <Button asChild className="h-11 text-base">
              <Link href="/vendor">Open vendor dashboard</Link>
            </Button>
          </div>
        </section>
      )}

      {nextAction === "suspended" && (
        <Notice tone="danger" title="This store is suspended">
          Your store is currently suspended, so the vendor dashboard is
          unavailable. Please contact support if you believe this is a mistake.
        </Notice>
      )}

      {nextAction === "rejected" && (
        <Notice tone="danger" title="Your store application was not approved">
          {application.rejection_reason ? (
            <p className="mt-1 whitespace-pre-line">{application.rejection_reason}</p>
          ) : (
            <p className="mt-1">The application did not meet our requirements.</p>
          )}
          <div className="mt-4">
            <Button variant="outline" asChild>
              <Link href="/account/store/apply">Submit a new application</Link>
            </Button>
          </div>
        </Notice>
      )}

      <section className="rounded-2xl bg-background p-6 ring-1 ring-foreground/10">
        <h3 className="font-semibold">Application details</h3>
        <dl className="mt-4 grid gap-x-6 gap-y-3 text-sm sm:grid-cols-2">
          <Detail label="Business email" value={application.business_email} />
          <Detail label="Contact phone" value={application.contact_phone} />
          <Detail
            label="Business type"
            value={BUSINESS_TYPE_LABELS[application.business_type] ?? application.business_type}
          />
          {application.registered_business_name && (
            <Detail label="Registered name" value={application.registered_business_name} />
          )}
          <Detail
            label="Address"
            value={[
              application.full_address,
              application.district,
              application.division,
              application.country,
            ]
              .filter(Boolean)
              .join(", ")}
          />
          {application.website_url && (
            <Detail label="Website" value={application.website_url} />
          )}
        </dl>
      </section>
    </div>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="break-words">{value}</dd>
    </div>
  );
}

function Notice({
  tone,
  title,
  children,
}: {
  tone: "info" | "danger";
  title: string;
  children: ReactNode;
}) {
  return (
    <section
      role={tone === "danger" ? "alert" : undefined}
      className={
        tone === "danger"
          ? "rounded-2xl bg-destructive/10 p-6 ring-1 ring-destructive/20"
          : "rounded-2xl bg-primary/5 p-6 ring-1 ring-primary/15"
      }
    >
      <div className="flex items-center gap-2">
        <Info
          className={tone === "danger" ? "size-4 text-destructive" : "size-4 text-primary"}
          aria-hidden="true"
        />
        <h3 className="font-semibold">{title}</h3>
      </div>
      <div className="mt-2 text-sm text-muted-foreground">{children}</div>
    </section>
  );
}
