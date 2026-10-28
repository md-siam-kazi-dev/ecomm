"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import type { ReactNode } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Check, Loader2, Save, X } from "lucide-react";
import { toast } from "sonner";

import {
  ApplicationStatusBadge,
  PaymentStatusBadge,
} from "@/components/store/application-status-badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { apiPost } from "@/lib/api-client";
import { BUSINESS_TYPE_LABELS } from "@/lib/store-application-labels";
import { formatDate } from "@/lib/formatters";
import type { AdminApplicationDetail } from "@/types/store-application";

export function AdminStoreApplicationDetail({
  application,
}: {
  application: AdminApplicationDetail;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState<"approve" | "reject" | null>(null);
  const [rejectOpen, setRejectOpen] = useState(false);
  const [reason, setReason] = useState("");
  const [note, setNote] = useState(application.admin_review_notes ?? "");
  const [savingNote, setSavingNote] = useState(false);

  const pending = application.status === "pending";

  async function approve() {
    setBusy("approve");
    const res = await apiPost(`/api/v1/admin/store-applications/${application.id}/approve`);
    setBusy(null);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Application approved");
    router.refresh();
  }

  async function reject() {
    if (!reason.trim()) {
      toast.error("A rejection reason is required.");
      return;
    }
    setBusy("reject");
    const res = await apiPost(
      `/api/v1/admin/store-applications/${application.id}/reject`,
      { reason: reason.trim() }
    );
    setBusy(null);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    setRejectOpen(false);
    toast.success("Application rejected");
    router.refresh();
  }

  async function saveNote() {
    setSavingNote(true);
    const res = await apiPost(
      `/api/v1/admin/store-applications/${application.id}/notes`,
      { note }
    );
    setSavingNote(false);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Note saved");
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-6">
      <Link
        href="/admin/store-applications"
        className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        All applications
      </Link>

      <header className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="font-mono text-xs text-muted-foreground">
            {application.application_number}
          </p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight">{application.store_name}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {application.applicant_account.name} · {application.applicant_account.email} ·
            submitted {formatDate(application.submitted_at)}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <ApplicationStatusBadge status={application.status} />
          {application.payment && <PaymentStatusBadge status={application.payment.status} />}
        </div>
      </header>

      <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
        <div className="flex flex-col gap-6">
          <Section title="Store information">
            {(application.logo_url || application.banner_url) && (
              <div className="mb-4 flex flex-col gap-4 sm:flex-row sm:items-start">
                {application.logo_url && (
                  <figure className="shrink-0">
                    <figcaption className="mb-1.5 text-xs text-muted-foreground">
                      Store logo
                    </figcaption>
                    <Image
                      src={application.logo_url}
                      alt={`${application.store_name} logo`}
                      width={72}
                      height={72}
                      unoptimized
                      className="size-18 rounded-xl object-cover ring-1 ring-foreground/10"
                    />
                  </figure>
                )}
                {application.banner_url && (
                  <figure className="min-w-0 flex-1">
                    <figcaption className="mb-1.5 text-xs text-muted-foreground">
                      Store banner
                    </figcaption>
                    <Image
                      src={application.banner_url}
                      alt={`${application.store_name} banner`}
                      width={640}
                      height={160}
                      unoptimized
                      className="h-24 w-full max-w-lg rounded-xl object-cover ring-1 ring-foreground/10"
                    />
                  </figure>
                )}
              </div>
            )}
            <Grid>
              <Field label="Store name" value={application.store_name} />
              <Field label="Store URL" value={application.store_slug} />
              <Field label="Category" value={application.category} />
              <Field label="Description" value={application.description} full />
              {application.website_url && <Field label="Website" value={application.website_url} />}
            </Grid>
          </Section>

          <Section title="Contact">
            <Grid>
              <Field label="Business email" value={application.business_email} />
              <Field label="Contact phone" value={application.contact_phone} />
              {application.support_phone && <Field label="Support phone" value={application.support_phone} />}
            </Grid>
          </Section>

          <Section title="Business address">
            <Grid>
              <Field label="Country" value={application.country} />
              <Field label="Division" value={application.division} />
              <Field label="District" value={application.district} />
              {application.postal_code && <Field label="Postal code" value={application.postal_code} />}
              <Field label="Full address" value={application.full_address} full />
            </Grid>
          </Section>

          <Section title="Business & verification">
            <Grid>
              <Field
                label="Business type"
                value={BUSINESS_TYPE_LABELS[application.business_type] ?? application.business_type}
              />
              {application.registered_business_name && (
                <Field label="Registered name" value={application.registered_business_name} />
              )}
              {application.trade_license_number && (
                <Field label="Trade license" value={application.trade_license_number} />
              )}
            </Grid>
          </Section>

          <Section title="Applicant account">
            <Grid>
              <Field label="Name" value={application.applicant_account.name} />
              <Field label="Email" value={application.applicant_account.email} />
              <Field label="Account status" value={application.applicant_account.status} />
              <Field
                label="Registered"
                value={formatDate(application.applicant_account.created_at)}
              />
              <Field label="User ID" value={application.applicant_user_id} full mono />
            </Grid>
          </Section>

          <Section title="History">
            {application.history.length === 0 ? (
              <p className="text-sm text-muted-foreground">No recorded activity yet.</p>
            ) : (
              <ol className="flex flex-col gap-3">
                {application.history.map((entry, index) => (
                  <li key={`${entry.action}-${index}`} className="flex items-start gap-3 text-sm">
                    <span className="mt-1.5 size-2 shrink-0 rounded-full bg-primary" aria-hidden="true" />
                    <span>
                      <span className="font-medium capitalize">
                        {entry.action.replace(/[._]/g, " ").replace("store application ", "")}
                      </span>
                      <span className="block text-xs text-muted-foreground">
                        {entry.actor_name} · {formatDate(entry.created_at)}
                      </span>
                    </span>
                  </li>
                ))}
              </ol>
            )}
          </Section>
        </div>

        <aside className="flex flex-col gap-6">
          <div className="rounded-xl bg-card p-5 ring-1 ring-foreground/10">
            <h2 className="font-semibold">Decision</h2>
            {application.status === "rejected" && application.rejection_reason && (
              <p className="mt-3 rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {application.rejection_reason}
              </p>
            )}
            {application.status === "approved" && (
              <p className="mt-3 text-sm text-muted-foreground">
                Approved{application.reviewed_at ? ` on ${formatDate(application.reviewed_at)}` : ""}.
                The applicant can now complete the application fee.
              </p>
            )}

            {pending ? (
              <div className="mt-4 flex flex-col gap-2">
                <Button onClick={approve} disabled={busy !== null}>
                  {busy === "approve" ? (
                    <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                  ) : (
                    <Check className="size-4" aria-hidden="true" />
                  )}
                  Approve application
                </Button>

                <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
                  <DialogTrigger asChild>
                    <Button variant="destructive" disabled={busy !== null}>
                      <X className="size-4" aria-hidden="true" />
                      Reject application
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Reject this application?</DialogTitle>
                      <DialogDescription>
                        Provide a reason. It will be shown to the applicant.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="flex flex-col gap-1.5">
                      <Label htmlFor="reject-reason">Rejection reason</Label>
                      <textarea
                        id="reject-reason"
                        value={reason}
                        onChange={(e) => setReason(e.target.value)}
                        rows={4}
                        className="min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
                        placeholder="e.g. Business details could not be verified."
                      />
                    </div>
                    <DialogFooter>
                      <Button variant="outline" onClick={() => setRejectOpen(false)}>
                        Cancel
                      </Button>
                      <Button
                        variant="destructive"
                        onClick={reject}
                        disabled={busy !== null || !reason.trim()}
                      >
                        {busy === "reject" && <Loader2 className="size-4 animate-spin" aria-hidden="true" />}
                        Reject application
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </div>
            ) : (
              <p className="mt-3 text-xs text-muted-foreground">
                This application has already been reviewed.
              </p>
            )}
          </div>

          <div className="rounded-xl bg-card p-5 ring-1 ring-foreground/10">
            <h2 className="font-semibold">Internal notes</h2>
            <p className="mt-1 text-xs text-muted-foreground">
              Not visible to the applicant.
            </p>
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              rows={4}
              className="mt-3 min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              placeholder="Add a note for other admins…"
            />
            <Button
              variant="outline"
              className="mt-3 w-full"
              onClick={saveNote}
              disabled={savingNote}
            >
              {savingNote ? (
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
              ) : (
                <Save className="size-4" aria-hidden="true" />
              )}
              Save note
            </Button>
          </div>
        </aside>
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-xl bg-card p-5 ring-1 ring-foreground/10 sm:p-6">
      <h2 className="font-semibold">{title}</h2>
      <div className="mt-4">{children}</div>
    </section>
  );
}

function Grid({ children }: { children: ReactNode }) {
  return <dl className="grid gap-x-6 gap-y-3 text-sm sm:grid-cols-2">{children}</dl>;
}

function Field({
  label,
  value,
  full,
  mono,
}: {
  label: string;
  value: string;
  full?: boolean;
  mono?: boolean;
}) {
  return (
    <div className={full ? "sm:col-span-2" : undefined}>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className={`break-words ${mono ? "font-mono text-xs" : ""}`}>{value || "—"}</dd>
    </div>
  );
}
