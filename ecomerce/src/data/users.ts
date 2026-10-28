import type { User } from "@/types/user";

/**
 * Demo customers (PRD §47 — 5 customers). These back the read-only admin
 * customer views and the interim demo session in `src/lib/session.ts` while
 * authentication moves to the Go backend.
 */
export const USERS: User[] = [
  {
    id: "u-001",
    name: "Amelia Hart",
    email: "amelia.hart@example.com",
    phone: "+1 555 010 42",
    addresses: [
      {
        id: "addr-1",
        fullName: "Amelia Hart",
        line1: "418 Maple Grove Avenue",
        line2: "Apt 6B",
        city: "Portland",
        state: "OR",
        postalCode: "97205",
        country: "United States",
        phone: "+1 555 010 42",
        isDefault: true,
      },
      {
        id: "addr-2",
        fullName: "Amelia Hart",
        line1: "92 Harbor Point Road",
        city: "Portland",
        state: "OR",
        postalCode: "97210",
        country: "United States",
        isDefault: false,
      },
    ],
    memberSince: "2025-11-03",
  },
  {
    id: "u-002",
    name: "Daniel Osei",
    email: "daniel.osei@example.com",
    addresses: [
      {
        id: "addr-3",
        fullName: "Daniel Osei",
        line1: "77 Birchwood Court",
        city: "Austin",
        state: "TX",
        postalCode: "78701",
        country: "United States",
        isDefault: true,
      },
    ],
    memberSince: "2026-01-22",
  },
  {
    id: "u-003",
    name: "Priya Raman",
    email: "priya.raman@example.com",
    phone: "+1 555 017 88",
    addresses: [
      {
        id: "addr-4",
        fullName: "Priya Raman",
        line1: "1210 Sunset Terrace",
        line2: "Unit 12",
        city: "Seattle",
        state: "WA",
        postalCode: "98101",
        country: "United States",
        phone: "+1 555 017 88",
        isDefault: true,
      },
    ],
    memberSince: "2025-08-15",
  },
  {
    id: "u-004",
    name: "Lucas Meyer",
    email: "lucas.meyer@example.com",
    addresses: [
      {
        id: "addr-5",
        fullName: "Lucas Meyer",
        line1: "309 Foxglove Lane",
        city: "Denver",
        state: "CO",
        postalCode: "80202",
        country: "United States",
        isDefault: true,
      },
    ],
    memberSince: "2026-03-09",
  },
  {
    id: "u-005",
    name: "Sofia Marino",
    email: "sofia.marino@example.com",
    addresses: [
      {
        id: "addr-6",
        fullName: "Sofia Marino",
        line1: "15 Windmill Street",
        line2: "Floor 3",
        city: "Chicago",
        state: "IL",
        postalCode: "60601",
        country: "United States",
        isDefault: true,
      },
      {
        id: "addr-7",
        fullName: "Sofia Marino",
        line1: "801 Lakeshore Drive",
        city: "Chicago",
        state: "IL",
        postalCode: "60611",
        country: "United States",
        isDefault: false,
      },
    ],
    memberSince: "2025-12-30",
  },
];
