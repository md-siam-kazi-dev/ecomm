"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Loader2, Truck } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { apiPut } from "@/lib/api-client";
import { ANNOUNCEMENT_MESSAGE } from "@/lib/constants";
import { cn } from "@/lib/utils";
import type { SiteSettings } from "@/types/site-settings";

const MAX_ANNOUNCEMENT_LENGTH = 280;

/**
 * Admin editor for the storefront announcement bar (PRD §24). The message is
 * persisted by the Go API and read back by the server-rendered storefront bar.
 */
export function SiteSettingsManager({
  initialSettings,
}: {
  initialSettings: SiteSettings | null;
}) {
  const router = useRouter();
  const [enabled, setEnabled] = useState(initialSettings?.announcement_enabled ?? true);
  const [message, setMessage] = useState(
    initialSettings?.announcement_message ?? ANNOUNCEMENT_MESSAGE
  );
  const [isSaving, setIsSaving] = useState(false);

  const trimmed = message.trim();
  const tooLong = trimmed.length > MAX_ANNOUNCEMENT_LENGTH;
  const messageRequired = enabled && trimmed.length === 0;
  const canSave = !isSaving && !tooLong && !messageRequired;

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canSave) return;

    setIsSaving(true);
    const res = await apiPut<SiteSettings>("/api/v1/admin/settings", {
      announcement_enabled: enabled,
      announcement_message: trimmed,
    });
    setIsSaving(false);

    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }

    setEnabled(res.data.announcement_enabled);
    setMessage(res.data.announcement_message);
    toast.success("Storefront settings saved");
    router.refresh();
  }

  return (
    <form
      onSubmit={onSubmit}
      noValidate
      className="flex max-w-2xl flex-col gap-6"
      aria-labelledby="announcement-heading"
    >
      {initialSettings === null ? (
        <p
          role="alert"
          className="rounded-lg bg-destructive/10 px-3 py-2 text-sm font-medium text-destructive"
        >
          Could not load the current settings. Saving will still update the
          storefront.
        </p>
      ) : null}

      <section className="rounded-xl border bg-card p-5 sm:p-6">
        <h2 id="announcement-heading" className="text-base font-semibold">
          Announcement bar
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          The message shown above the header across the storefront.
        </p>

        <label
          htmlFor="announcement-enabled"
          className="mt-4 flex cursor-pointer items-center gap-3 rounded-xl border p-3.5 transition-colors hover:bg-muted has-focus-visible:ring-2 has-focus-visible:ring-ring/50"
        >
          <input
            id="announcement-enabled"
            type="checkbox"
            checked={enabled}
            onChange={(event) => setEnabled(event.target.checked)}
            className="size-4 shrink-0 accent-primary"
          />
          <span className="text-sm font-medium">Show the announcement bar</span>
        </label>

        <div className="mt-4 flex flex-col gap-1.5">
          <Label htmlFor="announcement-message">Message</Label>
          <textarea
            id="announcement-message"
            value={message}
            onChange={(event) => setMessage(event.target.value)}
            rows={3}
            maxLength={MAX_ANNOUNCEMENT_LENGTH * 2}
            disabled={!enabled}
            aria-invalid={tooLong || messageRequired ? true : undefined}
            aria-describedby="announcement-message-help"
            className={cn(
              "w-full resize-y rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none transition-colors placeholder:text-muted-text focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:bg-input/30"
            )}
          />
          <p
            id="announcement-message-help"
            className={cn(
              "text-xs",
              tooLong || messageRequired
                ? "font-medium text-destructive"
                : "text-muted-foreground"
            )}
          >
            {messageRequired
              ? "Enter a message, or turn the bar off."
              : tooLong
                ? `Message is too long (${trimmed.length}/${MAX_ANNOUNCEMENT_LENGTH}).`
                : `${trimmed.length}/${MAX_ANNOUNCEMENT_LENGTH} characters`}
          </p>
        </div>

        <div className="mt-5">
          <Button type="submit" disabled={!canSave}>
            {isSaving ? (
              <Loader2 className="size-4 animate-spin" aria-hidden="true" />
            ) : null}
            {isSaving ? "Saving…" : "Save changes"}
          </Button>
        </div>
      </section>

      <section className="rounded-xl border bg-card p-5 sm:p-6">
        <h2 className="text-base font-semibold">Preview</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          How the bar appears at the top of the storefront.
        </p>

        <div className="mt-4 overflow-hidden rounded-lg" aria-hidden="true">
          {enabled && trimmed.length > 0 ? (
            <div className="bg-primary text-primary-foreground">
              <div className="flex items-center justify-center gap-2 px-4 py-2">
                <Truck className="size-4 shrink-0" />
                <p className="text-center text-xs font-medium tracking-wide sm:text-sm">
                  {trimmed}
                </p>
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center border border-dashed px-4 py-2 text-xs text-muted-foreground">
              Announcement bar is hidden
            </div>
          )}
        </div>
      </section>
    </form>
  );
}
