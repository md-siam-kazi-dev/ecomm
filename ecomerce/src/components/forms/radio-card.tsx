import type { ComponentProps, ReactNode } from "react";

import { cn } from "@/lib/utils";

interface RadioCardProps {
  /** Radio group name — all cards in one group share it. */
  name: string;
  value: string;
  label: string;
  description?: string;
  /** Trailing content, e.g. a price. */
  trailing?: ReactNode;
  checked: boolean;
  /** RHF registration props (onChange/onBlur/name/ref). */
  registration: ComponentProps<"input">;
}

/**
 * Native radio input styled as a selectable card — stays fully accessible
 * (real radio semantics, arrow-key group navigation) while matching the
 * design system.
 */
export function RadioCard({
  name,
  value,
  label,
  description,
  trailing,
  checked,
  registration,
}: RadioCardProps) {
  const inputId = `radio-${name}-${value}`;

  return (
    <label
      htmlFor={inputId}
      className={cn(
        "flex cursor-pointer items-start gap-3 rounded-xl border p-4 transition-colors has-focus-visible:ring-2 has-focus-visible:ring-ring/50",
        checked ? "border-primary bg-primary/5" : "hover:bg-muted",
        "has-disabled:cursor-not-allowed has-disabled:opacity-50"
      )}
    >
      <input
        id={inputId}
        type="radio"
        name={name}
        value={value}
        checked={checked}
        className="mt-0.5 size-4 shrink-0 accent-primary"
        {...registration}
      />
      <span className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="text-sm font-medium">{label}</span>
        {description && (
          <span className="text-xs text-muted-foreground">{description}</span>
        )}
      </span>
      {trailing && (
        <span className="shrink-0 text-sm font-semibold">{trailing}</span>
      )}
    </label>
  );
}
