import path from "node:path";

import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactCompiler: true,
  images: {
    // Serve modern formats first and cache optimised images for 31 days to cut
    // repeat optimisations (SEO plan, Part 1 §8).
    formats: ["image/avif", "image/webp"],
    minimumCacheTTL: 2678400,
    remotePatterns: [
      // Deterministic placeholder photography for the demo catalog.
      { protocol: "https", hostname: "picsum.photos" },
      { protocol: "https", hostname: "fastly.picsum.photos" },
      // Profile pictures hosted on imgbb.
      { protocol: "https", hostname: "i.ibb.co" },
    ],
  },
  turbopack: {
    // Pin the workspace root (absolute) so Next.js never resolves stray
    // lockfiles (e.g. a package-lock.json in the user's home directory).
    root: path.resolve(),
  },
  async rewrites() {
    // Proxy the Go API through the Next.js origin so auth cookies are
    // same-site on localhost. `credentials: "include"` then works without
    // SameSite=None + HTTPS (which plain-HTTP localhost cannot satisfy).
    const backend = process.env.BACKEND_API_URL ?? "http://localhost:8080";
    return [
      {
        source: "/api/v1/:path*",
        destination: `${backend}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
