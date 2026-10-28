import { Badge } from "@/components/ui/badge";
import {
  APPLICATION_STATUS_LABELS,
  PAYMENT_STATUS_LABELS,
} from "@/lib/store-application-labels";
import { cn } from "@/lib/utils";
import type {
  ApplicationStatus,
  PaymentStatus,
} from "@/types/store-application";

const APPLICATION_TONE: Record<string, string> = {
  pending: "bg-warning/15 text-warning",
  approved: "bg-success/15 text-success",
  rejected: "bg-destructive/10 text-destructive",
};

const PAYMENT_TONE: Record<string, string> = {
  pending: "bg-warning/15 text-warning",
  processing: "bg-primary/10 text-primary-dark",
  paid: "bg-success/15 text-success",
  failed: "bg-destructive/10 text-destructive",
  cancelled: "bg-muted text-muted-foreground",
  refunded: "bg-muted text-muted-foreground",
  partially_refunded: "bg-muted text-muted-foreground",
};

export function ApplicationStatusBadge({
  status,
  className,
}: {
  status: ApplicationStatus | string;
  className?: string;
}) {
  return (
    <Badge
      variant="outline"
      className={cn(
        "border-transparent font-medium",
        APPLICATION_TONE[status] ?? "bg-muted text-muted-foreground",
        className
      )}
    >
      {APPLICATION_STATUS_LABELS[status] ?? status}
    </Badge>
  );
}

export function PaymentStatusBadge({
  status,
  className,
}: {
  status: PaymentStatus | string;
  className?: string;
}) {
  if (!status) return null;
  return (
    <Badge
      variant="outline"
      className={cn(
        "border-transparent font-medium",
        PAYMENT_TONE[status] ?? "bg-muted text-muted-foreground",
        className
      )}
    >
      {PAYMENT_STATUS_LABELS[status] ?? status}
    </Badge>
  );
}
