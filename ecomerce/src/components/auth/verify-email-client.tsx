"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import { apiPost } from "@/lib/api-client";

type Status = "pending" | "success" | "error";

export function VerifyEmailClient() {
  const token = useSearchParams().get("token");
  // Derived from the token so the missing-token case never needs an effect.
  const [status, setStatus] = useState<Status>(token ? "pending" : "error");
  const [message, setMessage] = useState(
    token ? "Verifying your email…" : "This verification link is missing its token."
  );

  useEffect(() => {
    if (!token) return;

    let active = true;
    void apiPost("/api/v1/auth/verify-email", { token }).then((res) => {
      if (!active) return;
      if (res.ok) {
        setStatus("success");
        setMessage("Your email has been verified. You can now sign in.");
      } else {
        setStatus("error");
        setMessage(res.error.message);
      }
    });

    return () => {
      active = false;
    };
  }, [token]);

  const tone =
    status === "error"
      ? "border-destructive/30 bg-destructive/10 text-destructive"
      : "border-primary/30 bg-primary/10 text-primary-dark";

  return (
    <div className="flex flex-col gap-4">
      <p role="status" className={`rounded-lg border px-3 py-2 text-sm font-medium ${tone}`}>
        {message}
      </p>
      {status !== "pending" && (
        <Button asChild className="h-10 w-full">
          <Link href="/auth/sign-in">Go to sign in</Link>
        </Button>
      )}
    </div>
  );
}
