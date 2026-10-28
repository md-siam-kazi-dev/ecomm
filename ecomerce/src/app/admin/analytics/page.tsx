import type { Metadata } from "next";
import { Boxes, Receipt, TrendingUp, Users } from "lucide-react";

import { BarList } from "@/components/admin/bar-list";
import { SalesChart } from "@/components/admin/sales-chart";
import { StatCard } from "@/components/admin/stat-card";
import { StatusBreakdown } from "@/components/admin/status-breakdown";
import { PageHeader } from "@/components/dashboard/page-header";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  getAdminMetrics,
  getCategorySales,
  getCustomerSummaries,
  getOrderStatusBreakdown,
  getRevenueSeries,
  getTopProducts,
} from "@/lib/admin";
import { formatPrice } from "@/lib/formatters";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Admin Analytics",
  description: "Sales, category and customer analytics for ShopKart.",
  path: "/admin/analytics",
  noIndex: true,
});

const TOP_PRODUCT_COUNT = 6;

export default async function AdminAnalyticsPage() {
  const [metrics, revenue, categorySales, topProducts, statuses, customers] =
    await Promise.all([
      getAdminMetrics(),
      getRevenueSeries(6),
      getCategorySales(),
      getTopProducts(TOP_PRODUCT_COUNT),
      getOrderStatusBreakdown(),
      getCustomerSummaries(),
    ]);

  const activeCustomers = customers.filter(
    (customer) => customer.orderCount > 0
  ).length;

  return (
    <>
      <PageHeader
        title="Analytics"
        description="Revenue, product and customer performance from the demo dataset."
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
          label="Average order value"
          value={formatPrice(metrics.averageOrderValue)}
          count={{ to: metrics.averageOrderValue, format: "currency" }}
          icon={Receipt}
        />
        <StatCard
          label="Units sold"
          value={String(metrics.unitsSold)}
          count={{ to: metrics.unitsSold }}
          icon={Boxes}
        />
        <StatCard
          label="Active customers"
          value={`${activeCustomers}/${customers.length}`}
          icon={Users}
          hint="Customers with at least one order"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Revenue trend</CardTitle>
        </CardHeader>
        <CardContent>
          <SalesChart points={revenue} />
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Revenue by category</CardTitle>
          </CardHeader>
          <CardContent>
            <BarList
              items={categorySales.map((entry) => ({
                label: entry.category,
                value: entry.revenue,
                display: formatPrice(entry.revenue),
              }))}
              emptyLabel="No category sales yet."
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Top products</CardTitle>
          </CardHeader>
          <CardContent>
            <BarList
              items={topProducts.map((product) => ({
                label: product.name,
                value: product.revenue,
                display: `${product.unitsSold} sold · ${formatPrice(
                  product.revenue
                )}`,
              }))}
              emptyLabel="No product sales yet."
            />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Fulfilment breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <StatusBreakdown entries={statuses} />
        </CardContent>
      </Card>
    </>
  );
}
