import { Skeleton } from "@/components/ui/skeleton";

export default function AccountOrderDetailLoading() {
  return (
    <div aria-hidden="true" className="flex flex-col gap-6">
      <Skeleton className="h-8 w-52" />
      <Skeleton className="h-64 rounded-2xl" />
      <div className="grid gap-4 sm:grid-cols-2">
        <Skeleton className="h-40 rounded-2xl" />
        <Skeleton className="h-40 rounded-2xl" />
      </div>
    </div>
  );
}
