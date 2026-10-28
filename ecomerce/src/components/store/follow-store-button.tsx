"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { UserCheck, UserPlus } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { apiDelete, apiPost } from "@/lib/api-client";
import { useAuthStore } from "@/store/auth-store";

interface FollowStoreButtonProps {
  storeSlug: string;
  /** Server-resolved follow state; null when the viewer is signed out. */
  initialFollowing: boolean | null;
  initialFollowerCount: number;
}

interface FollowResponse {
  following: boolean;
  follower_count: number;
}

/** Follow / unfollow a store. Optimistic — flips instantly, no loading state. */
export function FollowStoreButton({
  storeSlug,
  initialFollowing,
  initialFollowerCount,
}: FollowStoreButtonProps) {
  const router = useRouter();
  const status = useAuthStore((state) => state.status);
  const [following, setFollowing] = useState(initialFollowing ?? false);
  const [count, setCount] = useState(initialFollowerCount);
  const inFlight = useRef(false);

  async function toggle() {
    if (status !== "authenticated") {
      toast.error("Sign in to follow this store");
      router.push("/auth/sign-in");
      return;
    }
    // Ignore repeat clicks while a request is in flight (no visible loading).
    if (inFlight.current) return;
    inFlight.current = true;

    const next = !following;
    const delta = next ? 1 : -1;

    // Optimistic: flip immediately.
    setFollowing(next);
    setCount((prev) => Math.max(0, prev + delta));

    const path = `/api/v1/stores/${encodeURIComponent(storeSlug)}/follow`;
    const res = next
      ? await apiPost<FollowResponse>(path)
      : await apiDelete<FollowResponse>(path);

    inFlight.current = false;

    if (!res.ok) {
      // Revert the optimistic change on failure.
      setFollowing(!next);
      setCount((prev) => Math.max(0, prev - delta));
      toast.error(res.error.message, {
        description: `HTTP ${res.status} · ${res.error.code}`,
      });
      return;
    }

    // Reconcile with the server's authoritative values.
    setFollowing(res.data.following);
    setCount(res.data.follower_count);
  }

  return (
    <Button
      type="button"
      variant={following ? "outline" : "default"}
      onClick={toggle}
      aria-pressed={following}
    >
      {following ? (
        <UserCheck className="size-4" aria-hidden="true" />
      ) : (
        <UserPlus className="size-4" aria-hidden="true" />
      )}
      {following ? "Following" : "Follow"}
      <span className="tabular-nums">· {count}</span>
    </Button>
  );
}
