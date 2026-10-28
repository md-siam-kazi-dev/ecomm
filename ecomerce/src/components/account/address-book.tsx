"use client";

import { useCallback, useEffect, useState } from "react";
import { AlertTriangle, MapPin, Pencil, Plus, Star, Trash2 } from "lucide-react";
import { Controller, useForm, type Control } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { z } from "zod";

import { TextField } from "@/components/forms/text-field";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { DIVISIONS, districtsOf } from "@/data/bd-locations";
import {
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type ApiResult,
} from "@/lib/api-client";
import { cn } from "@/lib/utils";
import type { Address } from "@/types/user";

const ADDRESSES_PATH = "/api/v1/account/addresses";

/** Shape returned by the Go API. */
interface AddressApi {
  id: string;
  full_name: string;
  phone: string;
  division: string;
  district: string;
  upazila: string;
  village_town: string;
  clarification: string;
  is_default: boolean;
}

function toAddress(a: AddressApi): Address {
  return {
    id: a.id,
    fullName: a.full_name,
    phone: a.phone,
    division: a.division,
    district: a.district,
    upazila: a.upazila,
    villageTown: a.village_town,
    clarification: a.clarification,
    isDefault: a.is_default,
  };
}

const addressSchema = z.object({
  fullName: z.string().trim().min(2, "Enter the recipient's full name"),
  phone: z.string().trim().min(6, "Enter a valid phone number"),
  division: z.string().min(1, "Select a division"),
  district: z.string().min(1, "Select a district"),
  upazila: z.string().trim().min(2, "Enter an upazila / thana"),
  villageTown: z.string().trim().min(2, "Enter a village or town"),
  clarification: z.string().trim().max(500).optional(),
  isDefault: z.boolean(),
});

type AddressFormValues = z.infer<typeof addressSchema>;

const EMPTY_DEFAULTS: AddressFormValues = {
  fullName: "",
  phone: "",
  division: "",
  district: "",
  upazila: "",
  villageTown: "",
  clarification: "",
  isDefault: false,
};

/**
 * Address book (PRD §23): a user can save many addresses. Backed by the Go API
 * (`/api/v1/account/addresses`); the first address becomes the default.
 */
export function AddressBook() {
  const [addresses, setAddresses] = useState<Address[]>([]);
  const [status, setStatus] = useState<"loading" | "ready" | "error">("loading");
  const [error, setError] = useState("");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  // Mirrors the division field so the district options can react to it without
  // reading back from react-hook-form's watch().
  const [division, setDivision] = useState("");

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    control,
    formState: { errors, isSubmitting },
  } = useForm<AddressFormValues>({
    resolver: zodResolver(addressSchema),
    mode: "onTouched",
    defaultValues: EMPTY_DEFAULTS,
  });

  const load = useCallback(() => apiGet<AddressApi[]>(ADDRESSES_PATH), []);

  const apply = useCallback((res: ApiResult<AddressApi[]>) => {
    if (res.ok) {
      setAddresses(res.data.map(toAddress));
      setStatus("ready");
    } else {
      setError(res.error.message);
      setStatus("error");
    }
  }, []);

  useEffect(() => {
    let active = true;
    void load().then((res) => {
      if (active) apply(res);
    });
    return () => {
      active = false;
    };
  }, [load, apply]);

  function openCreate() {
    setEditingId(null);
    setDivision("");
    reset({ ...EMPTY_DEFAULTS, isDefault: addresses.length === 0 });
    setIsDialogOpen(true);
  }

  function openEdit(address: Address) {
    setEditingId(address.id);
    setDivision(address.division ?? "");
    reset({
      fullName: address.fullName,
      phone: address.phone ?? "",
      division: address.division ?? "",
      district: address.district ?? "",
      upazila: address.upazila ?? "",
      villageTown: address.villageTown ?? "",
      clarification: address.clarification ?? "",
      isDefault: address.isDefault,
    });
    setIsDialogOpen(true);
  }

  const onSubmit = handleSubmit(async (values) => {
    const payload = {
      full_name: values.fullName,
      phone: values.phone,
      division: values.division,
      district: values.district,
      upazila: values.upazila,
      village_town: values.villageTown,
      clarification: values.clarification ?? "",
      is_default: values.isDefault,
    };

    const res = editingId
      ? await apiPut(`${ADDRESSES_PATH}/${editingId}`, payload)
      : await apiPost(ADDRESSES_PATH, payload);

    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }

    toast.success(editingId ? "Address updated" : "Address saved");
    setIsDialogOpen(false);
    apply(await load());
  });

  async function removeAddress(addressId: string) {
    const res = await apiDelete(`${ADDRESSES_PATH}/${addressId}`);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Address removed");
    apply(await load());
  }

  async function setDefault(addressId: string) {
    const res = await apiPost(`${ADDRESSES_PATH}/${addressId}/default`);
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Default address updated");
    apply(await load());
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between gap-4">
        <p className="text-sm text-muted-foreground">
          {status === "ready"
            ? `${addresses.length} saved ${addresses.length === 1 ? "address" : "addresses"}`
            : "\u00a0"}
        </p>
        <Button onClick={openCreate}>
          <Plus className="size-4" aria-hidden="true" />
          Add address
        </Button>
      </div>

      {status === "loading" ? (
        <div className="grid gap-4 sm:grid-cols-2" role="status" aria-label="Loading addresses">
          {Array.from({ length: 2 }).map((_, index) => (
            <Skeleton key={index} className="h-44 rounded-2xl" />
          ))}
        </div>
      ) : status === "error" ? (
        <div
          role="alert"
          className="flex flex-col items-center gap-3 rounded-xl border border-dashed px-6 py-12 text-center"
        >
          <AlertTriangle className="size-6 text-destructive" aria-hidden="true" />
          <p className="text-sm text-muted-foreground">{error}</p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setStatus("loading");
              void load().then(apply);
            }}
          >
            Try again
          </Button>
        </div>
      ) : addresses.length === 0 ? (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed px-6 py-16 text-center">
          <MapPin className="size-7 text-muted-foreground" aria-hidden="true" />
          <p className="text-sm text-muted-foreground">
            You have no saved addresses yet.
          </p>
        </div>
      ) : (
        <ul className="grid gap-4 sm:grid-cols-2">
          {addresses.map((address) => (
            <li
              key={address.id}
              className={cn(
                "flex flex-col rounded-2xl border p-5",
                address.isDefault && "border-primary/40 bg-primary/5"
              )}
            >
              <div className="flex items-start justify-between gap-3">
                <span className="font-medium">{address.fullName}</span>
                {address.isDefault && (
                  <Badge
                    variant="outline"
                    className="border-transparent bg-primary/10 text-primary-dark"
                  >
                    Default
                  </Badge>
                )}
              </div>
              <address className="mt-2 flex-1 text-sm not-italic leading-relaxed text-muted-foreground">
                {address.villageTown}
                {address.upazila && (
                  <>
                    <br />
                    {address.upazila} Upazila
                  </>
                )}
                <br />
                {address.district}, {address.division}
                <br />
                Bangladesh
                {address.clarification && (
                  <>
                    <br />
                    <span className="text-xs">{address.clarification}</span>
                  </>
                )}
                {address.phone && (
                  <>
                    <br />
                    {address.phone}
                  </>
                )}
              </address>

              <div className="mt-4 flex flex-wrap gap-2 border-t pt-3">
                {!address.isDefault && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setDefault(address.id)}
                    aria-label={`Set ${address.fullName}'s address as default`}
                  >
                    <Star className="size-3.5" aria-hidden="true" />
                    Set as default
                  </Button>
                )}
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => openEdit(address)}
                  aria-label={`Edit ${address.fullName}'s address`}
                >
                  <Pencil className="size-3.5" aria-hidden="true" />
                  Edit
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => removeAddress(address.id)}
                  aria-label={`Remove ${address.fullName}'s address`}
                  className="text-destructive hover:text-destructive"
                >
                  <Trash2 className="size-3.5" aria-hidden="true" />
                  Remove
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingId ? "Edit address" : "Add a new address"}</DialogTitle>
            <DialogDescription>
              Saved to your account. The first address becomes your default.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={onSubmit} noValidate className="grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <TextField
                label="Recipient name"
                autoComplete="name"
                error={errors.fullName?.message}
                {...register("fullName")}
              />
            </div>
            <div className="sm:col-span-2">
              <TextField
                label="Phone"
                type="tel"
                autoComplete="tel"
                error={errors.phone?.message}
                {...register("phone")}
              />
            </div>

            <SelectField
              label="Division"
              control={control}
              name="division"
              placeholder="Select division"
              options={DIVISIONS.map((d) => ({ value: d.name, label: d.name }))}
              error={errors.division?.message}
              onChange={(value) => {
                setDivision(value);
                setValue("district", "");
              }}
            />
            <SelectField
              label="District"
              control={control}
              name="district"
              placeholder={division ? "Select district" : "Select division first"}
              options={districtsOf(division).map((d) => ({ value: d, label: d }))}
              error={errors.district?.message}
              disabled={!division}
            />

            <TextField
              label="Upazila / Thana"
              error={errors.upazila?.message}
              {...register("upazila")}
            />
            <TextField
              label="Village / Town"
              error={errors.villageTown?.message}
              {...register("villageTown")}
            />

            <div className="sm:col-span-2">
              <TextField
                label="Address clarification (optional)"
                placeholder="Landmark, house, road…"
                error={errors.clarification?.message}
                {...register("clarification")}
              />
            </div>

            <label className="flex items-center gap-2 text-sm sm:col-span-2">
              <input
                type="checkbox"
                className="size-4 rounded border-input accent-primary"
                {...register("isDefault")}
              />
              Use as my default address
            </label>

            <DialogFooter className="sm:col-span-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsDialogOpen(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {editingId ? "Save changes" : "Save address"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

interface SelectFieldProps {
  label: string;
  control: Control<AddressFormValues>;
  name: "division" | "district";
  placeholder: string;
  options: { value: string; label: string }[];
  error?: string;
  disabled?: boolean;
  onChange?: (value: string) => void;
}

function SelectField({
  label,
  control,
  name,
  placeholder,
  options,
  error,
  disabled,
  onChange,
}: SelectFieldProps) {
  const fieldId = `address-${name}`;

  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={fieldId}>{label}</Label>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <Select
            value={field.value}
            disabled={disabled}
            onValueChange={(value) => {
              field.onChange(value);
              onChange?.(value);
            }}
          >
            <SelectTrigger id={fieldId} className="w-full" aria-invalid={error ? true : undefined}>
              <SelectValue placeholder={placeholder} />
            </SelectTrigger>
            <SelectContent>
              {options.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      />
      {error && (
        <p role="alert" className="text-xs font-medium text-destructive">
          {error}
        </p>
      )}
    </div>
  );
}
