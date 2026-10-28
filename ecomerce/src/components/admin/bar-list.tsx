import { cn } from "@/lib/utils";

export interface BarListItem {
  label: string;
  /** Drives the relative bar length. */
  value: number;
  /** Preformatted value shown on the right, e.g. `"$1,204"`. */
  display: string;
}

interface BarListProps {
  items: BarListItem[];
  emptyLabel?: string;
  className?: string;
}

/**
 * Horizontal bar list for rankings (top products, category sales). The bars
 * are decorative — the label and formatted value carry the meaning.
 */
export function BarList({
  items,
  emptyLabel = "No data yet.",
  className,
}: BarListProps) {
  if (items.length === 0) {
    return <p className="text-sm text-muted-foreground">{emptyLabel}</p>;
  }

  const max = Math.max(...items.map((item) => item.value), 1);

  return (
    <ul className={cn("flex flex-col gap-4", className)}>
      {items.map((item) => (
        <li key={item.label} className="flex flex-col gap-1.5">
          <div className="flex items-baseline justify-between gap-4 text-sm">
            <span className="min-w-0 truncate font-medium">{item.label}</span>
            <span className="shrink-0 tabular-nums text-muted-foreground">
              {item.display}
            </span>
          </div>
          <div
            aria-hidden="true"
            className="h-2 overflow-hidden rounded-full bg-muted"
          >
            <div
              className="h-full rounded-full bg-primary"
              style={{ width: `${Math.max(4, (item.value / max) * 100)}%` }}
            />
          </div>
        </li>
      ))}
    </ul>
  );
}
