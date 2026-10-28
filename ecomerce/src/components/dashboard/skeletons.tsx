import type { ReactNode } from "react";

import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

interface LoadingRegionProps {
  label: string;
  children: ReactNode;
  className?: string;
}

/**
 * Wraps a skeleton set in a live region so assistive tech announces the wait
 * once, while the skeleton blocks themselves stay decorative.
 */
function LoadingRegion({ label, children, className }: LoadingRegionProps) {
  return (
    <div
      role="status"
      aria-busy="true"
      aria-live="polite"
      className={cn("flex flex-col gap-6", className)}
    >
      <span className="sr-only">{label}</span>
      {children}
    </div>
  );
}

interface PageHeaderSkeletonProps {
  withAction?: boolean;
}

export function PageHeaderSkeleton({ withAction = false }: PageHeaderSkeletonProps) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between" aria-hidden="true">
      <div className="flex flex-col gap-2">
        <Skeleton className="h-8 w-52" />
        <Skeleton className="h-5 w-80 max-w-full" />
      </div>
      {withAction && <Skeleton className="h-9 w-32" />}
    </div>
  );
}

interface KpiGridSkeletonProps {
  count?: number;
}

export function KpiGridSkeleton({ count = 4 }: KpiGridSkeletonProps) {
  return (
    <div
      className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4"
      aria-hidden="true"
    >
      {Array.from({ length: count }).map((_, index) => (
        <Skeleton key={index} className="h-32 rounded-xl" />
      ))}
    </div>
  );
}

interface ChartSkeletonProps {
  className?: string;
}

export function ChartSkeleton({ className }: ChartSkeletonProps) {
  return (
    <Skeleton className={cn("h-72 rounded-xl", className)} aria-hidden="true" />
  );
}

interface TableSkeletonProps {
  rows?: number;
  columns?: number;
  /** Show a search/filter bar above the rows. */
  withToolbar?: boolean;
  /** When set, announces the wait in a live region instead of hiding it. */
  label?: string;
}

export function TableSkeleton({
  rows = 6,
  columns = 4,
  withToolbar = true,
  label,
}: TableSkeletonProps) {
  const table = (
    <div
      aria-hidden={label ? undefined : true}
      className="flex flex-col gap-4 rounded-xl bg-card p-4 ring-1 ring-foreground/10"
    >
      {withToolbar && (
        <div className="flex flex-wrap items-center gap-3">
          <Skeleton className="h-9 flex-1 sm:max-w-72" />
          <Skeleton className="h-9 w-32" />
        </div>
      )}

      <div className="flex flex-col gap-3">
        {Array.from({ length: rows }).map((_, rowIndex) => (
          <div key={rowIndex} className="flex items-center gap-4">
            {Array.from({ length: columns }).map((__, columnIndex) => (
              <Skeleton
                key={columnIndex}
                className={cn("h-10 flex-1", columnIndex === 0 && "flex-[1.6]")}
              />
            ))}
          </div>
        ))}
      </div>
    </div>
  );

  if (!label) return table;

  return (
    <div role="status" aria-busy="true" aria-live="polite">
      <span className="sr-only">{label}</span>
      {table}
    </div>
  );
}

interface CardListSkeletonProps {
  count?: number;
  className?: string;
}

export function CardListSkeleton({ count = 4, className }: CardListSkeletonProps) {
  return (
    <div className={cn("grid gap-4 md:grid-cols-2", className)} aria-hidden="true">
      {Array.from({ length: count }).map((_, index) => (
        <Skeleton key={index} className="h-40 rounded-xl" />
      ))}
    </div>
  );
}

interface ListPageSkeletonProps {
  label: string;
  columns?: number;
  rows?: number;
  withToolbar?: boolean;
}

/** Standard header + table skeleton for the dashboard list routes. */
export function ListPageSkeleton({
  label,
  columns = 4,
  rows = 8,
  withToolbar = true,
}: ListPageSkeletonProps) {
  return (
    <div role="status" aria-busy="true" aria-live="polite" className="flex flex-col gap-6">
      <span className="sr-only">{label}</span>
      <PageHeaderSkeleton />
      <TableSkeleton rows={rows} columns={columns} withToolbar={withToolbar} />
    </div>
  );
}

export { LoadingRegion };
