"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import {
  ArrowLeft,
  BadgeCheck,
  CircleAlert,
  CreditCard,
  PackageCheck,
  Pencil,
  Star,
  Truck,
} from "lucide-react";

import { ReviewFormDialog } from "@/components/account/review-form-dialog";
import { OrderStatusBadge } from "@/components/orders/order-status-badge";
import { OrderTimeline } from "@/components/orders/order-timeline";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useHasMounted } from "@/hooks/use-has-mounted";
import { formatDate, formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import {
  selectOrderById,
  useOrderStore,
  type ShippingAddressSnapshot,
} from "@/store/order-store";
import type { Order, OrderItemReview } from "@/types/order";
import type { Address } from "@/types/user";

interface OrderDetailViewProps {
  orderId: string;
  /** Demo order resolved on the server; null when it was placed this session. */
  order: Order | null;
  defaultAddress: Address | null;
}

function toSnapshot(address: Address | null): ShippingAddressSnapshot | null {
  if (!address) return null;
  // Addresses may be the structured BD shape or the legacy line shape.
  return {
    fullName: address.fullName,
    line1: address.line1 ?? address.villageTown ?? "",
    line2: address.line2,
    city: address.city ?? address.district ?? "",
    state: address.state ?? address.division ?? "",
    postalCode: address.postalCode ?? "",
    country: address.country ?? "Bangladesh",
  };
}

function OrderDetailSkeleton() {
  return (
    <div aria-hidden="true" className="flex flex-col gap-6">
      <Skeleton className="h-8 w-52" />
      <Skeleton className="h-64 rounded-2xl" />
      <div className="grid gap-4 sm:grid-cols-2">
        <Skeleton className="h-40 rounded-2xl" />
        <Skeleton className="h-40 rounded-2xl" />
      </div>
    </div>
  );
}

/**
 * Order detail (PRD §23): items, totals, shipping, delivery and payment.
 * Resolves session orders from the store, falling back to the demo order
 * passed in from the server.
 */
export function OrderDetailView({
  orderId,
  order: demoOrder,
  defaultAddress,
}: OrderDetailViewProps) {
  const hasMounted = useHasMounted();
  const placed = useOrderStore(selectOrderById(orderId));
  const [reviews, setReviews] = useState<Record<string, OrderItemReview>>({});
  const [target, setTarget] = useState<{
    orderItemId: string;
    name: string;
    existing?: OrderItemReview;
  } | null>(null);

  if (!hasMounted) {
    return <OrderDetailSkeleton />;
  }

  const order = placed?.order ?? demoOrder;

  if (!order) {
    return (
      <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed px-6 py-16 text-center">
        <span className="flex size-14 items-center justify-center rounded-full bg-muted">
          <CircleAlert
            className="size-7 text-muted-foreground"
            aria-hidden="true"
          />
        </span>
        <div>
          <h2 className="text-lg font-semibold">Order not found</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            We could not find an order with that reference.
          </p>
        </div>
        <Link href="/account/orders" className={cn(buttonVariants())}>
          Back to orders
        </Link>
      </div>
    );
  }

  const shippingAddress =
    placed?.shippingAddress ?? order.shippingAddress ?? toSnapshot(defaultAddress);

  return (
    <div className="flex flex-col gap-6">
      <div>
        <Link
          href="/account/orders"
          className="inline-flex items-center gap-1 rounded-md text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2"
        >
          <ArrowLeft className="size-3.5" aria-hidden="true" />
          All orders
        </Link>

        <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-xl font-semibold">Order {order.orderNumber}</h2>
            <p className="text-sm text-muted-foreground">
              Placed {formatDate(order.createdAt)}
            </p>
          </div>
          <OrderStatusBadge status={order.status} />
        </div>
      </div>

      <OrderTimeline status={order.status} createdAt={order.createdAt} />

      <section
        aria-labelledby="items-heading"
        className="rounded-2xl border p-5 sm:p-6"
      >
        <h3 id="items-heading" className="text-base font-semibold">
          Items
        </h3>
        <ul className="mt-4 flex flex-col gap-4">
          {order.items.map((item) => {
            const key =
              item.orderItemId ?? `${item.productId}:${item.variantId ?? "base"}`;
            const review = reviews[key] ?? item.review;
            const reviewed = review !== undefined || item.reviewed === true;
            const canReview =
              order.status === "delivered" &&
              Boolean(item.productId) &&
              Boolean(item.orderItemId);

            return (
              <li key={key} className="flex items-start gap-3 sm:gap-4">
                <div className="relative size-14 shrink-0 overflow-hidden rounded-lg bg-muted">
                  {item.imageUrl ? (
                    <Image
                      src={item.imageUrl}
                      alt=""
                      fill
                      sizes="56px"
                      className="object-cover"
                    />
                  ) : (
                    <span className="flex h-full w-full items-center justify-center">
                      <PackageCheck
                        className="size-5 text-muted-foreground"
                        aria-hidden="true"
                      />
                    </span>
                  )}
                </div>

                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium">{item.name}</p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {item.sku ? `${item.sku} · ` : ""}
                    Qty {item.quantity}
                  </p>

                  {canReview && (
                    <div className="mt-2 flex flex-wrap items-center gap-2">
                      {reviewed ? (
                        <>
                          <Badge variant="secondary" className="gap-1">
                            <BadgeCheck className="size-3.5" aria-hidden="true" />
                            Reviewed
                          </Badge>
                          {review && (
                            <span
                              className="flex items-center gap-0.5"
                              aria-label={`${review.rating} out of 5 stars`}
                            >
                              {Array.from({ length: 5 }).map((_, starIndex) => (
                                <Star
                                  key={starIndex}
                                  className={cn(
                                    "size-3.5",
                                    starIndex < review.rating
                                      ? "fill-amber-400 text-amber-400"
                                      : "fill-muted text-muted-foreground/30"
                                  )}
                                  aria-hidden="true"
                                />
                              ))}
                            </span>
                          )}
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() =>
                              setTarget({
                                orderItemId: item.orderItemId as string,
                                name: item.name,
                                existing: review,
                              })
                            }
                          >
                            <Pencil className="size-3.5" aria-hidden="true" />
                            Edit review
                          </Button>
                        </>
                      ) : (
                        <Button
                          type="button"
                          size="sm"
                          onClick={() =>
                            setTarget({
                              orderItemId: item.orderItemId as string,
                              name: item.name,
                            })
                          }
                        >
                          <Star className="size-3.5" aria-hidden="true" />
                          Write a review
                        </Button>
                      )}
                    </div>
                  )}
                </div>

                <span className="shrink-0 text-sm font-medium tabular-nums">
                  {formatPrice(item.price * item.quantity)}
                </span>
              </li>
            );
          })}
        </ul>

        <dl className="mt-4 flex flex-col gap-2 border-t pt-4 text-sm">
          <div className="flex items-baseline justify-between">
            <dt className="text-muted-foreground">Subtotal</dt>
            <dd className="font-medium">{formatPrice(order.subtotal)}</dd>
          </div>
          <div className="flex items-baseline justify-between">
            <dt className="text-muted-foreground">Shipping</dt>
            <dd className="font-medium">
              {order.shipping === 0 ? "Free" : formatPrice(order.shipping)}
            </dd>
          </div>
          {order.tax > 0 && (
            <div className="flex items-baseline justify-between">
              <dt className="text-muted-foreground">Tax</dt>
              <dd className="font-medium">{formatPrice(order.tax)}</dd>
            </div>
          )}
          <div className="flex items-baseline justify-between border-t pt-2 text-base">
            <dt className="font-semibold">Total</dt>
            <dd className="font-bold">{formatPrice(order.total)}</dd>
          </div>
        </dl>
      </section>

      <div className="grid gap-4 sm:grid-cols-2">
        <section aria-labelledby="shipping-heading" className="rounded-2xl border p-5">
          <h3
            id="shipping-heading"
            className="flex items-center gap-2 text-sm font-semibold"
          >
            <Truck className="size-4 text-primary" aria-hidden="true" />
            Shipping address
          </h3>
          {shippingAddress ? (
            <address className="mt-3 text-sm not-italic leading-relaxed text-muted-foreground">
              {shippingAddress.fullName}
              <br />
              {shippingAddress.line1}
              {shippingAddress.line2 && (
                <>
                  <br />
                  {shippingAddress.line2}
                </>
              )}
              <br />
              {shippingAddress.city}, {shippingAddress.state}{" "}
              {shippingAddress.postalCode}
              <br />
              {shippingAddress.country}
            </address>
          ) : (
            <p className="mt-3 text-sm text-muted-foreground">
              No shipping address on file for this order.
            </p>
          )}
        </section>

        <section aria-labelledby="delivery-heading" className="rounded-2xl border p-5">
          <h3
            id="delivery-heading"
            className="flex items-center gap-2 text-sm font-semibold"
          >
            <PackageCheck className="size-4 text-primary" aria-hidden="true" />
            Delivery
          </h3>
          <p className="mt-3 text-sm font-medium">
            {placed?.deliveryLabel ?? "Standard Delivery"}
          </p>
          {order.estimatedDelivery && (
            <p className="mt-1 text-sm text-muted-foreground">
              Estimated arrival: {formatDate(order.estimatedDelivery)}
            </p>
          )}
          {placed?.paymentLabel && (
            <p className="mt-3 flex items-center gap-1.5 text-xs text-muted-foreground">
              <CreditCard className="size-3.5" aria-hidden="true" />
              Payment: {placed.paymentLabel}
            </p>
          )}
        </section>
      </div>

      {target && (
        <ReviewFormDialog
          key={`${target.orderItemId}:${target.existing?.id ?? "new"}`}
          open={target !== null}
          onOpenChange={(open) => {
            if (!open) setTarget(null);
          }}
          orderItemId={target.orderItemId}
          productName={target.name}
          existing={target.existing}
          onSaved={(review) =>
            setReviews((prev) => ({ ...prev, [target.orderItemId]: review }))
          }
        />
      )}
    </div>
  );
}
