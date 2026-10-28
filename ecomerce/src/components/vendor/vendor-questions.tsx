"use client";

import { useCallback, useEffect, useState } from "react";
import Image from "next/image";
import {
  MessageCircleQuestion,
  MessageSquareText,
  SearchX,
} from "lucide-react";
import { toast } from "sonner";

import { TablePagination } from "@/components/admin/table-pagination";
import { FilterTabs } from "@/components/dashboard/filter-tabs";
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
  DialogFooter,
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
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { apiGet, apiPost, type PaginationMeta } from "@/lib/api-client";
import { formatDate } from "@/lib/formatters";
import type { VendorQuestion } from "@/lib/vendor";

const PAGE_SIZE = 10;

const FILTERS = [
  { value: "all", label: "All" },
  { value: "unanswered", label: "Unanswered" },
  { value: "answered", label: "Answered" },
];

type State =
  | { key: string; ok: true; items: VendorQuestion[]; pagination: PaginationMeta }
  | { key: string; ok: false; message: string };

/** Vendor questions: filter, search, paginate and answer product questions. */
export function VendorQuestions() {
  const [searchInput, setSearchInput] = useState("");
  const search = useDebouncedValue(searchInput, 350);
  const [status, setStatus] = useState("all");
  const [page, setPage] = useState(1);
  const [state, setState] = useState<State | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  const requestUrl =
    `/api/v1/vendor/questions?page=${page}&limit=${PAGE_SIZE}` +
    (status !== "all" ? `&status=${status}` : "") +
    (search ? `&search=${encodeURIComponent(search)}` : "");
  const queryKey = `${requestUrl}#${reloadToken}`;
  const load = useCallback(() => apiGet<VendorQuestion[]>(requestUrl), [requestUrl]);

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
  const hasFilters = search !== "" || status !== "all";

  // Answer dialog.
  const [answering, setAnswering] = useState<VendorQuestion | null>(null);
  const [answerText, setAnswerText] = useState("");
  const [saving, setSaving] = useState(false);

  function openAnswer(question: VendorQuestion) {
    setAnswering(question);
    setAnswerText(question.answer ?? "");
  }

  async function submitAnswer() {
    if (!answering) return;
    const trimmed = answerText.trim();
    if (!trimmed) {
      toast.error("Write an answer first");
      return;
    }
    setSaving(true);
    const res = await apiPost<VendorQuestion>(
      `/api/v1/vendor/questions/${answering.id}/answer`,
      { answer: trimmed }
    );
    setSaving(false);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Answer saved", { description: answering.product_name });
    setAnswering(null);
    setReloadToken((token) => token + 1);
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-3">
        <FilterTabs
          label="Filter questions"
          value={status}
          onValueChange={(next) => {
            setStatus(next);
            setPage(1);
          }}
          tabs={FILTERS}
        />

        <div className="sm:max-w-xs">
          <label htmlFor="vendor-question-search" className="sr-only">
            Search questions
          </label>
          <Input
            id="vendor-question-search"
            type="search"
            placeholder="Search by product or question"
            value={searchInput}
            onChange={(event) => {
              setSearchInput(event.target.value);
              setPage(1);
            }}
          />
        </div>
      </div>

      {isLoading ? (
        <TableSkeleton rows={5} columns={5} label="Loading questions" />
      ) : failed ? (
        <ErrorState
          title="Couldn't load questions"
          description={(state as { message: string }).message}
          onRetry={() => setReloadToken((token) => token + 1)}
        />
      ) : items.length === 0 ? (
        <TableEmpty
          icon={hasFilters ? SearchX : MessageCircleQuestion}
          title="No questions"
          description={
            hasFilters
              ? "No questions match your filters."
              : "No questions on your products yet."
          }
          action={
            hasFilters ? (
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSearchInput("");
                  setStatus("all");
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
                  <TableHead>Question</TableHead>
                  <TableHead className="hidden lg:table-cell">Customer</TableHead>
                  <TableHead className="hidden lg:table-cell">Date</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((question) => (
                  <TableRow key={question.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        {question.product_image ? (
                          <Image
                            src={question.product_image}
                            alt=""
                            width={40}
                            height={40}
                            className="size-10 shrink-0 rounded-md object-cover"
                          />
                        ) : (
                          <span className="size-10 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                        )}
                        <span className="max-w-[180px] truncate font-medium">
                          {question.product_name}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <p className="max-w-[280px] truncate">{question.question}</p>
                      {question.answered && question.answer && (
                        <p className="mt-0.5 max-w-[280px] truncate text-xs text-muted-foreground">
                          {question.answer}
                        </p>
                      )}
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground lg:table-cell">
                      {question.author}
                    </TableCell>
                    <TableCell className="hidden text-muted-foreground lg:table-cell">
                      {formatDate(question.created_at)}
                    </TableCell>
                    <TableCell>
                      {question.answered ? (
                        <Badge className="bg-success text-success-foreground">Answered</Badge>
                      ) : (
                        <Badge variant="secondary">Unanswered</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => openAnswer(question)}
                        aria-label={`${question.answered ? "Edit" : "Answer"} question about ${question.product_name}`}
                      >
                        <MessageSquareText className="size-3.5" aria-hidden="true" />
                        {question.answered ? "Edit" : "Answer"}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Mobile: cards */}
          <ul className="flex flex-col divide-y md:hidden">
            {items.map((question) => (
              <li key={question.id} className="flex items-start gap-3 p-4">
                {question.product_image ? (
                  <Image
                    src={question.product_image}
                    alt=""
                    width={44}
                    height={44}
                    className="size-11 shrink-0 rounded-md object-cover"
                  />
                ) : (
                  <span className="size-11 shrink-0 rounded-md bg-muted" aria-hidden="true" />
                )}
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{question.product_name}</p>
                  <p className="mt-1 text-sm">{question.question}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {question.author} · {formatDate(question.created_at)}
                  </p>
                  <div className="mt-2 flex items-center gap-2">
                    {question.answered ? (
                      <Badge className="bg-success text-success-foreground">Answered</Badge>
                    ) : (
                      <Badge variant="secondary">Unanswered</Badge>
                    )}
                  </div>
                  {question.answered && question.answer && (
                    <p className="mt-2 rounded-lg border-l-2 border-primary bg-muted/50 p-2 text-sm">
                      {question.answer}
                    </p>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    className="mt-3 w-full"
                    onClick={() => openAnswer(question)}
                  >
                    <MessageSquareText className="size-3.5" aria-hidden="true" />
                    {question.answered ? "Edit answer" : "Answer question"}
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
              itemLabel="questions"
            />
          </div>
        </Card>
      )}

      <Dialog
        open={answering !== null}
        onOpenChange={(open) => {
          if (!open) setAnswering(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {answering?.answered ? "Edit answer" : "Answer question"}
            </DialogTitle>
            <DialogDescription>
              {answering?.product_name} · asked by {answering?.author}
            </DialogDescription>
          </DialogHeader>

          <div className="rounded-lg border bg-muted/50 p-3">
            <p className="text-xs font-medium text-muted-foreground">Question</p>
            <p className="mt-1 text-sm">{answering?.question}</p>
          </div>

          <label htmlFor="vendor-answer" className="sr-only">
            Your answer
          </label>
          <textarea
            id="vendor-answer"
            rows={4}
            value={answerText}
            disabled={saving}
            onChange={(event) => setAnswerText(event.target.value)}
            placeholder="Yes — it's dishwasher safe on the top rack."
            className="w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-60"
          />

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setAnswering(null)}
              disabled={saving}
            >
              Cancel
            </Button>
            <Button type="button" onClick={submitAnswer} disabled={saving}>
              {saving ? "Saving…" : "Save answer"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

