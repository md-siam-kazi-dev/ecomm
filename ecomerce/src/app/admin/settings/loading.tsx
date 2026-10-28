import {
  CardListSkeleton,
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
      <span className="sr-only">Loading store settings</span>
      <PageHeaderSkeleton />
      <CardListSkeleton count={2} className="md:grid-cols-1" />
    </div>
  );
}
