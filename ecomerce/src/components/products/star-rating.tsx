import { Star } from "lucide-react";

import { cn } from "@/lib/utils";

interface StarRatingProps {
  rating?: number;
  reviewCount: number;
  className?: string;
}

/**
 * 5-star display with numeric rating and review count. Products without
 * reviews render a quiet "No reviews yet" line instead of empty stars.
 */
export function StarRating({ rating, reviewCount, className }: StarRatingProps) {
  if (!rating || reviewCount === 0) {
    return <p className={cn("text-xs text-muted-foreground", className)}>No reviews yet</p>;
  }

  return (
    <div
      className={cn("flex items-center gap-1.5", className)}
      aria-label={`Rated ${rating} out of 5 stars from ${reviewCount} reviews`}
    >
      <div className="flex" aria-hidden="true">
        {Array.from({ length: 5 }).map((_, index) => (
          <Star
            key={index}
            className={cn(
              "size-3.5",
              index < Math.round(rating)
                ? "fill-amber-400 text-amber-400"
                : "fill-muted text-muted-foreground/30"
            )}
          />
        ))}
      </div>
      <span className="text-xs text-muted-foreground">
        {rating.toFixed(1)} ({reviewCount})
      </span>
    </div>
  );
}
