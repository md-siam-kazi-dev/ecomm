"use client";

import { useState, type ComponentProps } from "react";
import { Eye, EyeOff } from "lucide-react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface TextFieldProps extends ComponentProps<typeof Input> {
  label: string;
  /** Validation message shown below the field (PRD §18). */
  error?: string;
  hint?: string;
}

/**
 * Labeled input with accessible error/hint messaging (PRD §38/§18).
 *
 * Password fields get a show/hide toggle; the button never submits the form and
 * exposes its state through `aria-pressed` + a labelled action.
 */
export function TextField({
  label,
  error,
  hint,
  type,
  className,
  ...inputProps
}: TextFieldProps) {
  const [revealed, setRevealed] = useState(false);

  const fieldId =
    inputProps.id ?? `field-${String(inputProps.name ?? label).toLowerCase()}`;
  const isPassword = type === "password";
  const resolvedType = isPassword && revealed ? "text" : type;

  const input = (
    <Input
      id={fieldId}
      type={resolvedType}
      aria-invalid={error ? true : undefined}
      aria-describedby={error ? `${fieldId}-error` : undefined}
      className={cn(isPassword && "pr-10", className)}
      {...inputProps}
    />
  );

  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={fieldId}>{label}</Label>

      {isPassword ? (
        <div className="relative">
          {input}
          <button
            type="button"
            onClick={() => setRevealed((value) => !value)}
            aria-label={revealed ? "Hide password" : "Show password"}
            aria-pressed={revealed}
            aria-controls={fieldId}
            className="absolute inset-y-0 right-0 flex items-center rounded-md px-3 text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            {revealed ? (
              <EyeOff className="size-4" aria-hidden="true" />
            ) : (
              <Eye className="size-4" aria-hidden="true" />
            )}
          </button>
        </div>
      ) : (
        input
      )}

      {error ? (
        <p
          id={`${fieldId}-error`}
          role="alert"
          className="text-xs font-medium text-destructive"
        >
          {error}
        </p>
      ) : hint ? (
        <p className="text-xs text-muted-foreground">{hint}</p>
      ) : null}
    </div>
  );
}
