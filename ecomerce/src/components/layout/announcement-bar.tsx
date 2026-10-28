import { Truck } from "lucide-react";

import { ANNOUNCEMENT_MESSAGE } from "@/lib/constants";
import { getSiteSettings } from "@/lib/settings";

/**
 * Server-rendered announcement bar above the header. The message is
 * admin-editable (see /admin/settings); it falls back to the default constant
 * when the API is unreachable and hides entirely when disabled. Zero JavaScript.
 */
export async function AnnouncementBar() {
  const settings = await getSiteSettings();

  if (settings && !settings.announcement_enabled) {
    return null;
  }

  const message = settings?.announcement_message.trim() || ANNOUNCEMENT_MESSAGE;

  return (
    <div className="bg-primary text-primary-foreground">
      <div className="mx-auto flex w-full max-w-7xl items-center justify-center gap-2 px-4 py-2 sm:px-6 lg:px-8">
        <Truck className="size-4 shrink-0" aria-hidden="true" />
        <p className="text-center text-xs font-medium tracking-wide sm:text-sm">
          {message}
        </p>
      </div>
    </div>
  );
}
