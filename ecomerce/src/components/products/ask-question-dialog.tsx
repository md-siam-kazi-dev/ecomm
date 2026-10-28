"use client";

import { useState } from "react";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { apiPost } from "@/lib/api-client";

interface AskQuestionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  productSlug: string;
  onSubmitted: () => void;
}

/** Submit a product question (any signed-in user, no purchase required). */
export function AskQuestionDialog({
  open,
  onOpenChange,
  productSlug,
  onSubmitted,
}: AskQuestionDialogProps) {
  const [question, setQuestion] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    const trimmed = question.trim();
    if (trimmed.length < 5) {
      toast.error("Write a question of at least 5 characters");
      return;
    }

    setSubmitting(true);
    const res = await apiPost(`/api/v1/products/${encodeURIComponent(productSlug)}/questions`, {
      question: trimmed,
    });
    setSubmitting(false);

    if (!res.ok) {
      toast.error(res.error.message, {
        description: `HTTP ${res.status} · ${res.error.code}`,
      });
      return;
    }

    setQuestion("");
    toast.success("Question submitted", {
      description: "The seller has been notified.",
    });
    onSubmitted();
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Ask a question</DialogTitle>
          <DialogDescription>
            Ask the seller anything about this product. Answers are public.
          </DialogDescription>
        </DialogHeader>

        <label htmlFor="ask-question" className="sr-only">
          Your question
        </label>
        <textarea
          id="ask-question"
          rows={4}
          value={question}
          disabled={submitting}
          onChange={(event) => setQuestion(event.target.value)}
          placeholder="Is this dishwasher safe?"
          className="w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-60"
        />

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={submitting}
          >
            Cancel
          </Button>
          <Button type="button" onClick={submit} disabled={submitting}>
            {submitting && <Loader2 className="size-4 animate-spin" aria-hidden="true" />}
            Submit question
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
