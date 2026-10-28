// @ts-expect-error: bun's runtime ships `bun:test`, but this project does not
// install its type declarations (no new dependencies are permitted).
import { describe, expect, test } from "bun:test";

import { SITE_CONFIG } from "./constants";
import {
  buildSitemapEntries,
  nextSitemapCursor,
  type StaticSitemapPage,
} from "./sitemap-entries";

const STATIC: StaticSitemapPage[] = [
  { path: "/", priority: 1, changeFrequency: "weekly" },
  { path: "/products", priority: 0.9, changeFrequency: "daily" },
];

const absolute = (path: string): string =>
  new URL(path, SITE_CONFIG.url).toString();

/** base64url without padding, matching Go's `base64.RawURLEncoding`. */
function expectedCursor(payload: string): string {
  return Buffer.from(payload, "utf8")
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

describe("buildSitemapEntries", () => {
  test("emits absolute URLs ordered static → categories → stores → products", () => {
    const entries = buildSitemapEntries({
      staticPages: STATIC,
      categories: [{ slug: "home", updatedAt: "2026-01-01T00:00:00Z" }],
      stores: [{ slug: "freshmix" }],
      products: [{ slug: "blender", updatedAt: "2026-01-02T00:00:00Z" }],
    });
    const urls = entries.map((entry) => entry.url);

    expect(urls).toEqual([
      absolute("/"),
      absolute("/products"),
      absolute("/category/home"),
      absolute("/stores/freshmix"),
      absolute("/products/blender"),
    ]);
    for (const url of urls) {
      expect(url.startsWith(SITE_CONFIG.url)).toBe(true);
    }
  });

  test("de-duplicates repeated URLs, keeping the first occurrence", () => {
    const entries = buildSitemapEntries({
      staticPages: STATIC,
      categories: [{ slug: "home" }, { slug: "home" }],
      stores: [],
      products: [{ slug: "blender" }, { slug: "blender" }],
    });
    const urls = entries.map((entry) => entry.url);

    expect(urls.filter((url) => url === absolute("/category/home"))).toHaveLength(1);
    expect(urls.filter((url) => url === absolute("/products/blender"))).toHaveLength(1);
  });

  test("drops entries with an empty or whitespace-only slug", () => {
    const entries = buildSitemapEntries({
      staticPages: [],
      categories: [{ slug: "" }, { slug: "   " }],
      stores: [{ slug: "" }],
      products: [{ slug: "  " }],
    });
    expect(entries).toHaveLength(0);
  });

  test("emits lastModified only for a real timestamp", () => {
    const entries = buildSitemapEntries({
      staticPages: [],
      categories: [
        { slug: "with-date", updatedAt: "2026-05-01T12:00:00Z" },
        { slug: "no-date" },
        { slug: "bad-date", updatedAt: "not-a-date" },
      ],
      stores: [],
      products: [],
    });
    const byUrl = new Map(entries.map((entry) => [entry.url, entry]));

    expect(byUrl.get(absolute("/category/with-date"))?.lastModified).toBeInstanceOf(Date);
    expect(byUrl.get(absolute("/category/no-date"))?.lastModified).toBeUndefined();
    expect(byUrl.get(absolute("/category/bad-date"))?.lastModified).toBeUndefined();
  });

  test("assigns a change frequency and priority per type", () => {
    const entries = buildSitemapEntries({
      staticPages: STATIC,
      categories: [{ slug: "c" }],
      stores: [{ slug: "s" }],
      products: [{ slug: "p" }],
    });
    const byUrl = new Map(entries.map((entry) => [entry.url, entry]));

    expect(byUrl.get(absolute("/"))?.priority).toBe(1);
    expect(byUrl.get(absolute("/"))?.changeFrequency).toBe("weekly");
    expect(byUrl.get(absolute("/category/c"))?.priority).toBe(0.7);
    expect(byUrl.get(absolute("/category/c"))?.changeFrequency).toBe("weekly");
    expect(byUrl.get(absolute("/stores/s"))?.priority).toBe(0.6);
    expect(byUrl.get(absolute("/products/p"))?.priority).toBe(0.8);
  });
});

describe("nextSitemapCursor", () => {
  test("encodes base64url(`updated_at|id`) exactly like the backend", () => {
    const row = {
      id: "11111111-1111-1111-1111-111111111111",
      slug: "demo-blender",
      updated_at: "2026-01-02T03:04:05Z",
    };
    const cursor = nextSitemapCursor(row);

    expect(cursor).toBe(
      expectedCursor(
        "2026-01-02T03:04:05Z|11111111-1111-1111-1111-111111111111"
      )
    );
    expect(cursor).not.toContain("=");
    expect(cursor).not.toMatch(/[+/]/);
  });

  test("returns null when there is no row to continue from", () => {
    expect(nextSitemapCursor(null)).toBeNull();
    expect(nextSitemapCursor(undefined)).toBeNull();
    expect(
      nextSitemapCursor({ id: "", slug: "x", updated_at: "2026-01-02T03:04:05Z" })
    ).toBeNull();
    expect(nextSitemapCursor({ id: "abc", slug: "x", updated_at: "" })).toBeNull();
  });
});
