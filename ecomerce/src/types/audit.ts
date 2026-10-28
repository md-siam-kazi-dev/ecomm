/** Admin audit-trail types (mirror the Go audit API). */

export type AuditCategory = "security" | "admin" | "vendor" | "system";

export interface AuditActor {
  id: string | null;
  name: string;
  email: string | null;
}

export interface AuditEvent {
  id: string;
  created_at: string;
  action: string;
  action_label: string;
  category: AuditCategory;
  actor: AuditActor;
  actor_role: string | null;
  entity_type: string | null;
  entity_id: string | null;
  result: string | null;
  reason: string | null;
}

export interface AuditEventDetail extends AuditEvent {
  old_values: Record<string, unknown> | null;
  new_values: Record<string, unknown> | null;
  metadata: Record<string, unknown> | null;
  ip_address: string | null;
  user_agent: string | null;
  request_id: string | null;
}

export interface AuditActionOption {
  value: string;
  label: string;
}

export interface AuditSummary {
  total: number;
  admin_actions: number;
  vendor_actions: number;
  security_events: number;
  from?: string;
  to?: string;
  actions: AuditActionOption[];
}
