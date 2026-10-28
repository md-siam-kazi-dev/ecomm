"use client";

import { useRef, useState, type KeyboardEvent, type MouseEvent, type TouchEvent } from "react";
import Image from "next/image";
import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { Product } from "@/types/product";

interface ProductGalleryProps {
  product: Product;
}

const SWIPE_THRESHOLD_PX = 40;

/**
 * Product gallery (PRD §15): main image with click-to-zoom, thumbnails,
 * prev/next controls, arrow-key navigation and touch swipe. The fixed
 * aspect-ratio wrapper prevents layout shift (PRD §15/§34).
 */
export function ProductGallery({ product }: ProductGalleryProps) {
  const images = product.images;
  const [activeIndex, setActiveIndex] = useState(0);
  const [isZoomed, setIsZoomed] = useState(false);
  const [zoomOrigin, setZoomOrigin] = useState("50% 50%");
  const touchStartX = useRef<number | null>(null);

  const activeImage = images[activeIndex] ?? images[0];

  function showImage(index: number) {
    setActiveIndex((index + images.length) % images.length);
    setIsZoomed(false);
  }

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === "ArrowLeft") {
      event.preventDefault();
      showImage(activeIndex - 1);
    } else if (event.key === "ArrowRight") {
      event.preventDefault();
      showImage(activeIndex + 1);
    }
  }

  function handleTouchStart(event: TouchEvent<HTMLDivElement>) {
    touchStartX.current = event.touches[0].clientX;
  }

  function handleTouchEnd(event: TouchEvent<HTMLDivElement>) {
    if (touchStartX.current === null) return;
    const deltaX = event.changedTouches[0].clientX - touchStartX.current;
    if (Math.abs(deltaX) > SWIPE_THRESHOLD_PX) {
      showImage(activeIndex + (deltaX < 0 ? 1 : -1));
    }
    touchStartX.current = null;
  }

  function handleMouseMove(event: MouseEvent<HTMLDivElement>) {
    const rect = event.currentTarget.getBoundingClientRect();
    setZoomOrigin(
      `${((event.clientX - rect.left) / rect.width) * 100}% ${((event.clientY - rect.top) / rect.height) * 100}%`
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <div
        className="relative aspect-square overflow-hidden rounded-2xl border bg-muted"
        onKeyDown={handleKeyDown}
        onTouchStart={handleTouchStart}
        onTouchEnd={handleTouchEnd}
        onMouseMove={handleMouseMove}
      >
        <button
          type="button"
          onClick={() => setIsZoomed((zoomed) => !zoomed)}
          aria-pressed={isZoomed}
          aria-label={
            isZoomed ? "Zoom out of product image" : "Zoom in on product image"
          }
          className="absolute inset-0 h-full w-full cursor-zoom-in focus-visible:outline-2 focus-visible:-outline-offset-2"
        >
          <Image
            src={activeImage.url}
            alt={activeImage.alt}
            fill
            priority
            sizes="(max-width: 1024px) 100vw, 50vw"
            style={{ transformOrigin: zoomOrigin }}
            className={cn(
              "object-cover transition-transform duration-300",
              isZoomed && "scale-150 cursor-zoom-out"
            )}
          />
        </button>

        {images.length > 1 && (
          <>
            <Button
              type="button"
              variant="secondary"
              size="icon"
              onClick={() => showImage(activeIndex - 1)}
              aria-label="Previous image"
              className="absolute left-3 top-1/2 size-10 -translate-y-1/2 rounded-full shadow-md"
            >
              <ChevronLeft className="size-5" aria-hidden="true" />
            </Button>
            <Button
              type="button"
              variant="secondary"
              size="icon"
              onClick={() => showImage(activeIndex + 1)}
              aria-label="Next image"
              className="absolute right-3 top-1/2 size-10 -translate-y-1/2 rounded-full shadow-md"
            >
              <ChevronRight className="size-5" aria-hidden="true" />
            </Button>
          </>
        )}

        <span className="absolute bottom-3 right-3 rounded-full bg-background/80 px-2.5 py-1 text-xs font-medium tabular-nums">
          {activeIndex + 1} / {images.length}
        </span>
      </div>

      {images.length > 1 && (
        <div className="flex flex-wrap gap-2">
          {images.map((image, index) => (
            <button
              key={image.url}
              type="button"
              onClick={() => showImage(index)}
              aria-label={`View image ${index + 1}: ${image.alt}`}
              aria-current={index === activeIndex}
              className={cn(
                "relative size-16 overflow-hidden rounded-lg border-2 transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 sm:size-20",
                index === activeIndex
                  ? "border-primary"
                  : "border-transparent hover:border-muted-foreground/40"
              )}
            >
              <Image
                src={image.url}
                alt=""
                fill
                sizes="80px"
                className="object-cover"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
