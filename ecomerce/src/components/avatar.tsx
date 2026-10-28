import Image from "next/image";

import { cn } from "@/lib/utils";

function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  return (parts[0][0] + (parts[1]?.[0] ?? "")).toUpperCase();
}

interface AvatarProps {
  name: string;
  /** Profile picture URL. Falls back to the name's initials when absent. */
  src?: string | null;
  /** Size classes, e.g. `"size-10"`. */
  className?: string;
  /** Font size for the initials fallback, e.g. `"text-xl"`. */
  textClassName?: string;
}

/** Profile picture with a first-letter (initials) fallback. */
export function Avatar({ name, src, className, textClassName }: AvatarProps) {
  const base =
    "flex shrink-0 items-center justify-center overflow-hidden rounded-full bg-muted text-xs font-semibold uppercase text-muted-foreground ring-1 ring-border";

  if (src) {
    return (
      <Image
        src={src}
        alt=""
        width={96}
        height={96}
        unoptimized
        className={cn(base, "object-cover", className)}
      />
    );
  }

  return (
    <span className={cn(base, textClassName, className)} aria-hidden="true" title={name}>
      {initials(name)}
    </span>
  );
}
