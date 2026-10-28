"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { CreditCard, Info, Loader2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { apiPost } from "@/lib/api-client";
import type {
  ApplicationDetail,
  PaymentInitResponse,
} from "@/types/store-application";

/**
 * Application-fee payment action. No real provider is integrated yet, so when
 * the backend reports the provider is not configured the UI says so explicitly
 * — it never fakes a successful payment.
 */
export function PaymentAction({
  application,
}: {
  application: ApplicationDetail;
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const fee = application.fee;
  const payment = application.payment;

  if (!fee?.enabled) return null;

  const processing =
    payment && (payment.status === "pending" || payment.status === "processing");

  async function pay() {
    setBusy(true);
    const res = await apiPost<PaymentInitResponse>(
      `/api/v1/store-applications/${application.id}/payments`
    );
    setBusy(false);

    if (!res.ok) {
      if (res.error.code === "PAYMENT_NOT_CONFIGURED") {
        toast.error("Online payment isn't available yet", {
          description: res.error.message,
        });
        return;
      }
      toast.error(res.error.message);
      return;
    }

    if (res.data.redirect_url) {
      window.location.href = res.data.redirect_url;
      return;
    }
    toast.success("Payment started", {
      description: "We'll update your application once it's verified.",
    });
    router.refresh();
  }

  if (processing) {
    return (
      <p
        role="status"
        className="flex items-center gap-2 rounded-lg bg-warning/10 px-3 py-2 text-sm text-warning"
      >
        <Info className="size-4" aria-hidden="true" />
        Your payment is awaiting verification. We&apos;ll activate your store as
        soon as it&apos;s confirmed.
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <Button
        type="button"
        size="lg"
        className="h-11 w-full text-base sm:w-auto"
        onClick={pay}
        disabled={busy || !fee.provider_configured}
      >
        {busy ? (
          <Loader2 className="size-4 animate-spin" aria-hidden="true" />
        ) : (
          <CreditCard className="size-4" aria-hidden="true" />
        )}
        Pay application fee
      </Button>
      {!fee.provider_configured && (
        <p
          role="status"
          className="flex items-start gap-2 rounded-lg bg-muted px-3 py-2 text-xs text-muted-foreground"
        >
          <Info className="mt-0.5 size-3.5 shrink-0" aria-hidden="true" />
          Online payment isn&apos;t available yet — no payment provider is
          configured. Your approved application will be ready to activate once
          payments are enabled.
        </p>
      )}
    </div>
  );
}
