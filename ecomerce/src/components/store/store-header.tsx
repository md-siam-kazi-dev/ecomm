import type { LucideIcon } from "lucide-react";
import Image from "next/image";
import { Package, Star, ThumbsUp, Users } from "lucide-react";

import { FollowStoreButton } from "@/components/store/follow-store-button";
import { Badge } from "@/components/ui/badge";
import type { Store } from "@/lib/stores";

function initials(name: string): string {
  return name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

function StoreStat({
  icon: Icon,
  label,
  value,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center gap-2">
      <span className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary-dark">
        <Icon className="size-4" aria-hidden="true" />
      </span>
      <span className="flex flex-col leading-tight">
        <span className="text-sm font-semibold tabular-nums">{value}</span>
        <span className="text-xs text-muted-foreground">{label}</span>
      </span>
    </div>
  );
}

/** Store profile header: banner, logo, description, stats and follow action. */
export function StoreHeader({
  store,
  initialFollowing,
}: {
  store: Store;
  initialFollowing: boolean | null;
}) {
  return (
    <header className="overflow-hidden rounded-2xl bg-background ring-1 ring-foreground/10">
      <div className="relative h-40 w-full bg-muted sm:h-56">
        {store.bannerUrl ? (
          <Image
            src={store.bannerUrl}
            alt=""
            fill
            sizes="100vw"
            priority
            className="object-cover"
          />
        ) : null}
      </div>

      <div className="px-5 pb-5 sm:px-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex items-end gap-4">
            <span className="relative -mt-10 size-20 shrink-0 overflow-hidden rounded-2xl border-4 border-background bg-muted sm:-mt-12 sm:size-24">
              {store.logoUrl ? (
                <Image
                  src={store.logoUrl}
                  alt={`${store.name} logo`}
                  fill
                  sizes="96px"
                  className="object-cover"
                />
              ) : (
                <span
                  aria-hidden="true"
                  className="flex h-full w-full items-center justify-center text-lg font-semibold text-muted-foreground"
                >
                  {initials(store.name)}
                </span>
              )}
            </span>
            <div className="min-w-0 pb-1">
              <h1 className="truncate text-xl font-bold tracking-tight sm:text-2xl">
                {store.name}
              </h1>
              {store.category && (
                <Badge variant="secondary" className="mt-1">
                  {store.category}
                </Badge>
              )}
            </div>
          </div>

          <div className="sm:pb-1">
            <FollowStoreButton
              storeSlug={store.slug}
              initialFollowing={initialFollowing}
              initialFollowerCount={store.followerCount}
            />
          </div>
        </div>

        {store.description && (
          <p className="mt-4 max-w-3xl text-sm leading-relaxed text-muted-foreground">
            {store.description}
          </p>
        )}

        <dl className="mt-5 flex flex-wrap gap-x-8 gap-y-3">
          <StoreStat
            icon={ThumbsUp}
            label="Positive feedback"
            value={`${Math.round(store.positiveFeedback)}%`}
          />
          <StoreStat
            icon={Users}
            label="Followers"
            value={store.followerCount.toLocaleString()}
          />
          <StoreStat
            icon={Package}
            label="Products"
            value={store.productCount.toLocaleString()}
          />
          <StoreStat
            icon={Star}
            label="Reviews"
            value={store.reviewCount.toLocaleString()}
          />
        </dl>
      </div>
    </header>
  );
}
