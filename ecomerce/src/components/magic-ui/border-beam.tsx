"use client";

import { motion, useReducedMotion } from "motion/react";

import { cn } from "@/lib/utils";

interface BorderBeamProps {
  className?: string;
  /** Beam size in px along the path. */
  size?: number;
  /** Seconds for one full loop around the border. */
  duration?: number;
  delay?: number;
  colorFrom?: string;
  colorTo?: string;
  reverse?: boolean;
}

/**
 * A glowing beam that travels around the border of its parent — adapted from
 * Magic UI's BorderBeam, themed with design tokens and disabled under
 * prefers-reduced-motion. The parent must be relative, overflow-hidden and
 * rounded for the mask to follow the corners.
 */
export function BorderBeam({
  className,
  size = 60,
  duration = 6,
  delay = 0,
  colorFrom = "var(--primary)",
  colorTo = "var(--chart-3)",
  reverse = false,
}: BorderBeamProps) {
  const prefersReducedMotion = useReducedMotion();

  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 rounded-[inherit] border border-transparent [mask-clip:padding-box,border-box] [mask-composite:intersect] [mask-image:linear-gradient(transparent,transparent),linear-gradient(#000,#000)]"
    >
      {!prefersReducedMotion && (
        <motion.div
          className={cn("absolute aspect-square", className)}
          style={{
            width: size,
            offsetPath: `rect(0 auto auto 0 round ${size}px)`,
            background: `linear-gradient(to left, transparent, ${colorFrom}, ${colorTo}, transparent)`,
          }}
          initial={{ offsetDistance: reverse ? "100%" : "0%" }}
          animate={{
            offsetDistance: reverse ? ["100%", "0%"] : ["0%", "100%"],
          }}
          transition={{ repeat: Infinity, ease: "linear", duration, delay: -delay }}
        />
      )}
    </div>
  );
}
