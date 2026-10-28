import {
  FREE_SHIPPING_THRESHOLD,
  STANDARD_SHIPPING_RATE,
} from "@/lib/constants";
import { formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";

interface CartSummaryProps {
  subtotal: number;
  className?: string;
}

/**
 * Order summary (PRD §17): subtotal, shipping estimate and total, plus a
 * free-shipping progress indicator. Presentational — totals derive from the
 * cart subtotal via shared constants.
 */
export function CartSummary({ subtotal, className }: CartSummaryProps) {
  const freeShipping = subtotal >= FREE_SHIPPING_THRESHOLD;
  const shipping =
    subtotal > 0 && !freeShipping ? STANDARD_SHIPPING_RATE : 0;
  const total = subtotal + shipping;
  const progress = Math.min(1, subtotal / FREE_SHIPPING_THRESHOLD);
  const missing = FREE_SHIPPING_THRESHOLD - subtotal;

  return (
    <div className={cn("flex flex-col gap-3", className)}>
      <dl className="flex flex-col gap-2 text-sm">
        <div className="flex items-baseline justify-between">
          <dt className="text-muted-foreground">Subtotal</dt>
          <dd className="font-medium">{formatPrice(subtotal)}</dd>
        </div>
        <div className="flex items-baseline justify-between">
          <dt className="text-muted-foreground">Shipping estimate</dt>
          <dd className="font-medium">
            {shipping === 0 ? "Free" : formatPrice(shipping)}
          </dd>
        </div>
        <div className="flex items-baseline justify-between border-t pt-2 text-base">
          <dt className="font-semibold">Total</dt>
          <dd className="font-bold">{formatPrice(total)}</dd>
        </div>
      </dl>

      {subtotal > 0 && !freeShipping && (
        <div>
          <p className="text-xs text-muted-foreground">
            Add{" "}
            <span className="font-medium text-foreground">
              {formatPrice(missing)}
            </span>{" "}
            more for free shipping
          </p>
          <div
            role="progressbar"
            aria-label="Progress toward free shipping"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={Math.round(progress * 100)}
            className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted"
          >
            <div
              className="h-full rounded-full bg-primary transition-[width] duration-300"
              style={{ width: `${progress * 100}%` }}
            />
          </div>
        </div>
      )}

      {subtotal > 0 && freeShipping && (
        <p className="text-xs font-medium text-primary-dark">
          You have unlocked free shipping!
        </p>
      )}
    </div>
  );
}
