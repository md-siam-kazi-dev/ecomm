import { serializeJsonLd } from "@/lib/seo";

interface JsonLdProps {
  data: Record<string, unknown>;
}

/**
 * Renders schema.org JSON-LD structured data. Serialization and escaping
 * (`<`, `>`, `&`, U+2028/U+2029) are delegated to `serializeJsonLd` so the
 * escaping rules live in exactly one place (PRD §28).
 */
export function JsonLd({ data }: JsonLdProps) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: serializeJsonLd(data) }}
    />
  );
}
