"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Heart,
  LayoutDashboard,
  MapPin,
  Package,
  Settings,
  ShieldCheck,
  ShoppingBag,
  Store,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { useAuthStore } from "@/store/auth-store";

const ACCOUNT_LINKS = [
  { label: "Overview", href: "/account", icon: LayoutDashboard, exact: true },
  { label: "Orders", href: "/account/orders", icon: Package, exact: false },
  { label: "Wishlist", href: "/account/wishlist", icon: Heart, exact: false },
  { label: "Addresses", href: "/account/addresses", icon: MapPin, exact: false },
  { label: "Settings", href: "/account/settings", icon: Settings, exact: false },
];

// Store application / onboarding entry (Facebook-Page-style). Shown to every
// signed-in account; vendors additionally get the dashboard link below.
const SELLER_LINK = {
  label: "Your store",
  href: "/account/store",
  icon: ShoppingBag,
  exact: false,
};

const ADMIN_LINK = {
  label: "Admin dashboard",
  href: "/admin",
  icon: ShieldCheck,
  exact: false,
};

const VENDOR_LINK = {
  label: "Vendor dashboard",
  href: "/vendor",
  icon: Store,
  exact: false,
};

/**
 * Account section navigation (PRD §23). Horizontal scroll on mobile, a
 * vertical rail on large screens. Marks the active route for assistive tech.
 */
export function AccountNav() {
  const pathname = usePathname();
  const role = useAuthStore((state) => state.user?.role);
  // Admins and vendors get an extra link to their back-office area.
  const links = [
    ...ACCOUNT_LINKS,
    SELLER_LINK,
    ...(role === "admin" ? [ADMIN_LINK] : []),
    ...(role === "vendor" ? [VENDOR_LINK] : []),
  ];

  return (
    <nav aria-label="Account">
      <ul className="flex flex-wrap gap-1 lg:flex-col">
        {links.map((link) => {
          const isActive = link.exact
            ? pathname === link.href
            : pathname.startsWith(link.href);

          return (
            <li key={link.href} className="lg:shrink">
              <Link
                href={link.href}
                aria-current={isActive ? "page" : undefined}
                className={cn(
                  "flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2",
                  isActive
                    ? "bg-primary/10 text-primary-dark"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                )}
              >
                <link.icon className="size-4" aria-hidden="true" />
                {link.label}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
