/** Vendor inventory (per product variant). Mirrors the Go API's InventoryItem. */
export interface InventoryItem {
  variant_id: string;
  product_id: string;
  name: string;
  sku: string;
  category: string;
  image_url?: string;
  quantity: number;
  reserved: number;
  available: number;
  reorder_level: number;
}

export type StockFilter = "all" | "in" | "low" | "out" | "unpublished";

/** One entry in a variant's stock history (GET /api/v1/vendor/inventory/movements). */
export interface InventoryMovement {
  id: string;
  variant_id: string;
  product_id: string;
  name: string;
  sku: string;
  movement_type: string;
  quantity: number;
  quantity_before?: number;
  quantity_after?: number;
  notes?: string;
  reference_type?: string;
  order_number?: string;
  actor?: string;
  created_at: string;
}
