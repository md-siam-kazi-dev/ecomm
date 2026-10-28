"use client";

import Link from "next/link";
import type { LucideIcon } from "lucide-react";
import { Heart, Menu, Search, Shield, Store, UserRound } from "lucide-react";

import { Logo } from "@/components/layout/logo";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { NAV_LINKS } from "@/lib/constants";
import { useAuthStore } from "@/store/auth-store";
import { useUiStore } from "@/store/ui-store";

interface NavItem {
  label: string;
  href: string;
  icon: LucideIcon;
}

const SECONDARY_LINKS: NavItem[] = [
  { label: "Wishlist", href: "/wishlist", icon: Heart },
  { label: "Account", href: "/account", icon: UserRound },
  { label: "Search", href: "/search", icon: Search },
];

/**
 * Mobile navigation drawer (PRD §11): menu trigger + slide-in sheet with the
 * primary links and secondary account actions. Focus and Escape are handled
 * by the underlying Radix dialog.
 */
export function MobileNav() {
  const isOpen = useUiStore((state) => state.isMobileNavOpen);
  const openMobileNav = useUiStore((state) => state.openMobileNav);
  const closeMobileNav = useUiStore((state) => state.closeMobileNav);
  // Client-only UI hints; the backend remains authoritative for authorization.
  const status = useAuthStore((state) => state.status);
  const role = useAuthStore((state) => state.user?.role);

  // The header hides the dashboard icon below `sm`; surface it here so a
  // vendor/admin on a small device can still reach their dashboard.
  const dashboardLinks: NavItem[] = [];
  if (status === "authenticated" && role === "vendor") {
    dashboardLinks.push({ label: "Vendor dashboard", href: "/vendor", icon: Store });
  }
  if (status === "authenticated" && role === "admin") {
    dashboardLinks.push({ label: "Admin dashboard", href: "/admin", icon: Shield });
  }

  return (
    <Sheet
      open={isOpen}
      onOpenChange={(open) => (open ? openMobileNav() : closeMobileNav())}
    >
      <SheetTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="size-10 md:hidden"
          aria-label="Open navigation menu"
        >
          <Menu className="size-5" aria-hidden="true" />
        </Button>
      </SheetTrigger>

      <SheetContent side="left" className="w-80 p-0">
        <SheetHeader className="border-b p-4">
          <SheetTitle asChild>
            <Logo />
          </SheetTitle>
          <SheetDescription className="sr-only">
            Main navigation menu
          </SheetDescription>
        </SheetHeader>

        <nav aria-label="Mobile" className="flex flex-col gap-1 p-4">
          {NAV_LINKS.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              onClick={closeMobileNav}
              className="rounded-md px-3 py-2.5 text-base font-medium text-foreground transition-colors hover:bg-muted focus-visible:outline-2 focus-visible:outline-offset-2"
            >
              {link.label}
            </Link>
          ))}
        </nav>

        <div className="mt-auto flex flex-col gap-1 border-t p-4">
          {[...dashboardLinks, ...SECONDARY_LINKS].map((link) => (
            <Link
              key={link.label}
              href={link.href}
              onClick={closeMobileNav}
              className="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2"
            >
              <link.icon className="size-4" aria-hidden="true" />
              {link.label}
            </Link>
          ))}
        </div>
      </SheetContent>
    </Sheet>
  );
}
