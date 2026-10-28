import Link from "next/link";
import { notFound } from "next/navigation";
import type { Metadata } from "next";
import { Check, Store } from "lucide-react";

import { Breadcrumb } from "@/components/navigation/breadcrumb";
import { JsonLd } from "@/components/seo/json-ld";
import { ProductBuyBox } from "@/components/products/product-buy-box";
import { ProductGallery } from "@/components/products/product-gallery";
import { ProductGrid } from "@/components/products/product-grid";
import { ProductQuestions } from "@/components/products/product-questions";
import { ProductReviews } from "@/components/products/product-reviews";
import { StarRating } from "@/components/products/star-rating";
import { buttonVariants } from "@/components/ui/button";
import {
  getCategories,
  getProductBySlug,
  getProductQuestions,
  getProductRating,
  getProductReviews,
  getRelatedProducts,
} from "@/lib/products";
import { buildBreadcrumbJsonLd, buildMetadata, buildProductJsonLd } from "@/lib/seo";
import { sanitizeText } from "@/lib/seo-text";
import { cn } from "@/lib/utils";

interface ProductPageProps {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({
  params,
}: ProductPageProps): Promise<Metadata> {
  const { slug } = await params;
  // A backend/network failure throws from the getter and is surfaced as a 500
  // by the route's error boundary; only a genuine 404 reaches `notFound()`.
  const product = await getProductBySlug(slug);

  // Genuine 404. Calling notFound() here — before the route's loading boundary
  // starts streaming — makes the page answer with a real 404 status instead of a
  // 200 that merely renders the not-found UI.
  if (!product) notFound();

  const image = product.images[0] ?? product.thumbnail;

  // Per-product SEO overrides with sensible fallbacks (PRD §26/§27).
  return buildMetadata({
    title: product.seo?.title || product.name,
    description: sanitizeText(
      product.seo?.description ||
        product.shortDescription ||
        `Discover ${product.name}, including its key features, available price, and product details.`
    ),
    path: `/products/${product.slug}`,
    type: "product",
    keywords: product.seo?.keywords,
    image: image.url
      ? { url: image.url, alt: image.alt || product.seo?.imageAlt }
      : undefined,
  });
}

/** Product detail page (PRD §14/§60) — server-rendered with client buy box. */
export default async function ProductPage({ params }: ProductPageProps) {
  const { slug } = await params;
  const product = await getProductBySlug(slug);

  if (!product) notFound();

  const [reviewsResult, rating, questionsResult, related, categories] =
    await Promise.all([
      getProductReviews(product.slug, { page: 1, pageSize: 10 }),
      getProductRating(product.slug),
      getProductQuestions(product.slug, { page: 1, pageSize: 10 }),
      getRelatedProducts(product, 4),
      getCategories(),
    ]);

  const reviews = reviewsResult.reviews;

  const category = categories.find(
    (candidate) => candidate.slug === product.categorySlug
  );

  const breadcrumbItems = [
    { name: "Home", path: "/" },
    { name: "Products", path: "/products" },
    {
      name: category?.name ?? product.category,
      path: `/category/${product.categorySlug}`,
    },
    { name: product.name, path: `/products/${product.slug}` },
  ];

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <JsonLd data={buildProductJsonLd(product, reviews)} />
      <JsonLd data={buildBreadcrumbJsonLd(breadcrumbItems)} />

      <Breadcrumb
        items={breadcrumbItems.map((item, index) => ({
          label: item.name,
          href: index < breadcrumbItems.length - 1 ? item.path : undefined,
        }))}
      />

      <div className="mt-6 grid gap-8 lg:grid-cols-2 lg:gap-12">
        <ProductGallery product={product} />

        <div className="flex flex-col gap-4">
          <div>
            <p className="text-sm text-muted-foreground">{product.brand}</p>
            <h1 className="mt-1 text-3xl font-bold tracking-tight sm:text-4xl">
              {product.name}
            </h1>
          </div>

          {product.storeSlug && (
            <div className="flex flex-wrap items-center gap-3">
              <span className="flex items-center gap-1.5 text-sm text-muted-foreground">
                <Store className="size-4" aria-hidden="true" />
                Sold by{" "}
                <span className="font-medium text-foreground">
                  {product.storeName}
                </span>
              </span>
              <Link
                href={`/stores/${product.storeSlug}`}
                className={cn(buttonVariants({ variant: "outline", size: "sm" }))}
              >
                Visit store
              </Link>
            </div>
          )}

          <div className="flex items-center gap-3">
            <StarRating
              rating={product.rating}
              reviewCount={product.reviewCount}
            />
            {product.reviewCount > 0 && (
              <Link
                href="#reviews"
                className="text-sm text-muted-foreground underline-offset-2 transition-colors hover:text-foreground hover:underline"
              >
                See reviews
              </Link>
            )}
          </div>

          <ProductBuyBox product={product} />
        </div>
      </div>

      <div className="mt-12 grid gap-8 lg:grid-cols-2">
        <div className="flex flex-col gap-8">
          <section aria-labelledby="description-heading">
            <h2
              id="description-heading"
              className="text-xl font-semibold tracking-tight"
            >
              Description
            </h2>
            <p className="mt-3 leading-relaxed text-muted-foreground">
              {product.description}
            </p>
          </section>

          {product.features.length > 0 && (
            <section aria-labelledby="features-heading">
              <h2
                id="features-heading"
                className="text-xl font-semibold tracking-tight"
              >
                Features
              </h2>
              <ul className="mt-3 flex flex-col gap-2.5">
                {product.features.map((feature) => (
                  <li
                    key={feature}
                    className="flex items-start gap-2.5 text-sm text-muted-foreground"
                  >
                    <Check
                      className="mt-0.5 size-4 shrink-0 text-primary"
                      aria-hidden="true"
                    />
                    {feature}
                  </li>
                ))}
              </ul>
            </section>
          )}
        </div>

        {product.specifications.length > 0 && (
          <section aria-labelledby="specs-heading">
            <div className="rounded-2xl border bg-secondary p-6">
              <h2
                id="specs-heading"
                className="text-xl font-semibold tracking-tight"
              >
                Specifications
              </h2>
              <dl className="mt-4">
                {product.specifications.map((specification) => (
                  <div
                    key={specification.label}
                    className="flex items-baseline justify-between gap-4 border-b py-2.5 text-sm last:border-0"
                  >
                    <dt className="text-muted-foreground">
                      {specification.label}
                    </dt>
                    <dd className="text-right font-medium">
                      {specification.value}
                    </dd>
                  </div>
                ))}
              </dl>
            </div>
          </section>
        )}
      </div>

      <ProductReviews
        key={product.slug}
        productSlug={product.slug}
        reviews={reviews}
        pagination={reviewsResult.pagination}
        summary={rating}
        rating={product.rating}
        reviewCount={product.reviewCount}
      />

      <ProductQuestions
        key={product.slug}
        productSlug={product.slug}
        initialQuestions={questionsResult.questions}
        total={questionsResult.pagination?.total ?? questionsResult.questions.length}
      />

      <section aria-labelledby="related-heading" className="mt-12">
        <h2
          id="related-heading"
          className="text-2xl font-semibold tracking-tight"
        >
          You may also like
        </h2>
        <div className="mt-6">
          <ProductGrid products={related} />
        </div>
      </section>
    </div>
  );
}
