import type { Metadata } from "next";
import { Geist_Mono, Inter } from "next/font/google";

import { AnnouncementBar } from "@/components/layout/announcement-bar";
import { AuthProvider } from "@/components/auth/auth-provider";
import { CartDrawer } from "@/components/cart/cart-drawer";
import { CartProvider } from "@/components/cart/cart-provider";
import { Footer } from "@/components/layout/footer";
import { Header } from "@/components/layout/header";
import { SiteChrome } from "@/components/layout/site-chrome";
import { JsonLd } from "@/components/seo/json-ld";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { SITE_CONFIG } from "@/lib/constants";
import { buildOrganizationJsonLd, buildWebsiteJsonLd } from "@/lib/seo";

import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL(SITE_CONFIG.url),
  title: {
    default: `${SITE_CONFIG.name} — Modern Online Store`,
    template: `%s | ${SITE_CONFIG.name}`,
  },
  description: SITE_CONFIG.description,
  // Site-wide social defaults so a page that only sets title/description still
  // emits sensible Open Graph / Twitter tags (pages can override via
  // `buildMetadata`, which replaces these wholesale).
  openGraph: {
    siteName: SITE_CONFIG.name,
    type: "website",
    locale: "en_US",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
  },
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${inter.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="flex min-h-full flex-col">
        {/* Site-wide structured data (SEO plan, Part 5). */}
        <JsonLd data={buildOrganizationJsonLd()} />
        <JsonLd data={buildWebsiteJsonLd()} />
        <a
          href="#main-content"
          className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[100] focus:rounded-md focus:bg-primary focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-foreground focus:shadow-md"
        >
          Skip to main content
        </a>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <AuthProvider>
            <CartProvider>
              <TooltipProvider delayDuration={200}>
                <SiteChrome
                  announcement={<AnnouncementBar />}
                  header={<Header />}
                  footer={<Footer />}
                >
                  {children}
                </SiteChrome>
              </TooltipProvider>
            </CartProvider>
          </AuthProvider>
          <CartDrawer />
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}

