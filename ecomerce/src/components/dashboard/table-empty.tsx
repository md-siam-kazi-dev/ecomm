import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

interface TableEmptyProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  /** Optional recovery control, e.g. a "Clear filters" button. */
  action?: ReactNode;
}

/** Shared "nothing matched" state for the dashboard tables. */
export function TableEmpty({
  icon: Icon,
  title,
  description,
  action,
}: TableEmptyProps) {
  return (
    <Empty className="border-0 p-10">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Icon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        {description && <EmptyDescription>{description}</EmptyDescription>}
      </EmptyHeader>
      {action && <EmptyContent>{action}</EmptyContent>}
    </Empty>
  );
}
