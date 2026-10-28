import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { OrderStatus } from "@/types/order";

/** Human-readable labels for the order lifecycle (PRD §24). */
export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  pending: "Pending",
  confirmed: "Confirmed",
  processing: "Processing",
  packed: "Packed",
  shipped: "Shipped",
  delivered: "Delivered",
  cancelled: "Cancelled",
  returned: "Returned",
  refunded: "Refunded",
};

type BadgeVariant =
  | "default"
  | "secondary"
  | "destructive"
  | "success"
  | "warning"
  | "info"
  | "muted"
  | "outline";

/** Amber = needs action, blue = in flight, green = done, grey = closed. */
const STATUS_VARIANTS: Record<OrderStatus, BadgeVariant> = {
  pending: "warning",
  confirmed: "info",
  processing: "info",
  packed: "info",
  shipped: "info",
  delivered: "success",
  cancelled: "destructive",
  returned: "muted",
  refunded: "muted",
};

interface OrderStatusBadgeProps {
  status: OrderStatus;
  className?: string;
}

/** Shared status pill for account and admin order views (PRD §24). */
export function OrderStatusBadge({
  status,
  className,
}: OrderStatusBadgeProps) {
  return (
    <Badge variant={STATUS_VARIANTS[status]} className={cn("font-medium", className)}>
      {ORDER_STATUS_LABELS[status]}
    </Badge>
  );
}
