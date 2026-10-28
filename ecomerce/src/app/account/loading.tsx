import { Skeleton } from "@/components/ui/skeleton";

/** Route-level loading skeleton for the account section (PRD §35). */
export default function AccountLoading() {
  return (
    <div
      aria-hidden="true"
      className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8"
    >
      <Skeleton className="h-8 w-48" />
      <Skeleton className="mt-2 h-5 w-64" />
      <div className="mt-8 grid gap-8 lg:grid-cols-[220px_1fr]">
        <div className="flex flex-col gap-2">
          {Array.from({ length: 5 }).map((_, index) => (
            <Skeleton key={index} className="h-9 w-full rounded-lg" />
          ))}
        </div>
        <div className="flex flex-col gap-6">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-28 rounded-2xl" />
            ))}
          </div>
          <Skeleton className="h-64 rounded-2xl" />
        </div>
      </div>
    </div>
  );
}
