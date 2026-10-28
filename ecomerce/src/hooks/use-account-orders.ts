"use client";

import { useMemo } from "react";

import { useOrderStore, type PlacedOrder } from "@/store/order-store";
import type { Order } from "@/types/order";

export interface AccountOrder {
  order: Order;
  /** Present when the order was placed in this session (has full snapshot). */
  placed?: PlacedOrder;
}

/**
 * Merges the demo order history returned by the repository with orders placed
 * during this browser session (PRD §23). Session orders win on id collisions
 * so their shipping/payment snapshot is available.
 */
export function useAccountOrders(demoOrders: Order[]): AccountOrder[] {
  const placedOrders = useOrderStore((state) => state.orders);

  return useMemo(() => {
    const demoIds = new Set(demoOrders.map((order) => order.id));
    const placedById = new Map(
      placedOrders.map((placed) => [placed.order.id, placed])
    );

    const merged: AccountOrder[] = [
      ...placedOrders
        .filter((placed) => !demoIds.has(placed.order.id))
        .map((placed) => ({ order: placed.order, placed })),
      ...demoOrders.map((order) => ({
        order,
        placed: placedById.get(order.id),
      })),
    ];

    return merged.sort((a, b) =>
      b.order.createdAt.localeCompare(a.order.createdAt)
    );
  }, [demoOrders, placedOrders]);
}
