import { USERS } from "@/data/users";
import type { User } from "@/types/user";

/**
 * Customers repository — demo data now, Go API later (PRD §44).
 *
 * The authenticated session lives in `src/lib/session.ts` (Better Auth);
 * this module only feeds the read-only admin customer views.
 */

/** All customers (admin customers page). */
export async function getCustomers(): Promise<User[]> {
  return USERS;
}

export async function getCustomerById(id: string): Promise<User | null> {
  return USERS.find((user) => user.id === id) ?? null;
}
