"use client";

import Link from "next/link";
import { CircleCheck, PackageCheck, Truck } from "lucide-react";
import { motion, useReducedMotion } from "motion/react";

import { EASE_OUT, FadeUp } from "@/components/animations/fade-up";
import { buttonVariants } from "@/components/ui/button";
import { formatPrice } from "@/lib/formatters";
import type { Order } from "@/types/order";

/**
 * Order confirmation (PRD §19): animated success icon, order number, itemized
 * summary, shipping and status. Reads the order passed from the server.
 */
export function OrderSuccessView({ order }: { order: Order | null }) {
  const prefersReducedMotion = useReducedMotion();

  if (!order) {
    return (
      <div className="mx-auto flex w-full max-w-3xl flex-col items-center gap-4 px-4 py-24 text-center">
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
          No order found
        </h1>
        <p className="max-w-md text-sm text-muted-foreground sm:text-base">
          We couldn&apos;t find that order. If you just paid, give it a moment and
          check your order history.
        </p>
        <div className="flex flex-wrap justify-center gap-3">
          <Link href="/products" className={buttonVariants()}>
            Continue Shopping
          </Link>
          <Link href="/account/orders" className={buttonVariants({ variant: "outline" })}>
            View orders
          </Link>
        </div>
      </div>
    );
  }

  const shippingAddress = order.shippingAddress;

  return (
    <div className="mx-auto w-full max-w-3xl px-4 py-12 sm:px-6 sm:py-16">
      <div className="flex flex-col items-center text-center">
        <motion.span
          initial={prefersReducedMotion ? { opacity: 0 } : { opacity: 0, scale: 0.5 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: prefersReducedMotion ? 0 : 0.5, ease: EASE_OUT }}
          className="flex size-20 items-center justify-center rounded-full bg-primary-light"
        >
          <CircleCheck className="size-10 text-primary" aria-hidden="true" />
        </motion.span>

        <FadeUp delay={0.15}>
          <h1 className="mt-6 text-3xl font-bold tracking-tight sm:text-4xl">
            Thank you for your order!
          </h1>
        </FadeUp>

        <FadeUp delay={0.25}>
          <p className="mt-3 text-muted-foreground">
            Your order has been received and is being processed.
          </p>
        </FadeUp>

        <FadeUp delay={0.35}>
          <p className="mt-6 rounded-full border px-4 py-2 font-mono text-sm font-medium">
            Order {order.orderNumber}
          </p>
        </FadeUp>
      </div>

      <FadeUp delay={0.45}>
        <div className="mt-10 rounded-2xl border bg-card p-6 sm:p-8">
          <h2 className="text-lg font-semibold">Order Summary</h2>
          <ul className="mt-4 flex flex-col gap-3" aria-label="Items in this order">
            {order.items.map((item) => (
              <li
                key={`${item.productId}:${item.variantId ?? "base"}`}
                className="flex items-baseline justify-between gap-4 text-sm"
              >
                <span className="min-w-0">
                  <span className="font-medium">{item.name}</span>
                  <span className="text-muted-foreground"> × {item.quantity}</span>
                </span>
                <span className="shrink-0 font-medium tabular-nums">
                  {formatPrice(item.price * item.quantity)}
                </span>
              </li>
            ))}
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
            <div className="flex items-baseline justify-between border-t pt-2 text-base">
              <dt className="font-semibold">Total</dt>
              <dd className="font-bold">{formatPrice(order.total)}</dd>
            </div>
          </dl>
        </div>
      </FadeUp>

      <FadeUp delay={0.55}>
        <div className="mt-6 grid gap-4 sm:grid-cols-2">
          <div className="rounded-2xl border p-6">
            <h2 className="flex items-center gap-2 text-sm font-semibold">
              <Truck className="size-4 text-primary" aria-hidden="true" />
              Shipping Information
            </h2>
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
                {shippingAddress.city}
                {shippingAddress.state ? `, ${shippingAddress.state}` : ""}{" "}
                {shippingAddress.postalCode}
                <br />
                {shippingAddress.country}
              </address>
            ) : (
              <p className="mt-3 text-sm text-muted-foreground">
                No shipping address on file.
              </p>
            )}
          </div>

          <div className="rounded-2xl border p-6">
            <h2 className="flex items-center gap-2 text-sm font-semibold">
              <PackageCheck className="size-4 text-primary" aria-hidden="true" />
              Status
            </h2>
            <p className="mt-3 inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-1 text-xs font-medium capitalize text-primary-dark">
              <span aria-hidden="true" className="size-2 rounded-full bg-primary" />
              {order.status}
            </p>
          </div>
        </div>
      </FadeUp>

      <FadeUp delay={0.65}>
        <div className="mt-8 flex flex-wrap justify-center gap-3">
          <Link href="/products" className={buttonVariants({ size: "lg" })}>
            Continue Shopping
          </Link>
          <Link
            href="/account/orders"
            className={buttonVariants({ variant: "outline", size: "lg" })}
          >
            View Order
          </Link>
        </div>
      </FadeUp>
    </div>
  );
}
