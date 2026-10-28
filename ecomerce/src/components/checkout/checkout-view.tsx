"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { EmptyCart } from "@/components/cart/empty-cart";
import { CheckoutSummary } from "@/components/checkout/checkout-summary";
import { RadioCard } from "@/components/forms/radio-card";
import { TextField } from "@/components/forms/text-field";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useCartLines } from "@/hooks/use-cart-lines";
import {
  EXPRESS_SHIPPING_RATE,
  FREE_SHIPPING_THRESHOLD,
  STANDARD_SHIPPING_RATE,
} from "@/lib/constants";
import { formatPrice } from "@/lib/formatters";
import { apiPost } from "@/lib/api-client";
import {
  selectCartItemCount,
  selectCartSubtotal,
  useCartStore,
} from "@/store/cart-store";
import { toast } from "sonner";

const checkoutSchema = z.object({
  email: z.email("Enter a valid email address"),
  phone: z.string().optional(),
  fullName: z.string().trim().min(2, "Enter the recipient's full name"),
  address1: z.string().trim().min(4, "Enter a street address"),
  address2: z.string().optional(),
  city: z.string().trim().min(2, "Enter a city"),
  state: z.string().trim().min(2, "Enter a state or region"),
  postalCode: z.string().trim().min(3, "Enter a postal code"),
  country: z.string().trim().min(2, "Enter a country"),
  deliveryMethod: z.enum(["standard", "express"]),
  paymentMethod: z.enum(["card", "cod"]),
});

type CheckoutFormValues = z.infer<typeof checkoutSchema>;

interface CreateOrderResult {
  order: { id: string; order_number: string };
  checkout_url?: string;
}

/**
 * Checkout (PRD §18). Contact, shipping address, delivery and payment sections
 * with validated fields. The order is placed server-side: a card order returns
 * a Stripe Checkout URL to redirect to, a cash-on-delivery order goes straight
 * to the confirmation page.
 */
export function CheckoutView() {
  const router = useRouter();
  const itemCount = useCartStore(selectCartItemCount);
  const subtotal = useCartStore(selectCartSubtotal);
  const resetCart = useCartStore((state) => state.reset);
  const status = useCartStore((state) => state.status);
  const lines = useCartLines();
  const [isPlacing, setIsPlacing] = useState(false);

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<CheckoutFormValues>({
    resolver: zodResolver(checkoutSchema),
    mode: "onTouched",
    defaultValues: {
      country: "United States",
      deliveryMethod: "standard",
      // Default to card so "Place Order" goes to Stripe Checkout; customers who
      // want cash on delivery select it explicitly.
      paymentMethod: "card",
    },
  });

  const deliveryMethod = useWatch({ control, name: "deliveryMethod" });
  const paymentMethod = useWatch({ control, name: "paymentMethod" });
  const deliveryCost =
    deliveryMethod === "express"
      ? EXPRESS_SHIPPING_RATE
      : subtotal >= FREE_SHIPPING_THRESHOLD
        ? 0
        : STANDARD_SHIPPING_RATE;
  const deliveryLabel =
    deliveryMethod === "express"
      ? "Express (2–3 business days)"
      : "Standard (5–7 business days)";
  const total = subtotal + deliveryCost;

  const onSubmit = handleSubmit(async (values) => {
    if (isPlacing || itemCount === 0) return;
    setIsPlacing(true);

    // Line items are omitted on purpose: the server reads the caller's cart
    // (the same cart the guest/user sees) and converts it into the order.
    const res = await apiPost<CreateOrderResult>("/api/v1/orders", {
      email: values.email,
      shipping_address: {
        full_name: values.fullName,
        phone: values.phone ?? "",
        line1: values.address1,
        line2: values.address2 ?? "",
        city: values.city,
        state: values.state,
        postal_code: values.postalCode,
        country: values.country,
      },
      delivery_method: values.deliveryMethod,
      payment_method: values.paymentMethod,
    });

    if (!res.ok) {
      setIsPlacing(false);
      toast.error(res.error.message);
      return;
    }

    // The server already converted the cart; drop the local copy.
    resetCart();
    if (res.data.checkout_url) {
      // Redirect to the provider's hosted checkout.
      window.location.assign(res.data.checkout_url);
      return;
    }
    router.push(`/order/success?order=${res.data.order.id}`);
  });

  if (itemCount === 0 && (status === "idle" || status === "loading")) {
    return (
      <div
        aria-hidden="true"
        className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8"
      >
        <Skeleton className="h-9 w-40" />
        <div className="mt-8 grid gap-10 lg:grid-cols-[1fr_380px]">
          <Skeleton className="h-96 rounded-2xl" />
          <Skeleton className="h-72 rounded-2xl" />
        </div>
      </div>
    );
  }

  if (itemCount === 0) {
    return (
      <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
        <header className="mb-8">
          <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
            Checkout
          </h1>
        </header>
        <EmptyCart />
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          Checkout
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Complete your details and pay securely. Card payments are processed by
          Stripe.
        </p>
      </header>

      <form
        onSubmit={onSubmit}
        noValidate
        className="grid items-start gap-10 lg:grid-cols-[1fr_380px]"
      >
        <div className="flex flex-col gap-6">
          <section
            aria-labelledby="contact-heading"
            className="rounded-2xl border p-6"
          >
            <h2 id="contact-heading" className="text-lg font-semibold">
              Contact
            </h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              <TextField
                label="Email address"
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                error={errors.email?.message}
                {...register("email")}
              />
              <TextField
                label="Phone (optional)"
                type="tel"
                placeholder="+1 555 000 0000"
                autoComplete="tel"
                error={errors.phone?.message}
                {...register("phone")}
              />
            </div>
          </section>

          <section
            aria-labelledby="address-heading"
            className="rounded-2xl border p-6"
          >
            <h2 id="address-heading" className="text-lg font-semibold">
              Shipping Address
            </h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <TextField
                  label="Full name"
                  autoComplete="name"
                  error={errors.fullName?.message}
                  {...register("fullName")}
                />
              </div>
              <div className="sm:col-span-2">
                <TextField
                  label="Street address"
                  autoComplete="address-line1"
                  placeholder="Street and number"
                  error={errors.address1?.message}
                  {...register("address1")}
                />
              </div>
              <div className="sm:col-span-2">
                <TextField
                  label="Apartment, suite, etc. (optional)"
                  autoComplete="address-line2"
                  error={errors.address2?.message}
                  {...register("address2")}
                />
              </div>
              <TextField
                label="City"
                autoComplete="address-level2"
                error={errors.city?.message}
                {...register("city")}
              />
              <TextField
                label="State / Region"
                autoComplete="address-level1"
                error={errors.state?.message}
                {...register("state")}
              />
              <TextField
                label="Postal code"
                autoComplete="postal-code"
                error={errors.postalCode?.message}
                {...register("postalCode")}
              />
              <TextField
                label="Country"
                autoComplete="country-name"
                error={errors.country?.message}
                {...register("country")}
              />
            </div>
          </section>

          <section
            aria-labelledby="delivery-heading"
            className="rounded-2xl border p-6"
          >
            <h2 id="delivery-heading" className="text-lg font-semibold">
              Delivery Method
            </h2>
            <div className="mt-4 flex flex-col gap-3">
              <RadioCard
                name="deliveryMethod"
                value="standard"
                label="Standard Delivery"
                description="5–7 business days"
                trailing={
                  subtotal >= FREE_SHIPPING_THRESHOLD
                    ? "Free"
                    : formatPrice(STANDARD_SHIPPING_RATE)
                }
                checked={deliveryMethod === "standard"}
                registration={register("deliveryMethod")}
              />
              <RadioCard
                name="deliveryMethod"
                value="express"
                label="Express Delivery"
                description="2–3 business days"
                trailing={formatPrice(EXPRESS_SHIPPING_RATE)}
                checked={deliveryMethod === "express"}
                registration={register("deliveryMethod")}
              />
            </div>
          </section>

          <section
            aria-labelledby="payment-heading"
            className="rounded-2xl border p-6"
          >
            <h2 id="payment-heading" className="text-lg font-semibold">
              Payment Method
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">
              Card payments are handled by Stripe Checkout — no card details
              touch this site.
            </p>
            <div className="mt-4 flex flex-col gap-3">
              <RadioCard
                name="paymentMethod"
                value="card"
                label="Card payment"
                description="Pay securely with Stripe. You'll be redirected to Stripe Checkout."
                checked={paymentMethod === "card"}
                registration={register("paymentMethod")}
              />
              <RadioCard
                name="paymentMethod"
                value="cod"
                label="Cash on Delivery"
                description="Pay in cash when your order arrives."
                checked={paymentMethod === "cod"}
                registration={register("paymentMethod")}
              />
            </div>
          </section>
        </div>

        <aside>
          <div className="rounded-2xl border bg-secondary p-6 lg:sticky lg:top-24">
            <h2 className="text-lg font-semibold">Order Summary</h2>
            <div className="mt-4">
              <CheckoutSummary
                lines={lines}
                subtotal={subtotal}
                deliveryLabel={deliveryLabel}
                deliveryCost={deliveryCost}
              />
            </div>
            <Button
              type="submit"
              variant="default"
              size="lg"
              disabled={isPlacing}
              className="mt-5 h-11 w-full text-base"
            >
              {isPlacing && (
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
              )}
              {isPlacing
                ? "Placing order…"
                : `Place Order · ${formatPrice(total)}`}
            </Button>
            <p className="mt-3 text-center text-xs text-muted-foreground">
              By placing this order you agree to the store terms. Card payments
              are processed securely by Stripe.
            </p>
          </div>
        </aside>
      </form>
    </div>
  );
}
