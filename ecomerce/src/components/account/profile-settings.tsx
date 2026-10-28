"use client";

import { useRef, useState, type ChangeEvent } from "react";
import { Camera, Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { z } from "zod";

import { Avatar } from "@/components/avatar";
import { TextField } from "@/components/forms/text-field";
import { SignOutButton } from "@/components/auth/sign-out-button";
import { Button } from "@/components/ui/button";
import { apiPut, apiUpload } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/store/auth-store";
import type { User } from "@/types/user";

const profileSchema = z.object({
  name: z.string().trim().min(2, "Enter your full name"),
  phone: z
    .string()
    .trim()
    .refine((value) => value === "" || value.length >= 6, "Enter a valid phone number"),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

const NOTIFICATION_OPTIONS = [
  { id: "order-updates", label: "Order updates", defaultChecked: true },
  { id: "deals", label: "Deals and promotions", defaultChecked: true },
  { id: "product-news", label: "Product news", defaultChecked: false },
];

interface AvatarResponse {
  full_name: string;
  email: string;
  phone: string;
  avatar_url: string;
}

/**
 * Profile + notification preferences (PRD §23). Name, phone and the profile
 * picture are persisted by the backend; the picture is uploaded to imgbb and
 * only its URL is stored.
 */
export function ProfileSettings({ user }: { user: User }) {
  const initialize = useAuthStore((state) => state.initialize);
  const [avatarUrl, setAvatarUrl] = useState(user.avatarUrl ?? "");
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [notifications, setNotifications] = useState<Record<string, boolean>>(
    Object.fromEntries(
      NOTIFICATION_OPTIONS.map((option) => [option.id, option.defaultChecked])
    )
  );

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    mode: "onTouched",
    defaultValues: {
      name: user.name,
      phone: user.phone ?? "",
    },
  });

  const onSubmit = handleSubmit(async (values) => {
    const res = await apiPut<AvatarResponse>("/api/v1/account/profile", {
      full_name: values.name,
      phone: values.phone ?? "",
    });
    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    toast.success("Profile updated");
    // Refresh the header/account name from the server.
    await initialize();
  });

  async function handleAvatarChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    const formData = new FormData();
    formData.append("image", file);

    const res = await apiUpload<AvatarResponse>("/api/v1/account/avatar", formData);
    setIsUploading(false);
    if (fileInputRef.current) fileInputRef.current.value = "";

    if (!res.ok) {
      toast.error(res.error.message);
      return;
    }
    setAvatarUrl(res.data.avatar_url);
    toast.success("Profile picture updated");
    await initialize();
  }

  return (
    <div className="flex flex-col gap-6">
      <form
        onSubmit={onSubmit}
        noValidate
        className="rounded-2xl border p-5 sm:p-6"
        aria-labelledby="profile-heading"
      >
        <h2 id="profile-heading" className="text-base font-semibold">
          Profile details
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Update your name, phone number and profile picture.
        </p>

        {/* Profile picture */}
        <div className="mt-5 flex items-center gap-4">
          <Avatar
            name={user.name}
            src={avatarUrl}
            className="size-20"
            textClassName="text-xl"
          />
          <div>
            <input
              ref={fileInputRef}
              id="avatar-input"
              type="file"
              accept="image/png,image/jpeg,image/webp,image/gif"
              className="sr-only"
              onChange={handleAvatarChange}
            />
            <Button
              type="button"
              variant="outline"
              onClick={() => fileInputRef.current?.click()}
              disabled={isUploading}
            >
              {isUploading ? (
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
              ) : (
                <Camera className="size-4" aria-hidden="true" />
              )}
              {isUploading ? "Uploading…" : "Upload photo"}
            </Button>
            <p className="mt-1 text-xs text-muted-foreground">
              PNG, JPG, WebP or GIF · max 5 MB
            </p>
          </div>
        </div>

        <div className="mt-5 grid gap-4 sm:grid-cols-2">
          <div className="sm:col-span-2">
            <TextField
              label="Full name"
              autoComplete="name"
              error={errors.name?.message}
              {...register("name")}
            />
          </div>
          <TextField
            label="Email address"
            type="email"
            autoComplete="email"
            defaultValue={user.email}
            disabled
            hint="Email cannot be changed here."
          />
          <TextField
            label="Phone (optional)"
            type="tel"
            autoComplete="tel"
            placeholder="+8801XXXXXXXXX"
            error={errors.phone?.message}
            {...register("phone")}
          />
        </div>

        <div className="mt-5">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Saving…" : "Save changes"}
          </Button>
        </div>
      </form>

      <section
        aria-labelledby="notifications-heading"
        className="rounded-2xl border p-5 sm:p-6"
      >
        <h2 id="notifications-heading" className="text-base font-semibold">
          Email preferences
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Choose which emails you would like to receive.
        </p>

        <ul className="mt-4 flex flex-col gap-3">
          {NOTIFICATION_OPTIONS.map((option) => (
            <li key={option.id}>
              <label
                htmlFor={option.id}
                className={cn(
                  "flex cursor-pointer items-center gap-3 rounded-xl border p-3.5 transition-colors hover:bg-muted has-focus-visible:ring-2 has-focus-visible:ring-ring/50"
                )}
              >
                <input
                  id={option.id}
                  type="checkbox"
                  checked={notifications[option.id]}
                  onChange={(event) =>
                    setNotifications((previous) => ({
                      ...previous,
                      [option.id]: event.target.checked,
                    }))
                  }
                  className="size-4 shrink-0 accent-primary"
                />
                <span className="text-sm font-medium">{option.label}</span>
              </label>
            </li>
          ))}
        </ul>
      </section>

      <section
        aria-labelledby="session-heading"
        className="rounded-2xl border p-5 sm:p-6"
      >
        <h2 id="session-heading" className="text-base font-semibold">
          Session
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Sign out of your account on this device.
        </p>
        <div className="mt-4">
          <SignOutButton />
        </div>
      </section>
    </div>
  );
}
