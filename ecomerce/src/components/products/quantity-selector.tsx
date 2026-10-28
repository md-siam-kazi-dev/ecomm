"use client";

import { Minus, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";

interface QuantitySelectorProps {
  value: number;
  onChange: (value: number) => void;
  /** Maximum allowed quantity (usually stock). */
  max?: number;
  min?: number;
  label?: string;
}

/** Accessible quantity stepper (PRD §14) — reused by the cart later. */
export function QuantitySelector({
  value,
  onChange,
  max = 99,
  min = 1,
  label = "Quantity",
}: QuantitySelectorProps) {
  const clampedMax = Math.max(min, max);

  return (
    <div
      role="group"
      aria-label={label}
      className="inline-flex items-center rounded-lg border"
    >
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => onChange(Math.max(min, value - 1))}
        disabled={value <= min}
        aria-label="Decrease quantity"
        className="size-10 rounded-r-none"
      >
        <Minus className="size-4" aria-hidden="true" />
      </Button>
      <span
        aria-live="polite"
        aria-label={`${label}: ${value}`}
        className="w-10 text-center text-sm font-medium tabular-nums"
      >
        {value}
      </span>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={() => onChange(Math.min(clampedMax, value + 1))}
        disabled={value >= clampedMax}
        aria-label="Increase quantity"
        className="size-10 rounded-l-none"
      >
        <Plus className="size-4" aria-hidden="true" />
      </Button>
    </div>
  );
}
