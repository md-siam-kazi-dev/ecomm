"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  BarChart3,
  Boxes,
  ClipboardCheck,
  LayoutDashboard,
  MessageCircleQuestion,
  Package,
  ScrollText,
  Settings,
  ShoppingCart,
  Star,
  Store,
  TrendingUp,
  Users,
  type LucideIcon,
} from "lucide-react";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";

export type DashboardRole = "admin" | "vendor";

export interface DashboardNavItem {
  label: string;
  href: string;
  icon: LucideIcon;
  /** Match the pathname exactly instead of by prefix (used by overview links). */
  exact?: boolean;
}

export interface DashboardNavGroup {
  label: string;
  items: DashboardNavItem[];
}

export const ROLE_LABEL: Record<DashboardRole, string> = {
  admin: "Admin",
  vendor: "Vendor",
};

const ADMIN_NAV_GROUPS: DashboardNavGroup[] = [
  {
    label: "Overview",
    items: [
      { label: "Dashboard", href: "/admin", icon: LayoutDashboard, exact: true },
      { label: "Analytics", href: "/admin/analytics", icon: BarChart3 },
    ],
  },
  {
    label: "Catalog",
    items: [
      { label: "Products", href: "/admin/products", icon: Package, exact: true },
      { label: "Product review", href: "/admin/products/review", icon: ClipboardCheck },
      { label: "Inventory", href: "/admin/inventory", icon: Boxes },
    ],
  },
  {
    label: "Sales",
    items: [
      { label: "Orders", href: "/admin/orders", icon: ShoppingCart },
      { label: "Customers", href: "/admin/customers", icon: Users },
      { label: "Store applications", href: "/admin/store-applications", icon: Store },
    ],
  },
  {
    label: "System",
    items: [
      { label: "Audit logs", href: "/admin/audit-logs", icon: ScrollText },
      { label: "Store settings", href: "/admin/settings", icon: Settings },
    ],
  },
];

const VENDOR_NAV_GROUPS: DashboardNavGroup[] = [
  {
    label: "Overview",
    items: [
      { label: "Dashboard", href: "/vendor", icon: LayoutDashboard, exact: true },
      { label: "Revenue", href: "/vendor/revenue", icon: TrendingUp },
    ],
  },
  {
    label: "Catalog",
    items: [
      { label: "Products", href: "/vendor/products", icon: Package },
      { label: "Inventory", href: "/vendor/inventory", icon: Boxes },
    ],
  },
  {
    label: "Sales",
    items: [{ label: "Orders", href: "/vendor/orders", icon: ShoppingCart }],
  },
  {
    label: "Feedback",
    items: [
      { label: "Reviews", href: "/vendor/reviews", icon: Star },
      { label: "Questions", href: "/vendor/questions", icon: MessageCircleQuestion },
    ],
  },
];

export const DASHBOARD_NAV_GROUPS: Record<DashboardRole, DashboardNavGroup[]> = {
  admin: ADMIN_NAV_GROUPS,
  vendor: VENDOR_NAV_GROUPS,
};

/** Flat navigation list, used for breadcrumbs and other path lookups. */
export const DASHBOARD_NAV: Record<DashboardRole, DashboardNavItem[]> = {
  admin: ADMIN_NAV_GROUPS.flatMap((group) => group.items),
  vendor: VENDOR_NAV_GROUPS.flatMap((group) => group.items),
};

export function isNavItemActive(item: DashboardNavItem, pathname: string) {
  return item.exact ? pathname === item.href : pathname.startsWith(item.href);
}

interface DashboardNavProps {
  role: DashboardRole;
  onNavigate?: () => void;
}

export function DashboardNav({ role, onNavigate }: DashboardNavProps) {
  const pathname = usePathname();
  const { isMobile, setOpenMobile } = useSidebar();

  return (
    <>
      {DASHBOARD_NAV_GROUPS[role].map((group) => (
        <SidebarGroup key={group.label}>
          <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {group.items.map((item) => (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton
                    asChild
                    isActive={isNavItemActive(item, pathname)}
                    tooltip={item.label}
                  >
                    <Link
                      href={item.href}
                      onClick={() => {
                        onNavigate?.();
                        // On mobile the sidebar lives in a sheet; close it
                        // once a destination has been picked.
                        if (isMobile) setOpenMobile(false);
                      }}
                    >
                      <item.icon aria-hidden="true" />
                      <span>{item.label}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      ))}
    </>
  );
}
