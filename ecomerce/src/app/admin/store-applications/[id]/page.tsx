import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { AdminStoreApplicationDetail } from "@/components/admin/store-application-detail";
import { buildMetadata } from "@/lib/seo";
import { getAdminApplicationDetail } from "@/lib/store-applications";

export const metadata: Metadata = buildMetadata({
  title: "Application Review",
  description: "Review a store application.",
  path: "/admin/store-applications",
  noIndex: true,
});

export default async function StoreApplicationDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const detail = await getAdminApplicationDetail(id);
  if (!detail) notFound();

  return <AdminStoreApplicationDetail application={detail} />;
}
