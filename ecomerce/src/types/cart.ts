/**
 * Server cart API shapes (GET /api/v1/cart and every cart mutation return the
 * full cart, so the client never needs to reconstruct a cart locally).
 */

export interface CartApiItem {
  id: string;
  product_id: string;
  variant_id: string;
  name: string;
  slug: string;
  image_url: string;
  image_alt: string;
  variant_label?: string;
  unit_price: number;
  quantity: number;
  line_total: number;
  max_quantity: number;
  in_stock: boolean;
}

export interface CartApiResponse {
  id: string;
  currency: string;
  item_count: number;
  subtotal: number;
  items: CartApiItem[];
}
