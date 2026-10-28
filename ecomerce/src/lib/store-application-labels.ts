/** Client-safe label maps for store-application statuses (no server imports). */

export const APPLICATION_STATUS_LABELS: Record<string, string> = {
  pending: "Under review",
  approved: "Approved",
  rejected: "Not approved",
};

export const PAYMENT_STATUS_LABELS: Record<string, string> = {
  pending: "Payment pending",
  processing: "Processing",
  paid: "Paid",
  failed: "Failed",
  cancelled: "Cancelled",
  partially_refunded: "Partially refunded",
  refunded: "Refunded",
};

export const BUSINESS_TYPE_LABELS: Record<string, string> = {
  individual: "Individual",
  registered: "Registered business",
};
