import type { OrderStatus } from "./order";
import type { User } from "./user";

/** Headline KPIs for the admin dashboard (PRD §24). */
export interface AdminMetrics {
  /** Non-cancelled order total. */
  revenue: number;
  orderCount: number;
  customerCount: number;
  productCount: number;
  lowStockCount: number;
  averageOrderValue: number;
  /** Total units across non-cancelled orders. */
  unitsSold: number;
  /** Month-over-month change (percentage); 0 when there is no prior month. */
  revenueChangePct: number;
  orderChangePct: number;
}

/** One bucket in the revenue chart. */
export interface RevenuePoint {
  /** `yyyy-MM`. */
  month: string;
  /** Short axis label, e.g. `"Sep"`. */
  label: string;
  revenue: number;
  orders: number;
}

export interface TopProduct {
  productId: string;
  name: string;
  unitsSold: number;
  revenue: number;
}

export interface StatusBreakdownEntry {
  status: OrderStatus;
  count: number;
  revenue: number;
}

export interface CategorySalesEntry {
  categorySlug: string;
  category: string;
  unitsSold: number;
  revenue: number;
}

/** A customer row enriched with their order aggregates (admin customers). */
export interface CustomerSummary extends User {
  orderCount: number;
  totalSpent: number;
  lastOrderAt?: string;
}
