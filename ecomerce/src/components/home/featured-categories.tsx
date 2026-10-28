import Image from "next/image";
import Link from "next/link";

import { FadeUp, StaggerContainer, StaggerItem } from "@/components/animations";
import type { CategoryWithCount } from "@/lib/products";

interface FeaturedCategoriesProps {
  categories: CategoryWithCount[];
}

/** Shop-by-category grid with photos, names and product counts (PRD §9). */
export function FeaturedCategories({ categories }: FeaturedCategoriesProps) {
  return (
    <section className="mx-auto w-full max-w-7xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
      <FadeUp inView className="mb-8">
        <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
          Shop by category
        </h2>
        <p className="mt-2 max-w-xl text-sm text-muted-foreground sm:text-base">
          Eight curated collections, from kitchen upgrades to pet comforts.
        </p>
      </FadeUp>

      <StaggerContainer
        inView
        className="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 lg:grid-cols-4"
      >
        {categories.map((category) => (
          <StaggerItem key={category.id}>
            <Link
              href={`/category/${category.slug}`}
              className="group relative block overflow-hidden rounded-xl border focus-visible:outline-2 focus-visible:outline-offset-2"
            >
              <div className="relative aspect-square">
                <Image
                  src={category.image.url}
                  alt={category.image.alt}
                  fill
                  sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
                  className="object-cover transition-transform duration-300 group-hover:scale-[1.04]"
                />
                <div
                  aria-hidden="true"
                  className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent"
                />
                <div className="absolute inset-x-0 bottom-0 p-3 sm:p-4">
                  <h3 className="text-sm font-semibold text-white sm:text-base">
                    {category.name}
                  </h3>
                  <p className="text-xs text-white/80">
                    {category.productCount}{" "}
                    {category.productCount === 1 ? "product" : "products"}
                  </p>
                </div>
              </div>
            </Link>
          </StaggerItem>
        ))}
      </StaggerContainer>
    </section>
  );
}
