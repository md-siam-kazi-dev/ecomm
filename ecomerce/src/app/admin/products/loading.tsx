import { ListPageSkeleton } from "@/components/dashboard/skeletons";

export default function Loading() {
  return (
    <ListPageSkeleton label="Loading products" columns={5} withToolbar={false} />
  );
}
