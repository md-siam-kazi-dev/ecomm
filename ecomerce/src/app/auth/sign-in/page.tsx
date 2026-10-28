import type { Metadata } from "next";
import Link from "next/link";

import { AuthCard } from "@/components/auth/auth-card";
import { SignInForm } from "@/components/auth/sign-in-form";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Sign in",
  description:
    "Sign in to your ShopKart account to track orders and manage your details.",
  path: "/auth/sign-in",
  noIndex: true,
});

export default function SignInPage() {
  return (
    <AuthCard
      title="Welcome back"
      description="Sign in to your ShopKart account."
      footer={
        <>
          New to ShopKart?{" "}
          <Link
            href="/auth/sign-up"
            className="font-medium text-primary-dark underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
          >
            Create an account
          </Link>
        </>
      }
    >
      <SignInForm />
    </AuthCard>
  );
}
