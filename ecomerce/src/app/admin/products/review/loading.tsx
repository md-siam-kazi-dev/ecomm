import { ListPageSkeleton } from "@/components/dashboard/skeletons";

export default function Loading() {
  return (
    <ListPageSkeleton label="Loading the review queue" columns={4} withToolbar={false} />
  );
}
