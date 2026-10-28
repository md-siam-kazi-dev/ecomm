import type { Metadata } from "next";

import { SiteSettingsManager } from "@/components/admin/site-settings-manager";
import { PageHeader } from "@/components/dashboard/page-header";
import { buildMetadata } from "@/lib/seo";
import { getAdminSiteSettings } from "@/lib/settings";

export const metadata: Metadata = buildMetadata({
  title: "Store settings",
  description: "Manage the storefront announcement bar and other site settings.",
  path: "/admin/settings",
  noIndex: true,
});

export default async function AdminSettingsPage() {
  const settings = await getAdminSiteSettings();

  return (
    <>
      <PageHeader
        title="Store settings"
        description="Configure text shown across the storefront. Changes go live immediately."
      />

      <SiteSettingsManager initialSettings={settings} />
    </>
  );
}
