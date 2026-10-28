"use client";

import { useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { TextField } from "@/components/forms/text-field";
import { Button } from "@/components/ui/button";
import { apiPost } from "@/lib/api-client";

const schema = z
  .object({
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirmPassword: z.string(),
  })
  .refine((values) => values.password === values.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type Values = z.infer<typeof schema>;

export function ResetPasswordForm() {
  const token = useSearchParams().get("token");
  const [formError, setFormError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<Values>({
    resolver: zodResolver(schema),
    mode: "onTouched",
    defaultValues: { password: "", confirmPassword: "" },
  });

  const onSubmit = handleSubmit(async (values) => {
    if (!token) {
      setFormError("This reset link is missing or invalid.");
      return;
    }
    setFormError(null);

    const res = await apiPost("/api/v1/auth/reset-password", {
      token,
      new_password: values.password,
    });
    if (!res.ok) {
      setFormError(res.error.message);
      return;
    }
    setDone(true);
  });

  if (done) {
    return (
      <div className="flex flex-col gap-4">
        <p
          role="status"
          className="rounded-lg border border-primary/30 bg-primary/10 px-3 py-2 text-sm font-medium text-primary-dark"
        >
          Your password has been updated. All other sessions were signed out.
        </p>
        <Button asChild className="h-10 w-full">
          <Link href="/auth/sign-in">Sign in</Link>
        </Button>
      </div>
    );
  }

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {formError && (
        <p
          role="alert"
          className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm font-medium text-destructive"
        >
          {formError}
        </p>
      )}

      <TextField
        label="New password"
        type="password"
        autoComplete="new-password"
        placeholder="At least 8 characters"
        error={errors.password?.message}
        {...register("password")}
      />
      <TextField
        label="Confirm password"
        type="password"
        autoComplete="new-password"
        placeholder="Repeat your password"
        error={errors.confirmPassword?.message}
        {...register("confirmPassword")}
      />

      <Button type="submit" disabled={isSubmitting} className="h-10 w-full">
        {isSubmitting ? "Updating…" : "Update password"}
      </Button>
    </form>
  );
}
