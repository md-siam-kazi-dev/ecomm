export type OrderStatus =
  | "pending"
  | "confirmed"
  | "processing"
  | "packed"
  | "shipped"
  | "delivered"
  | "cancelled"
  | "returned"
  | "refunded";

/** The customer's existing review for an order line. */
export interface OrderItemReview {
  id: string;
  rating: number;
  title: string;
  comment: string;
  /** Optional photos attached to the review. */
  images?: string[];
  /** ISO date string. */
  createdAt: string;
}

export interface OrderItem {
  productId: string;
  variantId?: string;
  /** Order-line id — the anchor for submitting a review. */
  orderItemId?: string;
  /** Product name snapshot at purchase time. */
  name: string;
  /** Variant SKU snapshot, when present. */
  sku?: string;
  quantity: number;
  /** Unit price at purchase time. */
  price: number;
  /** Primary product image. */
  imageUrl?: string;
  /** Whether the customer has already reviewed this line. */
  reviewed?: boolean;
  /** The submitted review, for prefill/view. */
  review?: OrderItemReview;
}

/** Shipping-address snapshot captured with the order. */
export interface OrderShippingAddress {
  fullName: string;
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
}

export interface Order {
  id: string;
  /** Human-readable number shown to customers, e.g. `"SK-10234"`. */
  orderNumber: string;
  customerId: string;
  items: OrderItem[];
  subtotal: number;
  shipping: number;
  tax: number;
  total: number;
  status: OrderStatus;
  /** ISO date string. */
  createdAt: string;
  /** ISO date string. */
  estimatedDelivery?: string;
  /** Shipping address the order was placed with. */
  shippingAddress?: OrderShippingAddress;
}
