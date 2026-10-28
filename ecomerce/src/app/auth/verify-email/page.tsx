import type { Metadata } from "next";
import { Suspense } from "react";

import { AuthCard } from "@/components/auth/auth-card";
import { VerifyEmailClient } from "@/components/auth/verify-email-client";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Verify email",
  description: "Verify the email address on your ShopKart account.",
  path: "/auth/verify-email",
  noIndex: true,
});

export default function VerifyEmailPage() {
  return (
    <AuthCard
      title="Email verification"
      description="Confirming your email address."
    >
      <Suspense fallback={null}>
        <VerifyEmailClient />
      </Suspense>
    </AuthCard>
  );
}
