"use client";

import { useCallback, useEffect, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { Pencil, SearchX } from "lucide-react";

import { TablePagination } from "@/components/admin/table-pagination";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { ErrorState } from "@/components/errors/error-state";
import { ProductStatusBadge } from "@/components/vendor/product-status-badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
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
import { formatDate, formatPrice } from "@/lib/formatters";

interface VendorProduct {
  id: string;
  name: string;
  slug: string;
  status: string;
  approval_status: string;
  price: number;
  stock: number;
  image_url?: string;
  review_note?: string;
  submitted_at?: string | null;
  reviewed_at?: string | null;
}

const PAGE_SIZE = 20;

/** A product the vendor must correct and resubmit. */
function needsResubmit(product: VendorProduct): boolean {
  return product.approval_status === "rejected" || product.approval_status === "changes_requested";
}

type State =
  | { key: string; ok: true; products: VendorProduct[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

/** Vendor product list (DB-backed) with approval status and review feedback. */
export function VendorProductsManager() {
  const [searchInput, setSearchInput] = useState("");
  const search = useDebouncedValue(searchInput, 350);
  const [page, setPage] = useState(1);
  const [state, setState] = useState<State | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const requestUrl = `/api/v1/vendor/products?page=${page}&limit=${PAGE_SIZE}${
    search ? `&search=${encodeURIComponent(search)}` : ""
  }`;
  const queryKey = `${requestUrl}#${reloadToken}`;

  const load = useCallback(() => apiGet<VendorProduct[]>(requestUrl), [requestUrl]);

  useEffect(() => {
    let active = true;
    void load().then((res) => {
      if (!active) return;
      if (res.ok && res.pagination) {
        setState({ key: queryKey, ok: true, products: res.data, pagination: res.pagination });
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
  const products = !isLoading && state?.ok === true ? state.products : [];
  const pagination = !isLoading && state?.ok === true ? state.pagination : null;

  return (
    <div className="flex flex-col gap-4">
      <div className="sm:max-w-xs">
        <label htmlFor="vendor-product-search" className="sr-only">
          Search your products
        </label>
        <Input
          id="vendor-product-search"
          type="search"
          placeholder="Search your products"
          value={searchInput}
          onChange={(event) => {
            setSearchInput(event.target.value);
            setPage(1);
          }}
        />
      </div>

      {isLoading ? (
        <TableSkeleton rows={5} columns={5} label="Loading products" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load your products"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((t) => t + 1)}
        />
      ) : products.length === 0 ? (
        <TableEmpty
          icon={SearchX}
          title={search ? "No products match your search" : "You have no products yet"}
          description={
            search
              ? "Try a different search term."
              : "Add your first product to start selling."
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
                  <TableHead>Price</TableHead>
                  <TableHead>Stock</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="hidden lg:table-cell">Reviewer feedback</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {products.map((product) => (
                  <TableRow key={product.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        {product.image_url ? (
                          <Image src={product.image_url} alt="" width={40} height={40} className="size-10 shrink-0 rounded-md object-cover" />
                        ) : (
                          <span className="size-10 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                        )}
                        <span className="max-w-[240px] truncate font-medium">{product.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="tabular-nums">{formatPrice(product.price)}</TableCell>
                    <TableCell className="tabular-nums">{product.stock}</TableCell>
                    <TableCell>
                      <ProductStatusBadge product={product} />
                    </TableCell>
                    <TableCell className="hidden max-w-[240px] text-sm text-muted-foreground lg:table-cell">
                      <span className="block truncate">{product.review_note || "—"}</span>
                      {product.reviewed_at && (
                        <span className="block text-xs">Reviewed {formatDate(product.reviewed_at)}</span>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button asChild variant={needsResubmit(product) ? "default" : "outline"} size="sm">
                        <Link href={`/vendor/products/${product.id}/edit`} aria-label={`Edit ${product.name}`}>
                          <Pencil className="size-3.5" aria-hidden="true" />
                          {needsResubmit(product) ? "Fix & resubmit" : "Edit"}
                        </Link>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Mobile: cards */}
          <ul className="flex flex-col divide-y md:hidden">
            {products.map((product) => (
              <li key={product.id} className="flex items-start gap-3 p-4">
                {product.image_url ? (
                  <Image
                    src={product.image_url}
                    alt=""
                    width={44}
                    height={44}
                    className="size-11 shrink-0 rounded-md object-cover"
                  />
                ) : (
                  <span className="size-11 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                )}
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{product.name}</p>

                  <div className="mt-2 flex flex-wrap items-center gap-2">
                    <ProductStatusBadge product={product} />
                  </div>

                  <dl className="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                    <div className="flex flex-col">
                      <dt>Price</dt>
                      <dd className="text-foreground tabular-nums">{formatPrice(product.price)}</dd>
                    </div>
                    <div className="flex flex-col">
                      <dt>Stock</dt>
                      <dd className="text-foreground tabular-nums">{product.stock}</dd>
                    </div>
                    {product.review_note && (
                      <div className="col-span-2 flex justify-between gap-3">
                        <dt className="shrink-0">Feedback</dt>
                        <dd className="min-w-0 truncate text-foreground">{product.review_note}</dd>
                      </div>
                    )}
                    {product.reviewed_at && (
                      <div className="col-span-2 flex justify-between gap-3">
                        <dt className="shrink-0">Reviewed</dt>
                        <dd className="text-foreground">{formatDate(product.reviewed_at)}</dd>
                      </div>
                    )}
                  </dl>

                  <Button
                    asChild
                    variant={needsResubmit(product) ? "default" : "outline"}
                    size="sm"
                    className="mt-3 w-full"
                  >
                    <Link href={`/vendor/products/${product.id}/edit`} aria-label={`Edit ${product.name}`}>
                      <Pencil className="size-3.5" aria-hidden="true" />
                      {needsResubmit(product) ? "Fix & resubmit" : "Edit"}
                    </Link>
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
              itemLabel="products"
            />
          </div>
        </Card>
      )}
    </div>
  );
}
