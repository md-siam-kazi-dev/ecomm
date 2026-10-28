"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, History } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { apiGet } from "@/lib/api-client";
import { formatDate } from "@/lib/formatters";
import type { InventoryMovement } from "@/types/inventory";

const PAGE_SIZE = 20;

type State =
  | { key: string; ok: true; items: InventoryMovement[]; total: number }
  | { key: string; ok: false; message: string };

/**
 * Inventory activity log: the store's stock movements (newest first) with the
 * level before and after each change. Optionally scoped to one variant.
 */
export function InventoryHistory({
  open,
  onOpenChange,
  variantId,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  variantId?: string;
}) {
  const [page, setPage] = useState(1);
  const [state, setState] = useState<State | null>(null);

  const url =
    `/api/v1/vendor/inventory/movements?page=${page}&limit=${PAGE_SIZE}` +
    (variantId ? `&variant_id=${encodeURIComponent(variantId)}` : "");
  const key = url;

  useEffect(() => {
    if (!open) return;
    let active = true;
    void apiGet<InventoryMovement[]>(url).then((res) => {
      if (!active) return;
      if (res.ok && res.pagination) {
        setState({ key, ok: true, items: res.data, total: res.pagination.total });
      } else {
        setState({
          key,
          ok: false,
          message: res.ok ? "Pagination metadata was missing." : res.error.message,
        });
      }
    });
    return () => {
      active = false;
    };
  }, [open, url, key]);

  // Reset to the first page whenever the dialog opens.
  function handleOpenChange(next: boolean) {
    if (next) setPage(1);
    onOpenChange(next);
  }

  const loading = state?.key !== key;
  const failed = !loading && state?.ok === false;
  const items = !loading && state?.ok === true ? state.items : [];
  const total = !loading && state?.ok === true ? state.total : 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Inventory activity</DialogTitle>
          <DialogDescription>
            Recent stock movements{variantId ? " for this product" : ""}.
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex flex-col gap-2" role="status" aria-label="Loading activity">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton key={index} className="h-14 w-full" />
            ))}
          </div>
        ) : failed ? (
          <div role="alert" className="flex flex-col items-center gap-3 py-10 text-center">
            <AlertTriangle className="size-7 text-destructive" aria-hidden="true" />
            <p className="text-sm text-muted-foreground">{(state as { message: string }).message}</p>
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center gap-3 py-16 text-center">
            <History className="size-7 text-muted-foreground" aria-hidden="true" />
            <p className="text-sm text-muted-foreground">No stock movements yet.</p>
          </div>
        ) : (
          <ul className="flex flex-col divide-y">
            {items.map((movement) => (
              <li key={movement.id} className="flex flex-col gap-1 py-3">
                <div className="flex items-start justify-between gap-3">
                  <p className="min-w-0 truncate text-sm font-medium">{movement.name}</p>
                  <span className="shrink-0 text-xs capitalize text-muted-foreground">
                    {movement.movement_type.replace(/_/g, " ")}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground">
                  <span className="font-mono">{movement.sku}</span>
                  {" · "}
                  {movement.quantity_before != null && movement.quantity_after != null ? (
                    <span className="tabular-nums">
                      {movement.quantity_before} → {movement.quantity_after}
                    </span>
                  ) : (
                    <span className="tabular-nums">
                      {movement.movement_type === "sale" ? "-" : "+"}
                      {movement.quantity}
                    </span>
                  )}
                  {" · "}
                  {formatDate(movement.created_at)}
                </p>
                <p className="text-xs text-muted-foreground">
                  {movement.actor ? `by ${movement.actor}` : "system"}
                  {movement.order_number ? ` · order ${movement.order_number}` : ""}
                  {movement.notes ? ` · ${movement.notes}` : ""}
                </p>
              </li>
            ))}
          </ul>
        )}

        {pageCount > 1 && (
          <div className="flex items-center justify-between gap-3 border-t pt-3">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Previous
            </Button>
            <span className="text-xs text-muted-foreground">
              Page {page} of {pageCount}
            </span>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={page >= pageCount}
              onClick={() => setPage((p) => Math.min(pageCount, p + 1))}
            >
              Next
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
