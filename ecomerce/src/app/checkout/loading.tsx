import { Skeleton } from "@/components/ui/skeleton";

/** Route-level loading skeleton for /checkout (PRD §35). */
export default function CheckoutLoading() {
  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <Skeleton className="h-9 w-40" />
      <div className="mt-8 grid gap-10 lg:grid-cols-[1fr_380px]">
        <div aria-hidden="true" className="flex flex-col gap-6">
          {Array.from({ length: 4 }).map((_, index) => (
            <div key={index} className="flex flex-col gap-4 rounded-2xl border p-6">
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ))}
        </div>
        <Skeleton className="h-80 rounded-2xl" />
      </div>
    </div>
  );
}
