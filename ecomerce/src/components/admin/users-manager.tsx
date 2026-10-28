"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { Copy, MoreHorizontal, Users } from "lucide-react";
import { toast } from "sonner";

import { TablePagination } from "@/components/admin/table-pagination";
import { Avatar } from "@/components/avatar";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { ErrorState } from "@/components/errors/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { apiGet, type PaginationMeta } from "@/lib/api-client";
import { formatDate } from "@/lib/formatters";

interface AdminUser {
  id: string;
  name: string;
  email: string;
  avatar_url?: string;
  role: string;
  roles: string[];
  status: string;
  email_verified: boolean;
  created_at: string;
  last_login_at: string | null;
}

const PAGE_SIZES = [10, 20, 50, 100];
const DEFAULT_PAGE_SIZE = 20;
const ROLE_FILTERS = [
  { value: "all", label: "All users" },
  { value: "user", label: "Customer" },
  { value: "vendor", label: "Vendor" },
  { value: "admin", label: "Admin" },
];

type FetchState =
  | { key: string; ok: true; users: AdminUser[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

/**
 * Admin user management (PRD §24). Every filter — role, search, page size and
 * page — is applied by the backend; the client only issues paginated requests
 * and mirrors the state into the URL.
 */
export function UsersManager() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  // Query state is derived from the URL so it survives navigation/refresh.
  const page = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);
  const limitParam = Number(searchParams.get("limit"));
  const limit = PAGE_SIZES.includes(limitParam) ? limitParam : DEFAULT_PAGE_SIZE;
  const roleParam = searchParams.get("role") ?? "all";
  const role = ROLE_FILTERS.some((r) => r.value === roleParam) ? roleParam : "all";
  const search = searchParams.get("search") ?? "";

  const [searchInput, setSearchInput] = useState(search);
  const debouncedSearch = useDebouncedValue(searchInput, 350);
  const [state, setState] = useState<FetchState | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const updateQuery = useCallback(
    (updates: Record<string, string | number | null>) => {
      const params = new URLSearchParams(searchParams.toString());
      for (const [key, value] of Object.entries(updates)) {
        if (value === null || value === "") params.delete(key);
        else params.set(key, String(value));
      }
      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, { scroll: false });
    },
    [pathname, router, searchParams]
  );

  // Debounced search writes to the URL (resetting to page 1).
  useEffect(() => {
    if (debouncedSearch === search) return;
    updateQuery({ search: debouncedSearch || null, page: 1 });
  }, [debouncedSearch, search, updateQuery]);

  const requestUrl = useMemo(() => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (role !== "all") params.set("role", role);
    if (search) params.set("search", search);
    return `/api/v1/admin/users?${params.toString()}`;
  }, [page, limit, role, search]);

  const queryKey = `${requestUrl}#${reloadToken}`;

  useEffect(() => {
    let active = true;
    void apiGet<AdminUser[]>(requestUrl).then((res) => {
      if (!active) return;
      if (res.ok && res.pagination) {
        setState({ key: queryKey, ok: true, users: res.data, pagination: res.pagination });
      } else {
        setState({
          key: queryKey,
          ok: false,
          message: res.ok ? "Pagination metadata was missing." : res.error.message,
        });
      }
    });
    return () => {
      active = false;
    };
  }, [requestUrl, queryKey]);

  const isLoading = state?.key !== queryKey;
  const failed = !isLoading && state?.ok === false;
  const users = !isLoading && state?.ok === true ? state.users : [];
  const pagination = !isLoading && state?.ok === true ? state.pagination : null;

  const hasFilters = search !== "" || role !== "all";

  function clearFilters() {
    setSearchInput("");
    updateQuery({ search: null, role: null, page: 1 });
  }

  function copy(value: string, label: string) {
    void navigator.clipboard?.writeText(value).then(
      () => toast.success(label),
      () => toast.error("Copy failed")
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Toolbar */}
      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
        <div className="flex w-full flex-col gap-1.5 sm:max-w-xs">
          <label htmlFor="user-search" className="text-xs font-medium text-muted-foreground">
            Search
          </label>
          <Input
            id="user-search"
            type="search"
            placeholder="Search by name, email, or ID…"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium text-muted-foreground">Role</label>
          <Select
            value={role}
            onValueChange={(value) => updateQuery({ role: value === "all" ? null : value, page: 1 })}
          >
            <SelectTrigger className="w-full sm:w-40" aria-label="Filter by role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {ROLE_FILTERS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium text-muted-foreground">Per page</label>
          <Select
            value={String(limit)}
            onValueChange={(value) => updateQuery({ limit: value, page: 1 })}
          >
            <SelectTrigger className="w-full sm:w-28" aria-label="Rows per page">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PAGE_SIZES.map((size) => (
                <SelectItem key={size} value={String(size)}>
                  {size}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <TableSkeleton rows={6} columns={5} label="Loading users" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load users"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((token) => token + 1)}
        />
      ) : users.length === 0 ? (
        <TableEmpty
          icon={Users}
          title="No users found"
          description={
            hasFilters
              ? "No users match your search or filter."
              : "There are no users to display yet."
          }
          action={
            hasFilters ? (
              <Button variant="outline" size="sm" onClick={clearFilters}>
                Clear filters
              </Button>
            ) : undefined
          }
        />
      ) : (
        <>
          {/* Desktop / tablet: table */}
          <Card className="hidden [--card-spacing:0px] md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="hidden lg:table-cell">Registered</TableHead>
                  <TableHead className="hidden lg:table-cell">Last login</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar name={user.name} src={user.avatar_url} className="size-10" />
                        <span className="flex min-w-0 flex-col">
                          <span className="font-medium">{user.name}</span>
                          <span className="text-xs text-muted-foreground">{user.email}</span>
                          <span className="font-mono text-[11px] text-muted-foreground">
                            {user.id}
                          </span>
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <RoleBadge role={user.role} />
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={user.status} verified={user.email_verified} />
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground lg:table-cell">
                      {formatDate(user.created_at)}
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground lg:table-cell">
                      {user.last_login_at ? formatDate(user.last_login_at) : "Never"}
                    </TableCell>
                    <TableCell className="text-right">
                      <UserActions user={user} onCopy={copy} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Card>

          {/* Mobile: cards */}
          <ul className="flex flex-col gap-3 md:hidden">
            {users.map((user) => (
              <li
                key={user.id}
                className="rounded-xl bg-card p-4 ring-1 ring-foreground/10"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <Avatar name={user.name} src={user.avatar_url} className="size-10" />
                    <div className="min-w-0">
                      <p className="truncate font-medium">{user.name}</p>
                      <p className="truncate text-sm text-muted-foreground">{user.email}</p>
                    </div>
                  </div>
                  <UserActions user={user} onCopy={copy} />
                </div>

                <div className="mt-3 flex flex-wrap items-center gap-2">
                  <RoleBadge role={user.role} />
                  <StatusBadge status={user.status} verified={user.email_verified} />
                </div>

                <dl className="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                  <div className="col-span-2 flex justify-between gap-3">
                    <dt className="font-mono">ID</dt>
                    <dd className="truncate font-mono">{user.id}</dd>
                  </div>
                  <div className="flex flex-col">
                    <dt>Registered</dt>
                    <dd className="text-foreground">{formatDate(user.created_at)}</dd>
                  </div>
                  <div className="flex flex-col">
                    <dt>Last login</dt>
                    <dd className="text-foreground">
                      {user.last_login_at ? formatDate(user.last_login_at) : "Never"}
                    </dd>
                  </div>
                </dl>
              </li>
            ))}
          </ul>

          <Card className="[--card-spacing:0px]">
            <TablePagination
              page={pagination?.page ?? page}
              pageCount={Math.max(1, pagination?.total_pages ?? 1)}
              total={pagination?.total ?? 0}
              onPageChange={(next) => updateQuery({ page: next })}
              itemLabel="users"
            />
          </Card>
        </>
      )}
    </div>
  );
}

function RoleBadge({ role }: { role: string }) {
  const label = role === "user" ? "Customer" : role === "vendor" ? "Vendor" : "Admin";
  const variant = role === "admin" ? "default" : role === "vendor" ? "secondary" : "outline";
  return <Badge variant={variant}>{label}</Badge>;
}

const STATUS_VARIANTS: Record<string, "info" | "warning" | "destructive" | "muted"> = {
  active: "info",
  pending: "warning",
  suspended: "destructive",
  blocked: "destructive",
};

function StatusBadge({ status, verified }: { status: string; verified: boolean }) {
  return (
    <span className="inline-flex flex-wrap items-center gap-1.5">
      <Badge variant={STATUS_VARIANTS[status] ?? "muted"} className="capitalize">
        {status}
      </Badge>
      {verified && (
        <span className="text-[11px] text-muted-foreground" title="Email verified">
          verified
        </span>
      )}
    </span>
  );
}

function UserActions({
  user,
  onCopy,
}: {
  user: AdminUser;
  onCopy: (value: string, label: string) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="size-8"
          aria-label={`Actions for ${user.name}`}
        >
          <MoreHorizontal className="size-4" aria-hidden="true" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-44">
        <DropdownMenuLabel>User actions</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => onCopy(user.email, "Email copied")}>
          <Copy className="size-4" aria-hidden="true" />
          Copy email
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => onCopy(user.id, "User ID copied")}>
          <Copy className="size-4" aria-hidden="true" />
          Copy user ID
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <Link href="/admin/orders">View orders</Link>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

