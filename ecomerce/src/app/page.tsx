import type { Metadata } from "next";

import { Hero } from "@/components/hero/hero";
import { FeaturedCategories } from "@/components/home/featured-categories";
import { NewsletterSection } from "@/components/home/newsletter-section";
import { ProductSection } from "@/components/home/product-section";
import { PromoBanner } from "@/components/home/promo-banner";
import { TestimonialsSection } from "@/components/home/testimonials-section";
import { TrustSection } from "@/components/home/trust-section";
import { ProductGrid } from "@/components/products/product-grid";
import { SITE_CONFIG } from "@/lib/constants";
import {
  getBestSellers,
  getCategories,
  getFeaturedProducts,
  getTopReviews,
} from "@/lib/products";
import { buildMetadata } from "@/lib/seo";

/**
 * Homepage metadata (SEO plan, Part 1 §1). A plain title here is *not* wrapped
 * by the layout template (the template does not apply to the same segment), so
 * the document title is just the site name.
 */
export function generateMetadata(): Metadata {
  return buildMetadata({
    title: SITE_CONFIG.name,
    description: SITE_CONFIG.description,
    path: "/",
  });
}

// Homepage sections in PRD §9 order: hero → categories → featured products
// → promotion → best sellers → trust → reviews → newsletter. All data is
// fetched server-side through the catalog repository.
export default async function HomePage() {
  const [categories, featured, bestSellers, topReviews] = await Promise.all([
    getCategories(),
    getFeaturedProducts(8),
    getBestSellers(8),
    getTopReviews(3),
  ]);

  const [heroProduct] = featured;

  return (
    <>
      <Hero product={heroProduct} />
      {categories.length > 0 && <FeaturedCategories categories={categories} />}

      {featured.length > 0 && (
        <ProductSection
          title="Featured Products"
          description="Hand-picked favorites our customers keep coming back for."
          href="/products"
          linkLabel="Shop all products"
        >
          <ProductGrid products={featured} />
        </ProductSection>
      )}

      <PromoBanner />

      {bestSellers.length > 0 && (
        <ProductSection
          title="Best Sellers"
          description="The products everyone is talking about, backed by customer reviews."
          href="/products"
          linkLabel="Shop all products"
        >
          <ProductGrid products={bestSellers} />
        </ProductSection>
      )}

      <TrustSection />
      {topReviews.length > 0 && <TestimonialsSection reviews={topReviews} />}
      <NewsletterSection />
    </>
  );
}

