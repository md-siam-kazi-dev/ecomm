"use client";

import { useState } from "react";
import Link from "next/link";
import { HelpCircle, Loader2, MessageCircleQuestion, Store } from "lucide-react";

import { AskQuestionDialog } from "@/components/products/ask-question-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { apiGet } from "@/lib/api-client";
import {
  toQuestion,
  type ApiQuestion,
  type ProductQuestion,
} from "@/lib/catalog";
import { formatDate } from "@/lib/formatters";
import { useAuthStore } from "@/store/auth-store";

interface ProductQuestionsProps {
  productSlug: string;
  initialQuestions: ProductQuestion[];
  total: number;
}

const PAGE_SIZE = 10;

/** Product Q&A: public questions with vendor answers, plus ask-a-question. */
export function ProductQuestions({
  productSlug,
  initialQuestions,
  total,
}: ProductQuestionsProps) {
  const status = useAuthStore((state) => state.status);
  const [items, setItems] = useState<ProductQuestion[]>(initialQuestions);
  const [count, setCount] = useState(total);
  const [page, setPage] = useState(1);
  const [hasNext, setHasNext] = useState(initialQuestions.length < total);
  const [loading, setLoading] = useState(false);
  const [askOpen, setAskOpen] = useState(false);

  async function reloadFirstPage() {
    const res = await apiGet<ApiQuestion[]>(
      `/api/v1/products/${encodeURIComponent(productSlug)}/questions?page=1&limit=${PAGE_SIZE}`
    );
    if (!res.ok) return;
    setItems(res.data.map(toQuestion));
    setPage(1);
    setHasNext(res.pagination?.has_next ?? false);
    setCount(res.pagination?.total ?? res.data.length);
  }

  async function loadMore() {
    setLoading(true);
    const next = page + 1;
    const res = await apiGet<ApiQuestion[]>(
      `/api/v1/products/${encodeURIComponent(productSlug)}/questions?page=${next}&limit=${PAGE_SIZE}`
    );
    setLoading(false);
    if (!res.ok) return;
    setItems((prev) => [...prev, ...res.data.map(toQuestion)]);
    setPage(next);
    setHasNext(res.pagination?.has_next ?? false);
  }

  return (
    <section
      id="questions"
      aria-labelledby="questions-heading"
      className="mt-12 scroll-mt-24"
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2
            id="questions-heading"
            className="text-2xl font-semibold tracking-tight"
          >
            Questions &amp; Answers
          </h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {count} {count === 1 ? "question" : "questions"} about this product
          </p>
        </div>

        {status === "authenticated" ? (
          <Button type="button" onClick={() => setAskOpen(true)}>
            <MessageCircleQuestion className="size-4" aria-hidden="true" />
            Ask a question
          </Button>
        ) : (
          <Button asChild variant="outline">
            <Link href="/auth/sign-in">Sign in to ask</Link>
          </Button>
        )}
      </div>

      {items.length === 0 ? (
        <div className="mt-6 flex flex-col items-center gap-2 rounded-xl border border-dashed p-8 text-center">
          <HelpCircle className="size-7 text-muted-foreground" aria-hidden="true" />
          <p className="text-sm text-muted-foreground">
            No questions yet. Be the first to ask.
          </p>
        </div>
      ) : (
        <>
          <ul className="mt-6 flex flex-col gap-4">
            {items.map((question) => (
              <li key={question.id} className="rounded-xl border bg-card p-5">
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <p className="text-sm font-semibold">{question.question}</p>
                  {question.answers.length === 0 && (
                    <Badge variant="secondary">Awaiting answer</Badge>
                  )}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  Asked by {question.author} · {formatDate(question.createdAt)}
                </p>

                {question.answers.map((answer) => (
                  <div
                    key={answer.id}
                    className="mt-3 rounded-lg border-l-2 border-primary bg-muted/50 p-3"
                  >
                    <p className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground">
                      {answer.isOfficialVendorAnswer ? (
                        <>
                          <Store className="size-3.5" aria-hidden="true" />
                          Seller · {answer.author}
                        </>
                      ) : (
                        <>Answer · {answer.author}</>
                      )}
                      {" · "}
                      {formatDate(answer.createdAt)}
                    </p>
                    <p className="mt-1 text-sm">{answer.answer}</p>
                  </div>
                ))}
              </li>
            ))}
          </ul>

          {hasNext && (
            <div className="mt-6 flex justify-center">
              <Button
                type="button"
                variant="outline"
                onClick={loadMore}
                disabled={loading}
              >
                {loading && (
                  <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                )}
                Load more questions
              </Button>
            </div>
          )}
        </>
      )}

      <AskQuestionDialog
        open={askOpen}
        onOpenChange={setAskOpen}
        productSlug={productSlug}
        onSubmitted={reloadFirstPage}
      />
    </section>
  );
}
