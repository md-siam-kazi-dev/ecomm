# Taste

## Frontend / Next.js
- Prefers Next.js performance best practices: use `next/image` everywhere, dynamically import heavy client components, and keep `"use client"` usage minimal (server components by default). Confidence: 0.6
- Insists on real accessibility: semantic headings and labels, visible focus, sufficient contrast, and dialogs that close on Esc with a proper focus trap. Confidence: 0.6
- Wants complete UI states rather than bare content — loading skeletons on async pages, error boundaries, and deliberate empty states. Confidence: 0.6
- Expects client-side search to be debounced, keyboard/ARIA accessible, and driven by a URL query param. Confidence: 0.5
- Favors reusable, validated form components (labels, error messages, loading/disabled states) over ad-hoc inputs. Confidence: 0.5
