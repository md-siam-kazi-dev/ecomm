import type { Metadata } from "next";
import Link from "next/link";
import { Package, ShoppingCart, TrendingUp, Users } from "lucide-react";

import { BarList } from "@/components/admin/bar-list";
import { SalesChart } from "@/components/admin/sales-chart";
import { StatCard } from "@/components/admin/stat-card";
import { StatusBreakdown } from "@/components/admin/status-breakdown";
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
import {
  getAdminMetrics,
  getOrderStatusBreakdown,
  getRevenueSeries,
  getTopProducts,
} from "@/lib/admin";
import { formatPrice } from "@/lib/formatters";
import { getOrders } from "@/lib/orders";
import { getLowStockProducts } from "@/lib/products";
import { buildMetadata } from "@/lib/seo";
import { getCustomers } from "@/lib/users";

export const metadata: Metadata = buildMetadata({
  title: "Admin Dashboard",
  description: "Revenue, orders, customers and inventory at a glance.",
  path: "/admin",
  noIndex: true,
});

const RECENT_ORDER_COUNT = 5;
const TOP_PRODUCT_COUNT = 5;
const LOW_STOCK_PREVIEW = 6;

export default async function AdminDashboardPage() {
  const [metrics, revenue, statuses, topProducts, orders, customers, lowStock] =
    await Promise.all([
      getAdminMetrics(),
      getRevenueSeries(6),
      getOrderStatusBreakdown(),
      getTopProducts(TOP_PRODUCT_COUNT),
      getOrders(),
      getCustomers(),
      getLowStockProducts(),
    ]);

  const customerNames = new Map(
    customers.map((customer) => [customer.id, customer.name])
  );
  const recentOrders = orders.slice(0, RECENT_ORDER_COUNT);

  return (
    <>
      <PageHeader
        title="Dashboard"
        description="Store performance at a glance. All figures come from the demo data."
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Revenue"
          value={formatPrice(metrics.revenue)}
          count={{ to: metrics.revenue, format: "currency" }}
          icon={TrendingUp}
          trend={metrics.revenueChangePct}
        />
        <StatCard
          label="Orders"
          value={String(metrics.orderCount)}
          count={{ to: metrics.orderCount }}
          icon={ShoppingCart}
          trend={metrics.orderChangePct}
          hint={`${formatPrice(metrics.averageOrderValue)} average order`}
        />
        <StatCard
          label="Customers"
          value={String(metrics.customerCount)}
          count={{ to: metrics.customerCount }}
          icon={Users}
        />
        <StatCard
          label="Products"
          value={String(metrics.productCount)}
          count={{ to: metrics.productCount }}
          icon={Package}
          hint={`${metrics.lowStockCount} low on stock`}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Sales overview</CardTitle>
          </CardHeader>
          <CardContent>
            <SalesChart points={revenue} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Orders by status</CardTitle>
          </CardHeader>
          <CardContent>
            <StatusBreakdown entries={statuses} />
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Recent orders</CardTitle>
            <CardAction>
              <Button variant="ghost" size="sm" asChild>
                <Link href="/admin/orders">View all</Link>
              </Button>
            </CardAction>
          </CardHeader>
          <CardContent>
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Order</TableHead>
                    <TableHead className="hidden sm:table-cell">Customer</TableHead>
                    <TableHead>Total</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {recentOrders.map((order) => (
                    <TableRow key={order.id}>
                      <TableCell className="font-medium">
                        {order.orderNumber}
                      </TableCell>
                      <TableCell className="hidden text-muted-foreground sm:table-cell">
                        {customerNames.get(order.customerId) ?? order.customerId}
                      </TableCell>
                      <TableCell className="tabular-nums">
                        {formatPrice(order.total)}
                      </TableCell>
                      <TableCell>
                        <OrderStatusBadge status={order.status} />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <ul className="flex flex-col gap-3 md:hidden">
              {recentOrders.map((order) => (
                <li
                  key={order.id}
                  className="rounded-xl bg-muted/50 p-4 ring-1 ring-foreground/5"
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="min-w-0 truncate font-medium">
                      {order.orderNumber}
                    </p>
                    <OrderStatusBadge status={order.status} />
                  </div>

                  <dl className="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                    <div className="flex flex-col">
                      <dt>Customer</dt>
                      <dd className="truncate text-foreground">
                        {customerNames.get(order.customerId) ?? order.customerId}
                      </dd>
                    </div>
                    <div className="flex flex-col">
                      <dt>Total</dt>
                      <dd className="font-medium text-foreground tabular-nums">
                        {formatPrice(order.total)}
                      </dd>
                    </div>
                  </dl>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Low stock</CardTitle>
            <CardAction>
              <Button variant="ghost" size="sm" asChild>
                <Link href="/admin/inventory">Manage</Link>
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
                    <Badge variant="warning" className="shrink-0 tabular-nums">
                      {product.stock} left
                    </Badge>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Top products by revenue</CardTitle>
        </CardHeader>
        <CardContent>
          <BarList
            items={topProducts.map((product) => ({
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
