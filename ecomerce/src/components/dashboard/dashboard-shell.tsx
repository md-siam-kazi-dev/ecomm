"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import {
  DASHBOARD_NAV,
  DashboardNav,
  ROLE_LABEL,
  isNavItemActive,
  type DashboardRole,
} from "@/components/dashboard/dashboard-nav";
import { DashboardUserMenu } from "@/components/dashboard/dashboard-user-menu";
import { LogoMark } from "@/components/layout/logo";
import { ThemeToggle } from "@/components/theme-toggle";
import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { SITE_CONFIG } from "@/lib/constants";

interface Crumb {
  label: string;
  href?: string;
}

const titleCase = (value: string) =>
  value
    .split("-")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");

function useCrumbs(role: DashboardRole): Crumb[] {
  const pathname = usePathname();
  const home = `/${role}`;
  const match = DASHBOARD_NAV[role].find((item) => isNavItemActive(item, pathname));
  const sections = ROLE_LABEL[role];

  if (pathname === home) {
    return [{ label: match?.label ?? sections }];
  }

  const crumbs: Crumb[] = [{ label: sections, href: home }];

  if (match && match.href !== home) {
    const rest = pathname
      .slice(match.href.length)
      .split("/")
      .filter(Boolean);

    crumbs.push({
      label: match.label,
      href: rest.length > 0 ? match.href : undefined,
    });

    for (const segment of rest) {
      crumbs.push({ label: titleCase(segment) });
    }

    return crumbs;
  }

  for (const segment of pathname.split("/").filter(Boolean).slice(1)) {
    crumbs.push({ label: titleCase(segment) });
  }

  return crumbs;
}

function DashboardBreadcrumb({ role }: { role: DashboardRole }) {
  const crumbs = useCrumbs(role);

  return (
    <Breadcrumb className="min-w-0">
      <BreadcrumbList className="flex-nowrap">
        {crumbs.map((crumb, index) => {
          const isLast = index === crumbs.length - 1;

          return (
            <span key={`${crumb.label}-${index}`} className="contents">
              {index > 0 && <BreadcrumbSeparator />}
              <BreadcrumbItem className={isLast ? "min-w-0" : undefined}>
                {crumb.href && !isLast ? (
                  <BreadcrumbLink asChild>
                    <Link href={crumb.href}>{crumb.label}</Link>
                  </BreadcrumbLink>
                ) : (
                  <BreadcrumbPage className="truncate">
                    {crumb.label}
                  </BreadcrumbPage>
                )}
              </BreadcrumbItem>
            </span>
          );
        })}
      </BreadcrumbList>
    </Breadcrumb>
  );
}

interface DashboardShellProps {
  role: DashboardRole;
  name: string;
  email: string;
  avatarUrl?: string;
  children: ReactNode;
}

/**
 * Shared chrome for the admin and vendor areas, built on the shadcn sidebar
 * primitives: a grouped, collapsible icon sidebar on desktop, the built-in
 * sheet drawer on mobile, and a sticky breadcrumb bar in the inset content.
 */
export function DashboardShell({
  role,
  name,
  email,
  avatarUrl,
  children,
}: DashboardShellProps) {
  const label = ROLE_LABEL[role];

  return (
    <SidebarProvider>
      <Sidebar collapsible="icon" variant="inset">
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild tooltip={label}>
                <Link href={`/${role}`}>
                  <LogoMark />
                  <span className="grid min-w-0 flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-semibold">{label}</span>
                    <span className="truncate text-xs text-muted-foreground">
                      {SITE_CONFIG.name}
                    </span>
                  </span>
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>

        <SidebarContent>
          <DashboardNav role={role} />
        </SidebarContent>

        <SidebarFooter>
          <DashboardUserMenu
            name={name}
            email={email}
            avatarUrl={avatarUrl}
          />
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>

      <SidebarInset id="main-content">
        <header className="sticky top-0 z-30 flex h-16 shrink-0 items-center gap-2 rounded-t-xl border-b bg-background/80 px-4 backdrop-blur-md sm:px-6">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-1 !h-4" />
          <DashboardBreadcrumb role={role} />
          <div className="ml-auto flex items-center gap-1">
            <ThemeToggle className="size-8" />
            <Badge variant="secondary" className="hidden sm:inline-flex">
              {label}
            </Badge>
          </div>
        </header>
        <div className="flex min-w-0 flex-1 flex-col gap-6 p-4 sm:p-6 lg:p-8">
          {children}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
