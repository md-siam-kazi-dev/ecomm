import type { Metadata } from "next";

import { ProductsTable } from "@/components/admin/products-table";
import { PageHeader } from "@/components/dashboard/page-header";
import { getProducts } from "@/lib/products";
import { buildMetadata } from "@/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Admin Products",
  description: "Manage the ShopKart product catalogue.",
  path: "/admin/products",
  noIndex: true,
});

export default async function AdminProductsPage() {
  const { items } = await getProducts({ pageSize: 48 });

  return (
    <>
      <PageHeader
        title="Products"
        description={`${items.length} products in the catalogue.`}
      />

      <ProductsTable products={items} />
    </>
  );
}
