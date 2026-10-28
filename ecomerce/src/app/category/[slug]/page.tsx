import Image from "next/image";
import { notFound } from "next/navigation";
import type { Metadata } from "next";

import { Breadcrumb } from "@/components/navigation/breadcrumb";
import { CatalogBrowser } from "@/components/products/catalog-browser";
import { JsonLd } from "@/components/seo/json-ld";
import { CATALOG_PAGE_SIZE } from "@/lib/constants";
import { getCategoryBySlug, getProducts } from "@/lib/products";
import {
  hasActiveFilters,
  parseCatalogFilters,
  type CatalogSearchParams,
} from "@/lib/query-params";
import {
  buildBreadcrumbJsonLd,
  buildItemListJsonLd,
  buildMetadata,
} from "@/lib/seo";
import { sanitizeText } from "@/lib/seo-text";

interface CategoryPageProps {
  params: Promise<{ slug: string }>;
  searchParams: Promise<CatalogSearchParams>;
}

/**
 * Category metadata. Uses the same indexation matrix as `/products` (SEO plan,
 * Part 4) — content filters are `noindex, follow` with a base-path canonical,
 * deeper unfiltered pages are self-canonical `?page=N`.
 */
export async function generateMetadata({
  params,
  searchParams,
}: CategoryPageProps): Promise<Metadata> {
  const [{ slug }, rawSearchParams] = await Promise.all([params, searchParams]);
  const category = await getCategoryBySlug(slug);

  // Genuine 404, raised before the route's loading boundary streams so the
  // response carries a real 404 status.
  if (!category) notFound();

  const basePath = `/category/${category.slug}`;
  const filters = parseCatalogFilters(rawSearchParams);
  // The category is the page itself, so a `category` param is not a filter here.
  const hasContentFilters = hasActiveFilters(filters, false);
  const isPaged = !hasContentFilters && filters.page > 1;

  const description =
    sanitizeText(category.description) ||
    `Shop ${category.name} at ShopKart — compare prices, read reviews and find the best deals.`;

  return buildMetadata({
    title: isPaged ? `${category.name} — Page ${filters.page}` : category.name,
    description,
    path: isPaged ? `${basePath}?page=${filters.page}` : basePath,
    noIndex: hasContentFilters,
    follow: true,
    image: category.image.url
      ? { url: category.image.url, alt: category.image.alt }
      : undefined,
  });
}

/** Category page (PRD §21): banner, description, filters and product grid. */
export default async function CategoryPage({
  params,
  searchParams,
}: CategoryPageProps) {
  const [{ slug }, rawSearchParams] = await Promise.all([params, searchParams]);
  const category = await getCategoryBySlug(slug);

  if (!category) notFound();

  const filters = parseCatalogFilters(rawSearchParams);

  // The same slice of products the grid streams (CatalogResults) — reused for
  // the ItemList structured data.
  const result = await getProducts({
    categorySlug: category.slug,
    query: filters.query,
    minPrice: filters.minPrice,
    maxPrice: filters.maxPrice,
    minRating: filters.rating,
    inStockOnly: filters.inStock,
    onSaleOnly: filters.onSale,
    sort: filters.sort,
    page: 1,
    pageSize: CATALOG_PAGE_SIZE * filters.page,
  });

  const breadcrumbItems = [
    { name: "Home", path: "/" },
    { name: category.name, path: `/category/${category.slug}` },
  ];

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <JsonLd data={buildBreadcrumbJsonLd(breadcrumbItems)} />
      {result.items.length > 0 && (
        <JsonLd
          data={buildItemListJsonLd(
            result.items.map((product) => ({
              name: product.name,
              slug: product.slug,
            }))
          )}
        />
      )}

      <Breadcrumb
        items={breadcrumbItems.map((item, index) => ({
          label: item.name,
          href: index < breadcrumbItems.length - 1 ? item.path : undefined,
        }))}
      />

      <div className="relative mt-4 h-40 w-full overflow-hidden rounded-2xl border sm:h-56">
        <Image
          src={category.image.url}
          alt={category.image.alt}
          fill
          priority
          sizes="100vw"
          className="object-cover"
        />
      </div>

      <header className="mt-6">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          {category.name}
        </h1>
        <p className="mt-2 max-w-2xl text-sm text-muted-foreground sm:text-base">
          {category.description}
        </p>
      </header>

      <div className="mt-8">
        <CatalogBrowser
          filters={filters}
          categories={[]}
          basePath={`/category/${category.slug}`}
          showCategoryFilter={false}
          fixedCategorySlug={category.slug}
          fixedCategoryName={category.name}
        />
      </div>
    </div>
  );
}
