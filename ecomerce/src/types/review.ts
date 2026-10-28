export interface Review {
  id: string;
  productId: string;
  author: string;
  /** The reviewer's profile picture, when available. */
  avatarUrl?: string;
  /** Rating from 1 to 5. */
  rating: number;
  title: string;
  comment: string;
  /** ISO date string. */
  createdAt: string;
  isVerifiedPurchase: boolean;
  helpfulCount: number;
  /** Optional photos attached to the review. */
  images?: string[];
  /** Product context, when the review is shown outside the product page. */
  productName?: string;
  productSlug?: string;
  /** The vendor's public reply, if any. */
  vendorReply?: string;
  /** ISO date string of the vendor reply. */
  vendorRepliedAt?: string;
}
