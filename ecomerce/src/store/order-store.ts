import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

import type { Order } from "@/types/order";

export interface ShippingAddressSnapshot {
  fullName: string;
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
}

export interface PlacedOrder {
  order: Order;
  email: string;
  shippingAddress: ShippingAddressSnapshot;
  deliveryLabel: string;
  paymentLabel: string;
}

/** Keeps the demo session from growing without bound. */
const MAX_STORED_ORDERS = 50;

interface OrderState {
  /** Orders placed in this browser session, newest first (PRD §19/§49). */
  orders: PlacedOrder[];
  /** Records a newly placed order at the top of the history. */
  placeOrder: (placedOrder: PlacedOrder) => void;
  clearOrders: () => void;
}

/**
 * Orders placed during the current session (PRD §19/§23). Session-scoped on
 * purpose (PRD §51) — it survives navigation and refresh but disappears when
 * the browser session ends. Account history merges these with the demo
 * orders returned by the repository.
 */
export const useOrderStore = create<OrderState>()(
  persist(
    (set) => ({
      orders: [],
      placeOrder: (placedOrder) =>
        set((state) => ({
          orders: [placedOrder, ...state.orders].slice(0, MAX_STORED_ORDERS),
        })),
      clearOrders: () => set({ orders: [] }),
    }),
    {
      name: "shopkart-orders",
      version: 1,
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({ orders: state.orders }),
    }
  )
);

/** Most recently placed order — drives the confirmation page (PRD §19). */
export const selectLastOrder = (state: OrderState): PlacedOrder | null =>
  state.orders[0] ?? null;

/** Looks up a session order by its id. */
export const selectOrderById =
  (orderId: string) =>
  (state: OrderState): PlacedOrder | undefined =>
    state.orders.find((placed) => placed.order.id === orderId);
