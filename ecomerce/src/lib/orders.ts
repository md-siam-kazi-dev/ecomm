import { ORDERS } from "@/data/orders";
import type { Order, OrderStatus } from "@/types/order";

/**
 * Orders repository — demo data now, Go API later (PRD §44).
 */

/** All orders, newest first (admin dashboard). */
export async function getOrders(): Promise<Order[]> {
  return [...ORDERS].sort((a, b) => b.createdAt.localeCompare(a.createdAt));
}

export async function getOrdersByStatus(status: OrderStatus): Promise<Order[]> {
  return (await getOrders()).filter((order) => order.status === status);
}

export async function getOrderByNumber(
  orderNumber: string
): Promise<Order | null> {
  return ORDERS.find((order) => order.orderNumber === orderNumber) ?? null;
}

export async function getOrderById(id: string): Promise<Order | null> {
  return ORDERS.find((order) => order.id === id) ?? null;
}

/** Order history for the account pages (newest first). */
export async function getOrdersByCustomer(customerId: string): Promise<Order[]> {
  return (await getOrders()).filter((order) => order.customerId === customerId);
}
