"use client";

import { useMemo, useState } from "react";
import { Eye, SearchX } from "lucide-react";

import { TablePagination } from "@/components/admin/table-pagination";
import { FilterTabs } from "@/components/dashboard/filter-tabs";
import { TableEmpty } from "@/components/dashboard/table-empty";
import {
  ORDER_STATUS_LABELS,
  OrderStatusBadge,
} from "@/components/orders/order-status-badge";
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatDate, formatPrice } from "@/lib/formatters";
import type { Order, OrderStatus } from "@/types/order";
import type { User } from "@/types/user";

const PAGE_SIZE = 8;
const STATUSES: OrderStatus[] = [
  "pending",
  "processing",
  "shipped",
  "delivered",
  "cancelled",
];

type StatusFilter = "all" | OrderStatus;

interface OrdersTableProps {
  orders: Order[];
  customers: User[];
}

/** Admin orders (PRD §24): status filters, search and a detail dialog. */
export function OrdersTable({ orders, customers }: OrdersTableProps) {
  const [status, setStatus] = useState<StatusFilter>("all");
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [activeOrder, setActiveOrder] = useState<Order | null>(null);

  const customerNames = useMemo(
    () => new Map(customers.map((customer) => [customer.id, customer.name])),
    [customers]
  );

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();

    return orders.filter((order) => {
      if (status !== "all" && order.status !== status) return false;
      if (term) {
        const customerName = customerNames.get(order.customerId) ?? "";
        const haystack =
          `${order.orderNumber} ${customerName} ${order.customerId}`.toLowerCase();
        if (!haystack.includes(term)) return false;
      }
      return true;
    });
  }, [orders, status, query, customerNames]);

  const pageCount = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const safePage = Math.min(page, pageCount);
  const pageItems = filtered.slice(
    (safePage - 1) * PAGE_SIZE,
    safePage * PAGE_SIZE
  );

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3">
        <FilterTabs
          label="Filter orders by status"
          value={status}
          onValueChange={(next) => {
            setStatus(next as StatusFilter);
            setPage(1);
          }}
          tabs={[
            { value: "all", label: "All" },
            ...STATUSES.map((value) => ({
              value,
              label: ORDER_STATUS_LABELS[value],
            })),
          ]}
        />

        <div className="sm:max-w-xs">
          <label htmlFor="order-search" className="sr-only">
            Search orders
          </label>
          <Input
            id="order-search"
            type="search"
            placeholder="Search by order number or customer"
            value={query}
            onChange={(event) => {
              setQuery(event.target.value);
              setPage(1);
            }}
          />
        </div>
      </div>

      <Card className="hidden [--card-spacing:0px] md:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Order</TableHead>
              <TableHead>Customer</TableHead>
              <TableHead className="hidden md:table-cell">Date</TableHead>
              <TableHead className="hidden sm:table-cell">Items</TableHead>
              <TableHead>Total</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pageItems.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="py-6">
                  <TableEmpty
                    icon={SearchX}
                    title="No orders match your filters"
                    description="Try a different status or search term."
                  />
                </TableCell>
              </TableRow>
            ) : (
              pageItems.map((order) => (
                <TableRow key={order.id}>
                  <TableCell className="font-medium">
                    {order.orderNumber}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {customerNames.get(order.customerId) ?? order.customerId}
                  </TableCell>
                  <TableCell className="hidden text-muted-foreground md:table-cell">
                    {formatDate(order.createdAt)}
                  </TableCell>
                  <TableCell className="hidden tabular-nums sm:table-cell">
                    {order.items.length}
                  </TableCell>
                  <TableCell className="font-medium tabular-nums">
                    {formatPrice(order.total)}
                  </TableCell>
                  <TableCell>
                    <OrderStatusBadge status={order.status} />
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setActiveOrder(order)}
                      aria-label={`View order ${order.orderNumber}`}
                    >
                      <Eye className="size-3.5" aria-hidden="true" />
                      View
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

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

      {pageItems.length === 0 && (
        <div className="md:hidden">
          <TableEmpty
            icon={SearchX}
            title="No orders match your filters"
            description="Try a different status or search term."
          />
        </div>
      )}

      <ul className="flex flex-col gap-3 md:hidden">
        {pageItems.map((order) => (
          <li
            key={order.id}
            className="rounded-xl bg-card p-4 ring-1 ring-foreground/10"
          >
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="truncate font-medium">{order.orderNumber}</p>
                <p className="truncate text-sm text-muted-foreground">
                  {customerNames.get(order.customerId) ?? order.customerId}
                </p>
              </div>
              <OrderStatusBadge status={order.status} />
            </div>

            <dl className="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
              <div className="flex flex-col">
                <dt>Date</dt>
                <dd className="text-foreground">{formatDate(order.createdAt)}</dd>
              </div>
              <div className="flex flex-col">
                <dt>Items</dt>
                <dd className="text-foreground tabular-nums">
                  {order.items.length}
                </dd>
              </div>
              <div className="flex flex-col">
                <dt>Total</dt>
                <dd className="font-medium text-foreground tabular-nums">
                  {formatPrice(order.total)}
                </dd>
              </div>
            </dl>

            <Button
              variant="outline"
              size="sm"
              className="mt-3 w-full"
              onClick={() => setActiveOrder(order)}
              aria-label={`View order ${order.orderNumber}`}
            >
              <Eye className="size-3.5" aria-hidden="true" />
              View order
            </Button>
          </li>
        ))}
      </ul>

      <Dialog
        open={activeOrder !== null}
        onOpenChange={(open) => {
          if (!open) setActiveOrder(null);
        }}
      >
        <DialogContent className="sm:max-w-lg">
          {activeOrder && (
            <>
              <DialogHeader>
                <DialogTitle>Order {activeOrder.orderNumber}</DialogTitle>
                <DialogDescription>
                  Placed {formatDate(activeOrder.createdAt)} ·{" "}
                  {customerNames.get(activeOrder.customerId) ??
                    activeOrder.customerId}
                </DialogDescription>
              </DialogHeader>

              <div className="flex items-center justify-between">
                <OrderStatusBadge status={activeOrder.status} />
                <span className="text-sm text-muted-foreground">
                  {activeOrder.items.length}{" "}
                  {activeOrder.items.length === 1 ? "item" : "items"}
                </span>
              </div>

              <ul className="flex flex-col gap-2">
                {activeOrder.items.map((item) => (
                  <li
                    key={`${item.productId}:${item.variantId ?? "base"}`}
                    className="flex items-baseline justify-between gap-4 text-sm"
                  >
                    <span className="min-w-0">
                      <span className="font-medium">{item.name}</span>
                      <span className="text-muted-foreground">
                        {" "}
                        × {item.quantity}
                      </span>
                    </span>
                    <span className="shrink-0 tabular-nums">
                      {formatPrice(item.price * item.quantity)}
                    </span>
                  </li>
                ))}
              </ul>

              <dl className="flex flex-col gap-2 border-t pt-3 text-sm">
                <div className="flex justify-between">
                  <dt className="text-muted-foreground">Subtotal</dt>
                  <dd className="tabular-nums">
                    {formatPrice(activeOrder.subtotal)}
                  </dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-muted-foreground">Shipping</dt>
                  <dd className="tabular-nums">
                    {activeOrder.shipping === 0
                      ? "Free"
                      : formatPrice(activeOrder.shipping)}
                  </dd>
                </div>
                {activeOrder.tax > 0 && (
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Tax</dt>
                    <dd className="tabular-nums">
                      {formatPrice(activeOrder.tax)}
                    </dd>
                  </div>
                )}
                <div className="flex justify-between border-t pt-2 text-base">
                  <dt className="font-semibold">Total</dt>
                  <dd className="font-bold tabular-nums">
                    {formatPrice(activeOrder.total)}
                  </dd>
                </div>
              </dl>
            </>
          )}
        </DialogContent>
      </Dialog>
      <div className="mt-3 px-2 pb-2 md:hidden">
        <TablePagination
          page={safePage}
          pageCount={pageCount}
          total={filtered.length}
          onPageChange={setPage}
          itemLabel="orders"
        />
      </div>
    </div>
  );
}
