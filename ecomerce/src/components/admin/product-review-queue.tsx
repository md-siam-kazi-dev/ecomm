"use client";

import { useCallback, useEffect, useState } from "react";
import Image from "next/image";
import { Check, SearchX, X } from "lucide-react";
import { toast } from "sonner";

import { TablePagination } from "@/components/admin/table-pagination";
import { FilterTabs } from "@/components/dashboard/filter-tabs";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { ErrorState } from "@/components/errors/error-state";
import { ProductStatusBadge } from "@/components/vendor/product-status-badge";
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
import { apiGet, apiPost, type PaginationMeta } from "@/lib/api-client";
import { formatDate, formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";

interface ReviewProduct {
  id: string;
  name: string;
  vendor_name: string;
  status: string;
  approval_status: string;
  price: number;
  stock: number;
  image_url?: string;
  review_note?: string;
  submitted_at?: string | null;
  created_at: string;
}

interface ReviewDetail {
  id: string;
  name: string;
  description: string;
  short_description: string;
  status: string;
  approval_status: string;
  category: string;
  brand: string;
  sku: string;
  price: number;
  sale_price?: number | null;
  stock: number;
  low_stock_threshold: number;
  images: { url: string; is_primary: boolean }[];
  attributes: Record<string, string>;
  review_note?: string;
  vendor_name: string;
  submitted_at?: string | null;
  approved_at?: string | null;
  created_at: string;
}

const PAGE_SIZE = 20;
const FILTERS = [
  { value: "pending", label: "Pending" },
  { value: "changes_requested", label: "Changes requested" },
  { value: "rejected", label: "Rejected" },
  { value: "all", label: "All" },
];

type ListState =
  | { key: string; ok: true; products: ReviewProduct[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

type DecisionKind = "reject" | "changes";

/** Admin product moderation queue (PRD §24): approve / reject / request changes. */
export function ProductReviewQueue() {
  const [status, setStatus] = useState("pending");
  const [searchInput, setSearchInput] = useState("");
  const search = useDebouncedValue(searchInput, 350);
  const [page, setPage] = useState(1);
  const [state, setState] = useState<ListState | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const [detail, setDetail] = useState<ReviewDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [decision, setDecision] = useState<{ kind: DecisionKind; text: string } | null>(null);
  const [busy, setBusy] = useState(false);

  const requestUrl = `/api/v1/admin/products?status=${status}&page=${page}&limit=${PAGE_SIZE}${
    search ? `&search=${encodeURIComponent(search)}` : ""
  }`;
  const queryKey = `${requestUrl}#${reloadToken}`;

  const load = useCallback(() => apiGet<ReviewProduct[]>(requestUrl), [requestUrl]);

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

  async function openDetail(productId: string) {
    setDetailLoading(true);
    setDetail({ id: productId } as ReviewDetail);
    const res = await apiGet<ReviewDetail>(`/api/v1/admin/products/${productId}`);
    setDetailLoading(false);
    if (!res.ok) {
      toast.error(res.error.message);
      setDetail(null);
      return;
    }
    setDetail(res.data);
  }

  async function approve(productId: string) {
    setBusy(true);
    const res = await apiPost(`/api/v1/admin/products/${productId}/approve`);
    setBusy(false);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Product approved and published");
    setDetail(null);
    setReloadToken((t) => t + 1);
  }

  async function submitDecision() {
    if (!detail || !decision) return;
    const text = decision.text.trim();
    if (!text) {
      toast.error(decision.kind === "reject" ? "A reason is required" : "Feedback is required");
      return;
    }
    const path = decision.kind === "reject" ? "reject" : "changes";
    const body = decision.kind === "reject" ? { reason: text } : { note: text };

    setBusy(true);
    const res = await apiPost(`/api/v1/admin/products/${detail.id}/${path}`, body);
    setBusy(false);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success(decision.kind === "reject" ? "Product rejected" : "Changes requested");
    setDecision(null);
    setDetail(null);
    setReloadToken((t) => t + 1);
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
        <FilterTabs
          label="Filter by review status"
          value={status}
          onValueChange={(next) => {
            setStatus(next);
            setPage(1);
          }}
          tabs={FILTERS}
        />
        <div className="min-w-0 sm:ml-auto sm:max-w-xs sm:flex-1">
          <label htmlFor="review-search" className="sr-only">
            Search products
          </label>
          <Input
            id="review-search"
            type="search"
            placeholder="Search products"
            value={searchInput}
            onChange={(event) => {
              setSearchInput(event.target.value);
              setPage(1);
            }}
          />
        </div>
      </div>

      {isLoading ? (
        <TableSkeleton rows={5} columns={4} label="Loading products" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load the review queue"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((t) => t + 1)}
        />
      ) : products.length === 0 ? (
        <TableEmpty
          icon={SearchX}
          title="No products in this queue"
          description="Nothing is waiting on review right now."
        />
      ) : (
        <>
          <Card className="hidden [--card-spacing:0px] md:block">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Product</TableHead>
                  <TableHead className="hidden md:table-cell">Vendor</TableHead>
                  <TableHead>Price</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="hidden lg:table-cell">Submitted</TableHead>
                  <TableHead className="text-right">Action</TableHead>
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
                    <TableCell className="hidden text-muted-foreground md:table-cell">{product.vendor_name}</TableCell>
                    <TableCell className="tabular-nums">{formatPrice(product.price)}</TableCell>
                    <TableCell>
                      <ProductStatusBadge product={product} />
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground lg:table-cell">
                      {product.submitted_at ? formatDate(product.submitted_at) : "—"}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="outline" size="sm" onClick={() => openDetail(product.id)}>
                        Review
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
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

          <ul className="flex flex-col gap-3 md:hidden">
            {products.map((product) => (
              <li
                key={product.id}
                className="rounded-xl bg-card p-4 ring-1 ring-foreground/10"
              >
                <div className="flex items-start gap-3">
                  {product.image_url ? (
                    <Image src={product.image_url} alt="" width={44} height={44} className="size-11 shrink-0 rounded-md object-cover" />
                  ) : (
                    <span className="size-11 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                  )}
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">{product.name}</p>
                    <p className="truncate text-xs text-muted-foreground">{product.vendor_name}</p>
                    <div className="mt-2">
                      <ProductStatusBadge product={product} />
                    </div>
                  </div>
                </div>

                <dl className="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-muted-foreground">
                  <div className="flex flex-col">
                    <dt>Price</dt>
                    <dd className="font-medium text-foreground tabular-nums">
                      {formatPrice(product.price)}
                    </dd>
                  </div>
                  <div className="flex flex-col">
                    <dt>Submitted</dt>
                    <dd className="text-foreground">
                      {product.submitted_at ? formatDate(product.submitted_at) : "—"}
                    </dd>
                  </div>
                </dl>

                <Button
                  variant="outline"
                  size="sm"
                  className="mt-3 w-full"
                  onClick={() => openDetail(product.id)}
                >
                  Review product
                </Button>
              </li>
            ))}
          </ul>
        </>
      )}

      {/* Detail / moderation dialog */}
      <Dialog open={detail !== null} onOpenChange={(open) => !open && setDetail(null)}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          {detailLoading ? (
            <div className="flex flex-col gap-3" role="status" aria-label="Loading product">
              <Skeleton className="h-6 w-40" />
              <Skeleton className="h-40 w-full rounded-xl" />
            </div>
          ) : detail ? (
            <>
              <DialogHeader>
                <DialogTitle>{detail.name}</DialogTitle>
                <DialogDescription>
                  {detail.vendor_name} · {detail.category} · submitted{" "}
                  {detail.submitted_at ? formatDate(detail.submitted_at) : "—"}
                </DialogDescription>
              </DialogHeader>

              <div className="flex flex-wrap items-center gap-2">
                <ProductStatusBadge product={detail} />
                {detail.review_note && (
                  <span className="text-sm text-muted-foreground">Last note: {detail.review_note}</span>
                )}
              </div>

              {detail.images.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {detail.images.map((img, index) => (
                    <Image
                      key={img.url}
                      src={img.url}
                      alt={`${detail.name} image ${index + 1}`}
                      width={72}
                      height={72}
                      className="size-18 rounded-lg object-cover"
                    />
                  ))}
                </div>
              )}

              <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm sm:grid-cols-3">
                <Detail label="Price" value={formatPrice(detail.price)} />
                <Detail label="Sale price" value={detail.sale_price ? formatPrice(detail.sale_price) : "—"} />
                <Detail label="Stock" value={String(detail.stock)} />
                <Detail label="Brand" value={detail.brand || "—"} />
                <Detail label="SKU" value={detail.sku || "—"} />
                <Detail label="Low-stock at" value={String(detail.low_stock_threshold)} />
              </dl>

              <div>
                <p className="text-sm font-medium">Description</p>
                <p className="mt-1 whitespace-pre-line text-sm text-muted-foreground">{detail.description}</p>
              </div>

              {Object.keys(detail.attributes ?? {}).length > 0 && (
                <div>
                  <p className="text-sm font-medium">Attributes & shipping</p>
                  <dl className="mt-1 grid grid-cols-2 gap-x-4 gap-y-1 text-sm sm:grid-cols-3">
                    {Object.entries(detail.attributes).map(([key, value]) => (
                      <div key={key} className="flex flex-col">
                        <dt className="text-muted-foreground">{key.replace(/_/g, " ")}</dt>
                        <dd>{value}</dd>
                      </div>
                    ))}
                  </dl>
                </div>
              )}

              <DialogFooter className="flex-wrap gap-2 sm:justify-start">
                <Button onClick={() => approve(detail.id)} disabled={busy}>
                  <Check className="size-4" aria-hidden="true" />
                  Approve
                </Button>
                <Button variant="outline" onClick={() => setDecision({ kind: "changes", text: "" })} disabled={busy}>
                  Request changes
                </Button>
                <Button variant="destructive" onClick={() => setDecision({ kind: "reject", text: "" })} disabled={busy}>
                  <X className="size-4" aria-hidden="true" />
                  Reject
                </Button>
              </DialogFooter>
            </>
          ) : null}
        </DialogContent>
      </Dialog>

      {/* Decision dialog */}
      <Dialog open={decision !== null} onOpenChange={(open) => !open && setDecision(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{decision?.kind === "reject" ? "Reject product" : "Request changes"}</DialogTitle>
            <DialogDescription>
              {decision?.kind === "reject"
                ? "Tell the vendor why this product is rejected. The product stays unpublished."
                : "Describe the corrections needed. The vendor can edit and resubmit."}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="decision-text">{decision?.kind === "reject" ? "Reason" : "Feedback"}</Label>
            <textarea
              id="decision-text"
              rows={4}
              value={decision?.text ?? ""}
              onChange={(event) => setDecision((prev) => (prev ? { ...prev, text: event.target.value } : prev))}
              className={cn(
                "min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none",
                "focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              )}
            />
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDecision(null)} disabled={busy}>
              Cancel
            </Button>
            <Button onClick={submitDecision} disabled={busy} variant={decision?.kind === "reject" ? "destructive" : "default"}>
              {decision?.kind === "reject" ? "Reject product" : "Send feedback"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <div className="mt-3 px-2 pb-2 md:hidden">
        <TablePagination
          page={pagination?.page ?? page}
          pageCount={Math.max(1, pagination?.total_pages ?? 1)}
          total={pagination?.total ?? 0}
          onPageChange={setPage}
          itemLabel="products"
        />
      </div>
    </div>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  );
}
