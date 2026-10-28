"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Heart, Search, Shield, ShoppingCart, Store, UserRound } from "lucide-react";

import { Logo } from "@/components/layout/logo";
import { MobileNav } from "@/components/navigation/mobile-nav";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useHasMounted } from "@/hooks/use-has-mounted";
import { NAV_LINKS } from "@/lib/constants";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/store/auth-store";
import { selectCartItemCount, useCartStore } from "@/store/cart-store";
import { useUiStore } from "@/store/ui-store";
import { selectWishlistCount, useWishlistStore } from "@/store/wishlist-store";

export function Header() {
  const pathname = usePathname();
  const sentinelRef = useRef<HTMLDivElement>(null);
  const [isScrolled, setIsScrolled] = useState(false);
  const hasMounted = useHasMounted();
  const cartCount = useCartStore(selectCartItemCount);
  const wishlistCount = useWishlistStore(selectWishlistCount);
  const openCart = useUiStore((state) => state.openCart);
  // Client-only UI hints; the backend remains authoritative for authorization.
  const isAuthenticated = useAuthStore((state) => state.status === "authenticated");
  const isAdmin = useAuthStore((state) => state.user?.role === "admin");
  const isVendor = useAuthStore((state) => state.user?.role === "vendor");

  const cartLabel =
    hasMounted && cartCount > 0
      ? `Shopping cart, ${cartCount} ${cartCount === 1 ? "item" : "items"}`
      : "Shopping cart";

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    // The sentinel sits at the very top of the document. Once it scrolls out
    // of view, the header gains its solid/blurred background (PRD §11).
    // An observer avoids scroll-event jank.
    const observer = new IntersectionObserver(
      ([entry]) => setIsScrolled(!entry.isIntersecting),
      { threshold: 0 }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, []);

  return (
    <>
      <div
        ref={sentinelRef}
        aria-hidden="true"
        className="pointer-events-none absolute top-0 h-px w-full"
      />
      <header
        className={cn(
          "sticky top-0 z-50 w-full border-b transition-all duration-300",
          isScrolled
            ? "border-border bg-background/90 shadow-sm backdrop-blur-md"
            : "border-transparent bg-background"
        )}
      >
        <div className="mx-auto flex h-16 w-full max-w-7xl items-center gap-1 px-4 sm:px-6 lg:px-8">
          <MobileNav />
          <Logo className="mr-2 sm:mr-6" />

          <nav aria-label="Main" className="hidden items-center gap-1 md:flex">
            {NAV_LINKS.map((link) => {
              const isActive = pathname === link.href;

              return (
                <Link
                  key={link.label}
                  href={link.href}
                  aria-current={isActive ? "page" : undefined}
                  className={cn(
                    "rounded-md px-3 py-2 text-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2",
                    isActive
                      ? "text-foreground"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  )}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>

          <div className="ml-auto flex items-center gap-0.5 sm:gap-1">
            <ThemeToggle />
            <HeaderIconButton
              href="/search"
              label="Search products"
              className="hidden sm:inline-flex"
            >
              <Search className="size-5" aria-hidden="true" />
            </HeaderIconButton>
            {isVendor && (
              <HeaderIconButton
                href="/vendor"
                label="Vendor dashboard"
                className="hidden sm:inline-flex"
              >
                <Store className="size-5" aria-hidden="true" />
              </HeaderIconButton>
            )}
            {isAdmin && (
              <HeaderIconButton
                href="/admin"
                label="Admin dashboard"
                className="hidden sm:inline-flex"
              >
                <Shield className="size-5" aria-hidden="true" />
              </HeaderIconButton>
            )}
            <HeaderIconButton
              href={isAuthenticated ? "/account" : "/auth/sign-in"}
              label={isAuthenticated ? "Your account" : "Sign in"}
              className="hidden sm:inline-flex"
            >
              <UserRound className="size-5" aria-hidden="true" />
            </HeaderIconButton>
            <HeaderIconButton
              href="/wishlist"
              label={
                hasMounted && wishlistCount > 0
                  ? `Wishlist, ${wishlistCount} saved ${wishlistCount === 1 ? "item" : "items"}`
                  : "Wishlist"
              }
              className="hidden sm:inline-flex"
              badge={hasMounted ? wishlistCount : 0}
            >
              <Heart className="size-5" aria-hidden="true" />
            </HeaderIconButton>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={openCart}
                  aria-label={cartLabel}
                  aria-haspopup="dialog"
                  className="relative size-10 text-muted-foreground hover:text-foreground"
                >
                  <ShoppingCart className="size-5" aria-hidden="true" />
                  {hasMounted && cartCount > 0 && (
                    <span
                      aria-hidden="true"
                      className="absolute right-0.5 top-0.5 flex min-w-4.5 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-semibold leading-4 text-primary-foreground"
                    >
                      {cartCount > 9 ? "9+" : cartCount}
                    </span>
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent side="bottom">{cartLabel}</TooltipContent>
            </Tooltip>
          </div>
        </div>
      </header>
    </>
  );
}

interface HeaderIconButtonProps {
  href: string;
  /** Full accessible name; the visible icon alone is not enough (PRD §38). */
  label: string;
  children: ReactNode;
  /** Count shown in the floating badge; rendered only when > 0. */
  badge?: number;
  className?: string;
}

function HeaderIconButton({
  href,
  label,
  children,
  badge = 0,
  className,
}: HeaderIconButtonProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          asChild
          variant="ghost"
          size="icon"
          className={cn(
            "relative size-10 text-muted-foreground hover:text-foreground",
            className
          )}
        >
          <Link href={href} aria-label={label}>
            {children}
            {badge > 0 && (
              <span
                aria-hidden="true"
                className="absolute right-0.5 top-0.5 flex min-w-4.5 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-semibold leading-4 text-primary-foreground"
              >
                {badge > 9 ? "9+" : badge}
              </span>
            )}
          </Link>
        </Button>
      </TooltipTrigger>
      <TooltipContent side="bottom">{label}</TooltipContent>
    </Tooltip>
  );
}
