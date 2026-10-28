import type { Metadata } from "next";
import { PackageX, TriangleAlert } from "lucide-react";

import { InventoryTable } from "@/components/admin/inventory-table";
import { StatCard } from "@/components/admin/stat-card";
import { PageHeader } from "@/components/dashboard/page-header";
import { LOW_STOCK_THRESHOLD } from "@/lib/constants";
import { getProducts } from "@/lib/products";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Admin Inventory",
  description: "Monitor and adjust ShopKart stock levels.",
  path: "/admin/inventory",
  noIndex: true,
});

export default async function AdminInventoryPage() {
  const { items } = await getProducts({ pageSize: 48 });

  const outOfStock = items.filter((product) => product.stock <= 0).length;
  const lowStock = items.filter(
    (product) => product.stock > 0 && product.stock <= LOW_STOCK_THRESHOLD
  ).length;

  return (
    <>
      <PageHeader
        title="Inventory"
        description="Track stock levels and flag items that need restocking."
      />

      <div className="grid gap-4 sm:grid-cols-2">
        <StatCard
          label="Out of stock"
          value={String(outOfStock)}
          count={{ to: outOfStock }}
          icon={PackageX}
        />
        <StatCard
          label={`Low stock (≤ ${LOW_STOCK_THRESHOLD})`}
          value={String(lowStock)}
          count={{ to: lowStock }}
          icon={TriangleAlert}
        />
      </div>

      <InventoryTable products={items} />
    </>
  );
}
