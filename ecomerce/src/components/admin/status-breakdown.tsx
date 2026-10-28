import { OrderStatusBadge } from "@/components/orders/order-status-badge";
import { formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { StatusBreakdownEntry } from "@/types/admin";

interface StatusBreakdownProps {
  entries: StatusBreakdownEntry[];
  className?: string;
}

/** Order counts and revenue by status (PRD §24). */
export function StatusBreakdown({
  entries,
  className,
}: StatusBreakdownProps) {
  return (
    <ul className={cn("flex flex-col gap-3", className)}>
      {entries.map((entry) => (
        <li
          key={entry.status}
          className="flex items-center justify-between gap-4"
        >
          <OrderStatusBadge status={entry.status} />
          <span className="flex items-baseline gap-3 text-sm">
            <span className="text-muted-foreground tabular-nums">
              {entry.count} {entry.count === 1 ? "order" : "orders"}
            </span>
            <span className="w-20 text-right font-medium tabular-nums">
              {formatPrice(entry.revenue)}
            </span>
          </span>
        </li>
      ))}
    </ul>
  );
}
