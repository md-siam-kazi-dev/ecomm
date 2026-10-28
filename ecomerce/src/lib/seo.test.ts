// @ts-expect-error: bun's runtime ships `bun:test`, but this project does not
// install its type declarations (no new dependencies are permitted).
import { describe, expect, test } from "bun:test";
import type { Metadata } from "next";

import { SITE_CONFIG } from "./constants";
import {
  availabilityForStock,
  buildBreadcrumbJsonLd,
  buildItemListJsonLd,
  buildMetadata,
  buildOrganizationJsonLd,
  buildProductJsonLd,
  buildWebsiteJsonLd,
  serializeJsonLd,
} from "./seo";
import type { Product } from "../types/product";
import type { Review } from "../types/review";

// Loose shapes for the metadata fields under test (Next's Metadata unions are
// awkward to index directly).
type OgShape = {
  type?: string;
  locale?: string;
  title?: string;
  description?: string;
  siteName?: string;
  url?: string;
  images?: { url: string; alt?: string; width?: number; height?: number }[];
};
type TwitterShape = { card?: string; images?: string[]; title?: string };
type RobotsShape = { index?: boolean; follow?: boolean };

const ogOf = (meta: Metadata): OgShape => meta.openGraph as unknown as OgShape;
const twitterOf = (meta: Metadata): TwitterShape =>
  meta.twitter as unknown as TwitterShape;
const robotsOf = (meta: Metadata): RobotsShape | undefined =>
  meta.robots as unknown as RobotsShape | undefined;

function makeProduct(overrides: Partial<Product> = {}): Product {
  return {
    id: "p-1",
    slug: "demo-blender",
    name: "Demo Blender",
    shortDescription: "Short copy",
    description: "Long copy",
    price: 39.99,
    currency: "USD",
    images: [{ url: "/img/blender.png", alt: "Blender" }],
    thumbnail: { url: "/img/blender.png", alt: "Blender" },
    category: "Home & Kitchen",
    categorySlug: "home-kitchen",
    brand: "FreshMix",
    storeName: "FreshMix Store",
    storeSlug: "freshmix",
    rating: 4.5,
    reviewCount: 3,
    stock: 12,
    sku: "SKU-1",
    tags: [],
    features: [],
    specifications: [],
    variants: [],
    isFeatured: true,
    isNew: true,
    isBestSeller: true,
    createdAt: "2026-01-02T00:00:00Z",
    updatedAt: "2026-01-03T00:00:00Z",
    ...overrides,
  };
}

describe("buildMetadata", () => {
  test("sanitises and falls back for empty title/description", () => {
    const meta = buildMetadata({ title: "<b>  </b>", description: "", path: "/" });
    expect(meta.title).toBe(SITE_CONFIG.name);
    expect(meta.description).toBe(SITE_CONFIG.description);
  });

  test("truncates a long description on a word boundary", () => {
    const meta = buildMetadata({
      title: "Title",
      description: "a ".repeat(200),
      path: "/",
    });
    const description = String(meta.description);
    expect(description.length).toBeLessThanOrEqual(155);
    expect(description.endsWith("…")).toBe(true);
    expect(description.startsWith("a a")).toBe(true);
  });

  test("canonical is absolute and derived from SITE_CONFIG", () => {
    const meta = buildMetadata({
      title: "Title",
      description: "Desc",
      path: "/products/blender",
    });
    expect(String(meta.alternates?.canonical)).toBe(
      new URL("/products/blender", SITE_CONFIG.url).toString()
    );
  });

  test("emits OG type, locale, site name and image dimensions", () => {
    const meta = buildMetadata({
      title: "Blender",
      description: "Desc",
      path: "/products/blender",
      type: "product",
      locale: "en_GB",
      image: { url: "/img/p.png", alt: "Product", width: 1200, height: 630 },
    });
    const og = ogOf(meta);
    // `product` is not a valid Open Graph type for Next's renderer, so it maps
    // onto `website` (the Product JSON-LD carries product semantics).
    expect(og.type).toBe("website");
    expect(og.locale).toBe("en_GB");
    expect(og.siteName).toBe(SITE_CONFIG.name);
    expect(og.images).toEqual([
      {
        url: new URL("/img/p.png", SITE_CONFIG.url).toString(),
        alt: "Product",
        width: 1200,
        height: 630,
      },
    ]);
  });

  test("switches the twitter card depending on an image", () => {
    const withImage = buildMetadata({
      title: "T",
      description: "D",
      path: "/",
      image: { url: "/i.png" },
    });
    expect(twitterOf(withImage).card).toBe("summary_large_image");
    expect(twitterOf(withImage).images).toEqual([
      new URL("/i.png", SITE_CONFIG.url).toString(),
    ]);

    const withoutImage = buildMetadata({ title: "T", description: "D", path: "/" });
    expect(twitterOf(withoutImage).card).toBe("summary");
    expect(twitterOf(withoutImage).images).toBeUndefined();
  });

  test("controls robots index/follow directives", () => {
    expect(
      robotsOf(buildMetadata({ title: "T", description: "D", path: "/" }))
    ).toBeUndefined();

    expect(
      robotsOf(
        buildMetadata({ title: "T", description: "D", path: "/search", noIndex: true })
      )
    ).toEqual({ index: false, follow: true });

    expect(
      robotsOf(
        buildMetadata({
          title: "T",
          description: "D",
          path: "/search",
          noIndex: true,
          follow: false,
        })
      )
    ).toEqual({ index: false, follow: false });

    expect(
      robotsOf(
        buildMetadata({ title: "T", description: "D", path: "/", follow: false })
      )
    ).toEqual({ index: true, follow: false });
  });
});

describe("availabilityForStock", () => {
  test("maps stock to a schema.org availability URL", () => {
    expect(availabilityForStock(5)).toBe("https://schema.org/InStock");
    expect(availabilityForStock(0)).toBe("https://schema.org/OutOfStock");
  });
});

describe("buildProductJsonLd", () => {
  test("maps price, currency and availability", () => {
    const offers = buildProductJsonLd(makeProduct()).offers as Record<
      string,
      unknown
    >;
    expect(offers.price).toBe("39.99");
    expect(offers.priceCurrency).toBe("USD");
    expect(offers.availability).toBe("https://schema.org/InStock");
  });

  test("reports OutOfStock when stock is zero", () => {
    const offers = buildProductJsonLd(makeProduct({ stock: 0 })).offers as Record<
      string,
      unknown
    >;
    expect(offers.availability).toBe("https://schema.org/OutOfStock");
  });

  test("omits empty brand and sku", () => {
    const data = buildProductJsonLd(makeProduct({ brand: "", sku: "" }));
    expect("brand" in data).toBe(false);
    expect("sku" in data).toBe(false);

    const withBrand = buildProductJsonLd(makeProduct());
    expect(withBrand.brand).toEqual({ "@type": "Brand", name: "FreshMix" });
    expect(withBrand.sku).toBe("SKU-1");
  });

  test("only emits aggregateRating when rating and reviewCount are genuine", () => {
    expect(
      "aggregateRating" in
        buildProductJsonLd(makeProduct({ rating: undefined, reviewCount: 0 }))
    ).toBe(false);
    expect(
      "aggregateRating" in
        buildProductJsonLd(makeProduct({ rating: 4.5, reviewCount: 0 }))
    ).toBe(false);

    const rated = buildProductJsonLd(makeProduct({ rating: 4.5, reviewCount: 3 }));
    expect(rated.aggregateRating).toMatchObject({
      "@type": "AggregateRating",
      ratingValue: 4.5,
      reviewCount: 3,
    });
  });

  test("uses the description fallback chain", () => {
    const seo = { title: "T", description: "SEO copy", keywords: [], imageAlt: "" };
    expect(buildProductJsonLd(makeProduct({ seo })).description).toBe("SEO copy");
    expect(
      buildProductJsonLd(makeProduct({ shortDescription: "Short", description: "Long" }))
        .description
    ).toBe("Short");
    expect(
      buildProductJsonLd(makeProduct({ shortDescription: "", description: "Long" }))
        .description
    ).toBe("Long");
  });

  test("uses absolute image URLs from the gallery", () => {
    const data = buildProductJsonLd(
      makeProduct({
        images: [{ url: "/img/a.png", alt: "A" }],
        thumbnail: { url: "", alt: "" },
      })
    );
    expect(data.image).toEqual([new URL("/img/a.png", SITE_CONFIG.url).toString()]);
  });

  test("falls back to the thumbnail when the gallery is empty", () => {
    const data = buildProductJsonLd(
      makeProduct({
        images: [],
        thumbnail: { url: "/img/t.png", alt: "T" },
      })
    );
    expect(data.image).toEqual([new URL("/img/t.png", SITE_CONFIG.url).toString()]);
  });

  test("includes genuine reviews", () => {
    const reviews: Review[] = [
      {
        id: "r-1",
        productId: "p-1",
        author: "Sam",
        rating: 5,
        title: "Great",
        comment: "Loved it",
        createdAt: "2026-01-01T00:00:00Z",
        isVerifiedPurchase: true,
        helpfulCount: 2,
      },
    ];
    const data = buildProductJsonLd(makeProduct(), reviews);
    const list = data.review as Array<Record<string, unknown>>;
    expect(list).toHaveLength(1);
    expect((list[0].reviewRating as Record<string, unknown>).ratingValue).toBe(5);
  });
});

describe("serializeJsonLd", () => {
  test("escapes script-breaking characters and line separators", () => {
    const out = serializeJsonLd({
      html: "<script>alert('&')</script>",
      sep: "\u2028\u2029",
    });
    expect(out).not.toContain("<");
    expect(out).not.toContain(">");
    expect(out).not.toContain("&");
    expect(out).not.toContain("\u2028");
    expect(out).not.toContain("\u2029");
    expect(out).toContain("\\u003c");
    expect(out).toContain("\\u003e");
    expect(out).toContain("\\u0026");
    expect(out).toContain("\\u2028");
    expect(out).toContain("\\u2029");
  });

  test("round-trips through JSON.parse", () => {
    const payload = { name: "A & B", html: "<b>x</b>", sep: "line\u2028sep" };
    expect(JSON.parse(serializeJsonLd(payload))).toEqual(payload);
  });
});

describe("structured data builders", () => {
  test("buildBreadcrumbJsonLd positions items with absolute URLs", () => {
    const data = buildBreadcrumbJsonLd([
      { name: "Home", path: "/" },
      { name: "Products", path: "/products" },
    ]);
    const items = data.itemListElement as Array<Record<string, unknown>>;
    expect(items).toHaveLength(2);
    expect(items[0]).toMatchObject({
      position: 1,
      name: "Home",
      item: new URL("/", SITE_CONFIG.url).toString(),
    });
    expect(items[1].position).toBe(2);
  });

  test("buildItemListJsonLd links to product pages", () => {
    const data = buildItemListJsonLd([{ name: "Blender", slug: "blender" }]);
    const items = data.itemListElement as Array<Record<string, unknown>>;
    expect(items[0].url).toBe(new URL("/products/blender", SITE_CONFIG.url).toString());
  });

  test("buildOrganizationJsonLd exposes the site identity", () => {
    const data = buildOrganizationJsonLd();
    expect(data["@type"]).toBe("Organization");
    expect(data.name).toBe(SITE_CONFIG.name);
    expect(data.url).toBe(SITE_CONFIG.url);
  });

  test("buildWebsiteJsonLd exposes a sitelinks search action", () => {
    const data = buildWebsiteJsonLd();
    expect(data["@type"]).toBe("WebSite");
    const action = data.potentialAction as Record<string, unknown>;
    expect(String((action.target as Record<string, unknown>).urlTemplate)).toContain(
      "/search"
    );
  });
});
