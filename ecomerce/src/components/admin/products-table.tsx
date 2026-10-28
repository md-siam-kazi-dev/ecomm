"use client";

import { useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { Copy, ExternalLink, MoreHorizontal, SearchX } from "lucide-react";
import { toast } from "sonner";

import { StockBadge, stockLevel } from "@/components/admin/stock-badge";
import { TablePagination } from "@/components/admin/table-pagination";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { calculateDiscountPercentage, formatPrice } from "@/lib/formatters";
import type { Product } from "@/types/product";

const PAGE_SIZE = 8;

const STOCK_OPTIONS = [
  { value: "all", label: "All stock" },
  { value: "in", label: "In stock" },
  { value: "low", label: "Low stock" },
  { value: "out", label: "Out of stock" },
];

/** Admin product catalogue (PRD §24): search, filter, paginate and act. */
export function ProductsTable({ products }: { products: Product[] }) {
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("all");
  const [stock, setStock] = useState("all");
  const [page, setPage] = useState(1);

  const categories = useMemo(() => {
    const map = new Map<string, string>();
    for (const product of products) map.set(product.categorySlug, product.category);
    return [...map.entries()]
      .map(([slug, name]) => ({ slug, name }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [products]);

  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase();

    return products.filter((product) => {
      if (category !== "all" && product.categorySlug !== category) return false;
      if (stock !== "all" && stockLevel(product.stock) !== stock) return false;
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
  }, [products, query, category, stock]);

  const pageCount = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const safePage = Math.min(page, pageCount);
  const pageItems = filtered.slice(
    (safePage - 1) * PAGE_SIZE,
    safePage * PAGE_SIZE
  );

  async function copySku(sku: string) {
    try {
      await navigator.clipboard.writeText(sku);
      toast.success("SKU copied", { description: sku });
    } catch {
      toast.error("Could not copy SKU");
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
        <div className="min-w-0 sm:max-w-xs sm:flex-1">
          <label htmlFor="product-search" className="sr-only">
            Search products
          </label>
          <Input
            id="product-search"
            type="search"
            placeholder="Search by name, SKU or brand"
            value={query}
            onChange={(event) => {
              setQuery(event.target.value);
              setPage(1);
            }}
          />
        </div>

        <Select
          value={category}
          onValueChange={(value) => {
            setCategory(value);
            setPage(1);
          }}
        >
          <SelectTrigger aria-label="Filter by category" className="w-full sm:w-48">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All categories</SelectItem>
            {categories.map((option) => (
              <SelectItem key={option.slug} value={option.slug}>
                {option.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={stock}
          onValueChange={(value) => {
            setStock(value);
            setPage(1);
          }}
        >
          <SelectTrigger aria-label="Filter by stock" className="w-full sm:w-40">
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
              <TableHead>Price</TableHead>
              <TableHead>Stock</TableHead>
              <TableHead className="hidden sm:table-cell">Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pageItems.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="py-6">
                  <TableEmpty
                    icon={SearchX}
                    title="No products match your filters"
                    description="Try a different search term or clear the filters."
                  />
                </TableCell>
              </TableRow>
            ) : (
              pageItems.map((product) => {
                const discount = calculateDiscountPercentage(
                  product.price,
                  product.compareAtPrice
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
                          <span className="max-w-[220px] truncate font-medium">
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
                      <span className="flex items-baseline gap-1.5">
                        <span className="font-medium tabular-nums">
                          {formatPrice(product.price)}
                        </span>
                        {discount && (
                          <span className="text-xs text-muted-foreground">
                            -{discount}%
                          </span>
                        )}
                      </span>
                    </TableCell>
                    <TableCell className="tabular-nums">
                      {product.stock}
                    </TableCell>
                    <TableCell className="hidden sm:table-cell">
                      <div className="flex flex-wrap gap-1">
                        <StockBadge stock={product.stock} />
                        {product.isFeatured && (
                          <Badge variant="secondary">Featured</Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            aria-label={`Actions for ${product.name}`}
                          >
                            <MoreHorizontal className="size-4" aria-hidden="true" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/products/${product.slug}`}
                              target="_blank"
                              rel="noreferrer"
                            >
                              <ExternalLink className="size-4" aria-hidden="true" />
                              View in store
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onSelect={() => void copySku(product.sku)}
                          >
                            <Copy className="size-4" aria-hidden="true" />
                            Copy SKU
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
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
            title="No products match your filters"
            description="Try a different search term or clear the filters."
          />
        </div>
      )}

      <ul className="flex flex-col gap-3 md:hidden">
        {pageItems.map((product) => {
          const discount = calculateDiscountPercentage(
            product.price,
            product.compareAtPrice
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
                  <p className="font-mono text-xs text-muted-foreground">
                    {product.sku}
                  </p>
                  <div className="mt-2 flex flex-wrap items-center gap-1">
                    <StockBadge stock={product.stock} />
                    {product.isFeatured && (
                      <Badge variant="secondary">Featured</Badge>
                    )}
                  </div>
                </div>
              </div>

              <dl className="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                <div className="flex flex-col">
                  <dt>Category</dt>
                  <dd className="text-foreground">{product.category}</dd>
                </div>
                <div className="flex flex-col">
                  <dt>Price</dt>
                  <dd className="flex items-baseline gap-1.5 font-medium text-foreground tabular-nums">
                    {formatPrice(product.price)}
                    {discount && (
                      <span className="text-xs font-normal text-muted-foreground">
                        -{discount}%
                      </span>
                    )}
                  </dd>
                </div>
                <div className="flex flex-col">
                  <dt>Stock</dt>
                  <dd className="text-foreground tabular-nums">
                    {product.stock}
                  </dd>
                </div>
              </dl>

              <div className="mt-3 flex gap-2">
                <Button variant="outline" size="sm" className="flex-1" asChild>
                  <Link
                    href={`/products/${product.slug}`}
                    target="_blank"
                    rel="noreferrer"
                  >
                    <ExternalLink className="size-3.5" aria-hidden="true" />
                    View in store
                  </Link>
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="flex-1"
                  onClick={() => void copySku(product.sku)}
                >
                  <Copy className="size-3.5" aria-hidden="true" />
                  Copy SKU
                </Button>
              </div>
            </li>
          );
        })}
      </ul>
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
