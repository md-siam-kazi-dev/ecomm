import { Skeleton } from "@/components/ui/skeleton";

export default function Loading() {
  return (
    <div className="flex flex-col gap-6" role="status" aria-label="Loading your store">
      <Skeleton className="h-8 w-40" />
      <div className="rounded-2xl bg-background p-6 ring-1 ring-foreground/10">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="mt-2 h-7 w-56" />
        <Skeleton className="mt-6 h-4 w-full max-w-md" />
      </div>
      <div className="rounded-2xl bg-background p-6 ring-1 ring-foreground/10">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="mt-4 h-24 w-full" />
      </div>
    </div>
  );
}
