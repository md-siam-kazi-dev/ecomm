import Link from "next/link";
import { ArrowRight, Sparkles } from "lucide-react";

import { BorderBeam } from "@/components/magic-ui/border-beam";
import { Button } from "@/components/ui/button";

/** Sale promotion banner with an animated border beam (Magic UI, PRD §41). */
export function PromoBanner() {
  return (
    <section className="mx-auto w-full max-w-7xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
      <div className="relative overflow-hidden rounded-3xl border bg-gradient-to-br from-primary-light/60 via-background to-primary-light/60 px-6 py-12 text-center sm:px-14 sm:py-16">
        <BorderBeam size={70} duration={7} />
        <div className="relative mx-auto max-w-xl">
          <span className="inline-flex items-center gap-1.5 rounded-full border bg-background/80 px-3 py-1 text-xs font-medium text-muted-foreground">
            <Sparkles className="size-3.5 text-primary" aria-hidden="true" />
            Limited time
          </span>
          <h2 className="mt-4 text-balance text-3xl font-bold tracking-tight sm:text-4xl">
            Summer Sale — up to{" "}
            <span className="text-primary">33% off</span>
          </h2>
          <p className="mx-auto mt-3 max-w-md text-sm text-muted-foreground sm:text-base">
            Refresh your essentials before the season ends. Discounted picks
            across every category, while stock lasts.
          </p>
          <Button asChild size="lg" className="mt-6 h-11 px-6 text-base">
            <Link href="/products?sort=discount">
              Shop the deals
              <ArrowRight className="size-4" aria-hidden="true" />
            </Link>
          </Button>
        </div>
      </div>
    </section>
  );
}
