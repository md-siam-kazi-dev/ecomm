"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
} from "@/components/ui/pagination";

interface TablePaginationProps {
  page: number;
  pageCount: number;
  total: number;
  onPageChange: (page: number) => void;
  /** Noun for the count summary, e.g. `"products"`. */
  itemLabel: string;
}

const MAX_WINDOW = 7;

/** First/last/current-neighbouring page numbers, with ellipses between. */
function pageList(page: number, pageCount: number): (number | "ellipsis")[] {
  if (pageCount <= MAX_WINDOW) {
    return Array.from({ length: pageCount }, (_, index) => index + 1);
  }

  const wanted = [1, pageCount, page - 1, page, page + 1]
    .filter((value) => value >= 1 && value <= pageCount)
    .sort((a, b) => a - b);

  const out: (number | "ellipsis")[] = [];
  let previous = 0;

  for (const value of wanted) {
    if (previous && value - previous > 1) out.push("ellipsis");
    out.push(value);
    previous = value;
  }

  return out;
}

/** Shared pagination footer for the dashboard tables (PRD §24). */
export function TablePagination({
  page,
  pageCount,
  total,
  onPageChange,
  itemLabel,
}: TablePaginationProps) {
  if (pageCount <= 1) {
    return (
      <p className="border-t px-2 pt-3 text-xs text-muted-foreground">
        {total} {itemLabel}
      </p>
    );
  }

  return (
    <div className="flex flex-col-reverse items-center gap-3 border-t px-2 pt-3 sm:flex-row sm:justify-between">
      <p className="text-xs text-muted-foreground">
        Page {page} of {pageCount} · {total} {itemLabel}
      </p>

      <Pagination className="mx-0 w-auto justify-end">
        <PaginationContent>
          <PaginationItem>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label="Go to previous page"
              onClick={() => onPageChange(page - 1)}
              disabled={page <= 1}
            >
              <ChevronLeft aria-hidden="true" />
            </Button>
          </PaginationItem>

          {pageList(page, pageCount).map((entry, index) =>
            entry === "ellipsis" ? (
              <PaginationItem key={`ellipsis-${index}`}>
                <PaginationEllipsis />
              </PaginationItem>
            ) : (
              <PaginationItem key={entry}>
                <Button
                  type="button"
                  variant={entry === page ? "outline" : "ghost"}
                  size="icon-sm"
                  className="tabular-nums"
                  aria-label={`Go to page ${entry}`}
                  aria-current={entry === page ? "page" : undefined}
                  onClick={() => onPageChange(entry)}
                >
                  {entry}
                </Button>
              </PaginationItem>
            )
          )}

          <PaginationItem>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label="Go to next page"
              onClick={() => onPageChange(page + 1)}
              disabled={page >= pageCount}
            >
              <ChevronRight aria-hidden="true" />
            </Button>
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  );
}
