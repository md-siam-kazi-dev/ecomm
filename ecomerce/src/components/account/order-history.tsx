"use client";

import Link from "next/link";
import { Package, PackageSearch } from "lucide-react";

import { OrderStatusBadge } from "@/components/orders/order-status-badge";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useAccountOrders } from "@/hooks/use-account-orders";
import { useHasMounted } from "@/hooks/use-has-mounted";
import { formatDate, formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { Order } from "@/types/order";

const PREVIEW_ITEM_COUNT = 3;

function OrderHistorySkeleton() {
  return (
    <div aria-hidden="true" className="flex flex-col gap-4">
      {Array.from({ length: 3 }).map((_, index) => (
        <Skeleton key={index} className="h-44 rounded-2xl" />
      ))}
    </div>
  );
}

/** Order history list (PRD §23) — demo orders merged with session orders. */
export function OrderHistory({ orders }: { orders: Order[] }) {
  const hasMounted = useHasMounted();
  const accountOrders = useAccountOrders(orders);

  if (!hasMounted) {
    return <OrderHistorySkeleton />;
  }

  if (accountOrders.length === 0) {
    return (
      <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed px-6 py-16 text-center">
        <span className="flex size-14 items-center justify-center rounded-full bg-muted">
          <PackageSearch
            className="size-7 text-muted-foreground"
            aria-hidden="true"
          />
        </span>
        <div>
          <h2 className="text-lg font-semibold">No orders yet</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            When you place an order it will appear here with its status.
          </p>
        </div>
        <Link href="/products" className={cn(buttonVariants())}>
          Start shopping
        </Link>
      </div>
    );
  }

  return (
    <ul className="flex flex-col gap-4">
      {accountOrders.map(({ order }) => {
        const previewItems = order.items.slice(0, PREVIEW_ITEM_COUNT);
        const hiddenCount = order.items.length - previewItems.length;

        return (
          <li key={order.id} className="rounded-2xl border p-4 sm:p-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-semibold">Order {order.orderNumber}</p>
                <p className="text-xs text-muted-foreground">
                  Placed {formatDate(order.createdAt)}
                </p>
              </div>
              <OrderStatusBadge status={order.status} />
            </div>

            <ul className="mt-3 flex flex-col gap-1 text-sm text-muted-foreground">
              {previewItems.map((item) => (
                <li
                  key={`${item.productId}:${item.variantId ?? "base"}`}
                  className="flex items-baseline gap-1"
                >
                  <span className="line-clamp-1 min-w-0">{item.name}</span>
                  <span aria-hidden="true">×</span>
                  <span className="sr-only">quantity</span>
                  {item.quantity}
                </li>
              ))}
              {hiddenCount > 0 && (
                <li>
                  +{hiddenCount} more {hiddenCount === 1 ? "item" : "items"}
                </li>
              )}
            </ul>

            <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-3">
              <p className="text-sm">
                <span className="text-muted-foreground">Total </span>
                <span className="font-semibold tabular-nums">
                  {formatPrice(order.total)}
                </span>
              </p>
              <Link
                href={`/account/orders/${order.id}`}
                className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
              >
                <Package className="size-3.5" aria-hidden="true" />
                View details
              </Link>
            </div>
          </li>
        );
      })}
    </ul>
  );
}
