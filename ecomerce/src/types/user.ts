export interface Address {
  id: string;
  fullName: string;
  phone?: string;
  /* Bangladesh administrative structure (new addresses). */
  division?: string;
  district?: string;
  upazila?: string;
  villageTown?: string;
  /** Free-text clarification, e.g. a landmark (optional). */
  clarification?: string;
  /* Legacy structured lines — kept for earlier demo data / other consumers. */
  line1?: string;
  line2?: string;
  city?: string;
  state?: string;
  postalCode?: string;
  country?: string;
  isDefault: boolean;
}

/** Roles, ordered by privilege (PRD §5). */
export type UserRole = "user" | "vendor" | "admin";

export interface User {
  id: string;
  name: string;
  email: string;
  phone?: string;
  avatarUrl?: string;
  addresses: Address[];
  /** ISO date string. */
  memberSince: string;
  /**
   * Effective role. Use only to gate UI; the backend is authoritative for
   * authorization.
   */
  role?: UserRole;
}
