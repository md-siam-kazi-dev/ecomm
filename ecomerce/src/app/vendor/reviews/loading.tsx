import { ListPageSkeleton } from "@/components/dashboard/skeletons";

export default function Loading() {
  return (
    <ListPageSkeleton label="Loading reviews" columns={3} withToolbar={false} />
  );
}
