import { Check, Dot } from "lucide-react";

import { cn } from "@/lib/utils";
import type { ApplicationDetail } from "@/types/store-application";

type StepState = "done" | "current" | "upcoming" | "failed";

interface Step {
  label: string;
  state: StepState;
}

function buildSteps(app: ApplicationDetail): Step[] {
  const reviewed = Boolean(app.reviewed_at) || app.status !== "pending";
  const approved = app.status === "approved";
  const rejected = app.status === "rejected";
  const active = app.store?.status === "active";
  const paid = app.payment?.status === "paid";

  const steps: Step[] = [
    { label: "Application submitted", state: "done" },
    {
      label: "Admin review",
      state: rejected ? "failed" : reviewed ? "done" : "current",
    },
    {
      label: "Approved",
      state: rejected ? "failed" : approved ? "done" : "upcoming",
    },
  ];

  if (app.fee?.enabled) {
    steps.push({
      label: "Application fee paid",
      state: paid ? "done" : approved && !rejected ? "current" : "upcoming",
    });
  }
  steps.push({
    label: "Store activated",
    state: active ? "done" : approved ? "upcoming" : "upcoming",
  });

  return steps;
}

/** Horizontal progress timeline for the store application. */
export function ApplicationTimeline({
  application,
}: {
  application: ApplicationDetail;
}) {
  const steps = buildSteps(application);

  return (
    <ol className="flex flex-wrap gap-x-2 gap-y-3" aria-label="Application progress">
      {steps.map((step, index) => (
        <li key={step.label} className="flex items-center gap-2">
          <span
            className={cn(
              "flex size-5 items-center justify-center rounded-full text-[11px] font-semibold",
              step.state === "done" && "bg-success text-success-foreground",
              step.state === "current" && "bg-warning/20 text-warning",
              step.state === "failed" && "bg-destructive/15 text-destructive",
              step.state === "upcoming" && "bg-muted text-muted-foreground"
            )}
            aria-hidden="true"
          >
            {step.state === "done" ? <Check className="size-3" /> : <Dot className="size-4" />}
          </span>
          <span
            className={cn(
              "text-xs font-medium",
              step.state === "upcoming" ? "text-muted-foreground" : "text-foreground"
            )}
          >
            {step.label}
          </span>
          {index < steps.length - 1 && (
            <span className="mx-1 hidden h-px w-6 bg-border sm:block" aria-hidden="true" />
          )}
        </li>
      ))}
    </ol>
  );
}
