import {
  Check,
  ClipboardCheck,
  CircleX,
  Package,
  PackageCheck,
  Receipt,
  Truck,
} from "lucide-react";

import { formatDate } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { OrderStatus } from "@/types/order";

interface TimelineStep {
  label: string;
  description: string;
  icon: typeof Receipt;
}

/** The customer-facing fulfilment flow (matches the backend order_status). */
const STEPS: TimelineStep[] = [
  { label: "Order placed", description: "We received your order.", icon: Receipt },
  { label: "Confirmed", description: "The seller accepted your order.", icon: ClipboardCheck },
  { label: "Packed", description: "Your items are packed for shipment.", icon: Package },
  { label: "Shipped", description: "Your parcel is on its way.", icon: Truck },
  { label: "Delivered", description: "Enjoy your order!", icon: PackageCheck },
];

/** Highest completed step for each order status (processing = accepted, being prepared). */
const STATUS_PROGRESS: Partial<Record<OrderStatus, number>> = {
  pending: 0,
  confirmed: 1,
  processing: 1,
  packed: 2,
  shipped: 3,
  delivered: 4,
  cancelled: 0,
  returned: 4,
  refunded: 4,
};

const TERMINAL_NOTICE: Partial<Record<OrderStatus, string>> = {
  cancelled: "This order was cancelled.",
  returned: "This order was returned.",
  refunded: "This order was refunded.",
};

interface OrderTimelineProps {
  status: OrderStatus;
  createdAt?: string;
  className?: string;
}

/**
 * Animated fulfilment timeline for a customer order: each completed step fills
 * in, the next step pulses while it is in progress, and terminal states
 * (cancelled/returned/refunded) show a destructive notice.
 */
export function OrderTimeline({ status, createdAt, className }: OrderTimelineProps) {
  const notice = TERMINAL_NOTICE[status];
  const reached = STATUS_PROGRESS[status] ?? 0;

  return (
    <section
      aria-label="Order progress"
      className={cn("rounded-2xl border p-5 sm:p-6", className)}
    >
      <h3 className="text-base font-semibold">Order progress</h3>

      <ol className="mt-4 flex flex-col">
        {STEPS.map((step, index) => {
          const done = index <= reached;
          const current = index === reached + 1;
          const Icon = step.icon;

          return (
            <li
              key={step.label}
              className="relative flex gap-4 pb-6 last:pb-0 motion-reduce:transition-none animate-in fade-in slide-in-from-left-2 fill-mode-backwards duration-500 motion-reduce:animate-none"
              style={{ animationDelay: `${index * 120}ms` }}
            >
              {/* Connector to the next step */}
              {index < STEPS.length - 1 && (
                <span
                  aria-hidden="true"
                  className={cn(
                    "absolute left-[15px] top-8 h-[calc(100%-2rem)] w-0.5 rounded-full transition-colors duration-500",
                    index < reached ? "bg-primary" : "bg-border"
                  )}
                />
              )}

              {/* Step indicator */}
              <span
                aria-hidden="true"
                className={cn(
                  "relative z-10 flex size-8 shrink-0 items-center justify-center rounded-full border-2 transition-colors duration-500",
                  done && "border-primary bg-primary text-primary-foreground",
                  current && "border-primary bg-background text-primary",
                  !done && !current && "border-border bg-muted text-muted-foreground"
                )}
              >
                {current && (
                  <span className="absolute inset-0 animate-ping rounded-full bg-primary/40 motion-reduce:animate-none" />
                )}
                <span className="relative">
                  {done ? <Check className="size-4" /> : <Icon className="size-4" />}
                </span>
              </span>

              <div className="min-w-0 pt-1">
                <p
                  className={cn(
                    "text-sm font-medium",
                    !done && !current && "text-muted-foreground"
                  )}
                >
                  {step.label}
                  {current && (
                    <span className="ml-2 rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-semibold text-primary-dark">
                      In progress
                    </span>
                  )}
                </p>
                <p className="mt-0.5 text-xs text-muted-foreground">{step.description}</p>
                {index === 0 && createdAt && (
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {formatDate(createdAt)}
                  </p>
                )}
              </div>
            </li>
          );
        })}
      </ol>

      {notice && (
        <p className="mt-4 flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm font-medium text-destructive">
          <CircleX className="size-4 shrink-0" aria-hidden="true" />
          {notice}
        </p>
      )}
    </section>
  );
}
