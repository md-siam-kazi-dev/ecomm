import { create } from "zustand";
import { persist } from "zustand/middleware";

const MAX_RECENT_SEARCHES = 5;

interface SearchState {
  recent: string[];
  /** Records a search query — deduped case-insensitively, capped, trimmed. */
  addRecentSearch: (query: string) => void;
  clearRecentSearches: () => void;
}

/** Recent search queries (PRD §20) — persisted locally, nothing sensitive. */
export const useSearchStore = create<SearchState>()(
  persist(
    (set) => ({
      recent: [],
      addRecentSearch: (query) =>
        set((state) => {
          const trimmed = query.trim();
          if (!trimmed) return state;

          const next = [
            trimmed,
            ...state.recent.filter(
              (item) => item.toLowerCase() !== trimmed.toLowerCase()
            ),
          ];

          return { recent: next.slice(0, MAX_RECENT_SEARCHES) };
        }),
      clearRecentSearches: () => set({ recent: [] }),
    }),
    {
      name: "shopkart-search",
      version: 1,
      partialize: (state) => ({ recent: state.recent }),
    }
  )
);
