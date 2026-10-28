import {
  CardListSkeleton,
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
      <span className="sr-only">Loading analytics</span>
      <PageHeaderSkeleton />
      <KpiGridSkeleton />
      <ChartSkeleton />
      <CardListSkeleton count={2} />
    </div>
  );
}
