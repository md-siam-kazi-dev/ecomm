"use client";

import { useMemo, useState } from "react";
import Image from "next/image";
import { PackagePlus, SearchX } from "lucide-react";
import { toast } from "sonner";

import { StockBadge, stockLevel } from "@/components/admin/stock-badge";
import { TablePagination } from "@/components/admin/table-pagination";
import { TableEmpty } from "@/components/dashboard/table-empty";
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Product } from "@/types/product";

const PAGE_SIZE = 8;
/** Stock at which the level bar reads as full. */
const FULL_STOCK = 40;

const STOCK_OPTIONS = [
  { value: "all", label: "All stock" },
  { value: "in", label: "In stock" },
  { value: "low", label: "Low stock" },
  { value: "out", label: "Out of stock" },
];

/** Admin inventory (PRD §24): stock levels plus a demo restock dialog. */
export function InventoryTable({ products }: { products: Product[] }) {
  const [query, setQuery] = useState("");
  const [stock, setStock] = useState("all");
  const [page, setPage] = useState(1);
  const [stockById, setStockById] = useState<Record<string, number>>({});
  const [restocking, setRestocking] = useState<Product | null>(null);
  const [restockValue, setRestockValue] = useState("");

  const stockOf = (product: Product) => stockById[product.id] ?? product.stock;

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();

    return products.filter((product) => {
      const currentStock = stockById[product.id] ?? product.stock;
      if (stock !== "all" && stockLevel(currentStock) !== stock) return false;
      if (
        term &&
        !`${product.name} ${product.sku} ${product.brand}`
          .toLowerCase()
          .includes(term)
      ) {
        return false;
      }
      return true;
    });
  }, [products, query, stock, stockById]);

  const pageCount = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const safePage = Math.min(page, pageCount);
  const pageItems = filtered.slice(
    (safePage - 1) * PAGE_SIZE,
    safePage * PAGE_SIZE
  );

  function openRestock(product: Product) {
    setRestocking(product);
    setRestockValue(String(stockOf(product)));
  }

  function confirmRestock() {
    if (!restocking) return;

    const next = Number.parseInt(restockValue, 10);
    if (!Number.isFinite(next) || next < 0) {
      toast.error("Enter a valid stock count");
      return;
    }

    setStockById((previous) => ({ ...previous, [restocking.id]: next }));
    toast.success("Stock updated", {
      description: `${restocking.name} set to ${next} units (demo only).`,
    });
    setRestocking(null);
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
        <div className="min-w-0 sm:max-w-xs sm:flex-1">
          <label htmlFor="inventory-search" className="sr-only">
            Search inventory
          </label>
          <Input
            id="inventory-search"
            type="search"
            placeholder="Search by name or SKU"
            value={query}
            onChange={(event) => {
              setQuery(event.target.value);
              setPage(1);
            }}
          />
        </div>

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
      </div>

      <Card className="hidden [--card-spacing:0px] md:block">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Product</TableHead>
              <TableHead className="hidden md:table-cell">Category</TableHead>
              <TableHead>Stock level</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pageItems.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="py-6">
                  <TableEmpty
                    icon={SearchX}
                    title="No inventory items match your filters"
                    description="Try a different stock level or search term."
                  />
                </TableCell>
              </TableRow>
            ) : (
              pageItems.map((product) => {
                const currentStock = stockOf(product);
                const percentage = Math.min(
                  100,
                  Math.round((currentStock / FULL_STOCK) * 100)
                );

                return (
                  <TableRow key={product.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Image
                          src={product.thumbnail.url}
                          alt=""
                          width={40}
                          height={40}
                          className="size-10 shrink-0 rounded-md object-cover"
                        />
                        <span className="flex min-w-0 flex-col">
                          <span className="max-w-[200px] truncate font-medium">
                            {product.name}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {product.sku}
                          </span>
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground md:table-cell">
                      {product.category}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <span
                          aria-hidden="true"
                          className="h-2 w-24 overflow-hidden rounded-full bg-muted"
                        >
                          <span
                            className="block h-full rounded-full bg-primary"
                            style={{ width: `${Math.max(2, percentage)}%` }}
                          />
                        </span>
                        <span className="text-sm tabular-nums">
                          {currentStock}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <StockBadge stock={currentStock} />
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => openRestock(product)}
                        aria-label={`Adjust stock for ${product.name}`}
                      >
                        <PackagePlus className="size-3.5" aria-hidden="true" />
                        Restock
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>

        <div className="px-2 pb-2">
          <TablePagination
            page={safePage}
            pageCount={pageCount}
            total={filtered.length}
            onPageChange={setPage}
            itemLabel="products"
          />
        </div>
      </Card>

      {pageItems.length === 0 && (
        <div className="md:hidden">
          <TableEmpty
            icon={SearchX}
            title="No inventory items match your filters"
            description="Try a different stock level or search term."
          />
        </div>
      )}

      <ul className="flex flex-col gap-3 md:hidden">
        {pageItems.map((product) => {
          const currentStock = stockOf(product);
          const percentage = Math.min(
            100,
            Math.round((currentStock / FULL_STOCK) * 100)
          );

          return (
            <li
              key={product.id}
              className="rounded-xl bg-card p-4 ring-1 ring-foreground/10"
            >
              <div className="flex items-start gap-3">
                <Image
                  src={product.thumbnail.url}
                  alt=""
                  width={44}
                  height={44}
                  className="size-11 shrink-0 rounded-md object-cover"
                />
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{product.name}</p>
                  <p className="text-xs text-muted-foreground">{product.sku}</p>
                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    <StockBadge stock={currentStock} />
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {currentStock} in stock
                    </span>
                  </div>
                </div>
              </div>

              <dl className="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                <div className="flex flex-col">
                  <dt>Category</dt>
                  <dd className="text-foreground">{product.category}</dd>
                </div>
                <div className="flex flex-col">
                  <dt>Stock level</dt>
                  <dd className="flex items-center gap-2">
                    <span
                      aria-hidden="true"
                      className="h-2 w-16 overflow-hidden rounded-full bg-muted"
                    >
                      <span
                        className="block h-full rounded-full bg-primary"
                        style={{ width: `${Math.max(2, percentage)}%` }}
                      />
                    </span>
                    <span className="text-foreground tabular-nums">
                      {currentStock}
                    </span>
                  </dd>
                </div>
              </dl>

              <Button
                variant="outline"
                size="sm"
                className="mt-3 w-full"
                onClick={() => openRestock(product)}
                aria-label={`Adjust stock for ${product.name}`}
              >
                <PackagePlus className="size-3.5" aria-hidden="true" />
                Restock
              </Button>
            </li>
          );
        })}
      </ul>

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
              {restocking?.name}. Demo only — changes are not persisted.
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
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setRestocking(null)}
            >
              Cancel
            </Button>
            <Button type="button" onClick={confirmRestock}>
              Save stock
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <div className="mt-3 px-2 pb-2 md:hidden">
        <TablePagination
          page={safePage}
          pageCount={pageCount}
          total={filtered.length}
          onPageChange={setPage}
          itemLabel="products"
        />
      </div>
    </div>
  );
}
