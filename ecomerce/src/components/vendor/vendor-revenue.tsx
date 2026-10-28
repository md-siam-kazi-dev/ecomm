import { Package, ShoppingCart, TrendingUp } from "lucide-react";

import { BarList } from "@/components/admin/bar-list";
import { SalesChart } from "@/components/admin/sales-chart";
import { StatCard } from "@/components/admin/stat-card";
import { StatusBreakdown } from "@/components/admin/status-breakdown";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatPrice } from "@/lib/formatters";
import type { VendorRevenue as VendorRevenueData } from "@/lib/vendor";
import type { OrderStatus } from "@/types/order";

function monthLabel(month: string): string {
  const date = new Date(`${month}-01T00:00:00Z`);
  return Number.isNaN(date.getTime())
    ? month
    : date.toLocaleString("en-US", { month: "short", timeZone: "UTC" });
}

/** Store revenue overview: totals, monthly series, top products and statuses. */
export function VendorRevenue({ revenue }: { revenue: VendorRevenueData }) {
  return (
    <div className="flex flex-col gap-6">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Revenue"
          value={formatPrice(revenue.revenue)}
          count={{ to: revenue.revenue, format: "currency" }}
          icon={TrendingUp}
        />
        <StatCard
          label="Orders"
          value={String(revenue.order_count)}
          count={{ to: revenue.order_count }}
          icon={ShoppingCart}
        />
        <StatCard
          label="Units sold"
          value={String(revenue.units_sold)}
          count={{ to: revenue.units_sold }}
          icon={Package}
        />
        <StatCard
          label="Average order"
          value={formatPrice(revenue.average_order_value)}
          count={{ to: revenue.average_order_value, format: "currency" }}
          icon={TrendingUp}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Revenue trend</CardTitle>
        </CardHeader>
        <CardContent>
          <SalesChart
            points={revenue.series.map((point) => ({
              month: point.month,
              label: monthLabel(point.month),
              revenue: point.revenue,
              orders: point.orders,
            }))}
          />
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Top products by revenue</CardTitle>
          </CardHeader>
          <CardContent>
            <BarList
              items={revenue.top_products.map((product) => ({
                label: product.name,
                value: product.revenue,
                display: formatPrice(product.revenue),
              }))}
              emptyLabel="No sales recorded yet."
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Orders by status</CardTitle>
          </CardHeader>
          <CardContent>
            <StatusBreakdown
              entries={revenue.statuses.map((entry) => ({
                status: entry.status as OrderStatus,
                count: entry.count,
                revenue: entry.revenue,
              }))}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
