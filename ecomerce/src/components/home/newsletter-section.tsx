import { NewsletterForm } from "@/components/layout/newsletter-form";
import { FadeUp } from "@/components/animations";

/** Newsletter call-to-action reusing the footer signup form (PRD §9). */
export function NewsletterSection() {
  return (
    <section className="mx-auto w-full max-w-7xl px-4 pb-16 pt-4 sm:px-6 sm:pb-20 lg:px-8">
      <FadeUp inView>
        <div className="mx-auto max-w-3xl rounded-3xl border bg-primary-light p-8 text-center sm:p-12">
          <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
            Join the ShopKart newsletter
          </h2>
          <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground sm:text-base">
            Get 10% off your first order, early access to deals and zero
            spam. Unsubscribe anytime.
          </p>
          <div className="mx-auto mt-6 max-w-md text-left">
            <NewsletterForm />
          </div>
        </div>
      </FadeUp>
    </section>
  );
}
