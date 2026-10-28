import Image from "next/image";

import type { CartLine } from "@/hooks/use-cart-lines";
import { formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";

interface CheckoutSummaryProps {
  lines: CartLine[];
  subtotal: number;
  deliveryLabel: string;
  deliveryCost: number;
  className?: string;
}

/** Read-only order summary for checkout (PRD §18) — items + totals. */
export function CheckoutSummary({
  lines,
  subtotal,
  deliveryLabel,
  deliveryCost,
  className,
}: CheckoutSummaryProps) {
  const total = subtotal + deliveryCost;

  return (
    <div className={cn("flex flex-col gap-5", className)}>
      <ul className="flex flex-col gap-3" aria-label="Items in this order">
        {lines.map((line) => (
          <li key={line.id} className="flex items-center gap-3 text-sm">
            <span className="relative size-12 shrink-0 overflow-hidden rounded-lg border">
              <Image
                src={line.imageUrl}
                alt={line.imageAlt}
                fill
                sizes="48px"
                className="object-cover"
              />
            </span>
            <span className="flex min-w-0 flex-1 flex-col">
              <span className="truncate font-medium">{line.name}</span>
              {line.variantLabel && (
                <span className="text-xs text-muted-foreground">
                  {line.variantLabel}
                </span>
              )}
              <span className="text-xs text-muted-foreground">
                Qty {line.quantity}
              </span>
            </span>
            <span className="font-medium tabular-nums">
              {formatPrice(line.lineTotal)}
            </span>
          </li>
        ))}
      </ul>

      <dl className="flex flex-col gap-2 border-t pt-4 text-sm">
        <div className="flex items-baseline justify-between">
          <dt className="text-muted-foreground">Subtotal</dt>
          <dd className="font-medium">{formatPrice(subtotal)}</dd>
        </div>
        <div className="flex items-baseline justify-between">
          <dt className="text-muted-foreground">Delivery — {deliveryLabel}</dt>
          <dd className="font-medium">
            {deliveryCost === 0 ? "Free" : formatPrice(deliveryCost)}
          </dd>
        </div>
        <div className="flex items-baseline justify-between border-t pt-2 text-base">
          <dt className="font-semibold">Total</dt>
          <dd className="font-bold">{formatPrice(total)}</dd>
        </div>
      </dl>
    </div>
  );
}
