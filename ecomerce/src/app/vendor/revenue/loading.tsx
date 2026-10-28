import {
  ChartSkeleton,
  KpiGridSkeleton,
  PageHeaderSkeleton,
} from "@/components/dashboard/skeletons";

export default function Loading() {
  return (
    <div
      role="status"
      aria-busy="true"
      aria-live="polite"
      className="flex flex-col gap-6"
    >
      <span className="sr-only">Loading revenue</span>
      <PageHeaderSkeleton />
      <KpiGridSkeleton />
      <ChartSkeleton />
    </div>
  );
}
