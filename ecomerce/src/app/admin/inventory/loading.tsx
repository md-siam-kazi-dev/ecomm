import {
  KpiGridSkeleton,
  PageHeaderSkeleton,
  TableSkeleton,
} from "@/components/dashboard/skeletons";

export default function Loading() {
  return (
    <div
      role="status"
      aria-busy="true"
      aria-live="polite"
      className="flex flex-col gap-6"
    >
      <span className="sr-only">Loading inventory</span>
      <PageHeaderSkeleton />
      <KpiGridSkeleton count={2} />
      <TableSkeleton rows={8} columns={5} withToolbar={false} />
    </div>
  );
}
