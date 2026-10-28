import type { Metadata } from "next";
import Link from "next/link";

import { AuthCard } from "@/components/auth/auth-card";
import { SignUpForm } from "@/components/auth/sign-up-form";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Create account",
  description:
    "Create a ShopKart account to check out faster and track your orders.",
  path: "/auth/sign-up",
  noIndex: true,
});

export default function SignUpPage() {
  return (
    <AuthCard
      title="Create your account"
      description="Join ShopKart to check out faster and track your orders."
      footer={
        <>
          Already have an account?{" "}
          <Link
            href="/auth/sign-in"
            className="font-medium text-primary-dark underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            Sign in
          </Link>
        </>
      }
    >
      <SignUpForm />
    </AuthCard>
  );
}
