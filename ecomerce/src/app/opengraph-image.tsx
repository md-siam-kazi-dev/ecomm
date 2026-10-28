import { ImageResponse } from "next/og";

import { SITE_CONFIG } from "@/lib/constants";

/** Default social-share card shown by any page without its own image. */
export const alt = `${SITE_CONFIG.name} — ${SITE_CONFIG.tagline}`;
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

/**
 * Static (no data fetching) Open Graph image: the site name and tagline on the
 * theme background, so pages that only set title/description still get a sane
 * social preview (SEO plan, Part 1 §10).
 */
export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          height: "100%",
          width: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          background: "#ffffff",
          padding: "80px",
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "20px" }}>
          <div
            style={{
              width: "64px",
              height: "64px",
              borderRadius: "16px",
              background: "#2563eb",
            }}
          />
          <div
            style={{ fontSize: "44px", fontWeight: 700, color: "#0f172a" }}
          >
            {SITE_CONFIG.name}
          </div>
        </div>

        <div
          style={{
            fontSize: "68px",
            fontWeight: 700,
            color: "#0f172a",
            lineHeight: 1.15,
            maxWidth: "960px",
          }}
        >
          {SITE_CONFIG.tagline}
        </div>

        <div style={{ fontSize: "28px", color: "#64748b", maxWidth: "960px" }}>
          {SITE_CONFIG.description}
        </div>
      </div>
    ),
    size
  );
}
