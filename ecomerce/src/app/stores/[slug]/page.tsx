import { notFound } from "next/navigation";
import type { Metadata } from "next";

import { Breadcrumb } from "@/components/navigation/breadcrumb";
import { ProductGrid } from "@/components/products/product-grid";
import { JsonLd } from "@/components/seo/json-ld";
import { StoreHeader } from "@/components/store/store-header";
import { buildBreadcrumbJsonLd, buildMetadata } from "@/lib/seo";
import { sanitizeText } from "@/lib/seo-text";
import { getStoreBySlug, getStoreFollow, getStoreProducts } from "@/lib/stores";

interface StorePageProps {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({
  params,
}: StorePageProps): Promise<Metadata> {
  const { slug } = await params;
  const store = await getStoreBySlug(slug);
  // Genuine 404, raised before the route's loading boundary streams so the
  // response carries a real 404 status.
  if (!store) notFound();

  const imageUrl = store.bannerUrl || store.logoUrl;

  return buildMetadata({
    title: store.name,
    description:
      sanitizeText(store.description) ||
      `Shop products from ${store.name} on ShopKart — compare prices and read reviews.`,
    path: `/stores/${store.slug}`,
    image: imageUrl ? { url: imageUrl, alt: store.name } : undefined,
  });
}

/** Public store profile (PRD store page): info, follow and the store's products. */
export default async function StorePage({ params }: StorePageProps) {
  const { slug } = await params;
  const store = await getStoreBySlug(slug);

  if (!store) notFound();

  const [{ items: products, pagination }, follow] = await Promise.all([
    getStoreProducts(store.slug, { page: 1, pageSize: 12, sort: "featured" }),
    getStoreFollow(store.slug),
  ]);

  const total = pagination?.total ?? products.length;

  const breadcrumbItems = [
    { name: "Home", path: "/" },
    { name: store.name, path: `/stores/${store.slug}` },
  ];

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <JsonLd data={buildBreadcrumbJsonLd(breadcrumbItems)} />

      <Breadcrumb
        items={breadcrumbItems.map((item, index) => ({
          label: item.name,
          href: index < breadcrumbItems.length - 1 ? item.path : undefined,
        }))}
      />

      <div className="mt-6">
        <StoreHeader store={store} initialFollowing={follow?.following ?? null} />
      </div>

      <section aria-labelledby="store-products-heading" className="mt-10">
        <h2
          id="store-products-heading"
          className="text-2xl font-semibold tracking-tight"
        >
          Products from {store.name}
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {total} {total === 1 ? "product" : "products"}
        </p>

        <div className="mt-6">
          {products.length === 0 ? (
            <p className="rounded-xl border border-dashed p-8 text-center text-sm text-muted-foreground">
              This store has no products yet.
            </p>
          ) : (
            <ProductGrid products={products} priorityCount={4} />
          )}
        </div>
      </section>
    </div>
  );
}
