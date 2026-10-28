"use client";

import { motion, useReducedMotion } from "motion/react";
import type { ReactNode } from "react";

import { EASE_OUT } from "./fade-up";

interface ScaleInProps {
  children: ReactNode;
  className?: string;
  /** Extra delay in seconds before the animation starts. */
  delay?: number;
  /** Duration in seconds. Micro interactions: 0.15–0.25s (PRD §40). */
  duration?: number;
}

/**
 * Scales content in from 96% — used for small reveals like popovers,
 * success icons and modals. Reduced-motion users get an instant opacity
 * change (PRD §40).
 */
export function ScaleIn({
  children,
  className,
  delay = 0,
  duration = 0.25,
}: ScaleInProps) {
  const prefersReducedMotion = useReducedMotion();

  const animationProps = prefersReducedMotion
    ? {}
    : { initial: { opacity: 0, scale: 0.96 } };

  return (
    <motion.div
      className={className}
      {...animationProps}
      animate={{ opacity: 1, scale: 1 }}
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
