"use client";

import { useState, type FormEvent } from "react";
import { Loader2, Send } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

/**
 * Demo newsletter signup (PRD §51): collects nothing real — submission is
 * simulated client-side and confirmed with a toast. Replace the timeout with
 * the future Go API call.
 */
export function NewsletterForm() {
  const [email, setEmail] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isSubmitting) return;

    setIsSubmitting(true);

    window.setTimeout(() => {
      setIsSubmitting(false);
      setEmail("");
      toast.success("You're subscribed!", {
        description:
          "Welcome to the ShopKart newsletter — demo only, no emails are stored.",
      });
    }, 800);
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-2">
      <Label htmlFor="newsletter-email" className="sr-only">
        Email address
      </Label>
      <div className="flex gap-2">
        <Input
          id="newsletter-email"
          type="email"
          required
          placeholder="you@example.com"
          autoComplete="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          disabled={isSubmitting}
        />
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? (
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          ) : (
            <Send className="size-4" aria-hidden="true" />
          )}
          Subscribe
        </Button>
      </div>
      <p className="text-xs text-muted-foreground">
        Get 10% off your first order plus early access to deals. Unsubscribe
        anytime.
      </p>
    </form>
  );
}
