import Link from "next/link";

import { Logo } from "@/components/layout/logo";
import { NewsletterForm } from "@/components/layout/newsletter-form";
import { NAV_LINKS, SITE_CONFIG } from "@/lib/constants";

// Support pages are static content planned for the polish phase; the links
// point at their final routes.
const SUPPORT_LINKS = [
  { label: "Contact", href: "/contact" },
  { label: "Shipping & Returns", href: "/shipping-returns" },
  { label: "FAQs", href: "/faq" },
  { label: "Privacy Policy", href: "/privacy-policy" },
  { label: "Terms of Service", href: "/terms-of-service" },
];

const PAYMENT_METHODS = [
  "Visa",
  "Mastercard",
  "PayPal",
  "Apple Pay",
  "Google Pay",
];

const footerLinkClassName =
  "text-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 rounded-sm";

/**
 * Site-wide footer (PRD §5/§59): brand blurb, shop links, support links,
 * newsletter signup and a bottom bar with payment methods.
 */
export function Footer() {
  const year = new Date().getFullYear();

  return (
    <footer className="border-t bg-secondary">
      <div className="mx-auto w-full max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="grid gap-10 py-12 sm:grid-cols-2 lg:grid-cols-[1.5fr_1fr_1fr_1.5fr] lg:gap-8">
          <div className="flex flex-col gap-4">
            <Logo />
            <p className="max-w-xs text-sm text-muted-foreground">
              {SITE_CONFIG.tagline}. Quality products, fast shipping and
              support you can count on.
            </p>
          </div>

          <nav aria-label="Shop">
            <h2 className="text-sm font-semibold text-foreground">Shop</h2>
            <ul className="mt-4 flex flex-col gap-2.5">
              {NAV_LINKS.map((link) => (
                <li key={link.label}>
                  <Link href={link.href} className={footerLinkClassName}>
                    {link.label}
                  </Link>
                </li>
              ))}
              <li>
                <Link href="/wishlist" className={footerLinkClassName}>
                  Wishlist
                </Link>
              </li>
            </ul>
          </nav>

          <nav aria-label="Support">
            <h2 className="text-sm font-semibold text-foreground">Support</h2>
            <ul className="mt-4 flex flex-col gap-2.5">
              {SUPPORT_LINKS.map((link) => (
                <li key={link.label}>
                  <Link href={link.href} className={footerLinkClassName}>
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>

          <div>
            <h2 className="text-sm font-semibold text-foreground">
              Join our newsletter
            </h2>
            <p className="mt-4 text-sm text-muted-foreground">
              New arrivals, exclusive deals and styling tips — straight to
              your inbox.
            </p>
            <div className="mt-4">
              <NewsletterForm />
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-4 border-t py-6 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-xs text-muted-foreground">
            © {year} {SITE_CONFIG.name}. All rights reserved. Demo store — no
            real orders are processed.
          </p>
          <ul
            aria-label="Accepted payment methods"
            className="flex flex-wrap items-center gap-1.5"
          >
            {PAYMENT_METHODS.map((method) => (
              <li
                key={method}
                className="rounded-md border bg-background px-2 py-1 text-[11px] font-medium text-muted-foreground"
              >
                {method}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </footer>
  );
}
