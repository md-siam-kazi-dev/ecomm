import { Badge } from "@/components/ui/badge";
import { LOW_STOCK_THRESHOLD } from "@/lib/constants";
import { cn } from "@/lib/utils";

export type StockLevel = "out" | "low" | "in";

const LEVEL_LABELS: Record<StockLevel, string> = {
  out: "Out of stock",
  low: "Low stock",
  in: "In stock",
};

const LEVEL_VARIANTS: Record<StockLevel, "destructive" | "warning" | "success"> = {
  out: "destructive",
  low: "warning",
  in: "success",
};

export function stockLevel(stock: number): StockLevel {
  if (stock <= 0) return "out";
  if (stock <= LOW_STOCK_THRESHOLD) return "low";
  return "in";
}

/** Stock availability pill shared by the admin product/inventory tables. */
export function StockBadge({
  stock,
  className,
}: {
  stock: number;
  className?: string;
}) {
  const level = stockLevel(stock);

  return (
    <Badge variant={LEVEL_VARIANTS[level]} className={cn("font-medium", className)}>
      {LEVEL_LABELS[level]}
    </Badge>
  );
}
