import { ListPageSkeleton } from "@/components/dashboard/skeletons";

export default function Loading() {
  return <ListPageSkeleton label="Loading audit logs" columns={5} />;
}
