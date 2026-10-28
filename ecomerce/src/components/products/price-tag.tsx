import { cn } from "@/lib/utils";
import { formatPrice } from "@/lib/formatters";

interface PriceTagProps {
  price: number;
  compareAtPrice?: number;
  className?: string;
}

/**
 * Current price with struck-through compare-at price when discounted.
 * The percentage lives on the card badge — this row stays compact.
 */
export function PriceTag({ price, compareAtPrice, className }: PriceTagProps) {
  const isDiscounted = !!compareAtPrice && compareAtPrice > price;

  return (
    <p className={cn("flex items-baseline gap-1.5", className)}>
      <span className="text-base font-semibold text-foreground sm:text-lg">
        {formatPrice(price)}
      </span>
      {isDiscounted && (
        <span className="text-xs text-muted-foreground line-through">
          {formatPrice(compareAtPrice)}
        </span>
      )}
    </p>
  );
}
