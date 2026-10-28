/** Store application / store / application-fee types (mirror the Go API). */

export type ApplicationStatus = "pending" | "approved" | "rejected";

export type StoreStatus = "pending_activation" | "active" | "suspended" | "closed";

export type PaymentStatus =
  | "pending"
  | "processing"
  | "paid"
  | "failed"
  | "cancelled"
  | "partially_refunded"
  | "refunded";

export type NextAction =
  | "apply"
  | "await_review"
  | "pay_fee"
  | "payment_processing"
  | "open_dashboard"
  | "rejected"
  | "suspended"
  | "none";

export interface ApplicationItem {
  id: string;
  application_number: string;
  applicant_name: string;
  applicant_email: string;
  store_name: string;
  category: string;
  status: ApplicationStatus;
  payment_status: string;
  submitted_at: string;
}

export interface StoreItem {
  id: string;
  name: string;
  slug: string;
  logo_url: string;
  status: StoreStatus;
  role: string;
  activated_at?: string;
}

export interface PaymentItem {
  reference: string;
  amount: string;
  currency: string;
  status: PaymentStatus;
  provider?: string;
  created_at: string;
  verified_at?: string;
}

export interface FeeInfo {
  enabled: boolean;
  amount: string;
  currency: string;
  provider_configured: boolean;
}

export interface ApplicationDetail {
  id: string;
  application_number: string;
  store_name: string;
  store_slug: string;
  description: string;
  category: string;
  logo_url: string;
  banner_url: string;

  business_email: string;
  contact_phone: string;
  support_phone: string;
  website_url: string;

  country: string;
  division: string;
  district: string;
  full_address: string;
  postal_code: string;

  business_type: string;
  registered_business_name: string;
  trade_license_number: string;

  status: ApplicationStatus;
  rejection_reason?: string;
  submitted_at: string;
  reviewed_at?: string;
  approved_at?: string;
  approval_letter_available: boolean;

  store?: StoreItem;
  fee?: FeeInfo;
  payment?: PaymentItem;
  next_action: NextAction;
}

export interface ApplicantAccount {
  id: string;
  email: string;
  name: string;
  status: string;
  created_at: string;
}

export interface HistoryItem {
  action: string;
  actor_name: string;
  created_at: string;
}

export interface AdminApplicationDetail extends ApplicationDetail {
  applicant_user_id: string;
  applicant_account: ApplicantAccount;
  admin_review_notes?: string;
  reviewed_by?: string;
  history: HistoryItem[];
}

export interface PaymentInitResponse {
  payment: PaymentItem;
  provider_configured: boolean;
  redirect_url?: string;
  client_payload?: Record<string, string>;
}

export interface PaymentAdminItem {
  id: string;
  reference: string;
  application_number: string;
  store_name: string;
  applicant_name: string;
  applicant_email: string;
  amount: string;
  currency: string;
  status: PaymentStatus;
  provider?: string;
  provider_transaction_id?: string;
  created_at: string;
  verified_at?: string;
}
