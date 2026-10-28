"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Check, Loader2, PackageSearch, SearchX, Truck } from "lucide-react";
import { toast } from "sonner";

import { TablePagination } from "@/components/admin/table-pagination";
import { FilterTabs } from "@/components/dashboard/filter-tabs";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { OrderStatusBadge } from "@/components/orders/order-status-badge";
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
import { Label } from "@/components/ui/label";
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
import { ExportMenu } from "@/components/vendor/export-menu";
import { apiGet, apiPost } from "@/lib/api-client";
import { formatDate, formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { OrderStatus } from "@/types/order";
import type { VendorOrderDetail, VendorOrderSummary } from "@/lib/vendor";

const PAGE_SIZE = 10;
const KNOWN_ORDER_STATUSES = new Set(["pending", "processing", "shipped", "delivered", "cancelled"]);

const STATUS_FILTERS = [
  { value: "all", label: "All" },
  { value: "pending", label: "Pending" },
  { value: "processing", label: "Processing" },
  { value: "packed", label: "Packed" },
  { value: "shipped", label: "Shipped" },
  { value: "delivered", label: "Delivered" },
  { value: "cancelled", label: "Cancelled" },
];

/** The shipment process steps, in order. */
const FLOW = [
  { status: "packed", label: "Packed", hint: "Items packed and ready for the carrier." },
  { status: "shipped", label: "Shipped", hint: "Handed to the carrier — add a tracking number." },
  { status: "out_for_delivery", label: "Out for delivery", hint: "On the final leg to the buyer." },
  { status: "delivered", label: "Delivered", hint: "Received by the buyer." },
] as const;
const FLOW_ORDER: string[] = FLOW.map((step) => step.status);

/** All valid shipment statuses (the flow plus exception states). */
const SHIPMENT_STATUSES = [
  ...FLOW.map((step) => ({ value: step.status, label: step.label })),
  { value: "in_transit", label: "In transit" },
  { value: "failed", label: "Delivery failed" },
  { value: "returned", label: "Returned" },
];

/** Normalize an intermediate status onto the flow (in transit == shipped). */
function normalizeStatus(status: string): string {
  return status === "in_transit" ? "shipped" : status;
}

/** How far along the flow a detail is (index into FLOW, -1 = nothing shipped yet). */
function flowProgress(detail: VendorOrderDetail): number {
  let reached = -1;
  const bump = (raw?: string) => {
    if (!raw) return;
    const index = FLOW_ORDER.indexOf(normalizeStatus(raw));
    if (index > reached) reached = index;
  };
  for (const event of detail.events ?? []) bump(event.status);
  if (detail.shipment) bump(detail.shipment.status);
  bump(detail.status);
  return reached;
}

interface ShipmentForm {
  carrier: string;
  trackingNumber: string;
  status: string;
  location: string;
  description: string;
}

/** Store order list + fulfilment: open an order to run its shipment process. */
export function VendorOrdersTable({ orders }: { orders: VendorOrderSummary[] }) {
  const router = useRouter();
  const [status, setStatus] = useState("all");
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);

  // Exports honour the active status filter.
  const exportUrl = `/api/v1/vendor/orders/export${status !== "all" ? `?status=${status}` : ""}`;

  const [detail, setDetail] = useState<VendorOrderDetail | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [form, setForm] = useState<ShipmentForm>({
    carrier: "",
    trackingNumber: "",
    status: "packed",
    location: "",
    description: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [statusUpdating, setStatusUpdating] = useState(false);

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();
    return orders.filter((order) => {
      if (status !== "all" && order.status !== status) return false;
      if (term && !order.order_number.toLowerCase().includes(term)) return false;
      return true;
    });
  }, [orders, status, query]);

  const pageCount = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const safePage = Math.min(page, pageCount);
  const pageItems = filtered.slice((safePage - 1) * PAGE_SIZE, safePage * PAGE_SIZE);

  const reached = detail ? flowProgress(detail) : -1;
  const nextStep = FLOW.find((_, index) => index > reached) ?? null;
  const targetStatus = form.status;
  const delivered = detail ? reached >= FLOW.length - 1 : false;
  const deliveredAt = detail
    ? detail.shipment?.delivered_at ??
      (detail.events ?? []).filter((event) => event.status === "delivered").at(-1)?.event_time
    : undefined;

  function applyDetail(data: VendorOrderDetail) {
    const next = FLOW.find((_, index) => index > flowProgress(data));
    setDetail(data);
    setForm({
      carrier: data.shipment?.carrier ?? "",
      trackingNumber: data.shipment?.tracking_number ?? "",
      status: next ? next.status : "delivered",
      location: "",
      description: "",
    });
  }

  async function openOrder(orderId: string) {
    if (!orderId) {
      toast.error("This order has no id — restart the API so it returns one.");
      return;
    }
    setDialogOpen(true);
    setDetailLoading(true);
    setDetail(null);
    const res = await apiGet<VendorOrderDetail>(`/api/v1/vendor/orders/${orderId}`);
    setDetailLoading(false);
    if (!res.ok) {
      toast.error(res.error.message, { description: `HTTP ${res.status} · ${res.error.code}` });
      setDialogOpen(false);
      return;
    }
    applyDetail(res.data);
  }

  async function submitShipment() {
    if (!detail) return;
    if ((targetStatus === "shipped" || targetStatus === "out_for_delivery") && !form.trackingNumber.trim()) {
      toast.error("Add a tracking number before marking this as shipped.");
      return;
    }

    setSubmitting(true);
    const res = await apiPost<VendorOrderDetail>(`/api/v1/vendor/orders/${detail.id}/shipment`, {
      carrier: form.carrier,
      tracking_number: form.trackingNumber,
      status: targetStatus,
      location: form.location,
      description: form.description,
    });
    setSubmitting(false);
    if (!res.ok) {
      toast.error(res.error.message, { description: `HTTP ${res.status} · ${res.error.code}` });
      return;
    }
    applyDetail(res.data);
    toast.success("Shipment updated");
    router.refresh();
  }

  const targetLabel = SHIPMENT_STATUSES.find((option) => option.value === targetStatus)?.label ?? targetStatus;

  /** Apply an order-level fulfilment transition (confirm / start processing). */
  async function setOrderStatus(next: string) {
    if (!detail) return;
    setStatusUpdating(true);
    const res = await apiPost<VendorOrderDetail>(`/api/v1/vendor/orders/${detail.id}/status`, {
      status: next,
    });
    setStatusUpdating(false);
    if (!res.ok) {
      toast.error(res.error.message, { description: `HTTP ${res.status} · ${res.error.code}` });
      return;
    }
    applyDetail(res.data);
    toast.success(`Order marked ${next}`);
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3">
        <FilterTabs
          label="Filter orders by status"
          value={status}
          onValueChange={(next) => {
            setStatus(next);
            setPage(1);
          }}
          tabs={STATUS_FILTERS}
        />

        <div className="flex flex-wrap items-center gap-2">
          <div className="sm:max-w-xs sm:flex-1">
            <label htmlFor="vendor-order-search" className="sr-only">
              Search your orders
            </label>
            <Input
              id="vendor-order-search"
              type="search"
              placeholder="Search by order number"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setPage(1);
              }}
            />
          </div>
          <ExportMenu baseUrl={exportUrl} />
        </div>
      </div>

      {orders.length === 0 ? (
        <TableEmpty
          icon={PackageSearch}
          title="No orders yet"
          description="Orders containing your products will appear here."
        />
      ) : filtered.length === 0 ? (
        <TableEmpty
          icon={SearchX}
          title="No orders match your filters"
          description="Try a different status or search term."
        />
      ) : (
        <Card className="[--card-spacing:0px]">
          {/* Desktop / tablet: table */}
          <div className="hidden md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Order</TableHead>
                  <TableHead className="hidden md:table-cell">Date</TableHead>
                  <TableHead>Items</TableHead>
                  <TableHead>Your revenue</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Shipment</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pageItems.map((order) => (
                  <TableRow key={order.id}>
                    <TableCell className="font-medium">{order.order_number}</TableCell>
                    <TableCell className="hidden text-muted-foreground md:table-cell">
                      {formatDate(order.created_at)}
                    </TableCell>
                    <TableCell className="tabular-nums">{order.item_count}</TableCell>
                    <TableCell className="font-medium tabular-nums">{formatPrice(order.total)}</TableCell>
                    <TableCell>
                      <OrderStatusCell status={order.status} />
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        size="sm"
                        onClick={() => openOrder(order.id)}
                        aria-label={`Manage shipment for order ${order.order_number}`}
                      >
                        <Truck className="size-3.5" aria-hidden="true" />
                        Manage
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Mobile: cards */}
          <ul className="flex flex-col divide-y md:hidden">
            {pageItems.map((order) => (
              <li key={order.id} className="flex flex-col gap-3 p-4">
                <div className="flex items-start justify-between gap-3">
                  <p className="min-w-0 truncate font-medium">{order.order_number}</p>
                  <OrderStatusCell status={order.status} />
                </div>

                <dl className="grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                  <div className="flex flex-col">
                    <dt>Date</dt>
                    <dd className="text-foreground">{formatDate(order.created_at)}</dd>
                  </div>
                  <div className="flex flex-col">
                    <dt>Items</dt>
                    <dd className="text-foreground tabular-nums">{order.item_count}</dd>
                  </div>
                  <div className="flex flex-col">
                    <dt>Your revenue</dt>
                    <dd className="font-medium text-foreground tabular-nums">
                      {formatPrice(order.total)}
                    </dd>
                  </div>
                </dl>

                <Button
                  size="sm"
                  className="w-full"
                  onClick={() => openOrder(order.id)}
                  aria-label={`Manage shipment for order ${order.order_number}`}
                >
                  <Truck className="size-3.5" aria-hidden="true" />
                  Manage shipment
                </Button>
              </li>
            ))}
          </ul>

          <div className="px-2 pb-2">
            <TablePagination
              page={safePage}
              pageCount={pageCount}
              total={filtered.length}
              onPageChange={setPage}
              itemLabel="orders"
            />
          </div>
        </Card>
      )}

      <Dialog
        open={dialogOpen}
        onOpenChange={(open) => {
          if (!open) {
            setDialogOpen(false);
            setDetail(null);
          }
        }}
      >
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          {detailLoading ? (
            <div className="flex flex-col gap-3" role="status" aria-label="Loading order">
              <Skeleton className="h-6 w-40" />
              <Skeleton className="h-40 w-full rounded-xl" />
            </div>
          ) : detail ? (
            <>
              <DialogHeader>
                <DialogTitle>Shipment · {detail.order_number}</DialogTitle>
                <DialogDescription>
                  Placed {formatDate(detail.created_at)} · your revenue {formatPrice(detail.vendor_revenue)}
                </DialogDescription>
              </DialogHeader>

              <div className="flex flex-wrap items-center gap-2">
                <OrderStatusCell status={detail.status} />
                {detail.shipment?.tracking_number && (
                  <span className="text-sm text-muted-foreground">
                    {detail.shipment.carrier ? `${detail.shipment.carrier} · ` : ""}
                    {detail.shipment.tracking_number}
                  </span>
                )}
              </div>

              {(detail.status === "pending" || detail.status === "confirmed") && (
                <div className="flex flex-wrap items-center gap-3 rounded-xl border border-dashed p-3">
                  <p className="mr-auto text-sm text-muted-foreground">
                    {detail.status === "pending"
                      ? "Confirm this order to begin fulfilment."
                      : "Begin picking and packing this order."}
                  </p>
                  <Button
                    type="button"
                    size="sm"
                    disabled={statusUpdating}
                    onClick={() => setOrderStatus(detail.status === "pending" ? "confirmed" : "processing")}
                  >
                    {statusUpdating && <Loader2 className="size-4 animate-spin" aria-hidden="true" />}
                    {detail.status === "pending" ? "Confirm order" : "Start processing"}
                  </Button>
                </div>
              )}

              {/* Shipment process stepper */}
              <ol className="flex flex-col gap-3 sm:flex-row sm:items-start sm:gap-2">
                {FLOW.map((step, index) => {
                  const done = index <= reached;
                  const current = index === reached + 1;
                  return (
                    <li key={step.status} className="flex flex-1 items-start gap-3 sm:flex-col sm:items-center sm:text-center">
                      <span
                        className={cn(
                          "flex size-8 shrink-0 items-center justify-center rounded-full border text-xs font-semibold",
                          done && "border-primary bg-primary text-primary-foreground",
                          current && "border-primary text-primary-dark",
                          !done && !current && "text-muted-foreground"
                        )}
                        aria-hidden="true"
                      >
                        {done ? <Check className="size-4" /> : index + 1}
                      </span>
                      <span className="min-w-0">
                        <span className={cn("block text-sm font-medium", done && "text-foreground")}>{step.label}</span>
                        <span className="hidden text-xs text-muted-foreground sm:block">{step.hint}</span>
                      </span>
                    </li>
                  );
                })}
              </ol>

              {delivered ? (
                <div className="flex items-center gap-3 rounded-xl border border-primary/30 bg-primary/5 p-4">
                  <span
                    aria-hidden="true"
                    className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground"
                  >
                    <Check className="size-5" />
                  </span>
                  <div>
                    <p className="text-sm font-semibold">Order delivery complete</p>
                    <p className="text-xs text-muted-foreground">
                      {deliveredAt
                        ? `Delivered on ${formatDate(deliveredAt)}.`
                        : "This order has been delivered."}
                    </p>
                  </div>
                </div>
              ) : (
                <div className="rounded-xl border p-4">
                  <p className="text-sm font-medium">
                    {nextStep ? `Next step: ${nextStep.label}` : "Shipment complete"}
                  </p>
                  {nextStep && <p className="mt-0.5 text-xs text-muted-foreground">{nextStep.hint}</p>}

                  <div className="mt-3 grid gap-3 sm:grid-cols-2">
                    <div className="flex flex-col gap-1.5">
                      <Label htmlFor="ship-carrier">Carrier</Label>
                      <Input
                        id="ship-carrier"
                        value={form.carrier}
                        onChange={(event) => setForm((p) => ({ ...p, carrier: event.target.value }))}
                        placeholder="DHL, Pathao…"
                      />
                    </div>
                    <div className="flex flex-col gap-1.5">
                      <Label htmlFor="ship-tracking">Tracking number</Label>
                      <Input
                        id="ship-tracking"
                        value={form.trackingNumber}
                        onChange={(event) => setForm((p) => ({ ...p, trackingNumber: event.target.value }))}
                      />
                    </div>
                    <div className="flex flex-col gap-1.5">
                      <Label htmlFor="ship-status">Status</Label>
                      <Select value={form.status} onValueChange={(value) => setForm((p) => ({ ...p, status: value }))}>
                        <SelectTrigger id="ship-status" className="w-full">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {SHIPMENT_STATUSES.map((option) => (
                            <SelectItem key={option.value} value={option.value}>
                              {option.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      {targetStatus === "failed" && (
                        <p className="text-xs text-muted-foreground">
                          Records a failed delivery attempt and notifies the buyer.
                        </p>
                      )}
                    </div>
                    <div className="flex flex-col gap-1.5">
                      <Label htmlFor="ship-location">Location (optional)</Label>
                      <Input
                        id="ship-location"
                        value={form.location}
                        onChange={(event) => setForm((p) => ({ ...p, location: event.target.value }))}
                        placeholder="Dhaka hub"
                      />
                    </div>
                    <div className="flex flex-col gap-1.5 sm:col-span-2">
                      <Label htmlFor="ship-description">Note (optional)</Label>
                      <textarea
                        id="ship-description"
                        rows={2}
                        value={form.description}
                        onChange={(event) => setForm((p) => ({ ...p, description: event.target.value }))}
                        className={cn(
                          "w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none",
                          "focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
                        )}
                      />
                    </div>
                  </div>
                  <Button type="button" onClick={submitShipment} disabled={submitting} className="mt-3">
                    {submitting && <Loader2 className="size-4 animate-spin" aria-hidden="true" />}
                    Mark as {targetLabel.toLowerCase()}
                  </Button>
                </div>
              )}

              {/* Items */}
              <div>
                <p className="text-sm font-medium">Your items</p>
                <ul className="mt-2 flex flex-col gap-2">
                  {(detail.items ?? []).map((item) => (
                    <li key={item.order_item_id} className="flex items-baseline justify-between gap-4 text-sm">
                      <span className="min-w-0">
                        <span className="font-medium">{item.name}</span>
                        <span className="text-muted-foreground"> × {item.quantity}</span>
                      </span>
                      <span className="shrink-0 tabular-nums">{formatPrice(item.line_total)}</span>
                    </li>
                  ))}
                </ul>
              </div>

              {/* Timeline */}
              <div>
                <p className="text-sm font-medium">Timeline</p>
                {(detail.events ?? []).length === 0 ? (
                  <p className="mt-2 text-sm text-muted-foreground">No tracking updates yet.</p>
                ) : (
                  <ol className="mt-2 flex flex-col gap-3 border-l pl-4">
                    {(detail.events ?? []).map((event) => (
                      <li key={event.id} className="relative">
                        <span className="absolute -left-[21px] top-1.5 size-2 rounded-full bg-primary" aria-hidden="true" />
                        <p className="text-sm font-medium capitalize">{event.status.replace(/_/g, " ")}</p>
                        <p className="text-xs text-muted-foreground">
                          {formatDate(event.event_time)}
                          {event.location ? ` · ${event.location}` : ""}
                        </p>
                        {event.description && <p className="text-sm text-muted-foreground">{event.description}</p>}
                      </li>
                    ))}
                  </ol>
                )}
              </div>

              {/* Delivery address */}
              <div className="rounded-xl border p-4">
                <p className="text-sm font-medium">Ship to</p>
                <address className="mt-1 text-sm not-italic leading-relaxed text-muted-foreground">
                  {detail.shipping_address.full_name}
                  <br />
                  {detail.shipping_address.line1}
                  {detail.shipping_address.line2 && <><br />{detail.shipping_address.line2}</>}
                  <br />
                  {detail.shipping_address.city}
                  {detail.shipping_address.state ? `, ${detail.shipping_address.state}` : ""}
                  {detail.shipping_address.postal_code ? ` ${detail.shipping_address.postal_code}` : ""}
                  <br />
                  {detail.shipping_address.country}
                  {detail.shipping_address.phone && <><br />{detail.shipping_address.phone}</>}
                </address>
              </div>
            </>
          ) : null}
        </DialogContent>
      </Dialog>
    </div>
  );
}

/** Order status pill — falls back to a plain badge for unknown statuses. */
function OrderStatusCell({ status }: { status: string }) {
  return KNOWN_ORDER_STATUSES.has(status) ? (
    <OrderStatusBadge status={status as OrderStatus} />
  ) : (
    <Badge variant="secondary" className="capitalize">
      {status}
    </Badge>
  );
}
