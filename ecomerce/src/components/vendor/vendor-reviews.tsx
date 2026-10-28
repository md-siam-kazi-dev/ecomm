"use client";

import { useCallback, useEffect, useState } from "react";
import Image from "next/image";
import { MessageSquareReply, Star, StarHalf } from "lucide-react";
import { toast } from "sonner";

import { BarList, type BarListItem } from "@/components/admin/bar-list";
import { StatCard } from "@/components/admin/stat-card";
import { TablePagination } from "@/components/admin/table-pagination";
import { FilterTabs } from "@/components/dashboard/filter-tabs";
import { TableSkeleton } from "@/components/dashboard/skeletons";
import { TableEmpty } from "@/components/dashboard/table-empty";
import { ErrorState } from "@/components/errors/error-state";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { apiGet, apiPost, type PaginationMeta } from "@/lib/api-client";
import { formatDate } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { VendorReview, VendorReviewSummary } from "@/lib/vendor";

const PAGE_SIZE = 10;

const FILTERS = [
  { value: "all", label: "All" },
  { value: "unreplied", label: "Awaiting reply" },
  { value: "replied", label: "Replied" },
];

type ListState =
  | { key: string; ok: true; items: VendorReview[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

/** Vendor reviews (DB-backed): star summary, distribution and replies. */
export function VendorReviews() {
  const [status, setStatus] = useState("all");
  const [page, setPage] = useState(1);
  const [state, setState] = useState<ListState | null>(null);
  const [summary, setSummary] = useState<VendorReviewSummary | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const requestUrl =
    `/api/v1/vendor/reviews?page=${page}&limit=${PAGE_SIZE}` +
    (status !== "all" ? `&status=${status}` : "");
  const queryKey = `${requestUrl}#${reloadToken}`;
  const load = useCallback(() => apiGet<VendorReview[]>(requestUrl), [requestUrl]);

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

  useEffect(() => {
    let active = true;
    void apiGet<VendorReviewSummary>("/api/v1/vendor/reviews/summary").then((res) => {
      if (active && res.ok) setSummary(res.data);
    });
    return () => {
      active = false;
    };
  }, [reloadToken]);

  const isLoading = state?.key !== queryKey;
  const failed = !isLoading && state?.ok === false;
  const items = !isLoading && state?.ok === true ? state.items : [];
  const pagination = !isLoading && state?.ok === true ? state.pagination : null;

  const replying = useReplyDialog(setReloadToken);

  const distribution: BarListItem[] = summary
    ? [5, 4, 3, 2, 1].map((star) => {
        const count = summary.distribution[String(star)] ?? 0;
        return { label: `${star} star`, value: count, display: String(count) };
      })
    : [];

  return (
    <div className="flex flex-col gap-6">
      {summary && summary.total > 0 && (
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-3">
            <StatCard
              label="Average rating"
              value={summary.average.toFixed(1)}
              icon={Star}
              hint={`From ${summary.total} ${summary.total === 1 ? "review" : "reviews"}`}
            />
            <StatCard
              label="Total reviews"
              value={String(summary.total)}
              icon={StarHalf}
            />
            <StatCard
              label="Awaiting reply"
              value={String(summary.unreplied)}
              icon={MessageSquareReply}
              hint={summary.unreplied > 0 ? "Reply to build trust" : "All caught up"}
            />
          </div>

          <div className="rounded-xl bg-card p-5 ring-1 ring-foreground/10">
            <p className="text-sm font-medium">Rating distribution</p>
            <div className="mt-3">
              <BarList items={distribution} />
            </div>
          </div>
        </div>
      )}

      <FilterTabs
        label="Filter reviews"
        value={status}
        onValueChange={(next) => {
          setStatus(next);
          setPage(1);
        }}
        tabs={FILTERS}
      />

      {isLoading ? (
        <TableSkeleton rows={4} columns={3} withToolbar={false} label="Loading reviews" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load reviews"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((token) => token + 1)}
        />
      ) : items.length === 0 ? (
        <TableEmpty
          icon={Star}
          title="No reviews yet"
          description={
            status === "all"
              ? "No reviews on your products yet."
              : "No reviews match this filter."
          }
        />
      ) : (
        <div className="flex flex-col gap-4">
          <ul className="flex flex-col gap-4">
            {items.map((review) => (
              <li
                key={review.id}
                className="rounded-xl bg-card p-5 ring-1 ring-foreground/10"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="flex min-w-0 items-start gap-3">
                    {review.product_image ? (
                      <Image
                        src={review.product_image}
                        alt=""
                        width={44}
                        height={44}
                        className="size-11 shrink-0 rounded-md object-cover"
                      />
                    ) : (
                      <span className="size-11 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                    )}
                    <div className="min-w-0">
                      <p className="text-sm font-medium">{review.author}</p>
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        {review.product_name}
                        {review.order_number ? ` · ${review.order_number}` : ""} ·{" "}
                        {formatDate(review.created_at)}
                      </p>
                    </div>
                  </div>
                  <span
                    className="flex items-center gap-0.5"
                    aria-label={`${review.rating} out of 5 stars`}
                  >
                    {Array.from({ length: 5 }).map((_, index) => (
                      <Star
                        key={index}
                        className={cn(
                          "size-3.5",
                          index < review.rating
                            ? "fill-primary text-primary"
                            : "text-muted-foreground"
                        )}
                        aria-hidden="true"
                      />
                    ))}
                  </span>
                </div>

                {review.title && <p className="mt-3 font-medium">{review.title}</p>}
                <p className="mt-1 text-sm text-muted-foreground">{review.comment}</p>

                {review.images && review.images.length > 0 && (
                  <ul className="mt-3 flex flex-wrap gap-2">
                    {review.images.map((url) => (
                      <li
                        key={url}
                        className="relative size-16 overflow-hidden rounded-lg border"
                      >
                        <Image
                          src={url}
                          alt="Review photo"
                          fill
                          sizes="64px"
                          className="object-cover"
                        />
                      </li>
                    ))}
                  </ul>
                )}

                {review.replied ? (
                  <div className="mt-3 rounded-lg border-l-2 border-primary bg-muted/50 p-3">
                    <p className="text-xs font-medium text-muted-foreground">
                      Your reply
                      {review.vendor_replied_at
                        ? ` · ${formatDate(review.vendor_replied_at)}`
                        : ""}
                    </p>
                    <p className="mt-1 text-sm">{review.vendor_reply}</p>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="mt-2 -ml-2"
                      onClick={() => replying.open(review)}
                    >
                      Edit reply
                    </Button>
                  </div>
                ) : (
                  <div className="mt-3">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => replying.open(review)}
                    >
                      <MessageSquareReply className="size-3.5" aria-hidden="true" />
                      Reply
                    </Button>
                  </div>
                )}
              </li>
            ))}
          </ul>

          <div className="rounded-xl bg-card px-2 pb-2 ring-1 ring-foreground/10">
            <TablePagination
              page={pagination?.page ?? page}
              pageCount={Math.max(1, pagination?.total_pages ?? 1)}
              total={pagination?.total ?? 0}
              onPageChange={setPage}
              itemLabel="reviews"
            />
          </div>
        </div>
      )}

      <Dialog
        open={replying.review !== null}
        onOpenChange={(open) => {
          if (!open) replying.close();
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {replying.review?.replied ? "Edit reply" : "Reply to review"}
            </DialogTitle>
            <DialogDescription>
              {replying.review?.author} · {replying.review?.product_name}
            </DialogDescription>
          </DialogHeader>

          <label htmlFor="vendor-reply" className="sr-only">
            Your reply
          </label>
          <textarea
            id="vendor-reply"
            rows={4}
            value={replying.text}
            disabled={replying.saving}
            onChange={(event) => replying.setText(event.target.value)}
            placeholder="Thanks for the feedback…"
            className="w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-60"
          />

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={replying.close}
              disabled={replying.saving}
            >
              Cancel
            </Button>
            <Button
              type="button"
              onClick={replying.submit}
              disabled={replying.saving}
            >
              {replying.saving ? "Saving…" : "Save reply"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

/** Reply dialog state + mutation, kept out of the main component body. */
function useReplyDialog(
  setReloadToken: (updater: (token: number) => number) => void
) {
  const [review, setReview] = useState<VendorReview | null>(null);
  const [text, setText] = useState("");
  const [saving, setSaving] = useState(false);

  function open(target: VendorReview) {
    setReview(target);
    setText(target.vendor_reply ?? "");
  }

  function close() {
    setReview(null);
    setText("");
  }

  async function submit() {
    if (!review) return;
    const trimmed = text.trim();
    if (!trimmed) {
      toast.error("Write a reply first");
      return;
    }
    setSaving(true);
    const res = await apiPost<VendorReview>(
      `/api/v1/vendor/reviews/${review.id}/reply`,
      { reply: trimmed }
    );
    setSaving(false);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Reply saved", { description: review.product_name });
    close();
    setReloadToken((token) => token + 1);
  }

  return { review, text, saving, open, close, submit, setText };
}

