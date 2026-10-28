"use client";

import { useEffect, useRef, useState, type ChangeEvent, type ReactNode } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { ArrowDown, ArrowUp, ImagePlus, Loader2, Trash2, Wand2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { z } from "zod";

import { ProductStatusBadge } from "@/components/vendor/product-status-badge";
import { TextField } from "@/components/forms/text-field";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { apiGet, apiPost, apiPut, apiUpload } from "@/lib/api-client";
import type { ApiCategory } from "@/lib/catalog";
import { formatDate } from "@/lib/formatters";
import type { CategoryOption } from "@/types/category";

interface ProductImage {
  url: string;
  is_primary: boolean;
  sort_order: number;
}

interface ProductDetailApi {
  id: string;
  name: string;
  description: string;
  short_description: string;
  category: string;
  brand: string;
  sku: string;
  price: number;
  sale_price?: number | null;
  stock: number;
  low_stock_threshold: number;
  status: string;
  approval_status: string;
  images: ProductImage[];
  attributes: Record<string, string>;
  review_note?: string;
  submitted_at?: string | null;
  reviewed_at?: string | null;
}

const schema = z
  .object({
    name: z.string().trim().min(2, "Enter a product name"),
    description: z.string().trim().min(1, "Add a description"),
    shortDescription: z.string().trim().optional(),
    category: z.string().min(1, "Select a category"),
    brand: z.string().trim().optional(),
    sku: z.string().trim().optional(),
    price: z.string().min(1, "Enter a price"),
    salePrice: z.string().optional(),
    stock: z.string().min(1, "Enter stock"),
    lowStock: z.string().optional(),
    color: z.string().optional(),
    size: z.string().optional(),
    material: z.string().optional(),
    weight: z.string().optional(),
    dimensions: z.string().optional(),
    warranty: z.string().optional(),
    shippingWeight: z.string().optional(),
    packageDimensions: z.string().optional(),
    shippingClass: z.string().optional(),
    handlingTime: z.string().optional(),
  })
  .superRefine((values, ctx) => {
    const price = Number(values.price);
    if (!Number.isFinite(price) || price <= 0) {
      ctx.addIssue({ path: ["price"], code: "custom", message: "Enter a positive price" });
    }
    if (values.salePrice) {
      const sale = Number(values.salePrice);
      if (!Number.isFinite(sale) || sale <= 0 || sale >= price) {
        ctx.addIssue({
          path: ["salePrice"],
          code: "custom",
          message: "Sale price must be less than the regular price",
        });
      }
    }
    const stock = Number(values.stock);
    if (!Number.isInteger(stock) || stock < 0) {
      ctx.addIssue({ path: ["stock"], code: "custom", message: "Enter a whole number (0 or more)" });
    }
    if (values.lowStock && (!Number.isInteger(Number(values.lowStock)) || Number(values.lowStock) < 0)) {
      ctx.addIssue({ path: ["lowStock"], code: "custom", message: "Enter a whole number" });
    }
  });

type FormValues = z.infer<typeof schema>;

const EMPTY: FormValues = {
  name: "", description: "", shortDescription: "", category: "", brand: "", sku: "",
  price: "", salePrice: "", stock: "", lowStock: "",
  color: "", size: "", material: "", weight: "", dimensions: "", warranty: "",
  shippingWeight: "", packageDimensions: "", shippingClass: "", handlingTime: "",
};

// Sample values for the "Fill test data" button (development only).
const SAMPLE: FormValues = {
  name: "Wireless Portable Blender",
  description:
    "A compact, USB-rechargeable blender for smoothies on the go: six stainless-steel blades, a 400 ml BPA-free jar and a leak-proof lid.",
  shortDescription: "USB-rechargeable portable blender, 400 ml.",
  category: "home-kitchen",
  brand: "FreshMix",
  sku: "",
  price: "39.99",
  salePrice: "29.99",
  stock: "50",
  lowStock: "10",
  color: "White",
  size: "400 ml",
  material: "BPA-free Tritan",
  weight: "0.6 kg",
  dimensions: "8 × 8 × 22 cm",
  warranty: "1 year",
  shippingWeight: "0.8 kg",
  packageDimensions: "10 × 10 × 24 cm",
  shippingClass: "Standard",
  handlingTime: "1–2 days",
};

// Placeholder image so the form is submittable without uploading manually.
const SAMPLE_IMAGE = "https://picsum.photos/seed/vendor-product/800/800";

// The test-fill affordance is hidden in production builds.
const SHOW_TEST_FILL = process.env.NODE_ENV !== "production";

interface ProductFormProps {
  productId?: string;
}

/**
 * Vendor product form (PRD §18). Sections: basic info, pricing & inventory,
 * images, attributes and shipping. Saves a draft or submits for admin review;
 * the backend sets all approval/publishing fields.
 */
export function ProductForm({ productId }: ProductFormProps) {
  const router = useRouter();
  const fileRef = useRef<HTMLInputElement>(null);

  const [images, setImages] = useState<string[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const [loading, setLoading] = useState(Boolean(productId));
  const [submitting, setSubmitting] = useState(false);
  const [header, setHeader] = useState<{ status: string; approval_status: string; submitted_at?: string | null; reviewed_at?: string | null } | null>(null);
  const [reviewNote, setReviewNote] = useState("");

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema), mode: "onTouched", defaultValues: EMPTY });

  // Mirrors the category field for the controlled Select.
  const [category, setCategory] = useState("");
  const [categories, setCategories] = useState<CategoryOption[]>([]);

  // Category options come from the API (no static catalog).
  useEffect(() => {
    let active = true;
    void apiGet<ApiCategory[]>("/api/v1/categories").then((res) => {
      if (active && res.ok) {
        setCategories(res.data.map((c) => ({ slug: c.slug, name: c.name })));
      }
    });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    if (!productId) return;
    let active = true;
    void apiGet<ProductDetailApi>(`/api/v1/vendor/products/${productId}`).then((res) => {
      if (!active) return;
      if (!res.ok) {
        toast.error(res.error.message);
        setLoading(false);
        return;
      }
      const p = res.data;
      const a = p.attributes ?? {};
      reset({
        name: p.name,
        description: p.description,
        shortDescription: p.short_description,
        category: p.category,
        brand: p.brand,
        sku: p.sku,
        price: String(p.price),
        salePrice: p.sale_price ? String(p.sale_price) : "",
        stock: String(p.stock),
        lowStock: p.low_stock_threshold ? String(p.low_stock_threshold) : "",
        color: a.color ?? "", size: a.size ?? "", material: a.material ?? "",
        weight: a.weight ?? "", dimensions: a.dimensions ?? "", warranty: a.warranty ?? "",
        shippingWeight: a.shipping_weight ?? "", packageDimensions: a.package_dimensions ?? "",
        shippingClass: a.shipping_class ?? "", handlingTime: a.handling_time ?? "",
      });
      setImages(p.images.map((img) => img.url));
      setCategory(p.category);
      setHeader({ status: p.status, approval_status: p.approval_status, submitted_at: p.submitted_at, reviewed_at: p.reviewed_at });
      setReviewNote(p.review_note ?? "");
      setLoading(false);
    });
    return () => {
      active = false;
    };
  }, [productId, reset]);

  async function handleImageUpload(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    if (files.length === 0) return;
    setIsUploading(true);
    for (const file of files) {
      const formData = new FormData();
      formData.append("image", file);
      const res = await apiUpload<{ url: string }>("/api/v1/vendor/products/images", formData);
      if (!res.ok) {
        toast.error(res.error.message);
        break;
      }
      setImages((prev) => [...prev, res.data.url]);
    }
    setIsUploading(false);
    if (fileRef.current) fileRef.current.value = "";
  }

  function fillSample() {
    reset(SAMPLE);
    setCategory(SAMPLE.category);
    setImages([SAMPLE_IMAGE]);
    toast.success("Test data filled");
  }

  function moveImage(index: number, direction: -1 | 1) {
    setImages((prev) => {
      const next = [...prev];
      const target = index + direction;
      if (target < 0 || target >= next.length) return prev;
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  }

  const save = (submit: boolean) =>
    handleSubmit(async (values) => {
      if (images.length === 0) {
        toast.error("Add at least the main product image");
        return;
      }
      setSubmitting(true);

      const attributes: Record<string, string> = {};
      const put = (key: string, value?: string) => {
        const trimmed = value?.trim();
        if (trimmed) attributes[key] = trimmed;
      };
      put("color", values.color);
      put("size", values.size);
      put("material", values.material);
      put("weight", values.weight);
      put("dimensions", values.dimensions);
      put("warranty", values.warranty);
      put("shipping_weight", values.shippingWeight);
      put("package_dimensions", values.packageDimensions);
      put("shipping_class", values.shippingClass);
      put("handling_time", values.handlingTime);

      const payload = {
        name: values.name,
        description: values.description,
        short_description: values.shortDescription ?? "",
        category: values.category,
        brand: values.brand ?? "",
        sku: values.sku ?? "",
        price: Number(values.price),
        sale_price: values.salePrice ? Number(values.salePrice) : null,
        stock: Number(values.stock),
        low_stock_threshold: values.lowStock ? Number(values.lowStock) : 0,
        images,
        attributes,
        submit,
      };

      const res = productId
        ? await apiPut<ProductDetailApi>(`/api/v1/vendor/products/${productId}`, payload)
        : await apiPost<ProductDetailApi>("/api/v1/vendor/products", payload);

      setSubmitting(false);
      if (!res.ok) {
        toast.error(res.error.message);
        return;
      }
      toast.success(submit ? "Submitted for review" : "Draft saved");
      router.push("/vendor/products");
      router.refresh();
    });

  if (loading) {
    return (
      <div className="flex flex-col gap-4" role="status" aria-label="Loading product">
        <Skeleton className="h-40 rounded-xl" />
        <Skeleton className="h-40 rounded-xl" />
      </div>
    );
  }

  return (
    <form className="flex flex-col gap-6" noValidate>
      {SHOW_TEST_FILL && (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-dashed border-input px-4 py-3">
          <p className="text-xs text-muted-foreground">
            Testing? Populate the form with sample values and a placeholder image.
          </p>
          <Button type="button" variant="outline" size="sm" onClick={fillSample}>
            <Wand2 className="size-3.5" aria-hidden="true" />
            Fill test data
          </Button>
        </div>
      )}

      {header && (
        <div className="flex flex-col gap-2 rounded-xl border p-4">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium">Status</span>
            <ProductStatusBadge product={header} />
          </div>
          {reviewNote && (
            <p className="text-sm text-muted-foreground">
              <span className="font-medium text-foreground">Admin feedback:</span> {reviewNote}
            </p>
          )}
          {header.reviewed_at && (
            <p className="text-xs text-muted-foreground">Last reviewed {formatDate(header.reviewed_at)}</p>
          )}
        </div>
      )}

      <Section title="Basic information">
        <div className="sm:col-span-2">
          <TextField label="Product name" error={errors.name?.message} {...register("name")} />
        </div>
        <div className="sm:col-span-2">
          <Label htmlFor="description">Description</Label>
          <textarea
            id="description"
            rows={4}
            className={textareaClass}
            {...register("description")}
          />
          {errors.description && <FieldError>{errors.description.message}</FieldError>}
        </div>
        <div className="sm:col-span-2">
          <TextField label="Short description (optional)" error={errors.shortDescription?.message} {...register("shortDescription")} />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="category">Category</Label>
          <Select
            value={category}
            onValueChange={(value) => {
              setCategory(value);
              setValue("category", value, { shouldValidate: true });
            }}
          >
            <SelectTrigger id="category" className="w-full" aria-invalid={errors.category ? true : undefined}>
              <SelectValue placeholder="Select a category" />
            </SelectTrigger>
            <SelectContent>
              {categories.map((c) => (
                <SelectItem key={c.slug} value={c.slug}>
                  {c.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.category && <FieldError>{errors.category.message}</FieldError>}
        </div>
        <TextField label="Brand (optional)" error={errors.brand?.message} {...register("brand")} />
        <TextField label="SKU (optional)" hint="Auto-generated when left blank" error={errors.sku?.message} {...register("sku")} />
      </Section>

      <Section title="Pricing & inventory">
        <TextField label="Regular price" type="number" step="0.01" min="0" inputMode="decimal" error={errors.price?.message} {...register("price")} />
        <TextField label="Sale price (optional)" type="number" step="0.01" min="0" inputMode="decimal" error={errors.salePrice?.message} {...register("salePrice")} />
        <TextField label="Stock quantity" type="number" min="0" inputMode="numeric" error={errors.stock?.message} {...register("stock")} />
        <TextField label="Low-stock threshold (optional)" type="number" min="0" inputMode="numeric" error={errors.lowStock?.message} {...register("lowStock")} />
      </Section>

      <Section title="Product images">
        <div className="sm:col-span-2 flex flex-col gap-3">
          <p className="text-sm text-muted-foreground">
            The first image is the main image. PNG, JPG, WebP or GIF · max 5 MB.
          </p>
          <div className="flex flex-wrap gap-3">
            {images.map((url, index) => (
              <div key={url} className="relative size-24 overflow-hidden rounded-xl border">
                <Image src={url} alt={`Product image ${index + 1}`} width={96} height={96} className="size-24 object-cover" />
                {index === 0 && (
                  <span className="absolute left-1 top-1 rounded bg-primary px-1 text-[10px] font-semibold text-primary-foreground">
                    Main
                  </span>
                )}
                <div className="absolute inset-x-0 bottom-0 flex items-center justify-between bg-black/50 px-1 py-0.5">
                  <button type="button" onClick={() => moveImage(index, -1)} disabled={index === 0} aria-label="Move image earlier" className="text-white disabled:opacity-30">
                    <ArrowUp className="size-3.5" />
                  </button>
                  <button type="button" onClick={() => setImages((prev) => prev.filter((u) => u !== url))} aria-label="Remove image" className="text-white">
                    <Trash2 className="size-3.5" />
                  </button>
                  <button type="button" onClick={() => moveImage(index, 1)} disabled={index === images.length - 1} aria-label="Move image later" className="text-white disabled:opacity-30">
                    <ArrowDown className="size-3.5" />
                  </button>
                </div>
              </div>
            ))}
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              disabled={isUploading}
              className="flex size-24 flex-col items-center justify-center gap-1 rounded-xl border border-dashed text-xs text-muted-foreground hover:bg-muted disabled:opacity-50"
            >
              {isUploading ? <Loader2 className="size-5 animate-spin" /> : <ImagePlus className="size-5" />}
              {isUploading ? "Uploading" : "Add image"}
            </button>
            <input ref={fileRef} type="file" accept="image/png,image/jpeg,image/webp,image/gif" multiple className="sr-only" onChange={handleImageUpload} />
          </div>
        </div>
      </Section>

      <Section title="Product attributes (optional)">
        <TextField label="Color" error={errors.color?.message} {...register("color")} />
        <TextField label="Size" error={errors.size?.message} {...register("size")} />
        <TextField label="Material" error={errors.material?.message} {...register("material")} />
        <TextField label="Weight" placeholder="e.g. 1.2 kg" error={errors.weight?.message} {...register("weight")} />
        <TextField label="Dimensions" placeholder="e.g. 20 × 10 × 8 cm" error={errors.dimensions?.message} {...register("dimensions")} />
        <TextField label="Warranty" placeholder="e.g. 1 year" error={errors.warranty?.message} {...register("warranty")} />
      </Section>

      <Section title="Shipping (optional)">
        <TextField label="Shipping weight" placeholder="e.g. 1.5 kg" error={errors.shippingWeight?.message} {...register("shippingWeight")} />
        <TextField label="Package dimensions" placeholder="e.g. 25 × 15 × 10 cm" error={errors.packageDimensions?.message} {...register("packageDimensions")} />
        <TextField label="Shipping class" placeholder="e.g. Standard" error={errors.shippingClass?.message} {...register("shippingClass")} />
        <TextField label="Handling time" placeholder="e.g. 1–2 days" error={errors.handlingTime?.message} {...register("handlingTime")} />
      </Section>

      <div className="flex flex-wrap gap-3">
        <Button type="button" disabled={submitting} onClick={save(true)} className="h-10">
          {submitting && <Loader2 className="size-4 animate-spin" aria-hidden="true" />}
          {productId ? "Save & resubmit for review" : "Submit for review"}
        </Button>
        <Button type="button" variant="outline" disabled={submitting} onClick={save(false)} className="h-10">
          Save draft
        </Button>
        <Button type="button" variant="ghost" onClick={() => router.push("/vendor/products")} className="h-10">
          Cancel
        </Button>
      </div>
    </form>
  );
}

const textareaClass =
  "min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50";

function FieldError({ children }: { children: ReactNode }) {
  return (
    <p role="alert" className="text-xs font-medium text-destructive">
      {children}
    </p>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-xl border p-5 sm:p-6" aria-label={title}>
      <h2 className="text-base font-semibold">{title}</h2>
      <div className="mt-4 grid gap-4 sm:grid-cols-2">{children}</div>
    </section>
  );
}
