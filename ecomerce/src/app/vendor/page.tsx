import type { Metadata } from "next";
import Link from "next/link";
import { Package, ShoppingCart, TrendingUp, Truck } from "lucide-react";

import { BarList } from "@/components/admin/bar-list";
import { SalesChart } from "@/components/admin/sales-chart";
import { StatCard } from "@/components/admin/stat-card";
import { StatusBreakdown } from "@/components/admin/status-breakdown";
import { StockBadge } from "@/components/admin/stock-badge";
import { PageHeader } from "@/components/dashboard/page-header";
import { OrderStatusBadge } from "@/components/orders/order-status-badge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatDate, formatPrice } from "@/lib/formatters";
import { buildMetadata } from "@/lib/seo";
import {
  getVendorLowStock,
  getVendorOrders,
  getVendorRevenue,
  getVendorSummary,
} from "@/lib/vendor";
import type { OrderStatus } from "@/types/order";

export const metadata: Metadata = buildMetadata({
  title: "Vendor Dashboard",
  description: "Your products, orders, inventory and revenue at a glance.",
  path: "/vendor",
  noIndex: true,
});

const RECENT_ORDER_COUNT = 5;
const LOW_STOCK_PREVIEW = 6;
const KNOWN_ORDER_STATUSES = new Set([
  "pending",
  "processing",
  "shipped",
  "delivered",
  "cancelled",
]);

function monthLabel(month: string): string {
  const date = new Date(`${month}-01T00:00:00Z`);
  return Number.isNaN(date.getTime())
    ? month
    : date.toLocaleString("en-US", { month: "short", timeZone: "UTC" });
}

function StatusPill({ status }: { status: string }) {
  return KNOWN_ORDER_STATUSES.has(status) ? (
    <OrderStatusBadge status={status as OrderStatus} />
  ) : (
    <Badge variant="secondary" className="capitalize">
      {status}
    </Badge>
  );
}

export default async function VendorDashboardPage() {
  const [summary, revenue, orders, lowStock] = await Promise.all([
    getVendorSummary(),
    getVendorRevenue(),
    getVendorOrders(),
    getVendorLowStock(),
  ]);

  const recentOrders = orders.slice(0, RECENT_ORDER_COUNT);

  return (
    <>
      <PageHeader
        title="Dashboard"
        description="Your store performance at a glance. Figures cover your products only."
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Your revenue"
          value={formatPrice(summary?.revenue ?? 0)}
          count={{ to: summary?.revenue ?? 0, format: "currency" }}
          icon={TrendingUp}
        />
        <StatCard
          label="Orders"
          value={String(summary?.order_count ?? 0)}
          count={{ to: summary?.order_count ?? 0 }}
          icon={ShoppingCart}
        />
        <StatCard
          label="Pending fulfilment"
          value={String(summary?.pending_fulfilment ?? 0)}
          count={{ to: summary?.pending_fulfilment ?? 0 }}
          icon={Truck}
        />
        <StatCard
          label="Your products"
          value={String(summary?.product_count ?? 0)}
          count={{ to: summary?.product_count ?? 0 }}
          icon={Package}
          hint={`${summary?.low_stock_count ?? 0} low on stock`}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Sales overview</CardTitle>
          </CardHeader>
          <CardContent>
            <SalesChart
              points={(revenue?.series ?? []).map((point) => ({
                month: point.month,
                label: monthLabel(point.month),
                revenue: point.revenue,
                orders: point.orders,
              }))}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Orders by status</CardTitle>
          </CardHeader>
          <CardContent>
            <StatusBreakdown
              entries={(revenue?.statuses ?? []).map((entry) => ({
                status: entry.status as OrderStatus,
                count: entry.count,
                revenue: entry.revenue,
              }))}
            />
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Recent orders</CardTitle>
            <CardAction>
              <Button variant="ghost" size="sm" asChild>
                <Link href="/vendor/orders">Fulfil orders</Link>
              </Button>
            </CardAction>
          </CardHeader>
          <CardContent>
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Order</TableHead>
                    <TableHead className="hidden sm:table-cell">Date</TableHead>
                    <TableHead>Items</TableHead>
                    <TableHead>Your revenue</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {recentOrders.length === 0 ? (
                    <TableRow>
                      <TableCell
                        colSpan={5}
                        className="py-8 text-center text-sm text-muted-foreground"
                      >
                        No orders yet.
                      </TableCell>
                    </TableRow>
                  ) : (
                    recentOrders.map((order) => (
                      <TableRow key={order.id}>
                        <TableCell className="font-medium">
                          {order.order_number}
                        </TableCell>
                        <TableCell className="hidden text-muted-foreground sm:table-cell">
                          {formatDate(order.created_at)}
                        </TableCell>
                        <TableCell className="tabular-nums">
                          {order.item_count}
                        </TableCell>
                        <TableCell className="tabular-nums">
                          {formatPrice(order.total)}
                        </TableCell>
                        <TableCell>
                          <StatusPill status={order.status} />
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>

            <ul className="flex flex-col divide-y md:hidden">
              {recentOrders.length === 0 ? (
                <li className="py-8 text-center text-sm text-muted-foreground">
                  No orders yet.
                </li>
              ) : (
                recentOrders.map((order) => (
                  <li key={order.id} className="flex flex-col gap-2 py-3">
                    <div className="flex items-start justify-between gap-3">
                      <p className="min-w-0 truncate font-medium">
                        {order.order_number}
                      </p>
                      <StatusPill status={order.status} />
                    </div>

                    <dl className="grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                      <div className="flex flex-col">
                        <dt>Date</dt>
                        <dd className="text-foreground">
                          {formatDate(order.created_at)}
                        </dd>
                      </div>
                      <div className="flex flex-col">
                        <dt>Items</dt>
                        <dd className="text-foreground tabular-nums">
                          {order.item_count}
                        </dd>
                      </div>
                      <div className="flex flex-col">
                        <dt>Your revenue</dt>
                        <dd className="text-foreground tabular-nums">
                          {formatPrice(order.total)}
                        </dd>
                      </div>
                    </dl>
                  </li>
                ))
              )}
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Low stock</CardTitle>
            <CardAction>
              <Button variant="ghost" size="sm" asChild>
                <Link href="/vendor/inventory">Adjust stock</Link>
              </Button>
            </CardAction>
          </CardHeader>
          <CardContent>
            {lowStock.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                Everything is well stocked.
              </p>
            ) : (
              <ul className="flex flex-col gap-4">
                {lowStock.slice(0, LOW_STOCK_PREVIEW).map((product) => (
                  <li
                    key={product.id}
                    className="flex items-center justify-between gap-4"
                  >
                    <span className="min-w-0">
                      <span className="block truncate text-sm font-medium">
                        {product.name}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {product.sku}
                      </span>
                    </span>
                    <StockBadge stock={product.stock} className="shrink-0" />
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Your top products by revenue</CardTitle>
        </CardHeader>
        <CardContent>
          <BarList
            items={(revenue?.top_products ?? []).map((product) => ({
              label: product.name,
              value: product.revenue,
              display: formatPrice(product.revenue),
            }))}
            emptyLabel="No sales recorded yet."
          />
        </CardContent>
      </Card>
    </>
  );
}
