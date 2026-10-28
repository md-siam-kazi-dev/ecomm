import { getOrders } from "@/lib/orders";
import { getLowStockProducts, getProducts } from "@/lib/products";
import { getCustomers } from "@/lib/users";
import type {
  AdminMetrics,
  CategorySalesEntry,
  CustomerSummary,
  RevenuePoint,
  StatusBreakdownEntry,
  TopProduct,
} from "@/types/admin";
import type { Order, OrderStatus } from "@/types/order";

/**
 * Admin analytics repository (PRD §24/§44). Every figure is aggregated from
 * the demo orders/products/customers — nothing is fabricated. Swap these
 * functions for Go API calls later without touching the admin UI.
 */

const ORDER_STATUSES: OrderStatus[] = [
  "pending",
  "processing",
  "shipped",
  "delivered",
  "cancelled",
];

function round(value: number): number {
  return Math.round(value * 100) / 100;
}

function monthKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
}

function isRevenue(order: Order): boolean {
  return order.status !== "cancelled";
}

/** Monthly revenue + order counts for the dashboard chart (oldest first). */
export async function getRevenueSeries(months = 6): Promise<RevenuePoint[]> {
  const orders = (await getOrders()).filter(isRevenue);
  const now = new Date();
  const points: RevenuePoint[] = [];

  for (let offset = months - 1; offset >= 0; offset -= 1) {
    const date = new Date(now.getFullYear(), now.getMonth() - offset, 1);
    points.push({
      month: monthKey(date),
      label: date.toLocaleString("en-US", { month: "short" }),
      revenue: 0,
      orders: 0,
    });
  }

  const byMonth = new Map(points.map((point) => [point.month, point]));
  for (const order of orders) {
    const point = byMonth.get(monthKey(new Date(order.createdAt)));
    if (!point) continue;
    point.revenue = round(point.revenue + order.total);
    point.orders += 1;
  }

  return points;
}

/** Best-selling products ranked by revenue across non-cancelled orders. */
export async function getTopProducts(limit = 5): Promise<TopProduct[]> {
  const orders = (await getOrders()).filter(isRevenue);
  const tally = new Map<
    string,
    { name: string; unitsSold: number; revenue: number }
  >();

  for (const order of orders) {
    for (const item of order.items) {
      // Orders carry a product-name snapshot, so no catalog lookup is needed.
      const entry = tally.get(item.productId) ?? {
        name: item.name,
        unitsSold: 0,
        revenue: 0,
      };
      entry.unitsSold += item.quantity;
      entry.revenue = round(entry.revenue + item.price * item.quantity);
      tally.set(item.productId, entry);
    }
  }

  return [...tally.entries()]
    .map(([productId, totals]) => ({ productId, ...totals }))
    .sort((a, b) => b.revenue - a.revenue)
    .slice(0, limit);
}

/** Order counts and revenue by status, in lifecycle order. */
export async function getOrderStatusBreakdown(): Promise<
  StatusBreakdownEntry[]
> {
  const orders = await getOrders();

  return ORDER_STATUSES.map((status) => {
    const matching = orders.filter((order) => order.status === status);
    return {
      status,
      count: matching.length,
      revenue: round(matching.reduce((sum, order) => sum + order.total, 0)),
    };
  });
}

/**
 * Revenue and units grouped by product category. Demo order items carry no
 * category, so this is empty until real orders/back-office data exist.
 */
export async function getCategorySales(): Promise<CategorySalesEntry[]> {
  return [];
}

/** Customers with their lifetime order count, spend and latest order date. */
export async function getCustomerSummaries(): Promise<CustomerSummary[]> {
  const [customers, orders] = await Promise.all([getCustomers(), getOrders()]);

  return customers
    .map((customer) => {
      const customerOrders = orders.filter(
        (order) => order.customerId === customer.id && isRevenue(order)
      );
      const lastOrderAt = customerOrders
        .map((order) => order.createdAt)
        .sort()
        .at(-1);

      return {
        ...customer,
        orderCount: customerOrders.length,
        totalSpent: round(
          customerOrders.reduce((sum, order) => sum + order.total, 0)
        ),
        lastOrderAt,
      };
    })
    .sort((a, b) => b.totalSpent - a.totalSpent);
}

/** Dashboard headline metrics (PRD §24). */
export async function getAdminMetrics(): Promise<AdminMetrics> {
  const [orders, customers, productResult, lowStock, series] = await Promise.all(
    [
      getOrders(),
      getCustomers(),
      getProducts({ pageSize: 1 }),
      getLowStockProducts(),
      getRevenueSeries(2),
    ]
  );

  const revenueOrders = orders.filter(isRevenue);
  const revenue = round(
    revenueOrders.reduce((sum, order) => sum + order.total, 0)
  );
  const unitsSold = revenueOrders.reduce(
    (sum, order) =>
      sum + order.items.reduce((units, item) => units + item.quantity, 0),
    0
  );

  const previous = series[0];
  const current = series[series.length - 1];
  const changePct = (from: number, to: number) =>
    from > 0 ? round(((to - from) / from) * 100) : 0;

  return {
    revenue,
    orderCount: orders.length,
    customerCount: customers.length,
    productCount: productResult.total,
    lowStockCount: lowStock.length,
    averageOrderValue:
      revenueOrders.length > 0 ? round(revenue / revenueOrders.length) : 0,
    unitsSold,
    revenueChangePct: changePct(previous.revenue, current.revenue),
    orderChangePct: changePct(previous.orders, current.orders),
  };
}
