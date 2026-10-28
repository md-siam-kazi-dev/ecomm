import type { Metadata } from "next";

import { SITE_CONFIG } from "./constants";
import { absoluteUrl, sanitizeText, truncate } from "./seo-text";
import type { Product } from "@/types/product";
import type { Review } from "@/types/review";

/** Meta descriptions longer than this are truncated on a word boundary. */
const DESCRIPTION_MAX_LENGTH = 155;

/** Default Open Graph locale. */
const DEFAULT_LOCALE = "en_US";

/** Page kinds whose `og:type` we emit. */
export type MetadataPageType = "website" | "product" | "article";

/**
 * Next's own `OpenGraph.type` union — derived from the `Metadata` type so it
 * stays in sync with the installed version. It follows the Open Graph protocol,
 * which has no `product` member, so product pages are mapped onto `website` in
 * {@link buildMetadata} (passing `product` fails Next's render-time validation).
 */
type NextOpenGraphType = Extract<
  NonNullable<Metadata["openGraph"]>,
  { type: unknown }
>["type"];

export interface BuildMetadataInput {
  title: string;
  description: string;
  /** Absolute path, e.g. `/products/wireless-portable-blender`. */
  path: string;
  /** Open Graph object type; defaults to `"website"`. */
  type?: MetadataPageType;
  keywords?: string[];
  image?: {
    url: string;
    alt?: string;
    width?: number;
    height?: number;
  };
  /** Set for private/demo pages that must stay out of search indexes. */
  noIndex?: boolean;
  /** Allow crawlers to follow links from this page; defaults to `true`. */
  follow?: boolean;
  /** Open Graph locale; defaults to `"en_US"`. */
  locale?: string;
}

/**
 * Builds consistent page metadata — sanitised title/description, canonical URL,
 * Open Graph and Twitter cards, indexation directives — on top of the root
 * layout defaults (PRD §26–§32).
 *
 * The canonical URL is always derived from `SITE_CONFIG.url`, never from the
 * request host, so it stays deterministic behind proxies and previews.
 */
export function buildMetadata(input: BuildMetadataInput): Metadata {
  const title = sanitizeText(input.title) || SITE_CONFIG.name;
  const description =
    truncate(sanitizeText(input.description), DESCRIPTION_MAX_LENGTH) ||
    SITE_CONFIG.description;

  const canonical = new URL(input.path, SITE_CONFIG.url).toString();
  const locale = input.locale ?? DEFAULT_LOCALE;
  const noIndex = input.noIndex ?? false;

  // Next validates `og:type` at render time against the Open Graph protocol's
  // own type set, which has no `product` member (passing it throws
  // "Invalid OpenGraph type"). Product pages therefore advertise `website`; the
  // Product JSON-LD carries the product-specific semantics instead.
  const openGraphType: NextOpenGraphType =
    input.type === "product" ? "website" : (input.type ?? "website");

  const image = input.image;
  const imageUrl = image ? absoluteUrl(image.url, SITE_CONFIG.url) : "";
  const hasImage = imageUrl.length > 0;

  let robots: { index: boolean; follow: boolean } | undefined;
  if (noIndex) {
    robots = { index: false, follow: input.follow ?? true };
  } else if (input.follow === false) {
    robots = { index: true, follow: false };
  }

  return {
    title,
    description,
    ...(input.keywords?.length ? { keywords: input.keywords } : {}),
    alternates: {
      canonical,
    },
    ...(robots ? { robots } : {}),
    openGraph: {
      title,
      description,
      url: canonical,
      siteName: SITE_CONFIG.name,
      type: openGraphType,
      locale,
      ...(hasImage && image
        ? {
            images: [
              {
                url: imageUrl,
                ...(image.alt ? { alt: image.alt } : {}),
                ...(image.width !== undefined ? { width: image.width } : {}),
                ...(image.height !== undefined ? { height: image.height } : {}),
              },
            ],
          }
        : {}),
    },
    twitter: {
      card: hasImage ? "summary_large_image" : "summary",
      title,
      description,
      ...(hasImage ? { images: [imageUrl] } : {}),
    },
  };
}

/**
 * schema.org availability URL for a stock count (PRD §28). Out-of-stock
 * products stay indexable and simply advertise `OutOfStock`.
 */
export function availabilityForStock(stock: number): string {
  return stock > 0
    ? "https://schema.org/InStock"
    : "https://schema.org/OutOfStock";
}

/** Description fallback chain: SEO override → short → long description. */
function productDescription(product: Product): string {
  return (
    product.seo?.description || product.shortDescription || product.description
  );
}

/**
 * Product structured data for search engines (PRD §28). Only fields that
 * actually exist in the product data are included — brands, SKUs, ratings,
 * reviews and prices are never fabricated.
 */
export function buildProductJsonLd(
  product: Product,
  reviews?: Review[]
): Record<string, unknown> {
  const url = new URL(`/products/${product.slug}`, SITE_CONFIG.url).toString();

  const images = (product.images ?? [])
    .map((entry) => absoluteUrl(entry?.url ?? "", SITE_CONFIG.url))
    .filter((entry) => entry.length > 0);

  if (images.length === 0 && product.thumbnail?.url) {
    const thumbnail = absoluteUrl(product.thumbnail.url, SITE_CONFIG.url);
    if (thumbnail) images.push(thumbnail);
  }

  const data: Record<string, unknown> = {
    "@context": "https://schema.org",
    "@type": "Product",
    name: product.name,
    description: productDescription(product),
    image: images,
    url,
    offers: {
      "@type": "Offer",
      url,
      price: product.price.toFixed(2),
      priceCurrency: product.currency,
      itemCondition: "https://schema.org/NewCondition",
      availability: availabilityForStock(product.stock),
    },
  };

  if (product.brand) {
    data.brand = { "@type": "Brand", name: product.brand };
  }
  if (product.sku) {
    data.sku = product.sku;
  }
  if (product.category) {
    data.category = product.category;
  }
  if (product.storeName) {
    data.seller = { "@type": "Organization", name: product.storeName };
  }

  if (product.rating != null && product.reviewCount > 0) {
    data.aggregateRating = {
      "@type": "AggregateRating",
      ratingValue: product.rating,
      reviewCount: product.reviewCount,
    };
  }

  if (reviews?.length) {
    data.review = reviews.slice(0, 5).map((review) => ({
      "@type": "Review",
      author: { "@type": "Person", name: review.author },
      datePublished: review.createdAt,
      name: review.title,
      reviewBody: review.comment,
      reviewRating: {
        "@type": "Rating",
        ratingValue: review.rating,
        bestRating: 5,
        worstRating: 1,
      },
    }));
  }

  return data;
}

/** Breadcrumb trail as `BreadcrumbList` JSON-LD with absolute item URLs. */
export function buildBreadcrumbJsonLd(
  items: { name: string; path: string }[]
): Record<string, unknown> {
  return {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      item: absoluteUrl(item.path, SITE_CONFIG.url),
    })),
  };
}

/** A listing of product URLs as `ItemList` JSON-LD (category/store pages). */
export function buildItemListJsonLd(
  items: { name: string; slug: string }[]
): Record<string, unknown> {
  return {
    "@context": "https://schema.org",
    "@type": "ItemList",
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      url: new URL(`/products/${item.slug}`, SITE_CONFIG.url).toString(),
    })),
  };
}

/**
 * Site-wide `Organization` JSON-LD. Only data we actually hold is emitted —
 * no invented logo, address or social profiles (Part 5 of the SEO plan).
 */
export function buildOrganizationJsonLd(): Record<string, unknown> {
  return {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: SITE_CONFIG.name,
    url: SITE_CONFIG.url,
    description: SITE_CONFIG.description,
  };
}

/** Site-wide `WebSite` JSON-LD with a sitelinks search action. */
export function buildWebsiteJsonLd(): Record<string, unknown> {
  return {
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: SITE_CONFIG.name,
    url: SITE_CONFIG.url,
    potentialAction: {
      "@type": "SearchAction",
      target: {
        "@type": "EntryPoint",
        urlTemplate: new URL(
          "/search?q={search_term_string}",
          SITE_CONFIG.url
        ).toString(),
      },
      "query-input": "required name=search_term_string",
    },
  };
}

/**
 * Serialises JSON-LD for embedding in a `<script type="application/ld+json">`.
 * Escapes `<`, `>`, `&` and the U+2028/U+2029 line separators so untrusted data
 * values cannot break out of the script element or the JS parser.
 */
export function serializeJsonLd(data: unknown): string {
  return JSON.stringify(data)
    .replace(/</g, "\\u003c")
    .replace(/>/g, "\\u003e")
    .replace(/&/g, "\\u0026")
    .replace(/\u2028/g, "\\u2028")
    .replace(/\u2029/g, "\\u2029");
}
