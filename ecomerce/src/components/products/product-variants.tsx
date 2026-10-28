"use client";

import { cn } from "@/lib/utils";
import type { ProductVariant } from "@/types/product";

interface ProductVariantsProps {
  /** Variant axis label, e.g. "Color" or "Size". */
  name: string;
  variants: ProductVariant[];
  selectedId?: string;
  onSelect: (variantId: string) => void;
}

/**
 * Variant chip selector for one axis (PRD §14). Out-of-stock options are
 * disabled and marked for screen readers.
 */
export function ProductVariants({
  name,
  variants,
  selectedId,
  onSelect,
}: ProductVariantsProps) {
  return (
    <fieldset>
      <legend className="mb-2 text-sm font-semibold">{name}</legend>
      <div className="flex flex-wrap gap-2">
        {variants.map((variant) => {
          const isSelected = variant.id === selectedId;
          const isSoldOut = variant.stock <= 0;

          return (
            <button
              key={variant.id}
              type="button"
              onClick={() => onSelect(variant.id)}
              disabled={isSoldOut}
              aria-pressed={isSelected}
              aria-label={`${name}: ${variant.value}${isSoldOut ? " (out of stock)" : ""}`}
              className={cn(
                "min-w-12 rounded-lg border px-3 py-1.5 text-sm transition-colors focus-visible:outline-2 focus-visible:outline-offset-2",
                isSelected
                  ? "border-primary bg-primary/10 font-medium text-primary"
                  : "hover:border-foreground/30 hover:bg-muted",
                isSoldOut && "cursor-not-allowed opacity-50 line-through"
              )}
            >
              {variant.value}
            </button>
          );
        })}
      </div>
    </fieldset>
  );
}
