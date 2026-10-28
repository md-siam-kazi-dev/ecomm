"use client";

import { ErrorState } from "@/components/errors/error-state";

/** Error boundary for the admin area (PRD §54) — renders inside DashboardShell. */
export default function AdminError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <ErrorState
      title="Could not load the dashboard"
      description="The admin view hit an unexpected error. Retry, or come back in a moment."
      onRetry={reset}
    />
  );
}
