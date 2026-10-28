"use client";

import {
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
} from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Clock, Search, X } from "lucide-react";

import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useHasMounted } from "@/hooks/use-has-mounted";
import { useSearchSuggestions } from "@/hooks/use-search-suggestions";
import type { CategoryWithCount } from "@/lib/products";
import { formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import { useSearchStore } from "@/store/search-store";

interface SearchInputProps {
  initialQuery: string;
}

type SuggestionItem =
  | { type: "product"; product: import("@/types/product").Product }
  | { type: "category"; category: CategoryWithCount };

const SUGGESTION_LIMIT = 8;

/**
 * Search combobox (PRD §20): debounced URL commits for the results area,
 * an instant suggestion dropdown (keyboard navigable, ARIA combobox
 * pattern), and persisted recent searches.
 */
export function SearchInput({ initialQuery }: SearchInputProps) {
  const router = useRouter();
  const hasMounted = useHasMounted();
  const recentSearches = useSearchStore((state) => state.recent);
  const addRecentSearch = useSearchStore((state) => state.addRecentSearch);
  const clearRecentSearches = useSearchStore((state) => state.clearRecentSearches);

  const wrapperRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const [query, setQuery] = useState(initialQuery);
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);

  const debouncedQuery = useDebouncedValue(query.trim(), 300);
  const suggestions = useSearchSuggestions(query);

  const trimmed = query.trim();
  const dropdownOpen = isOpen && trimmed.length > 0;

  const flatItems: SuggestionItem[] = suggestions
    ? [
        ...suggestions.products.map((product) => ({
          type: "product" as const,
          product,
        })),
        ...suggestions.categories.map((category) => ({
          type: "category" as const,
          category,
        })),
      ].slice(0, SUGGESTION_LIMIT)
    : [];

  // Debounced commit — keeps ?q= in the URL so results stay shareable
  // (PRD §45) without polluting history on every keystroke.
  useEffect(() => {
    if (debouncedQuery === initialQuery.trim()) return;
    router.replace(
      debouncedQuery
        ? `/search?q=${encodeURIComponent(debouncedQuery)}`
        : "/search",
      { scroll: false }
    );
  }, [debouncedQuery, initialQuery, router]);

  function commitSearch(value: string) {
    const committed = value.trim();
    setIsOpen(false);
    setActiveIndex(-1);
    if (committed) addRecentSearch(committed);
    router.push(
      committed ? `/search?q=${encodeURIComponent(committed)}` : "/search"
    );
  }

  function selectItem(item: SuggestionItem) {
    setIsOpen(false);
    setActiveIndex(-1);
    if (item.type === "product") {
      router.push(`/products/${item.product.slug}`);
    } else {
      addRecentSearch(item.category.name);
      router.push(`/category/${item.category.slug}`);
    }
  }

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setIsOpen(true);
      setActiveIndex((index) => Math.min(index + 1, flatItems.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActiveIndex((index) => Math.max(index - 1, -1));
    } else if (event.key === "Escape") {
      setIsOpen(false);
      setActiveIndex(-1);
    } else if (event.key === "Enter") {
      if (isOpen && activeIndex >= 0 && flatItems[activeIndex]) {
        event.preventDefault();
        selectItem(flatItems[activeIndex]);
      }
    }
  }

  function handleBlur(event: React.FocusEvent<HTMLDivElement>) {
    // Ignore blurs that stay inside the combobox (e.g. clicking a suggestion).
    if (wrapperRef.current?.contains(event.relatedTarget as Node | null)) {
      return;
    }
    setIsOpen(false);
    setActiveIndex(-1);
  }

  return (
    <div ref={wrapperRef} className="relative" onBlur={handleBlur}>
      <form
        role="search"
        onSubmit={(event: FormEvent<HTMLFormElement>) => {
          event.preventDefault();
          commitSearch(query);
        }}
      >
        <label htmlFor="search-input" className="sr-only">
          Search products and categories
        </label>
        <div className="relative">
          <Search
            className="pointer-events-none absolute left-4 top-1/2 size-5 -translate-y-1/2 text-muted-foreground"
            aria-hidden="true"
          />
          <input
            ref={inputRef}
            id="search-input"
            type="text"
            role="combobox"
            aria-expanded={dropdownOpen}
            aria-controls="search-suggestions"
            aria-autocomplete="list"
            aria-activedescendant={
              activeIndex >= 0 ? `search-option-${activeIndex}` : undefined
            }
            autoComplete="off"
            spellCheck={false}
            value={query}
            placeholder="Search products…"
            onChange={(event) => {
              setQuery(event.target.value);
              setIsOpen(true);
              setActiveIndex(-1);
            }}
            onFocus={() => setIsOpen(true)}
            onKeyDown={handleKeyDown}
            className="h-12 w-full rounded-xl border bg-background pl-11 pr-20 text-base outline-none transition-colors placeholder:text-muted-text focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
          />
          {query && (
            <button
              type="button"
              onClick={() => {
                setQuery("");
                inputRef.current?.focus();
              }}
              aria-label="Clear search"
              className="absolute right-14 top-1/2 -translate-y-1/2 rounded-md p-1 text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2"
            >
              <X className="size-4" aria-hidden="true" />
            </button>
          )}
          <button
            type="submit"
            aria-label="Search"
            className="absolute right-2.5 top-1/2 flex size-8 -translate-y-1/2 items-center justify-center rounded-lg bg-primary text-primary-foreground transition-colors hover:bg-primary-dark focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            <Search className="size-4" aria-hidden="true" />
          </button>
        </div>
      </form>

      {dropdownOpen && (
        <div
          id="search-suggestions"
          role="listbox"
          aria-label="Search suggestions"
          className="absolute inset-x-0 z-50 mt-2 overflow-hidden rounded-xl border bg-popover shadow-lg"
        >
          {!suggestions ? (
            <p className="p-4 text-sm text-muted-foreground" aria-live="polite">
              Searching…
            </p>
          ) : flatItems.length === 0 ? (
            <p className="p-4 text-sm text-muted-foreground">
              No matches for “{trimmed}” — press Enter to search anyway.
            </p>
          ) : (
            <ul>
              {flatItems.map((item, index) => (
                <li key={item.type === "product" ? item.product.id : item.category.id}>
                  <button
                    type="button"
                    role="option"
                    aria-selected={index === activeIndex}
                    id={`search-option-${index}`}
                    onMouseDown={(event) => event.preventDefault()}
                    onClick={() => selectItem(item)}
                    onMouseEnter={() => setActiveIndex(index)}
                    className={cn(
                      "flex w-full items-center gap-3 px-4 py-2.5 text-left text-sm transition-colors",
                      index === activeIndex && "bg-muted"
                    )}
                  >
                    {item.type === "product" ? (
                      <>
                        <span
                          aria-hidden="true"
                          className="relative size-10 shrink-0 overflow-hidden rounded-md border"
                        >
                          <Image
                            src={item.product.thumbnail.url}
                            alt=""
                            fill
                            sizes="40px"
                            className="object-cover"
                          />
                        </span>
                        <span className="flex min-w-0 flex-1 flex-col">
                          <span className="truncate font-medium">
                            {item.product.name}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {item.product.category} ·{" "}
                            {formatPrice(item.product.price)}
                          </span>
                        </span>
                      </>
                    ) : (
                      <>
                        <span
                          aria-hidden="true"
                          className="flex size-10 shrink-0 items-center justify-center rounded-md bg-muted"
                        >
                          <Search className="size-4 text-muted-foreground" />
                        </span>
                        <span className="flex min-w-0 flex-1 flex-col">
                          <span className="truncate font-medium">
                            {item.category.name}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            Category · {item.category.productCount}{" "}
                            {item.category.productCount === 1
                              ? "product"
                              : "products"}
                          </span>
                        </span>
                      </>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {hasMounted && !query && recentSearches.length > 0 && (
        <section aria-label="Recent searches" className="mt-6">
          <div className="flex items-center justify-between">
            <h2 className="flex items-center gap-1.5 text-sm font-semibold text-muted-foreground">
              <Clock className="size-3.5" aria-hidden="true" />
              Recent searches
            </h2>
            <button
              type="button"
              onClick={clearRecentSearches}
              className="rounded-md text-xs font-medium text-primary transition-colors hover:text-primary-dark focus-visible:outline-2 focus-visible:outline-offset-2"
            >
              Clear
            </button>
          </div>
          <div className="mt-2.5 flex flex-wrap gap-2">
            {recentSearches.map((recent) => (
              <button
                key={recent}
                type="button"
                onClick={() => commitSearch(recent)}
                className="inline-flex items-center gap-1.5 rounded-full border bg-secondary px-3 py-1.5 text-xs font-medium transition-colors hover:bg-muted focus-visible:outline-2 focus-visible:outline-offset-2"
              >
                <Clock className="size-3" aria-hidden="true" />
                {recent}
              </button>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
