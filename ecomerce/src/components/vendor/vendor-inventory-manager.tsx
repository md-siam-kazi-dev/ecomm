"use client";

import { useCallback, useEffect, useState } from "react";
import Image from "next/image";
import { History, PackagePlus, SearchX } from "lucide-react";
import { toast } from "sonner";

import { StockBadge, stockLevel } from "@/components/admin/stock-badge";
import { TablePagination } from "@/components/admin/table-pagination";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { ErrorState } from "@/components/errors/error-state";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { ExportMenu } from "@/components/vendor/export-menu";
import { InventoryHistory } from "@/components/vendor/inventory-history";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { apiGet, apiPost, apiPut, type PaginationMeta } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import type { InventoryItem } from "@/types/inventory";

const PAGE_SIZE = 20;
/** Stock at which the level bar reads as full. */
const FULL_STOCK = 40;

const STOCK_OPTIONS = [
  { value: "all", label: "All stock" },
  { value: "in", label: "In stock" },
  { value: "low", label: "Low stock" },
  { value: "out", label: "Out of stock" },
  { value: "unpublished", label: "Unpublished" },
];

type State =
  | { key: string; ok: true; items: InventoryItem[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

/**
 * Vendor inventory (DB-backed): stock per product variant with server-side
 * search, stock filter and pagination, plus a persisted restock action.
 */
export function VendorInventoryManager() {
  const [searchInput, setSearchInput] = useState("");
  const search = useDebouncedValue(searchInput, 350);
  const [stock, setStock] = useState("all");
  const [page, setPage] = useState(1);
  const [state, setState] = useState<State | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const requestUrl =
    `/api/v1/vendor/inventory?page=${page}&limit=${PAGE_SIZE}` +
    (stock !== "all" ? `&status=${stock}` : "") +
    (search ? `&search=${encodeURIComponent(search)}` : "");
  const queryKey = `${requestUrl}#${reloadToken}`;
  const load = useCallback(() => apiGet<InventoryItem[]>(requestUrl), [requestUrl]);

  const exportParams = new URLSearchParams();
  if (stock !== "all") exportParams.set("status", stock);
  if (search) exportParams.set("search", search);
  const exportQueryString = exportParams.toString();
  const exportUrl = `/api/v1/vendor/inventory/export${exportQueryString ? `?${exportQueryString}` : ""}`;

  useEffect(() => {
    let active = true;
    void load().then((res) => {
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
  }, [load, queryKey]);

  const isLoading = state?.key !== queryKey;
  const failed = !isLoading && state?.ok === false;
  const items = !isLoading && state?.ok === true ? state.items : [];
  const pagination = !isLoading && state?.ok === true ? state.pagination : null;
  const hasFilters = search !== "" || stock !== "all";

  // Restock + threshold dialog.
  const [restocking, setRestocking] = useState<InventoryItem | null>(null);
  const [restockValue, setRestockValue] = useState("");
  const [restockThreshold, setRestockThreshold] = useState("");
  const [saving, setSaving] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);

  function openRestock(item: InventoryItem) {
    setRestocking(item);
    setRestockValue(String(item.quantity));
    setRestockThreshold(String(item.reorder_level));
  }

  async function confirmRestock() {
    if (!restocking) return;
    const next = Number.parseInt(restockValue, 10);
    const threshold = Number.parseInt(restockThreshold, 10);
    if (!Number.isFinite(next) || next < 0) {
      toast.error("Enter a valid stock count");
      return;
    }
    if (!Number.isFinite(threshold) || threshold < 0) {
      toast.error("Enter a valid low-stock threshold");
      return;
    }
    setSaving(true);

    const thresholdRes = await apiPut<InventoryItem>(
      `/api/v1/vendor/inventory/${restocking.variant_id}/threshold`,
      { reorder_level: threshold }
    );
    if (!thresholdRes.ok) {
      setSaving(false);
      toast.error(thresholdRes.error.message);
      return;
    }

    const res = await apiPost<InventoryItem>(
      `/api/v1/vendor/inventory/${restocking.variant_id}/adjust`,
      { quantity: next }
    );
    setSaving(false);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Inventory updated", {
      description: `${restocking.name} set to ${next} units · threshold ${threshold}.`,
    });
    setRestocking(null);
    setReloadToken((token) => token + 1);
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="sm:max-w-xs sm:flex-1">
          <label htmlFor="vendor-inventory-search" className="sr-only">
            Search inventory
          </label>
          <Input
            id="vendor-inventory-search"
            type="search"
            placeholder="Search by name or SKU"
            value={searchInput}
            onChange={(event) => {
              setSearchInput(event.target.value);
              setPage(1);
            }}
          />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <Select
            value={stock}
            onValueChange={(value) => {
              setStock(value);
              setPage(1);
            }}
          >
            <SelectTrigger aria-label="Filter by stock" className="w-full sm:w-44">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {STOCK_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button variant="outline" size="sm" onClick={() => setHistoryOpen(true)}>
            <History className="size-3.5" aria-hidden="true" />
            Activity
          </Button>
          <ExportMenu baseUrl={exportUrl} />
        </div>
      </div>

      {isLoading ? (
        <TableSkeleton rows={5} columns={5} label="Loading inventory" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load inventory"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((token) => token + 1)}
        />
      ) : items.length === 0 ? (
        <TableEmpty
          icon={SearchX}
          title="No inventory items"
          description={
            hasFilters
              ? "No inventory items match your filters."
              : "You have no products to track yet."
          }
          action={
            hasFilters ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearchInput("");
                  setStock("all");
                  setPage(1);
                }}
              >
                Clear filters
              </Button>
            ) : undefined
          }
        />
      ) : (
        <Card className="[--card-spacing:0px]">
          {/* Desktop / tablet: table */}
          <div className="hidden md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Product</TableHead>
                  <TableHead className="hidden lg:table-cell">Category</TableHead>
                  <TableHead>Stock level</TableHead>
                  <TableHead className="hidden xl:table-cell">Threshold</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.variant_id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        {item.image_url ? (
                          <Image
                            src={item.image_url}
                            alt=""
                            width={40}
                            height={40}
                            className="size-10 shrink-0 rounded-md object-cover"
                          />
                        ) : (
                          <span className="size-10 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                        )}
                        <span className="flex min-w-0 flex-col">
                          <span className="max-w-[220px] truncate font-medium">{item.name}</span>
                          <span className="font-mono text-xs text-muted-foreground">{item.sku}</span>
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground lg:table-cell">
                      {item.category || "—"}
                    </TableCell>
                    <TableCell>
                      <StockBar quantity={item.quantity} />
                    </TableCell>
                    <TableCell className="hidden tabular-nums text-muted-foreground xl:table-cell">
                      {item.reorder_level}
                    </TableCell>
                    <TableCell>
                      <StockBadge stock={item.quantity} />
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => openRestock(item)}
                        aria-label={`Adjust stock for ${item.name}`}
                      >
                        <PackagePlus className="size-3.5" aria-hidden="true" />
                        Restock
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Mobile: cards */}
          <ul className="flex flex-col divide-y md:hidden">
            {items.map((item) => (
              <li key={item.variant_id} className="flex items-start gap-3 p-4">
                {item.image_url ? (
                  <Image
                    src={item.image_url}
                    alt=""
                    width={44}
                    height={44}
                    className="size-11 shrink-0 rounded-md object-cover"
                  />
                ) : (
                  <span className="size-11 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                )}
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{item.name}</p>
                  <p className="font-mono text-xs text-muted-foreground">{item.sku}</p>
                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    <StockBadge stock={item.quantity} />
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {item.quantity} in stock
                    </span>
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Low-stock threshold: {item.reorder_level}
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    className="mt-3 w-full"
                    onClick={() => openRestock(item)}
                  >
                    <PackagePlus className="size-3.5" aria-hidden="true" />
                    Adjust stock
                  </Button>
                </div>
              </li>
            ))}
          </ul>

          <div className="px-2 pb-2">
            <TablePagination
              page={pagination?.page ?? page}
              pageCount={Math.max(1, pagination?.total_pages ?? 1)}
              total={pagination?.total ?? 0}
              onPageChange={setPage}
              itemLabel="items"
            />
          </div>
        </Card>
      )}

      <Dialog
        open={restocking !== null}
        onOpenChange={(open) => {
          if (!open) setRestocking(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Adjust stock</DialogTitle>
            <DialogDescription>
              {restocking?.name} · {restocking?.sku}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="restock-value">Units in stock</Label>
            <Input
              id="restock-value"
              type="number"
              min={0}
              inputMode="numeric"
              value={restockValue}
              onChange={(event) => setRestockValue(event.target.value)}
            />
            {restocking && restocking.reserved > 0 && (
              <p className="text-xs text-muted-foreground">
                {restocking.reserved} unit{restocking.reserved === 1 ? "" : "s"} reserved for open
                orders — stock can&apos;t go below that.
              </p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="restock-threshold">Low-stock threshold</Label>
            <Input
              id="restock-threshold"
              type="number"
              min={0}
              inputMode="numeric"
              value={restockThreshold}
              onChange={(event) => setRestockThreshold(event.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              You&apos;ll be alerted when stock drops to this level or below.
            </p>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setRestocking(null)}
              disabled={saving}
            >
              Cancel
            </Button>
            <Button type="button" onClick={confirmRestock} disabled={saving}>
              Save stock
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <InventoryHistory open={historyOpen} onOpenChange={setHistoryOpen} />
    </div>
  );
}

function StockBar({ quantity }: { quantity: number }) {
  const percentage = Math.min(100, Math.round((quantity / FULL_STOCK) * 100));
  // Reuse the shared low-stock definition for the bar colour.
  const level = stockLevel(quantity);
  return (
    <div className="flex items-center gap-3">
      <span aria-hidden="true" className="h-2 w-24 overflow-hidden rounded-full bg-muted">
        <span
          className={cn(
            "block h-full rounded-full",
            level === "out" && "bg-destructive",
            level === "low" && "bg-warning",
            level === "in" && "bg-success"
          )}
          style={{ width: `${Math.max(2, percentage)}%` }}
        />
      </span>
      <span className="text-sm tabular-nums">{quantity}</span>
    </div>
  );
}

