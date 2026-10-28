"use client";

import { useState } from "react";
import { Loader2, Send } from "lucide-react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { z } from "zod";

import { TextField } from "@/components/forms/text-field";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

const contactSchema = z.object({
  name: z.string().trim().min(2, "Enter your name"),
  email: z.email("Enter a valid email address"),
  subject: z.string().trim().min(3, "Add a short subject"),
  message: z.string().trim().min(10, "Tell us a little more (10+ characters)"),
});

type ContactFormValues = z.infer<typeof contactSchema>;

const SEND_LATENCY_MS = 700;

/** Contact form (PRD §18/§50). Demo only — nothing is sent anywhere. */
export function ContactForm() {
  const [isSending, setIsSending] = useState(false);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ContactFormValues>({
    resolver: zodResolver(contactSchema),
    mode: "onTouched",
  });

  const onSubmit = handleSubmit(() => {
    if (isSending) return;
    setIsSending(true);

    window.setTimeout(() => {
      setIsSending(false);
      reset();
      toast.success("Message sent", {
        description: "Demo only — your message was not actually delivered.",
      });
    }, SEND_LATENCY_MS);
  });

  return (
    <form
      onSubmit={onSubmit}
      noValidate
      className="mt-8 flex flex-col gap-4 rounded-2xl border p-5 sm:p-6"
      aria-labelledby="contact-form-heading"
    >
      <div>
        <h2 id="contact-form-heading" className="text-base font-semibold">
          Send us a message
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          We usually reply within one business day.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <TextField
          label="Your name"
          autoComplete="name"
          error={errors.name?.message}
          {...register("name")}
        />
        <TextField
          label="Email address"
          type="email"
          autoComplete="email"
          error={errors.email?.message}
          {...register("email")}
        />
      </div>

      <TextField
        label="Subject"
        error={errors.subject?.message}
        {...register("subject")}
      />

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="contact-message">Message</Label>
        <textarea
          id="contact-message"
          rows={5}
          aria-invalid={errors.message ? true : undefined}
          aria-describedby={
            errors.message ? "contact-message-error" : undefined
          }
          className="min-h-28 w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none transition-colors placeholder:text-muted-text focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:bg-input/30"
          {...register("message")}
        />
        {errors.message && (
          <p
            id="contact-message-error"
            role="alert"
            className="text-xs font-medium text-destructive"
          >
            {errors.message.message}
          </p>
        )}
      </div>

      <div>
        <Button type="submit" disabled={isSending}>
          {isSending ? (
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          ) : (
            <Send className="size-4" aria-hidden="true" />
          )}
          {isSending ? "Sending…" : "Send message"}
        </Button>
      </div>
    </form>
  );
}
