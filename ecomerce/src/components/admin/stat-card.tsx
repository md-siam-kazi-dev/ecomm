import type { LucideIcon } from "lucide-react";
import { TrendingDown, TrendingUp } from "lucide-react";

import CountUp from "@/components/reactbits/count-up";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { CURRENCY } from "@/lib/constants";
import { cn } from "@/lib/utils";

const currencySymbol =
  new Intl.NumberFormat("en-US", { style: "currency", currency: CURRENCY })
    .formatToParts(0)
    .find((part) => part.type === "currency")?.value ?? "";

/** Optional count-up config. `decimals` follows formatPrice for currency. */
interface StatCount {
  to: number;
  format?: "currency" | "number";
}

interface StatCardProps {
  label: string;
  /** Final formatted value, used as the accessible text. */
  value: string;
  icon: LucideIcon;
  /** Supporting line, e.g. "3 low on stock". */
  hint?: string;
  /** Month-over-month change; omit to hide the trend indicator. */
  trend?: number;
  /** Animate the number from zero on first view. */
  count?: StatCount;
  className?: string;
}

/** Headline KPI tile for the dashboards (PRD §24). */
export function StatCard({
  label,
  value,
  icon: Icon,
  hint,
  trend,
  count,
  className,
}: StatCardProps) {
  const hasTrend = trend !== undefined && trend !== 0;
  const isUp = (trend ?? 0) > 0;

  return (
    <Card className={cn("gap-0 [--card-spacing:--spacing(5)]", className)}>
      <CardContent className="flex flex-col gap-3">
        <div className="flex items-start justify-between gap-3">
          <span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary-dark">
            <Icon className="size-4" aria-hidden="true" />
          </span>
          {hasTrend && (
            <Badge variant={isUp ? "success" : "destructive"}>
              {isUp ? (
                <TrendingUp aria-hidden="true" />
              ) : (
                <TrendingDown aria-hidden="true" />
              )}
              {Math.abs(trend)}%
            </Badge>
          )}
        </div>

        <div className="flex flex-col gap-0.5">
          <p className="text-sm text-muted-foreground">{label}</p>
          <p className="text-2xl font-semibold tracking-tight">
            {count ? (
              <>
                <span aria-hidden="true">
                  <CountUp
                    to={count.to}
                    duration={1.4}
                    easing="easeOut"
                    separator=","
                    prefix={count.format === "currency" ? currencySymbol : ""}
                    decimals={count.format === "currency" ? 2 : 0}
                  />
                </span>
                <span className="sr-only">{value}</span>
              </>
            ) : (
              value
            )}
          </p>
          {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
          {hasTrend && (
            <p className="text-xs text-muted-foreground">vs. previous month</p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
