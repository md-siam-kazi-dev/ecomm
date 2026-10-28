"use client";

import { useState } from "react";
import Image from "next/image";
import { BadgeCheck, Loader2, MessageSquareReply, Star } from "lucide-react";

import { StarRating } from "@/components/products/star-rating";
import { Button } from "@/components/ui/button";
import { apiGet } from "@/lib/api-client";
import {
  toReview,
  type ApiReview,
  type PaginationMeta,
  type ProductRatingSummary,
} from "@/lib/catalog";
import { formatDate } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { Review } from "@/types/review";

interface ProductReviewsProps {
  productSlug: string;
  reviews: Review[];
  pagination: PaginationMeta | null;
  summary: ProductRatingSummary | null;
  rating?: number;
  reviewCount: number;
}

const PAGE_SIZE = 10;

function initialsOf(author: string): string {
  return author
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

function StarRow({ rating }: { rating: number }) {
  return (
    <span
      role="img"
      aria-label={`${rating} out of 5 stars`}
      className="flex gap-0.5"
    >
      {Array.from({ length: 5 }).map((_, starIndex) => (
        <Star
          key={starIndex}
          className={cn(
            "size-4",
            starIndex < rating
              ? "fill-amber-400 text-amber-400"
              : "fill-muted text-muted-foreground/30"
          )}
          aria-hidden="true"
        />
      ))}
    </span>
  );
}

/**
 * Reviews section (PRD §14): rating summary + distribution, review cards with
 * vendor replies, and incremental loading.
 */
export function ProductReviews({
  productSlug,
  reviews,
  pagination,
  summary,
  rating,
  reviewCount,
}: ProductReviewsProps) {
  const [items, setItems] = useState<Review[]>(reviews);
  const [page, setPage] = useState(pagination?.page ?? 1);
  const [hasNext, setHasNext] = useState(pagination?.has_next ?? false);
  const [loading, setLoading] = useState(false);

  const total = summary?.total ?? reviewCount;
  const average = summary?.total ? summary.average : rating;

  async function loadMore() {
    setLoading(true);
    const next = page + 1;
    const res = await apiGet<ApiReview[]>(
      `/api/v1/products/${encodeURIComponent(productSlug)}/reviews?page=${next}&limit=${PAGE_SIZE}`
    );
    setLoading(false);
    if (!res.ok) return;
    setItems((prev) => [...prev, ...res.data.map(toReview)]);
    setPage(next);
    setHasNext(res.pagination?.has_next ?? false);
  }

  return (
    <section
      id="reviews"
      aria-labelledby="reviews-heading"
      className="mt-12 scroll-mt-24"
    >
      <h2 id="reviews-heading" className="text-2xl font-semibold tracking-tight">
        Reviews
      </h2>

      {total > 0 && average ? (
        <div className="mt-6 grid gap-6 sm:grid-cols-[minmax(0,200px)_1fr] sm:items-center">
          <div className="flex flex-col items-start gap-1">
            <span className="text-5xl font-bold tabular-nums">
              {average.toFixed(1)}
            </span>
            <StarRating rating={average} reviewCount={total} />
            <p className="text-xs text-muted-foreground">
              Based on {total} {total === 1 ? "review" : "reviews"}
            </p>
          </div>

          <ul className="flex flex-col gap-2">
            {[5, 4, 3, 2, 1].map((star) => {
              const count = summary?.distribution?.[String(star)] ?? 0;
              const pct = total > 0 ? (count / total) * 100 : 0;
              return (
                <li key={star} className="flex items-center gap-3 text-xs">
                  <span className="flex w-12 shrink-0 items-center gap-1 tabular-nums text-muted-foreground">
                    {star}
                    <Star
                      className="size-3 fill-amber-400 text-amber-400"
                      aria-hidden="true"
                    />
                  </span>
                  <span
                    aria-hidden="true"
                    className="h-2 flex-1 overflow-hidden rounded-full bg-muted"
                  >
                    <span
                      className="block h-full rounded-full bg-amber-400 transition-[width] duration-500"
                      style={{ width: `${pct}%` }}
                    />
                  </span>
                  <span className="w-8 shrink-0 text-right tabular-nums text-muted-foreground">
                    {count}
                  </span>
                </li>
              );
            })}
          </ul>
        </div>
      ) : null}

      {items.length === 0 ? (
        <p className="mt-6 rounded-xl border border-dashed p-8 text-center text-sm text-muted-foreground">
          No reviews yet — be the first to review this product.
        </p>
      ) : (
        <>
          <ul className="mt-6 flex flex-col gap-4">
            {items.map((review) => (
              <li key={review.id} className="rounded-xl border bg-card p-5">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="flex items-center gap-3">
                    {review.avatarUrl ? (
                      <span className="relative size-9 shrink-0 overflow-hidden rounded-full bg-muted">
                        <Image
                          src={review.avatarUrl}
                          alt=""
                          fill
                          sizes="36px"
                          className="object-cover"
                        />
                      </span>
                    ) : (
                      <span
                        aria-hidden="true"
                        className="flex size-9 items-center justify-center rounded-full bg-primary-light text-xs font-semibold text-accent-foreground"
                      >
                        {initialsOf(review.author)}
                      </span>
                    )}
                    <div>
                      <p className="text-sm font-medium">{review.author}</p>
                      <p className="flex items-center gap-1 text-xs text-muted-foreground">
                        {review.isVerifiedPurchase && (
                          <>
                            <BadgeCheck
                              className="size-3.5 text-primary"
                              aria-hidden="true"
                            />
                            Verified purchase ·
                          </>
                        )}
                        {formatDate(review.createdAt)}
                      </p>
                    </div>
                  </div>
                  <StarRow rating={review.rating} />
                </div>

                {review.title && (
                  <h3 className="mt-3 text-sm font-semibold">{review.title}</h3>
                )}
                <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                  {review.comment}
                </p>

                {review.images && review.images.length > 0 && (
                  <ul className="mt-3 flex flex-wrap gap-2">
                    {review.images.map((url) => (
                      <li
                        key={url}
                        className="relative size-20 overflow-hidden rounded-lg border"
                      >
                        <Image
                          src={url}
                          alt="Review photo"
                          fill
                          sizes="80px"
                          className="object-cover"
                        />
                      </li>
                    ))}
                  </ul>
                )}

                {review.vendorReply && (
                  <div className="mt-3 rounded-lg border-l-2 border-primary bg-muted/50 p-3">
                    <p className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
                      <MessageSquareReply
                        className="size-3.5"
                        aria-hidden="true"
                      />
                      Reply from the seller
                      {review.vendorRepliedAt
                        ? ` · ${formatDate(review.vendorRepliedAt)}`
                        : ""}
                    </p>
                    <p className="mt-1 text-sm">{review.vendorReply}</p>
                  </div>
                )}

                {review.helpfulCount > 0 && (
                  <p className="mt-2 text-xs text-muted-foreground">
                    {review.helpfulCount} people found this helpful
                  </p>
                )}
              </li>
            ))}
          </ul>

          {hasNext && (
            <div className="mt-6 flex justify-center">
              <Button
                type="button"
                variant="outline"
                onClick={loadMore}
                disabled={loading}
              >
                {loading && (
                  <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                )}
                Load more reviews
              </Button>
            </div>
          )}
        </>
      )}
    </section>
  );
}
