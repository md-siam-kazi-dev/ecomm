"use client";

import { useSyncExternalStore } from "react";

const emptySubscribe = () => () => {};

/**
 * Returns `true` after the component has mounted on the client. Use it to
 * gate persisted Zustand values (cart/wishlist counts) so server-rendered
 * HTML and the first client render match (avoids hydration mismatches).
 *
 * Implemented with `useSyncExternalStore` server/client snapshots — React
 * renders the server snapshot during hydration, then switches to the client
 * snapshot, without effects or manual state.
 */
export function useHasMounted(): boolean {
  return useSyncExternalStore(
    emptySubscribe,
    () => true, // client snapshot
    () => false // server snapshot
  );
}

