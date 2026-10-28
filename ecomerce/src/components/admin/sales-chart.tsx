"use client";

import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";

import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";
import { formatPrice } from "@/lib/formatters";
import { cn } from "@/lib/utils";
import type { RevenuePoint } from "@/types/admin";

const chartConfig = {
  revenue: {
    label: "Revenue",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig;

interface SalesChartProps {
  points: RevenuePoint[];
  className?: string;
}

/**
 * Revenue trend (PRD §24). Interactive area chart with an accessible layer and
 * a rendered `<table>` of the same data for screen readers.
 */
export function SalesChart({ points, className }: SalesChartProps) {
  const total = points.reduce((sum, point) => sum + point.revenue, 0);

  return (
    <figure className={cn("flex flex-col gap-4", className)}>
      <div className="flex items-baseline justify-between gap-4">
        <p className="text-sm text-muted-foreground">
          Revenue · last {points.length} months
        </p>
        <p className="text-lg font-semibold tabular-nums">{formatPrice(total)}</p>
      </div>

      <ChartContainer
        config={chartConfig}
        className="h-48 w-full aspect-auto sm:h-56"
      >
        <AreaChart
          accessibilityLayer
          data={points}
          margin={{ left: 4, right: 4, top: 4, bottom: 0 }}
        >
          <CartesianGrid vertical={false} />
          <XAxis
            dataKey="label"
            tickLine={false}
            axisLine={false}
            tickMargin={8}
          />
          <ChartTooltip
            cursor={false}
            content={
              <ChartTooltipContent
                indicator="line"
                formatter={(value) => formatPrice(Number(value))}
              />
            }
          />
          <Area
            dataKey="revenue"
            type="monotone"
            stroke="var(--color-revenue)"
            strokeWidth={2.5}
            fill="var(--color-revenue)"
            fillOpacity={0.25}
          />
        </AreaChart>
      </ChartContainer>

      <figcaption className="sr-only">
        <table>
          <caption>Monthly revenue and order counts</caption>
          <thead>
            <tr>
              <th scope="col">Month</th>
              <th scope="col">Revenue</th>
              <th scope="col">Orders</th>
            </tr>
          </thead>
          <tbody>
            {points.map((point) => (
              <tr key={point.month}>
                <th scope="row">{point.month}</th>
                <td>{formatPrice(point.revenue)}</td>
                <td>{point.orders}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </figcaption>
    </figure>
  );
}
