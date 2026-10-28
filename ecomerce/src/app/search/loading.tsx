import { Skeleton } from "@/components/ui/skeleton";

/** Route-level loading skeleton for /search (PRD §35). */
export default function SearchLoading() {
  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <div className="mx-auto max-w-2xl text-center">
        <Skeleton className="mx-auto h-9 w-40" />
        <Skeleton className="mx-auto mt-4 h-12 w-full rounded-xl" />
      </div>
      <div className="mx-auto mt-12 max-w-5xl">
        <Skeleton className="h-5 w-48" />
        <div className="mt-5 grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-3 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <div key={index} className="flex flex-col gap-2.5">
              <Skeleton className="aspect-square rounded-xl" />
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
