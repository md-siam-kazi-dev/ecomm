import { ListPageSkeleton } from "@/components/dashboard/skeletons";

export default function Loading() {
  return <ListPageSkeleton label="Loading orders" columns={5} />;
}
