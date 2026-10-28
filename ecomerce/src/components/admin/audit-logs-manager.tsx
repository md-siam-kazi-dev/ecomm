"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { AlertTriangle, Inbox, ScrollText, Search, ShieldCheck, Users } from "lucide-react";

import { TablePagination } from "@/components/admin/table-pagination";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { ErrorState } from "@/components/errors/error-state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
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
import { formatDateTime } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { AuditEvent, AuditEventDetail, AuditSummary } from "@/types/audit";

const PAGE_SIZES = [10, 25, 50, 100];
const DEFAULT_PAGE_SIZE = 25;

const ROLE_FILTERS = [
  { value: "all", label: "All actors" },
  { value: "admin", label: "Admin" },
  { value: "vendor", label: "Vendor" },
  { value: "customer", label: "Customer" },
  { value: "system", label: "System" },
];

const ENTITY_FILTERS = [
  { value: "all", label: "All entities" },
  { value: "user", label: "User" },
  { value: "store", label: "Store" },
  { value: "store_application", label: "Store application" },
  { value: "product", label: "Product" },
  { value: "variant", label: "Variant" },
  { value: "order", label: "Order" },
  { value: "payment", label: "Payment" },
  { value: "review", label: "Review" },
  { value: "question", label: "Question" },
  { value: "inventory", label: "Inventory" },
];

const RESULT_FILTERS = [
  { value: "all", label: "All results" },
  { value: "success", label: "Success" },
  { value: "failure", label: "Failure" },
];

type ListState =
  | { key: string; ok: true; items: AuditEvent[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

/** Admin audit log: server-side filter/search/pagination over the audit trail. */
export function AuditLogsManager() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);
  const limitParam = Number(searchParams.get("limit"));
  const limit = PAGE_SIZES.includes(limitParam) ? limitParam : DEFAULT_PAGE_SIZE;

  const search = searchParams.get("search") ?? "";
  const action = searchParams.get("action") ?? "all";
  const actorRole = searchParams.get("actor_role") ?? "all";
  const entityType = searchParams.get("entity_type") ?? "all";
  const result = searchParams.get("result") ?? "all";
  const from = searchParams.get("from") ?? "";
  const to = searchParams.get("to") ?? "";

  const [searchInput, setSearchInput] = useState(search);
  const debouncedSearch = useDebouncedValue(searchInput, 350);
  const [state, setState] = useState<ListState | null>(null);
  const [summary, setSummary] = useState<AuditSummary | null>(null);
  const [reloadToken, setReloadToken] = useState(0);
  const [selectedId, setSelectedId] = useState<string | null>(null);

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

  // Shared filters (date range + dropdowns) used by both list and summary.
  const filterParams = useMemo(() => {
    const params = new URLSearchParams();
    if (search) params.set("search", search);
    if (action !== "all") params.set("action", action);
    if (actorRole !== "all") params.set("actor_role", actorRole);
    if (entityType !== "all") params.set("entity_type", entityType);
    if (result !== "all") params.set("result", result);
    if (from) params.set("from", from);
    if (to) params.set("to", to);
    return params;
  }, [search, action, actorRole, entityType, result, from, to]);

  const listUrl = useMemo(() => {
    const params = new URLSearchParams(filterParams);
    params.set("page", String(page));
    params.set("limit", String(limit));
    return `/api/v1/admin/audit-logs?${params.toString()}`;
  }, [filterParams, page, limit]);

  const summaryUrl = useMemo(
    () => `/api/v1/admin/audit-logs/summary?${filterParams.toString()}`,
    [filterParams]
  );

  const queryKey = `${listUrl}#${reloadToken}`;

  useEffect(() => {
    let active = true;
    void apiGet<AuditEvent[]>(listUrl).then((res) => {
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
  }, [listUrl, queryKey]);

  useEffect(() => {
    let active = true;
    void apiGet<AuditSummary>(summaryUrl).then((res) => {
      if (active && res.ok) setSummary(res.data);
    });
    return () => {
      active = false;
    };
  }, [summaryUrl, reloadToken]);

  const isLoading = state?.key !== queryKey;
  const failed = !isLoading && state?.ok === false;
  const items = !isLoading && state?.ok === true ? state.items : [];
  const pagination = !isLoading && state?.ok === true ? state.pagination : null;

  const hasFilters =
    search !== "" ||
    action !== "all" ||
    actorRole !== "all" ||
    entityType !== "all" ||
    result !== "all" ||
    from !== "" ||
    to !== "";

  function clearFilters() {
    setSearchInput("");
    updateQuery({
      search: null,
      action: null,
      actor_role: null,
      entity_type: null,
      result: null,
      from: null,
      to: null,
      page: 1,
    });
  }

  return (
    <div className="flex flex-col gap-6">
      <SummaryCards summary={summary} loading={summary === null} />

      <div className="flex flex-col gap-3 rounded-xl bg-card p-4 ring-1 ring-foreground/10">
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div className="flex flex-col gap-1.5 sm:col-span-2">
            <label htmlFor="audit-search" className="text-xs font-medium text-muted-foreground">
              Search
            </label>
            <div className="relative">
              <Search
                className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
                aria-hidden="true"
              />
              <Input
                id="audit-search"
                type="search"
                className="pl-8"
                placeholder="Actor, email, action or entity id…"
                value={searchInput}
                onChange={(event) => setSearchInput(event.target.value)}
              />
            </div>
          </div>

          <FilterSelect
            label="Action"
            value={action}
            onChange={(value) => updateQuery({ action: value === "all" ? null : value, page: 1 })}
            options={[
              { value: "all", label: "All actions" },
              ...(summary?.actions ?? []).map((a) => ({ value: a.value, label: a.label })),
            ]}
          />

          <FilterSelect
            label="Actor role"
            value={actorRole}
            onChange={(value) =>
              updateQuery({ actor_role: value === "all" ? null : value, page: 1 })
            }
            options={ROLE_FILTERS}
          />

          <FilterSelect
            label="Entity"
            value={entityType}
            onChange={(value) =>
              updateQuery({ entity_type: value === "all" ? null : value, page: 1 })
            }
            options={ENTITY_FILTERS}
          />

          <FilterSelect
            label="Result"
            value={result}
            onChange={(value) => updateQuery({ result: value === "all" ? null : value, page: 1 })}
            options={RESULT_FILTERS}
          />

          <div className="flex flex-col gap-1.5">
            <label htmlFor="audit-from" className="text-xs font-medium text-muted-foreground">
              From
            </label>
            <Input
              id="audit-from"
              type="date"
              value={from}
              onChange={(event) => updateQuery({ from: event.target.value || null, page: 1 })}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="audit-to" className="text-xs font-medium text-muted-foreground">
              To
            </label>
            <Input
              id="audit-to"
              type="date"
              value={to}
              onChange={(event) => updateQuery({ to: event.target.value || null, page: 1 })}
            />
          </div>

          <div className="flex items-end gap-2">
            <Button
              variant="outline"
              size="sm"
              className="h-9"
              onClick={clearFilters}
              disabled={!hasFilters}
            >
              Clear filters
            </Button>
          </div>
        </div>
      </div>

      {isLoading ? (
        <TableSkeleton rows={8} columns={5} label="Loading audit events" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load audit events"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((token) => token + 1)}
        />
      ) : items.length === 0 ? (
        <TableEmpty
          icon={Inbox}
          title="No audit events"
          description={
            hasFilters
              ? "No events match your filters."
              : "Activity will appear here as admins, vendors and customers act."
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
                  <TableHead>Event</TableHead>
                  <TableHead>Actor</TableHead>
                  <TableHead>Entity</TableHead>
                  <TableHead>Result</TableHead>
                  <TableHead className="hidden text-right lg:table-cell">When</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((event) => (
                  <TableRow
                    key={event.id}
                    className="cursor-pointer"
                    onClick={() => setSelectedId(event.id)}
                  >
                    <TableCell>
                      <span className="flex flex-col">
                        <span className="font-medium">{event.action_label}</span>
                        <span className="font-mono text-[11px] text-muted-foreground">
                          {event.action}
                        </span>
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="flex flex-col">
                        <span className="text-sm">{event.actor.name}</span>
                        <span className="text-xs text-muted-foreground">
                          {event.actor.email ?? event.actor_role ?? "—"}
                        </span>
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="flex flex-col">
                        <span className="text-sm">{event.entity_type ?? "—"}</span>
                        {event.entity_id && (
                          <span className="font-mono text-[11px] text-muted-foreground">
                            {shortId(event.entity_id)}
                          </span>
                        )}
                      </span>
                    </TableCell>
                    <TableCell>
                      <ResultBadge result={event.result} />
                    </TableCell>
                    <TableCell className="hidden text-right text-xs text-muted-foreground lg:table-cell">
                      {formatDateTime(event.created_at)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Card>

          <ul className="flex flex-col gap-3 md:hidden">
            {items.map((event) => (
              <li key={event.id}>
                <button
                  type="button"
                  onClick={() => setSelectedId(event.id)}
                  className="w-full rounded-xl bg-card p-4 text-left ring-1 ring-foreground/10 transition-colors hover:bg-muted/40 focus-visible:outline-2 focus-visible:outline-offset-2"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="truncate font-medium">{event.action_label}</p>
                      <p className="truncate font-mono text-[11px] text-muted-foreground">
                        {event.action}
                      </p>
                    </div>
                    <ResultBadge result={event.result} />
                  </div>
                  <p className="mt-2 text-sm">{event.actor.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatDateTime(event.created_at)}
                  </p>
                </button>
              </li>
            ))}
          </ul>

          <Card className="[--card-spacing:0px]">
            <TablePagination
              page={pagination?.page ?? page}
              pageCount={Math.max(1, pagination?.total_pages ?? 1)}
              total={pagination?.total ?? 0}
              onPageChange={(next) => updateQuery({ page: next })}
              itemLabel="events"
            />
          </Card>
        </>
      )}

      <AuditDetailDialog
        eventId={selectedId}
        onOpenChange={(open) => {
          if (!open) setSelectedId(null);
        }}
      />
    </div>
  );
}

function FilterSelect({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-xs font-medium text-muted-foreground">{label}</label>
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger className="w-full" aria-label={`Filter by ${label.toLowerCase()}`}>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

function SummaryCards({ summary, loading }: { summary: AuditSummary | null; loading: boolean }) {
  const cards = [
    { label: "Total events", value: summary?.total, icon: ScrollText },
    { label: "Admin actions", value: summary?.admin_actions, icon: ShieldCheck },
    { label: "Vendor actions", value: summary?.vendor_actions, icon: Users },
    { label: "Security events", value: summary?.security_events, icon: AlertTriangle },
  ];

  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {cards.map((card) => (
        <div
          key={card.label}
          className="flex items-center gap-3 rounded-xl bg-card p-4 ring-1 ring-foreground/10"
        >
          <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary-dark">
            <card.icon className="size-4" aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <p className="text-xs font-medium text-muted-foreground">{card.label}</p>
            {loading ? (
              <Skeleton className="mt-1 h-6 w-12" />
            ) : (
              <p className="text-lg font-semibold tabular-nums">
                {(card.value ?? 0).toLocaleString()}
              </p>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

function ResultBadge({ result }: { result: string | null }) {
  if (result === "failure") {
    return <Badge variant="destructive">Failed</Badge>;
  }
  if (result === "success") {
    return <Badge variant="secondary">Success</Badge>;
  }
  return <Badge variant="outline">—</Badge>;
}

type DetailState =
  | { id: string; ok: true; data: AuditEventDetail }
  | { id: string; ok: false; message: string };

function AuditDetailDialog({
  eventId,
  onOpenChange,
}: {
  eventId: string | null;
  onOpenChange: (open: boolean) => void;
}) {
  const [state, setState] = useState<DetailState | null>(null);

  useEffect(() => {
    if (!eventId) return;
    let active = true;
    void apiGet<AuditEventDetail>(`/api/v1/admin/audit-logs/${eventId}`).then((res) => {
      if (!active) return;
      if (res.ok) setState({ id: eventId, ok: true, data: res.data });
      else setState({ id: eventId, ok: false, message: res.error.message });
    });
    return () => {
      active = false;
    };
  }, [eventId]);

  // Ignore a stale response from a previously selected event.
  const current = state?.id === eventId ? state : null;
  const detail = current?.ok ? current.data : null;
  const error = current && !current.ok ? current.message : null;

  return (
    <Dialog open={eventId !== null} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{detail?.action_label ?? "Audit event"}</DialogTitle>
          <DialogDescription>
            {detail ? (
              <span className="font-mono text-xs">{detail.action}</span>
            ) : (
              "Loading event details…"
            )}
          </DialogDescription>
        </DialogHeader>

        {error ? (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        ) : !detail ? (
          <DetailSkeleton />
        ) : (
          <div className="flex flex-col gap-4">
            <dl className="grid grid-cols-1 gap-x-6 gap-y-3 text-sm sm:grid-cols-2">
              <DetailField label="Time" value={formatDateTime(detail.created_at)} />
              <DetailField label="Result" value={detail.result ?? "—"} />
              <DetailField label="Actor" value={detail.actor.name} />
              <DetailField label="Actor role" value={detail.actor_role ?? "—"} />
              <DetailField label="Actor email" value={detail.actor.email ?? "—"} />
              <DetailField label="Actor id" value={detail.actor.id ?? "—"} mono />
              <DetailField
                label="Entity"
                value={detail.entity_type ? `${detail.entity_type}` : "—"}
              />
              <DetailField label="Entity id" value={detail.entity_id ?? "—"} mono />
              <DetailField label="IP address" value={detail.ip_address ?? "—"} mono />
              <DetailField label="Request id" value={detail.request_id ?? "—"} mono />
              <DetailField
                label="User agent"
                value={detail.user_agent ?? "—"}
                className="sm:col-span-2"
              />
              {detail.reason && (
                <DetailField label="Reason" value={detail.reason} className="sm:col-span-2" />
              )}
            </dl>

            <JSONBlock title="Changes before" value={detail.old_values} />
            <JSONBlock title="Changes after" value={detail.new_values} />
            <JSONBlock title="Metadata" value={detail.metadata} />
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

function DetailField({
  label,
  value,
  mono = false,
  className,
}: {
  label: string;
  value: string;
  mono?: boolean;
  className?: string;
}) {
  return (
    <div className={cn("flex flex-col gap-0.5", className)}>
      <dt className="text-xs font-medium text-muted-foreground">{label}</dt>
      <dd className={cn("break-words", mono && "font-mono text-xs")}>{value}</dd>
    </div>
  );
}

function JSONBlock({ title, value }: { title: string; value: Record<string, unknown> | null }) {
  if (!value || Object.keys(value).length === 0) return null;
  return (
    <div className="flex flex-col gap-1">
      <p className="text-xs font-medium text-muted-foreground">{title}</p>
      <pre className="overflow-x-auto rounded-lg bg-muted/60 p-3 text-xs leading-relaxed">
        {JSON.stringify(value, null, 2)}
      </pre>
    </div>
  );
}

function DetailSkeleton() {
  return (
    <div role="status" aria-label="Loading event" className="flex flex-col gap-3">
      {Array.from({ length: 5 }).map((_, index) => (
        <Skeleton key={index} className="h-4 w-full" />
      ))}
    </div>
  );
}

function shortId(id: string): string {
  return id.length > 12 ? `${id.slice(0, 8)}…` : id;
}
