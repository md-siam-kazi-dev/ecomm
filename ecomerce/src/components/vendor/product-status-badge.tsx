import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type BadgeVariant =
  | "default"
  | "secondary"
  | "destructive"
  | "success"
  | "warning"
  | "info"
  | "muted";

/** Minimal shape needed to derive a product's display status. */
export interface ProductStatusInput {
  status: string;
  approval_status: string;
  submitted_at?: string | null;
}

export interface ProductStatusView {
  label: string;
  variant: BadgeVariant;
}

/**
 * Maps the backend's (status, approval_status, submitted_at) onto the vendor-
 * facing label. Draft = never submitted; Pending Review = submitted; Published
 * = approved + live.
 */
export function productStatus(product: ProductStatusInput): ProductStatusView {
  if (product.status === "active" && product.approval_status === "approved") {
    return { label: "Published", variant: "success" };
  }
  switch (product.approval_status) {
    case "approved":
      return { label: "Approved", variant: "success" };
    case "rejected":
      return { label: "Rejected", variant: "destructive" };
    case "changes_requested":
      return { label: "Changes Requested", variant: "warning" };
    default:
      return product.submitted_at
        ? { label: "Pending Review", variant: "info" }
        : { label: "Draft", variant: "muted" };
  }
}

/** Colored status pill for the vendor product list. */
export function ProductStatusBadge({
  product,
  className,
}: {
  product: ProductStatusInput;
  className?: string;
}) {
  const { label, variant } = productStatus(product);

  return (
    <Badge variant={variant} className={cn("font-medium", className)}>
      {label}
    </Badge>
  );
}
