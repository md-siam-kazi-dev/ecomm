"use client";

import { useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { Check, ImageIcon, Loader2, Upload, Wand2, X } from "lucide-react";
import { toast } from "sonner";

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
import { DIVISIONS } from "@/data/bd-locations";
import { apiGet, apiPost, apiUpload } from "@/lib/api-client";
import type { ApiCategory } from "@/lib/catalog";
import type { CategoryOption } from "@/types/category";
import { cn } from "@/lib/utils";
import type { ApplicationDetail } from "@/types/store-application";

const MAX_IMAGE_BYTES = 5 * 1024 * 1024;

interface FormState {
  storeName: string;
  storeSlug: string;
  description: string;
  category: string;
  logoUrl: string;
  bannerUrl: string;
  businessEmail: string;
  contactPhone: string;
  supportPhone: string;
  websiteUrl: string;
  country: string;
  division: string;
  district: string;
  fullAddress: string;
  postalCode: string;
  businessType: "individual" | "registered";
  registeredBusinessName: string;
  tradeLicenseNumber: string;
}

const INITIAL: FormState = {
  storeName: "",
  storeSlug: "",
  description: "",
  category: "",
  logoUrl: "",
  bannerUrl: "",
  businessEmail: "",
  contactPhone: "",
  supportPhone: "",
  websiteUrl: "",
  country: "Bangladesh",
  division: "",
  district: "",
  fullAddress: "",
  postalCode: "",
  businessType: "individual",
  registeredBusinessName: "",
  tradeLicenseNumber: "",
};

// Sample values used by the "Fill test data" button (development only).
const SAMPLE: FormState = {
  storeName: "Dhaka Gadgets",
  storeSlug: "dhaka-gadgets",
  description:
    "Quality electronics and everyday gadgets shipped across Bangladesh — audio, charging, lighting and portable devices.",
  category: "electronics",
  logoUrl: "",
  bannerUrl: "",
  businessEmail: "hello@dhakagadgets.test",
  contactPhone: "+8801712345678",
  supportPhone: "+8809612345678",
  websiteUrl: "https://dhakagadgets.test",
  country: "Bangladesh",
  division: "Dhaka",
  district: "Dhaka",
  fullAddress: "House 12, Road 5, Dhanmondi",
  postalCode: "1205",
  businessType: "registered",
  registeredBusinessName: "Dhaka Gadgets Ltd.",
  tradeLicenseNumber: "TRAD-123456",
};

// The test-fill affordance is hidden in production builds.
const SHOW_TEST_FILL = process.env.NODE_ENV !== "production";

function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function validate(form: FormState): Record<string, string> {
  const errors: Record<string, string> = {};
  if (form.storeName.trim().length < 2) errors.storeName = "Enter your store name.";
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(form.storeSlug) || form.storeSlug.length < 3)
    errors.storeSlug = "Use lowercase letters, numbers and hyphens (min 3).";
  if (!form.description.trim()) errors.description = "Describe your store.";
  if (!form.category) errors.category = "Choose a category.";
  if (!form.businessEmail.trim() || !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(form.businessEmail))
    errors.businessEmail = "Enter a valid business email.";
  if (!form.contactPhone.trim()) errors.contactPhone = "Enter a contact phone.";
  if (form.websiteUrl && !/^https?:\/\/.+/i.test(form.websiteUrl))
    errors.websiteUrl = "Enter a full URL starting with http:// or https://.";
  if (!form.division) errors.division = "Choose a division.";
  if (!form.district) errors.district = "Choose a district.";
  if (!form.fullAddress.trim()) errors.fullAddress = "Enter the full address.";
  if (form.businessType === "registered" && !form.registeredBusinessName.trim())
    errors.registeredBusinessName = "Enter the registered business name.";
  return errors;
}

const SECTION_CLASS = "rounded-2xl bg-background p-5 ring-1 ring-foreground/10 sm:p-6";

export function StoreApplicationForm() {
  const router = useRouter();
  const [form, setForm] = useState<FormState>(INITIAL);
  const [agreed, setAgreed] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [uploading, setUploading] = useState<"logo" | "banner" | null>(null);
  const slugTouched = useRef(false);

  const districts = useMemo(
    () => DIVISIONS.find((d) => d.name === form.division)?.districts ?? [],
    [form.division]
  );

  // Category options come from the API (no static catalog).
  const [categories, setCategories] = useState<CategoryOption[]>([]);
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

  function set<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  function onNameChange(value: string) {
    setForm((prev) => ({
      ...prev,
      storeName: value,
      storeSlug: slugTouched.current ? prev.storeSlug : slugify(value),
    }));
  }

  function fillSample() {
    slugTouched.current = true;
    setForm({ ...SAMPLE });
    setAgreed(true);
    setErrors({});
    toast.success("Test data filled");
  }

  async function upload(kind: "logo" | "banner", file: File) {
    if (file.size > MAX_IMAGE_BYTES) {
      toast.error("Image is too large (max 5 MB).");
      return;
    }
    setUploading(kind);
    const data = new FormData();
    data.append("image", file);
    const res = await apiUpload<{ url: string }>(
      "/api/v1/store-applications/images",
      data
    );
    setUploading(null);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    set(kind === "logo" ? "logoUrl" : "bannerUrl", res.data.url);
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    const found = validate(form);
    setErrors(found);
    if (Object.keys(found).length > 0) {
      toast.error("Please fix the highlighted fields.");
      return;
    }
    if (!agreed) {
      toast.error("Confirm the information and agree to the vendor terms.");
      return;
    }

    setSubmitting(true);
    const res = await apiPost<ApplicationDetail>("/api/v1/store-applications", {
      store_name: form.storeName.trim(),
      store_slug: form.storeSlug,
      description: form.description.trim(),
      category: form.category,
      logo_url: form.logoUrl,
      banner_url: form.bannerUrl,
      business_email: form.businessEmail.trim(),
      contact_phone: form.contactPhone.trim(),
      support_phone: form.supportPhone.trim(),
      website_url: form.websiteUrl.trim(),
      country: form.country,
      division: form.division,
      district: form.district,
      full_address: form.fullAddress.trim(),
      postal_code: form.postalCode.trim(),
      business_type: form.businessType,
      registered_business_name: form.registeredBusinessName.trim(),
      trade_license_number: form.tradeLicenseNumber.trim(),
      agreed_to_terms: agreed,
    });
    setSubmitting(false);

    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Application submitted", {
      description: "We'll review your store application shortly.",
    });
    router.push("/account/store");
    router.refresh();
  }

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-5" noValidate>
      {SHOW_TEST_FILL && (
        <div className="flex items-center justify-between gap-3 rounded-xl border border-dashed border-input px-4 py-3">
          <p className="text-xs text-muted-foreground">
            Testing? Populate the form with sample values.
          </p>
          <Button type="button" variant="outline" size="sm" onClick={fillSample}>
            <Wand2 className="size-3.5" aria-hidden="true" />
            Fill test data
          </Button>
        </div>
      )}

      {/* A. Basic store information */}
      <section className={SECTION_CLASS}>
        <h2 className="text-lg font-semibold">Store information</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          The basics customers will see on your storefront.
        </p>
        <div className="mt-5 grid gap-4 sm:grid-cols-2">
          <TextField
            label="Store name"
            name="storeName"
            value={form.storeName}
            onChange={(e) => onNameChange(e.target.value)}
            error={errors.storeName}
            placeholder="e.g. Dhaka Gadgets"
            required
          />
          <TextField
            label="Store URL"
            name="storeSlug"
            value={form.storeSlug}
            onChange={(e) => {
              slugTouched.current = true;
              set("storeSlug", slugify(e.target.value));
            }}
            error={errors.storeSlug}
            hint={`Your store link: /store/${form.storeSlug || "your-store"}`}
            required
          />
          <div className="flex flex-col gap-1.5 sm:col-span-2">
            <Label htmlFor="store-description">Store description</Label>
            <textarea
              id="store-description"
              value={form.description}
              onChange={(e) => set("description", e.target.value)}
              rows={4}
              aria-invalid={errors.description ? true : undefined}
              className="min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none transition-colors placeholder:text-muted-text focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
              placeholder="Tell customers what you sell."
            />
            <FieldError message={errors.description} />
          </div>
          <div className="flex flex-col gap-1.5 sm:col-span-2">
            <Label htmlFor="store-category">Store category</Label>
            <Select value={form.category} onValueChange={(v) => set("category", v)}>
              <SelectTrigger id="store-category" className="w-full" aria-label="Store category">
                <SelectValue placeholder="Choose a category" />
              </SelectTrigger>
              <SelectContent>
                {categories.map((category) => (
                  <SelectItem key={category.slug} value={category.slug}>
                    {category.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FieldError message={errors.category} />
          </div>
          <ImageUpload
            label="Store logo"
            kind="logo"
            url={form.logoUrl}
            uploading={uploading === "logo"}
            onPick={(file) => upload("logo", file)}
            onRemove={() => set("logoUrl", "")}
          />
          <ImageUpload
            label="Store banner"
            kind="banner"
            url={form.bannerUrl}
            uploading={uploading === "banner"}
            onPick={(file) => upload("banner", file)}
            onRemove={() => set("bannerUrl", "")}
          />
        </div>
      </section>

      {/* B. Contact information */}
      <section className={SECTION_CLASS}>
        <h2 className="text-lg font-semibold">Contact information</h2>
        <div className="mt-5 grid gap-4 sm:grid-cols-2">
          <TextField
            label="Business email"
            name="businessEmail"
            type="email"
            value={form.businessEmail}
            onChange={(e) => set("businessEmail", e.target.value)}
            error={errors.businessEmail}
            required
          />
          <TextField
            label="Contact phone"
            name="contactPhone"
            value={form.contactPhone}
            onChange={(e) => set("contactPhone", e.target.value)}
            error={errors.contactPhone}
            placeholder="+8801XXXXXXXXX"
            required
          />
          <TextField
            label="Customer support phone"
            name="supportPhone"
            value={form.supportPhone}
            onChange={(e) => set("supportPhone", e.target.value)}
            hint="Optional"
          />
          <TextField
            label="Website"
            name="websiteUrl"
            value={form.websiteUrl}
            onChange={(e) => set("websiteUrl", e.target.value)}
            error={errors.websiteUrl}
            placeholder="https://"
            hint="Optional"
          />
        </div>
      </section>

      {/* C. Business address */}
      <section className={SECTION_CLASS}>
        <h2 className="text-lg font-semibold">Business address</h2>
        <div className="mt-5 grid gap-4 sm:grid-cols-2">
          <TextField
            label="Country"
            name="country"
            value={form.country}
            onChange={(e) => set("country", e.target.value)}
            required
          />
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="store-division">Division / state</Label>
            <Select
              value={form.division}
              onValueChange={(v) => {
                set("division", v);
                set("district", "");
              }}
            >
              <SelectTrigger id="store-division" className="w-full" aria-label="Division">
                <SelectValue placeholder="Choose a division" />
              </SelectTrigger>
              <SelectContent>
                {DIVISIONS.map((division) => (
                  <SelectItem key={division.name} value={division.name}>
                    {division.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FieldError message={errors.division} />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="store-district">District / city</Label>
            <Select
              value={form.district}
              onValueChange={(v) => set("district", v)}
              disabled={!form.division}
            >
              <SelectTrigger id="store-district" className="w-full" aria-label="District">
                <SelectValue placeholder={form.division ? "Choose a district" : "Select a division first"} />
              </SelectTrigger>
              <SelectContent>
                {districts.map((district) => (
                  <SelectItem key={district} value={district}>
                    {district}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FieldError message={errors.district} />
          </div>
          <TextField
            label="Postal code"
            name="postalCode"
            value={form.postalCode}
            onChange={(e) => set("postalCode", e.target.value)}
            hint="Optional"
          />
          <div className="flex flex-col gap-1.5 sm:col-span-2">
            <Label htmlFor="store-address">Full address</Label>
            <textarea
              id="store-address"
              value={form.fullAddress}
              onChange={(e) => set("fullAddress", e.target.value)}
              rows={3}
              aria-invalid={errors.fullAddress ? true : undefined}
              className="min-h-20 w-full rounded-lg border border-input bg-transparent px-2.5 py-2 text-sm outline-none transition-colors placeholder:text-muted-text focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20"
              placeholder="Street, area, landmark"
            />
            <FieldError message={errors.fullAddress} />
          </div>
        </div>
      </section>

      {/* D. Business & verification */}
      <section className={SECTION_CLASS}>
        <h2 className="text-lg font-semibold">Business & verification</h2>
        <div className="mt-5 grid gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="business-type">Business type</Label>
            <Select
              value={form.businessType}
              onValueChange={(v) => set("businessType", v as FormState["businessType"])}
            >
              <SelectTrigger id="business-type" className="w-full" aria-label="Business type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="individual">Individual</SelectItem>
                <SelectItem value="registered">Registered business</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {form.businessType === "registered" && (
            <TextField
              label="Registered business name"
              name="registeredBusinessName"
              value={form.registeredBusinessName}
              onChange={(e) => set("registeredBusinessName", e.target.value)}
              error={errors.registeredBusinessName}
              required
            />
          )}
          <TextField
            label="Trade license number"
            name="tradeLicenseNumber"
            value={form.tradeLicenseNumber}
            onChange={(e) => set("tradeLicenseNumber", e.target.value)}
            hint="Optional"
          />
        </div>
        <p className="mt-4 text-xs text-muted-foreground">
          Verification documents are collected later, once private storage is
          available. We only ask for what we need right now.
        </p>
      </section>

      {/* E. Confirmation */}
      <section className={SECTION_CLASS}>
        <h2 className="text-lg font-semibold">Review & submit</h2>
        <dl className="mt-4 grid gap-x-6 gap-y-2 text-sm sm:grid-cols-2">
          <Summary label="Store" value={form.storeName} />
          <Summary label="Category" value={categories.find((c) => c.slug === form.category)?.name ?? "—"} />
          <Summary label="Business email" value={form.businessEmail} />
          <Summary label="Phone" value={form.contactPhone} />
          <Summary
            label="Location"
            value={[form.district, form.division, form.country].filter(Boolean).join(", ")}
          />
          <Summary label="Business type" value={form.businessType === "registered" ? "Registered" : "Individual"} />
        </dl>

        <label className="mt-5 flex items-start gap-3 text-sm">
          <input
            type="checkbox"
            checked={agreed}
            onChange={(e) => setAgreed(e.target.checked)}
            className="mt-0.5 size-4 rounded border-input"
          />
          <span>
            I confirm the information above is accurate and agree to the platform&apos;s
            vendor terms and store policies.
          </span>
        </label>

        <div className="mt-6">
          <Button
            type="submit"
            size="lg"
            className="h-11 w-full text-base sm:w-auto"
            disabled={submitting}
          >
            {submitting ? (
              <Loader2 className="size-4 animate-spin" aria-hidden="true" />
            ) : (
              <Check className="size-4" aria-hidden="true" />
            )}
            Submit application
          </Button>
        </div>
      </section>
    </form>
  );
}

function FieldError({ message }: { message?: string }) {
  if (!message) return null;
  return (
    <p role="alert" className="text-xs font-medium text-destructive">
      {message}
    </p>
  );
}

function Summary({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="break-words">{value || "—"}</dd>
    </div>
  );
}

function ImageUpload({
  label,
  kind,
  url,
  uploading,
  onPick,
  onRemove,
}: {
  label: string;
  kind: "logo" | "banner";
  url: string;
  uploading: boolean;
  onPick: (file: File) => void;
  onRemove: () => void;
}) {
  const inputId = `store-${kind}`;
  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={inputId}>{label}</Label>
      {url ? (
        <div className="flex items-center gap-3">
          <Image
            src={url}
            alt={`${label} preview`}
            width={64}
            height={kind === "banner" ? 40 : 64}
            className={cn("rounded-lg object-cover ring-1 ring-foreground/10", kind === "banner" ? "h-10 w-24" : "size-16")}
          />
          <Button type="button" variant="ghost" size="sm" onClick={onRemove}>
            <X className="size-3.5" aria-hidden="true" />
            Remove
          </Button>
        </div>
      ) : (
        <label
          htmlFor={inputId}
          className="flex cursor-pointer items-center gap-2 rounded-lg border border-dashed border-input px-3 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-muted"
        >
          {uploading ? (
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          ) : kind === "logo" ? (
            <ImageIcon className="size-4" aria-hidden="true" />
          ) : (
            <Upload className="size-4" aria-hidden="true" />
          )}
          {uploading ? "Uploading…" : "Upload (optional)"}
        </label>
      )}
      <input
        id={inputId}
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif"
        className="sr-only"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) onPick(file);
          e.target.value = "";
        }}
      />
    </div>
  );
}
