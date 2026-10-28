"use client";

import { motion, useReducedMotion, type Variants } from "motion/react";
import type { ReactNode } from "react";

import { EASE_OUT } from "./fade-up";

interface StaggerContainerProps {
  children: ReactNode;
  className?: string;
  /** Delay in seconds before the first child animates. */
  delay?: number;
  /** Gap in seconds between children. */
  gap?: number;
  /** Animate when scrolled into view instead of on mount. Runs once. */
  inView?: boolean;
}

interface StaggerItemProps {
  children: ReactNode;
  className?: string;
}

/**
 * Wraps a list of `StaggerItem` children and animates them one after another.
 * With reduced motion, items appear together instantly (PRD §40).
 */
export function StaggerContainer({
  children,
  className,
  delay = 0,
  gap = 0.08,
  inView = false,
}: StaggerContainerProps) {
  const prefersReducedMotion = useReducedMotion();

  const containerVariants: Variants = {
    hidden: {},
    visible: {
      transition: {
        staggerChildren: prefersReducedMotion ? 0 : gap,
        delayChildren: prefersReducedMotion ? 0 : delay,
      },
    },
  };

  const animationProps = prefersReducedMotion ? {} : { initial: "hidden" as const };

  return (
    <motion.div
      className={className}
      variants={containerVariants}
      {...animationProps}
      {...(inView
        ? { whileInView: "visible" as const, viewport: { once: true, margin: "-64px" } }
        : { animate: "visible" as const })}
    >
      {children}
    </motion.div>
  );
}

/**
 * A single child of `StaggerContainer`. Must be a direct child for variants
 * to propagate.
 */
export function StaggerItem({ children, className }: StaggerItemProps) {
  const prefersReducedMotion = useReducedMotion();

  const itemVariants: Variants = {
    hidden: { opacity: 0, y: prefersReducedMotion ? 0 : 16 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: prefersReducedMotion ? 0 : 0.45, ease: EASE_OUT },
    },
  };

  return (
    <motion.div className={className} variants={itemVariants}>
      {children}
    </motion.div>
  );
}
