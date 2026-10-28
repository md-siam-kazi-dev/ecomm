"use client";

/**
 * Last-resort error boundary (PRD §54). Replaces the root layout when it
 * itself fails, so it renders its own <html>/<body> with inline styles
 * because the stylesheet may not have loaded.
 */
export default function GlobalError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body
        style={{
          margin: 0,
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontFamily:
            "ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, sans-serif",
          background: "#1A120E",
          color: "#F8FAFC",
          padding: "24px",
        }}
      >
        <div style={{ maxWidth: "440px", textAlign: "center" }}>
          <h1 style={{ fontSize: "1.25rem", fontWeight: 600, margin: 0 }}>
            Something went wrong
          </h1>
          <p
            style={{
              marginTop: "8px",
              fontSize: "0.875rem",
              color: "#94A3B8",
              lineHeight: 1.6,
            }}
          >
            The store failed to load. Please try again — if the problem
            continues, come back a little later.
          </p>
          <button
            type="button"
            onClick={reset}
            style={{
              marginTop: "20px",
              padding: "10px 18px",
              borderRadius: "10px",
              border: "none",
              background: "#3B82F6",
              color: "#1A120E",
              fontSize: "0.875rem",
              fontWeight: 500,
              cursor: "pointer",
            }}
          >
            Try again
          </button>
        </div>
      </body>
    </html>
  );
}
