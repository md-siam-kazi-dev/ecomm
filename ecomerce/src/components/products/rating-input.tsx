"use client";

import { useState } from "react";
import { Star } from "lucide-react";

import { cn } from "@/lib/utils";

interface RatingInputProps {
  value: number;
  onChange: (value: number) => void;
  disabled?: boolean;
  id?: string;
}

/**
 * Accessible interactive 1–5 star picker (radiogroup semantics). Supports
 * hover/focus preview and keyboard selection.
 */
export function RatingInput({ value, onChange, disabled, id }: RatingInputProps) {
  const [preview, setPreview] = useState(0);
  const active = preview || value;

  return (
    <div
      id={id}
      role="radiogroup"
      aria-label="Star rating"
      className="flex items-center gap-1"
    >
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          role="radio"
          aria-checked={value === star}
          aria-label={`${star} star${star > 1 ? "s" : ""}`}
          disabled={disabled}
          onClick={() => onChange(star)}
          onMouseEnter={() => setPreview(star)}
          onMouseLeave={() => setPreview(0)}
          onFocus={() => setPreview(star)}
          onBlur={() => setPreview(0)}
          className="rounded-md p-0.5 transition-transform hover:scale-110 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-50"
        >
          <Star
            className={cn(
              "size-7 transition-colors",
              star <= active
                ? "fill-amber-400 text-amber-400"
                : "fill-muted text-muted-foreground/30"
            )}
            aria-hidden="true"
          />
        </button>
      ))}
    </div>
  );
}
