"use client";

import Link from "next/link";
import {
  ArrowRight,
  CalendarDays,
  Heart,
  MapPin,
  Package,
  Receipt,
} from "lucide-react";

import { OrderStatusBadge } from "@/components/orders/order-status-badge";
import { buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useAccountOrders } from "@/hooks/use-account-orders";
import { useHasMounted } from "@/hooks/use-has-mounted";
import { formatDate, formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import { selectWishlistCount, useWishlistStore } from "@/store/wishlist-store";
import type { Order } from "@/types/order";
import type { User } from "@/types/user";

interface AccountOverviewProps {
  user: User;
  /** Demo order history from the repository (PRD §44). */
  orders: Order[];
}

function OverviewSkeleton() {
  return (
    <div aria-hidden="true" className="flex flex-col gap-6">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} className="h-28 rounded-2xl" />
        ))}
      </div>
      <Skeleton className="h-64 rounded-2xl" />
    </div>
  );
}

interface OverviewStatProps {
  label: string;
  value: string;
  icon: typeof Package;
}

function OverviewStat({ label, value, icon: Icon }: OverviewStatProps) {
  return (
    <div className="rounded-2xl border p-5">
      <span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary-dark">
        <Icon className="size-4" aria-hidden="true" />
      </span>
      <p className="mt-3 text-sm text-muted-foreground">{label}</p>
      <p className="mt-0.5 text-xl font-semibold">{value}</p>
    </div>
  );
}

/**
 * Account dashboard (PRD §23): lifetime stats, recent orders and the default
 * shipping address. Reads the session order history through the store.
 */
export function AccountOverview({ user, orders }: AccountOverviewProps) {
  const hasMounted = useHasMounted();
  const accountOrders = useAccountOrders(orders);
  const wishlistCount = useWishlistStore(selectWishlistCount);

  if (!hasMounted) {
    return <OverviewSkeleton />;
  }

  const activeOrders = accountOrders.filter(
    ({ order }) => order.status !== "cancelled"
  );
  const totalSpent = activeOrders.reduce(
    (sum, { order }) => sum + order.total,
    0
  );
  const recentOrders = accountOrders.slice(0, 3);
  const defaultAddress =
    user.addresses.find((address) => address.isDefault) ?? user.addresses[0];

  return (
    <div className="flex flex-col gap-8">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <OverviewStat
          label="Total orders"
          value={String(accountOrders.length)}
          icon={Package}
        />
        <OverviewStat
          label="Total spent"
          value={formatPrice(totalSpent)}
          icon={Receipt}
        />
        <OverviewStat
          label="Wishlist items"
          value={String(wishlistCount)}
          icon={Heart}
        />
        <OverviewStat
          label="Member since"
          value={formatDate(user.memberSince)}
          icon={CalendarDays}
        />
      </div>

      <section aria-labelledby="recent-orders-heading">
        <div className="flex items-center justify-between gap-4">
          <h2 id="recent-orders-heading" className="text-lg font-semibold">
            Recent orders
          </h2>
          <Link
            href="/account/orders"
            className="inline-flex items-center gap-1 rounded-md text-sm font-medium text-primary-dark underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            View all
            <ArrowRight className="size-3.5" aria-hidden="true" />
          </Link>
        </div>

        {recentOrders.length === 0 ? (
          <div className="mt-4 flex flex-col items-center gap-3 rounded-2xl border border-dashed px-6 py-10 text-center">
            <Package
              className="size-7 text-muted-foreground"
              aria-hidden="true"
            />
            <p className="text-sm text-muted-foreground">
              You have not placed any orders yet.
            </p>
            <Link href="/products" className={cn(buttonVariants())}>
              Start shopping
            </Link>
          </div>
        ) : (
          <ul className="mt-4 flex flex-col gap-3">
            {recentOrders.map(({ order }) => (
              <li key={order.id}>
                <Link
                  href={`/account/orders/${order.id}`}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border p-4 transition-colors hover:bg-muted/50 focus-visible:outline-2 focus-visible:outline-offset-2"
                >
                  <span className="flex flex-col">
                    <span className="font-medium">Order {order.orderNumber}</span>
                    <span className="text-xs text-muted-foreground">
                      {formatDate(order.createdAt)} ·{" "}
                      {order.items.length}{" "}
                      {order.items.length === 1 ? "item" : "items"}
                    </span>
                  </span>
                  <span className="flex items-center gap-3">
                    <OrderStatusBadge status={order.status} />
                    <span className="text-sm font-semibold tabular-nums">
                      {formatPrice(order.total)}
                    </span>
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section aria-labelledby="address-heading">
        <h2 id="address-heading" className="text-lg font-semibold">
          Default shipping address
        </h2>
        {defaultAddress ? (
          <div className="mt-4 rounded-2xl border p-5">
            <address className="text-sm not-italic leading-relaxed text-muted-foreground">
              <span className="font-medium text-foreground">
                {defaultAddress.fullName}
              </span>
              {defaultAddress.villageTown || defaultAddress.district ? (
                <>
                  <br />
                  {defaultAddress.villageTown}
                  {defaultAddress.upazila && <>, {defaultAddress.upazila} Upazila</>}
                  <br />
                  {defaultAddress.district}, {defaultAddress.division}
                  <br />
                  Bangladesh
                </>
              ) : (
                <>
                  <br />
                  {defaultAddress.line1}
                  <br />
                  {defaultAddress.city}, {defaultAddress.state}{" "}
                  {defaultAddress.postalCode}
                  <br />
                  {defaultAddress.country}
                </>
              )}
            </address>
            <Link
              href="/account/addresses"
              className="mt-3 inline-flex items-center gap-1 rounded-md text-sm font-medium text-primary-dark underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
            >
              <MapPin className="size-3.5" aria-hidden="true" />
              Manage addresses
            </Link>
          </div>
        ) : (
          <p className="mt-4 text-sm text-muted-foreground">
            No address on file yet.
          </p>
        )}
      </section>
    </div>
  );
}
