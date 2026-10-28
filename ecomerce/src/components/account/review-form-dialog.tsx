"use client";

import { useState } from "react";
import Image from "next/image";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ImagePlus, Loader2, X } from "lucide-react";
import { z } from "zod";
import { toast } from "sonner";

import { TextField } from "@/components/forms/text-field";
import { RatingInput } from "@/components/products/rating-input";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { apiPost, apiPut, apiUpload } from "@/lib/api-client";
import type { OrderItemReview } from "@/types/order";

const MAX_IMAGES = 4;

const schema = z.object({
  title: z
    .string()
    .trim()
    .max(255, "Keep the title under 255 characters")
    .optional(),
  comment: z
    .string()
    .trim()
    .min(3, "Write at least a few words")
    .max(2000, "Keep it under 2000 characters"),
});

type FormValues = z.infer<typeof schema>;

interface ApiReview {
  id: string;
  rating: number;
  title: string;
  comment: string;
  images?: string[];
  created_at: string;
}

interface ReviewFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orderItemId: string;
  productName: string;
  /** Existing review → edit mode. */
  existing?: OrderItemReview;
  onSaved: (review: OrderItemReview) => void;
}

/**
 * Star + title + comment review form with optional photos. Creates a review for
 * the delivered order line, or updates the caller's existing review.
 */
export function ReviewFormDialog({
  open,
  onOpenChange,
  orderItemId,
  productName,
  existing,
  onSaved,
}: ReviewFormDialogProps) {
  const [rating, setRating] = useState(existing?.rating ?? 0);
  const [images, setImages] = useState<string[]>(existing?.images ?? []);
  const [uploading, setUploading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    mode: "onTouched",
    defaultValues: { title: existing?.title ?? "", comment: existing?.comment ?? "" },
  });

  async function addFiles(fileList: FileList | null) {
    const files = Array.from(fileList ?? []);
    if (files.length === 0) return;

    const room = MAX_IMAGES - images.length;
    if (room <= 0) {
      toast.error(`You can attach up to ${MAX_IMAGES} photos`);
      return;
    }

    setUploading(true);
    const added: string[] = [];
    for (const file of files.slice(0, room)) {
      const form = new FormData();
      form.append("image", file);
      const res = await apiUpload<{ url: string }>("/api/v1/reviews/images", form);
      if (!res.ok) {
        toast.error(res.error.message);
        continue;
      }
      added.push(res.data.url);
    }
    setUploading(false);
    if (added.length > 0) {
      setImages((prev) => [...prev, ...added].slice(0, MAX_IMAGES));
    }
  }

  const submit = handleSubmit(async (values) => {
    if (rating < 1) {
      toast.error("Select a star rating");
      return;
    }
    setSubmitting(true);
    const payload = {
      rating,
      title: values.title ?? "",
      comment: values.comment,
      images,
    };
    const res = existing
      ? await apiPut<ApiReview>(`/api/v1/reviews/${existing.id}`, payload)
      : await apiPost<ApiReview>("/api/v1/reviews", {
          order_item_id: orderItemId,
          ...payload,
        });
    setSubmitting(false);

    if (!res.ok) {
      toast.error(res.error.message, {
        description: `HTTP ${res.status} · ${res.error.code}`,
      });
      return;
    }

    onSaved({
      id: res.data.id,
      rating: res.data.rating,
      title: res.data.title,
      comment: res.data.comment,
      images: res.data.images,
      createdAt: res.data.created_at,
    });
    toast.success(existing ? "Review updated" : "Review submitted", {
      description: productName,
    });
    onOpenChange(false);
  });

  const busy = submitting || uploading;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{existing ? "Edit your review" : "Write a review"}</DialogTitle>
          <DialogDescription>{productName}</DialogDescription>
        </DialogHeader>

        <form onSubmit={submit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <span className="text-sm font-medium">Your rating</span>
            <RatingInput value={rating} onChange={setRating} disabled={busy} />
          </div>

          <TextField
            label="Title (optional)"
            placeholder="Sums it up in a few words"
            {...register("title")}
          />

          <div className="flex flex-col gap-1.5">
            <label htmlFor="review-comment" className="text-sm font-medium">
              Your review
            </label>
            <textarea
              id="review-comment"
              rows={4}
              disabled={busy}
              aria-invalid={errors.comment ? true : undefined}
              placeholder="What did you like or dislike?"
              className="w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-60"
              {...register("comment")}
            />
            {errors.comment && (
              <p role="alert" className="text-xs font-medium text-destructive">
                {errors.comment.message}
              </p>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <span className="text-sm font-medium">
              Photos (optional, up to {MAX_IMAGES})
            </span>
            <div className="flex flex-wrap items-center gap-2">
              {images.map((url) => (
                <span
                  key={url}
                  className="relative size-16 overflow-hidden rounded-lg border"
                >
                  <Image src={url} alt="" fill sizes="64px" className="object-cover" />
                  <button
                    type="button"
                    onClick={() => setImages((prev) => prev.filter((item) => item !== url))}
                    disabled={busy}
                    aria-label="Remove photo"
                    className="absolute right-0.5 top-0.5 flex size-5 items-center justify-center rounded-full bg-background/90 text-foreground shadow-sm focus-visible:outline-2 focus-visible:outline-offset-1"
                  >
                    <X className="size-3" aria-hidden="true" />
                  </button>
                </span>
              ))}

              {images.length < MAX_IMAGES && (
                <label
                  className="flex size-16 cursor-pointer flex-col items-center justify-center gap-1 rounded-lg border border-dashed text-muted-foreground transition-colors hover:bg-muted focus-within:outline-2 focus-within:outline-offset-2"
                  aria-label="Add a photo"
                >
                  {uploading ? (
                    <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                  ) : (
                    <ImagePlus className="size-4" aria-hidden="true" />
                  )}
                  <input
                    type="file"
                    accept="image/png,image/jpeg,image/webp,image/gif"
                    multiple
                    disabled={busy}
                    onChange={(event) => {
                      void addFiles(event.target.files);
                      event.target.value = "";
                    }}
                    className="sr-only"
                  />
                </label>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={busy}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={busy}>
              {submitting && <Loader2 className="size-4 animate-spin" aria-hidden="true" />}
              {existing ? "Save changes" : "Submit review"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
