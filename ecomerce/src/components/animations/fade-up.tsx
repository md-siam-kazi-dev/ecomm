"use client";

import { motion, useReducedMotion } from "motion/react";
import type { ReactNode } from "react";

/** Signature easing used across the animation system (smooth, decelerating). */
export const EASE_OUT: [number, number, number, number] = [
  0.21, 0.47, 0.32, 0.98,
];

/** Standard entrance distance in px — subtle, never exaggerated. */
const ENTRANCE_DISTANCE = 16;

interface FadeUpProps {
  children: ReactNode;
  className?: string;
  /** Extra delay in seconds before the animation starts. */
  delay?: number;
  /** Duration in seconds. Component entrances: 0.3–0.5s (PRD §40). */
  duration?: number;
  /** Animate when scrolled into view instead of on mount. Runs once. */
  inView?: boolean;
}

/**
 * Fades content in while moving it up by a small distance. Hero headings,
 * sections and cards all reuse this primitive. Reduced-motion users get an
 * instant opacity change with no movement (PRD §40).
 */
export function FadeUp({
  children,
  className,
  delay = 0,
  duration = 0.45,
  inView = false,
}: FadeUpProps) {
  const prefersReducedMotion = useReducedMotion();
  const y = prefersReducedMotion ? 0 : ENTRANCE_DISTANCE;

  const animationProps = prefersReducedMotion
    ? {}
    : { initial: { opacity: 0, y } };

  return (
    <motion.div
      className={className}
      {...animationProps}
      {...(inView
        ? { whileInView: { opacity: 1, y: 0 }, viewport: { once: true, margin: "-64px" } }
        : { animate: { opacity: 1, y: 0 } })}
      transition={{
        duration: prefersReducedMotion ? 0 : duration,
        delay: prefersReducedMotion ? 0 : delay,
        ease: EASE_OUT,
      }}
    >
      {children}
    </motion.div>
  );
}
