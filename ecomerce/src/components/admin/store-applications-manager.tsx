"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { Eye, Inbox } from "lucide-react";

import { TablePagination } from "@/components/admin/table-pagination";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { ErrorState } from "@/components/errors/error-state";
import {
  ApplicationStatusBadge,
  PaymentStatusBadge,
} from "@/components/store/application-status-badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
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
import type { ApplicationItem } from "@/types/store-application";

const PAGE_SIZES = [10, 20, 50, 100];
const DEFAULT_PAGE_SIZE = 20;
const STATUS_FILTERS = [
  { value: "all", label: "All statuses" },
  { value: "pending", label: "Pending review" },
  { value: "approved", label: "Approved" },
  { value: "rejected", label: "Rejected" },
];
const SORT_OPTIONS = [
  { value: "submitted_at_desc", label: "Newest first" },
  { value: "submitted_at_asc", label: "Oldest first" },
];

type FetchState =
  | { key: string; ok: true; items: ApplicationItem[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

/** Admin store-application queue: server-side filter/search/sort/pagination. */
export function StoreApplicationsManager() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);
  const limitParam = Number(searchParams.get("limit"));
  const limit = PAGE_SIZES.includes(limitParam) ? limitParam : DEFAULT_PAGE_SIZE;
  const statusParam = searchParams.get("status") ?? "all";
  const status = STATUS_FILTERS.some((s) => s.value === statusParam) ? statusParam : "all";
  const sortParam = searchParams.get("sort") ?? "submitted_at_desc";
  const sort = SORT_OPTIONS.some((s) => s.value === sortParam) ? sortParam : "submitted_at_desc";
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

  useEffect(() => {
    if (debouncedSearch === search) return;
    updateQuery({ search: debouncedSearch || null, page: 1 });
  }, [debouncedSearch, search, updateQuery]);

  const requestUrl = useMemo(() => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (status !== "all") params.set("status", status);
    if (sort !== "submitted_at_desc") params.set("sort", sort);
    if (search) params.set("search", search);
    return `/api/v1/admin/store-applications?${params.toString()}`;
  }, [page, limit, status, sort, search]);

  const queryKey = `${requestUrl}#${reloadToken}`;

  useEffect(() => {
    let active = true;
    void apiGet<ApplicationItem[]>(requestUrl).then((res) => {
      if (!active) return;
      if (res.ok && res.pagination) {
        setState({ key: queryKey, ok: true, items: res.data, pagination: res.pagination });
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
  const items = !isLoading && state?.ok === true ? state.items : [];
  const pagination = !isLoading && state?.ok === true ? state.pagination : null;
  const hasFilters = search !== "" || status !== "all";

  function clearFilters() {
    setSearchInput("");
    updateQuery({ search: null, status: null, sort: null, page: 1 });
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
        <div className="flex w-full flex-col gap-1.5 sm:max-w-xs">
          <label htmlFor="app-search" className="text-xs font-medium text-muted-foreground">
            Search
          </label>
          <Input
            id="app-search"
            type="search"
            placeholder="Name, email, app number or store…"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium text-muted-foreground">Status</label>
          <Select
            value={status}
            onValueChange={(value) => updateQuery({ status: value === "all" ? null : value, page: 1 })}
          >
            <SelectTrigger className="w-full sm:w-44" aria-label="Filter by status">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {STATUS_FILTERS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-medium text-muted-foreground">Sort</label>
          <Select
            value={sort}
            onValueChange={(value) =>
              updateQuery({ sort: value === "submitted_at_desc" ? null : value, page: 1 })
            }
          >
            <SelectTrigger className="w-full sm:w-40" aria-label="Sort applications">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {SORT_OPTIONS.map((option) => (
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

      {isLoading ? (
        <TableSkeleton rows={6} columns={5} label="Loading store applications" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load applications"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((token) => token + 1)}
        />
      ) : items.length === 0 ? (
        <TableEmpty
          icon={Inbox}
          title="No store applications"
          description={
            hasFilters
              ? "No applications match your filters."
              : "Applications will appear here once customers apply."
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
          <Card className="hidden [--card-spacing:0px] md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Application</TableHead>
                  <TableHead>Applicant</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Payment</TableHead>
                  <TableHead className="hidden lg:table-cell">Submitted</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <span className="flex flex-col">
                        <span className="font-medium">{item.store_name}</span>
                        <span className="font-mono text-[11px] text-muted-foreground">
                          {item.application_number}
                        </span>
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="flex flex-col">
                        <span className="text-sm">{item.applicant_name}</span>
                        <span className="text-xs text-muted-foreground">{item.applicant_email}</span>
                      </span>
                    </TableCell>
                    <TableCell>
                      <ApplicationStatusBadge status={item.status} />
                    </TableCell>
                    <TableCell>
                      {item.payment_status ? (
                        <PaymentStatusBadge status={item.payment_status} />
                      ) : (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground lg:table-cell">
                      {formatDate(item.submitted_at)}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="outline" size="sm" asChild>
                        <Link href={`/admin/store-applications/${item.id}`}>
                          <Eye className="size-3.5" aria-hidden="true" />
                          Review
                        </Link>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Card>

          <ul className="flex flex-col gap-3 md:hidden">
            {items.map((item) => (
              <li key={item.id} className="rounded-xl bg-card p-4 ring-1 ring-foreground/10">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate font-medium">{item.store_name}</p>
                    <p className="truncate text-xs text-muted-foreground">{item.applicant_email}</p>
                  </div>
                  <ApplicationStatusBadge status={item.status} />
                </div>
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  {item.payment_status && <PaymentStatusBadge status={item.payment_status} />}
                  <span className="text-xs text-muted-foreground">
                    {formatDate(item.submitted_at)}
                  </span>
                </div>
                <Button variant="outline" size="sm" className="mt-3 w-full" asChild>
                  <Link href={`/admin/store-applications/${item.id}`}>Review application</Link>
                </Button>
              </li>
            ))}
          </ul>

          <Card className="[--card-spacing:0px]">
            <TablePagination
              page={pagination?.page ?? page}
              pageCount={Math.max(1, pagination?.total_pages ?? 1)}
              total={pagination?.total ?? 0}
              onPageChange={(next) => updateQuery({ page: next })}
              itemLabel="applications"
            />
          </Card>
        </>
      )}
    </div>
  );
}

