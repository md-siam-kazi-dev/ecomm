import type { Metadata } from "next";

import { ProfileSettings } from "@/components/account/profile-settings";
import { buildMetadata } from "@/lib/seo";
import { requireCurrentUser } from "@/lib/session";

export const metadata: Metadata = buildMetadata({
  title: "Account Settings",
  description: "Update your ShopKart profile and email preferences.",
  path: "/account/settings",
  noIndex: true,
});

export default async function AccountSettingsPage() {
  const user = await requireCurrentUser();

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-xl font-semibold">Settings</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Manage your profile details and email preferences.
        </p>
      </div>
      <ProfileSettings user={user} />
    </div>
  );
}
