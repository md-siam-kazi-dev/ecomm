import Link from "next/link";
import { Star } from "lucide-react";

import { FadeUp } from "@/components/animations";
import type { ReviewWithProduct } from "@/lib/products";

interface TestimonialsSectionProps {
  reviews: ReviewWithProduct[];
}

/** Customer review cards pulled from real catalog review data (PRD §9). */
export function TestimonialsSection({ reviews }: TestimonialsSectionProps) {
  if (reviews.length === 0) return null;

  return (
    <section className="mx-auto w-full max-w-7xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
      <FadeUp inView className="mx-auto mb-8 max-w-2xl text-center">
        <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
          Loved by thousands
        </h2>
        <p className="mt-2 text-sm text-muted-foreground sm:text-base">
          Real reviews from verified ShopKart customers.
        </p>
      </FadeUp>

      <div className="grid gap-4 md:grid-cols-3">
        {reviews.map((review, index) => (
          <FadeUp key={review.id} inView delay={index * 0.08} className="h-full">
            <figure className="flex h-full flex-col gap-3 rounded-xl border bg-card p-5">
              <div
                className="flex gap-0.5"
                role="img"
                aria-label={`${review.rating} out of 5 stars`}
              >
                {Array.from({ length: 5 }).map((_, starIndex) => (
                  <Star
                    key={starIndex}
                    className="size-4 fill-amber-400 text-amber-400"
                    aria-hidden="true"
                  />
                ))}
              </div>
              <blockquote className="text-sm leading-relaxed text-foreground/90">
                “{review.comment}”
              </blockquote>
              <figcaption className="mt-auto flex flex-col gap-0.5 pt-1">
                <span className="text-sm font-medium">{review.author}</span>
                <span className="text-xs text-muted-foreground">
                  Verified purchase ·{" "}
                  <Link
                    href={`/products/${review.productSlug}`}
                    className="underline-offset-2 transition-colors hover:text-foreground hover:underline"
                  >
                    {review.productName}
                  </Link>
                </span>
              </figcaption>
            </figure>
          </FadeUp>
        ))}
      </div>
    </section>
  );
}
