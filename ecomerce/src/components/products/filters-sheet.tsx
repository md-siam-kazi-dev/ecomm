"use client";

import { useState } from "react";
import { SlidersHorizontal } from "lucide-react";

import { ProductFilters } from "@/components/products/product-filters";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import type { CategoryWithCount } from "@/lib/products";
import type { CatalogFilters } from "@/lib/query-params";

interface FiltersSheetProps {
  categories: CategoryWithCount[];
  current: CatalogFilters;
  showCategory?: boolean;
}

/** Mobile filter button + slide-in sheet (PRD §12). Closes on navigation. */
export function FiltersSheet({
  categories,
  current,
  showCategory,
}: FiltersSheetProps) {
  const [open, setOpen] = useState(false);
  const close = () => setOpen(false);

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="outline" className="lg:hidden">
          <SlidersHorizontal className="size-4" aria-hidden="true" />
          Filters
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-80 max-w-[85vw] overflow-y-auto p-0">
        <SheetHeader className="border-b p-4">
          <SheetTitle>Filters</SheetTitle>
          <SheetDescription className="sr-only">
            Narrow down products by category, price, rating and availability.
          </SheetDescription>
        </SheetHeader>
        <div className="p-4">
          <ProductFilters
            categories={categories}
            current={current}
            showCategory={showCategory}
            onNavigate={close}
          />
        </div>
      </SheetContent>
    </Sheet>
  );
}
