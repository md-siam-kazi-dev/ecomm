"use client";

import { ErrorState } from "@/components/errors/error-state";

/** Error boundary for the vendor area (PRD §54) — renders inside DashboardShell. */
export default function VendorError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <ErrorState
      title="Could not load the vendor dashboard"
      description="This view hit an unexpected error. Retry, or come back in a moment."
      onRetry={reset}
    />
  );
}
