import {
  ChartSkeleton,
  KpiGridSkeleton,
  PageHeaderSkeleton,
  TableSkeleton,
} from "@/components/dashboard/skeletons";

/** Route-level loading skeleton for the vendor dashboard (PRD §35). */
export default function VendorLoading() {
  return (
    <div
      role="status"
      aria-busy="true"
      aria-live="polite"
      className="flex flex-col gap-6"
    >
      <span className="sr-only">Loading dashboard</span>
      <PageHeaderSkeleton />
      <KpiGridSkeleton />
      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <ChartSkeleton />
        <ChartSkeleton />
      </div>
      <TableSkeleton rows={5} columns={4} withToolbar={false} />
    </div>
  );
}
