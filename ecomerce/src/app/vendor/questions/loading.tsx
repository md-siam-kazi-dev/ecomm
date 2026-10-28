import { ListPageSkeleton } from "@/components/dashboard/skeletons";

export default function Loading() {
  return <ListPageSkeleton label="Loading questions" columns={4} />;
}
