"use client";

import { useId } from "react";
import { motion, useReducedMotion } from "motion/react";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

export interface FilterTab {
  value: string;
  label: string;
}

interface FilterTabsProps {
  value: string;
  onValueChange: (value: string) => void;
  tabs: FilterTab[];
  /** Accessible name for the strip, e.g. "Filter orders by status". */
  label: string;
  className?: string;
}

const PILL = "absolute inset-0 rounded-md bg-background shadow-sm";

/**
 * Segmented filter tabs whose indicator glides between the active tabs.
 *
 * Built on the Radix-backed `Tabs` primitive — so arrow keys, roving focus and
 * the `tablist`/`tab` roles come for free — with `motion`'s shared-layout
 * animation driving the indicator. Users who prefer reduced motion get a static
 * indicator instead.
 */
export function FilterTabs({
  value,
  onValueChange,
  tabs,
  label,
  className,
}: FilterTabsProps) {
  const layoutId = `filter-tabs-${useId()}`;
  const reduceMotion = useReducedMotion();

  return (
    <Tabs
      value={value}
      onValueChange={onValueChange}
      className={cn("gap-0", className)}
    >
      <TabsList
        aria-label={label}
        className="h-auto! w-full flex-wrap justify-start! sm:w-fit"
      >
        {tabs.map((tab) => {
          const isActive = tab.value === value;

          return (
            <TabsTrigger
              key={tab.value}
              value={tab.value}
              className="relative flex-none px-2.5 py-1 whitespace-nowrap shadow-none! data-active:bg-transparent! dark:data-active:bg-transparent!"
            >
              {isActive &&
                (reduceMotion ? (
                  <span aria-hidden="true" className={PILL} />
                ) : (
                  <motion.span
                    aria-hidden="true"
                    layoutId={layoutId}
                    className={PILL}
                    transition={{ type: "spring", stiffness: 420, damping: 34 }}
                  />
                ))}
              <span className="relative">{tab.label}</span>
            </TabsTrigger>
          );
        })}
      </TabsList>
    </Tabs>
  );
}
