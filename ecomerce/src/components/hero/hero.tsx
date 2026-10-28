"use client";

import Image from "next/image";
import Link from "next/link";
import { motion, useReducedMotion } from "motion/react";
import { ArrowRight, Sparkles, Star, Truck } from "lucide-react";

import { EASE_OUT } from "@/components/animations/fade-up";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FREE_SHIPPING_THRESHOLD } from "@/lib/constants";
import { formatPrice } from "@/lib/formatters";
import type { Product } from "@/types/product";

interface HeroProps {
  /** Top featured product — drives the hero visual (real catalog data). */
  product?: Product;
}

/**
 * Homepage hero (PRD §10): staggered first-view animation — badge → heading
 * → copy → CTAs → visual → floating cards — all 0.4–0.8s and subtle.
 * Reduced-motion users get instant opacity-only reveals.
 */
export function Hero({ product }: HeroProps) {
  const prefersReducedMotion = useReducedMotion();

  const fadeUp = (delay: number) => ({
    initial: prefersReducedMotion ? { opacity: 0 } : { opacity: 0, y: 24 },
    animate: { opacity: 1, y: 0 },
    transition: {
      duration: prefersReducedMotion ? 0 : 0.55,
      delay: prefersReducedMotion ? 0 : delay,
      ease: EASE_OUT,
    },
  });

  const revealImage = {
    initial: prefersReducedMotion ? { opacity: 0 } : { opacity: 0, scale: 0.96 },
    animate: { opacity: 1, scale: 1 },
    transition: {
      duration: prefersReducedMotion ? 0 : 0.7,
      delay: prefersReducedMotion ? 0 : 0.3,
      ease: EASE_OUT,
    },
  };

  const floatIn = (delay: number) => ({
    initial: prefersReducedMotion ? { opacity: 0 } : { opacity: 0, y: 16 },
    animate: { opacity: 1, y: 0 },
    transition: {
      duration: prefersReducedMotion ? 0 : 0.5,
      delay: prefersReducedMotion ? 0 : delay,
      ease: EASE_OUT,
    },
  });

  return (
    <section className="relative overflow-hidden border-b bg-gradient-to-b from-primary-light/50 to-background">
      <div
        aria-hidden="true"
        className="absolute -top-24 right-0 size-96 rounded-full bg-primary/10 blur-3xl"
      />

      <div className="mx-auto grid w-full max-w-7xl items-center gap-10 px-4 py-12 sm:px-6 sm:py-16 lg:grid-cols-2 lg:gap-16 lg:px-8 lg:py-24">
        <div className="flex flex-col items-start gap-5">
          <motion.div {...fadeUp(0.05)}>
            <Badge variant="secondary" className="gap-1.5 rounded-full px-3 py-1">
              <Sparkles className="size-3.5 text-primary" aria-hidden="true" />
              Summer sale — up to 33% off
            </Badge>
          </motion.div>

          <motion.h1
            {...fadeUp(0.15)}
            className="text-balance text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl"
          >
            Everyday essentials,{" "}
            <span className="text-primary">delivered beautifully.</span>
          </motion.h1>

          <motion.p
            {...fadeUp(0.25)}
            className="max-w-lg text-base text-muted-foreground sm:text-lg"
          >
            From kitchen upgrades to travel gear — thousands of hand-picked
            products with fast shipping, easy returns and prices that make
            sense.
          </motion.p>

          <motion.div {...fadeUp(0.35)} className="flex flex-wrap gap-3">
            <Button asChild size="lg" className="h-11 px-6 text-base">
              <Link href="/products">
                Shop now
                <ArrowRight className="size-4" aria-hidden="true" />
              </Link>
            </Button>
            <Button asChild variant="outline" size="lg" className="h-11 px-6 text-base">
              <Link href="/products?sort=discount">Browse deals</Link>
            </Button>
          </motion.div>

          <motion.dl
            {...fadeUp(0.45)}
            className="mt-2 flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-muted-foreground"
          >
            <div className="flex items-center gap-1.5">
              <Star className="size-4 fill-amber-400 text-amber-400" aria-hidden="true" />
              <dd>
                <strong className="font-semibold text-foreground">4.8</strong>{" "}
                average rating
              </dd>
            </div>
            <div className="flex items-center gap-1.5">
              <Truck className="size-4 text-primary" aria-hidden="true" />
              <dd>
                Free shipping over{" "}
                <strong className="font-semibold text-foreground">
                  {formatPrice(FREE_SHIPPING_THRESHOLD)}
                </strong>
              </dd>
            </div>
          </motion.dl>
        </div>

        {product ? (
        <motion.div
          {...revealImage}
          className="relative mx-auto w-full max-w-md lg:max-w-none"
        >
          <div className="relative aspect-square overflow-hidden rounded-3xl border bg-muted shadow-xl">
            <Image
              src={product.thumbnail.url}
              alt={product.thumbnail.alt}
              fill
              priority
              sizes="(max-width: 1024px) 100vw, 50vw"
              className="object-cover"
            />
          </div>

          <motion.div
            {...floatIn(0.55)}
            className="absolute -left-4 top-8 hidden sm:block"
          >
            <Link
              href={`/products/${product.slug}`}
              className="flex items-center gap-3 rounded-xl border bg-card p-3 shadow-lg transition-shadow hover:shadow-md focus-visible:outline-2 focus-visible:outline-offset-2"
            >
              <span className="relative size-12 shrink-0 overflow-hidden rounded-lg">
                <Image
                  src={product.thumbnail.url}
                  alt=""
                  fill
                  sizes="48px"
                  className="object-cover"
                />
              </span>
              <span className="flex flex-col">
                <span className="max-w-36 truncate text-xs font-medium">
                  {product.name}
                </span>
                <span className="text-xs text-muted-foreground">
                  {formatPrice(product.price)}
                </span>
              </span>
            </Link>
          </motion.div>

          <motion.div
            {...floatIn(0.7)}
            className="absolute -bottom-4 right-6 hidden items-center gap-2 rounded-full border bg-card px-4 py-2.5 shadow-lg sm:flex"
          >
            <Truck className="size-4 text-primary" aria-hidden="true" />
            <span className="text-xs font-medium">Free 30-day returns</span>
          </motion.div>
        </motion.div>
        ) : (
          <div
            aria-hidden="true"
            className="relative mx-auto hidden aspect-square w-full max-w-md items-center justify-center rounded-3xl border bg-muted/40 lg:flex"
          >
            <Sparkles className="size-10 text-primary/40" />
          </div>
        )}
      </div>
    </section>
  );
}
