import { CircleAlert } from "lucide-react";

import { Button } from "@/components/ui/button";

interface ErrorStateProps {
  title?: string;
  description?: string;
  onRetry?: () => void;
  retryLabel?: string;
}

/** Reusable error state with a recovery action (PRD §36). */
export function ErrorState({
  title = "Something went wrong",
  description = "An unexpected error occurred while loading this page. Please try again.",
  onRetry,
  retryLabel = "Try again",
}: ErrorStateProps) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center gap-4 rounded-2xl border border-dashed px-6 py-16 text-center"
    >
      <span className="flex size-14 items-center justify-center rounded-full bg-destructive/10 text-destructive">
        <CircleAlert className="size-7" aria-hidden="true" />
      </span>
      <div>
        <h1 className="text-lg font-semibold">{title}</h1>
        <p className="mt-1 max-w-md text-sm text-muted-foreground">
          {description}
        </p>
      </div>
      {onRetry && (
        <Button type="button" onClick={onRetry}>
          {retryLabel}
        </Button>
      )}
    </div>
  );
}
