import type { MetadataRoute } from "next";

import { SITE_CONFIG } from "@/lib/constants";

/**
 * Robots (PRD §31): shop routes are crawlable; private areas are not.
 *
 * `/search`, `/cart` and `/wishlist` are deliberately NOT disallowed. Those
 * pages emit `robots: noindex, follow` themselves, and a crawler only reads
 * (and honours) that directive if it is allowed to fetch the page. Disallowing
 * them here would hide the noindex tag and risk the URLs being indexed on the
 * strength of external links alone.
 */
export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
        disallow: [
          "/admin/",
          "/account/",
          "/auth/",
          "/checkout/",
          "/order/",
          "/vendor/",
          "/api/",
        ],
      },
    ],
    sitemap: new URL("/sitemap.xml", SITE_CONFIG.url).toString(),
  };
}
