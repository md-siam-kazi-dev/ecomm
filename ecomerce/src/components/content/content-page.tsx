import type { ReactNode } from "react";

import { Breadcrumb } from "@/components/navigation/breadcrumb";
import { cn } from "@/lib/utils";

export interface ContentSection {
  heading: string;
  paragraphs: string[];
}

interface ContentPageProps {
  title: string;
  description: string;
  lastUpdated?: string;
  sections?: ContentSection[];
  /** Extra content rendered between the header and the sections. */
  children?: ReactNode;
  className?: string;
}

/** Shared layout for static informational pages (PRD §9 footer links). */
export function ContentPage({
  title,
  description,
  lastUpdated,
  sections = [],
  children,
  className,
}: ContentPageProps) {
  return (
    <div
      className={cn(
        "mx-auto w-full max-w-3xl px-4 py-10 sm:px-6 lg:px-8",
        className
      )}
    >
      <Breadcrumb items={[{ label: "Home", href: "/" }, { label: title }]} />

      <header className="mt-4">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          {title}
        </h1>
        <p className="mt-3 text-muted-foreground">{description}</p>
        {lastUpdated && (
          <p className="mt-2 text-xs text-muted-foreground">
            Last updated {lastUpdated}
          </p>
        )}
      </header>

      {children}

      {sections.length > 0 && (
        <div className="mt-10 flex flex-col gap-8">
          {sections.map((section) => (
            <section key={section.heading}>
              <h2 className="text-xl font-semibold tracking-tight">
                {section.heading}
              </h2>
              <div className="mt-3 flex flex-col gap-3 text-sm leading-relaxed text-muted-foreground">
                {section.paragraphs.map((paragraph) => (
                  <p key={paragraph}>{paragraph}</p>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}
