"use client";

import { motion, useReducedMotion } from "motion/react";
import type { ReactNode } from "react";

interface FadeInProps {
  children: ReactNode;
  className?: string;
  /** Extra delay in seconds before the animation starts. */
  delay?: number;
  /** Duration in seconds. Component entrances: 0.3–0.5s (PRD §40). */
  duration?: number;
  /**
   * Animate when scrolled into view instead of on mount. Runs once.
   * A negative margin keeps the element from triggering too early.
   */
  inView?: boolean;
}

/**
 * Fades content in without movement. Reduced-motion users get an instant,
 * opacity-only change (PRD §40).
 */
export function FadeIn({
  children,
  className,
  delay = 0,
  duration = 0.4,
  inView = false,
}: FadeInProps) {
  const prefersReducedMotion = useReducedMotion();

  const animationProps = prefersReducedMotion
    ? {}
    : { initial: { opacity: 0 } };

  return (
    <motion.div
      className={className}
      {...animationProps}
      {...(inView
        ? { whileInView: { opacity: 1 }, viewport: { once: true, margin: "-64px" } }
        : { animate: { opacity: 1 } })}
      transition={{
        duration: prefersReducedMotion ? 0 : duration,
        delay: prefersReducedMotion ? 0 : delay,
        ease: "easeOut",
      }}
    >
      {children}
    </motion.div>
  );
}
