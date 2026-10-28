"use client";

import { ThemeProvider as NextThemesProvider } from "next-themes";
import type { ComponentProps } from "react";

/**
 * Applies the theme by toggling a `.dark` class on <html> (matching the
 * `@custom-variant dark` in globals.css). Defaults to the OS preference and
 * persists the user's choice.
 */
export function ThemeProvider(props: ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props} />;
}
