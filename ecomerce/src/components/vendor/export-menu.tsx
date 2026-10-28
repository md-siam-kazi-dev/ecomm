"use client";

import { Download, FileSpreadsheet, FileText } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

/**
 * Export dropdown. The links are plain anchors to the vendor export endpoints
 * (same-origin, cookies sent) so the file downloads without client JS.
 */
export function ExportMenu({ baseUrl, disabled }: { baseUrl: string; disabled?: boolean }) {
  const separator = baseUrl.includes("?") ? "&" : "?";
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" disabled={disabled}>
          <Download className="size-3.5" aria-hidden="true" />
          Export
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem asChild>
          <a href={`${baseUrl}${separator}format=csv`}>
            <FileText className="size-4" aria-hidden="true" />
            Download CSV
          </a>
        </DropdownMenuItem>
        <DropdownMenuItem asChild>
          <a href={`${baseUrl}${separator}format=xlsx`}>
            <FileSpreadsheet className="size-4" aria-hidden="true" />
            Download Excel
          </a>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
