# Dropshipping E-commerce Platform — Frontend PRD

## 1. Product Overview

Build a modern, production-quality dropshipping e-commerce frontend using **Next.js + React**.

The first version will use **mock/demo data only**. There will be no real backend, payment processing, authentication server, supplier integration, or inventory synchronization yet.

The frontend must be architected so that demo data can later be replaced by a Go REST API + PostgreSQL backend without requiring a major UI rewrite.

The application should feel like a real commercial e-commerce platform rather than a tutorial/demo project.

---

# 2. Product Goals

## Primary Goals

* Build a high-quality modern e-commerce storefront.
* Create a scalable frontend architecture.
* Make the UI responsive across mobile, tablet, and desktop.
* Provide excellent product discovery and browsing.
* Create polished product detail pages.
* Implement smooth, purposeful animations.
* Implement SEO properly for every product page.
* Use reusable components rather than page-specific duplicated UI.
* Keep business/data access separate from presentation.
* Prepare the frontend for future Go API integration.
* Build an admin dashboard using mock data.
* Use Zustand for client-side application state.
* Follow senior-level frontend engineering practices.

## Non-Goals for Version 1

Do NOT implement:

* Real payment processing
* Real authentication
* Real user accounts
* Real supplier APIs
* Real inventory synchronization
* Real shipping APIs
* Real order processing
* Real database
* Real email delivery
* Real payment webhooks
* Real backend authorization

These will be implemented later.

---

# 3. Target Technology Stack

## Core

* Next.js
* React
* JavaScript/JSX unless TypeScript is already configured
* App Router
* Server Components by default
* Client Components only where interactivity is required

## Styling

* Tailwind CSS
* shadcn/ui
* CSS variables for design tokens

## UI / Animation

* shadcn/ui
* Magic UI
* Framer Motion / Motion
* Lucide icons

Use animation intentionally. Do not animate everything.

## State Management

* Zustand

Use Zustand for:

* Cart
* Wishlist
* UI state that genuinely needs global access
* Selected product options where appropriate

Do NOT put all server/mock data into Zustand.

Product/catalog data should remain accessible through a dedicated data/repository layer.

## Data

Initially:

```text
Mock data
↓
Repository/data-access functions
↓
React/Next.js UI
```

Future:

```text
Next.js
↓
API client
↓
Go REST API
↓
PostgreSQL
```

The UI should not directly depend on the implementation of the data source.

---

# 4. Design Direction

## Overall Design

Create a premium, modern, clean e-commerce interface.

Design characteristics:

* Minimal
* Professional
* High visual hierarchy
* Product-focused
* Generous whitespace
* Strong typography
* Subtle borders
* Soft shadows
* Rounded but not excessively rounded components
* Smooth micro-interactions
* Excellent mobile experience

Avoid:

* Excessive gradients
* Excessive glassmorphism
* Excessive animations
* Huge amounts of text
* Random decorative elements
* Inconsistent border radii
* Inconsistent spacing
* Generic template appearance

---

# 5. Color System

Use a neutral base with an emerald primary accent.

## Primary

```text
Emerald
#10B981
```

## Primary Dark

```text
#059669
```

## Primary Light

```text
#D1FAE5
```

## Background

```text
#FFFFFF
```

## Secondary Background

```text
#F8FAFC
```

## Foreground

```text
#0F172A
```

## Muted Foreground

```text
#64748B
```

## Border

```text
#E2E8F0
```

## Destructive

```text
#EF4444
```

Use these through Tailwind/shadcn design tokens rather than scattering raw hex values throughout components.

Dark mode should be architecturally possible, but it is not required for the first release.

---

# 6. Typography

Use a modern sans-serif typeface.

Recommended:

```text
Inter
```

Typography hierarchy:

```text
Hero heading
↓
Page heading
↓
Section heading
↓
Product title
↓
Body text
↓
Metadata
```

Avoid excessive font-weight variation.

Use typography to establish hierarchy rather than relying on colors and shadows.

---

# 7. Application Structure

Recommended structure:

```text
app/
├── layout.jsx
├── page.jsx
├── loading.jsx
├── not-found.jsx
├── globals.css
│
├── products/
│   ├── page.jsx
│   └── [slug]/
│       ├── page.jsx
│       ├── loading.jsx
│       └── not-found.jsx
│
├── category/
│   └── [slug]/
│       └── page.jsx
│
├── search/
│   └── page.jsx
│
├── cart/
│   └── page.jsx
│
├── checkout/
│   └── page.jsx
│
├── order/
│   └── success/
│       └── page.jsx
│
├── account/
│   ├── page.jsx
│   ├── orders/
│   ├── wishlist/
│   ├── addresses/
│   └── settings/
│
└── admin/
    ├── page.jsx
    ├── products/
    ├── orders/
    ├── customers/
    ├── inventory/
    └── analytics/
```

Components:

```text
components/
├── layout/
├── navigation/
├── hero/
├── products/
├── cart/
├── checkout/
├── account/
├── admin/
├── forms/
├── ui/
└── animations/
```

Data:

```text
data/
├── products.js
├── categories.js
├── reviews.js
├── orders.js
└── users.js
```

State:

```text
store/
├── cart-store.js
├── wishlist-store.js
└── ui-store.js
```

Utilities:

```text
lib/
├── utils.js
├── products.js
├── seo.js
├── constants.js
└── formatters.js
```

Types/interfaces:

```text
types/
├── product.js
├── order.js
├── user.js
└── category.js
```

---

# 8. Core Data Models

## Product

Every product should support:

```js
{
  id,
  slug,
  name,
  shortDescription,
  description,
  price,
  compareAtPrice,
  currency,
  images,
  thumbnail,
  category,
  categorySlug,
  brand,
  rating,
  reviewCount,
  stock,
  sku,
  tags,
  features,
  specifications,
  variants,
  isFeatured,
  isNew,
  isBestSeller,
  seo
}
```

## Product Variant

```js
{
  id,
  name,
  value,
  price,
  stock,
  sku
}
```

Examples:

```text
Color → Black
Color → White
Size → Small
Size → Medium
Size → Large
```

---

# 9. Homepage

The homepage should immediately communicate:

1. What the store sells
2. Why customers should care
3. Featured products
4. Categories
5. Trust signals
6. Current promotions

## Homepage Structure

```text
Announcement Bar
↓
Header / Navigation
↓
Hero
↓
Featured Categories
↓
Featured Products
↓
Promotion Section
↓
Best Sellers
↓
Why Shop With Us
↓
Customer Reviews
↓
Newsletter
↓
Footer
```

---

# 10. First-View Animation

The first viewport must have polished entrance animation.

Hero animation sequence:

```text
Hero container
    ↓
Heading fades + moves upward
    ↓
Description fades upward
    ↓
CTA buttons appear
    ↓
Hero image/product visual enters
    ↓
Secondary decorative elements appear
```

Animation should be:

* Fast
* Smooth
* Subtle
* Professional

Recommended duration:

```text
0.4s – 0.8s
```

Use staggered animation where appropriate.

Do NOT create long cinematic animations that delay content visibility.

Respect:

```text
prefers-reduced-motion
```

Users who prefer reduced motion should receive minimal/no decorative animation.

---

# 11. Header

Desktop:

```text
Logo
Categories
Shop
Deals
Search
Account
Wishlist
Cart
```

Mobile:

```text
Menu
Logo
Cart
```

Mobile navigation should use a drawer/sheet.

Header requirements:

* Sticky header
* Smooth background transition on scroll
* Accessible navigation
* Keyboard navigation
* Visible focus states
* Cart item count
* Wishlist indicator

Do not create unnecessary scroll-based animations.

---

# 12. Product Listing Page

Route:

```text
/products
```

Features:

* Product grid
* Search
* Category filter
* Price filter
* Rating filter
* Availability filter
* Sorting
* Pagination or load more
* Product count
* Empty state
* Loading state

Desktop:

```text
Sidebar filters | Product grid
```

Mobile:

```text
Products
Filter button
Sort button
Product grid
```

---

# 13. Product Card

Product cards are one of the most important reusable components.

Each card should contain:

```text
Product image
Badge
Wishlist button
Product name
Rating
Review count
Price
Compare-at price
Discount
Add to Cart
```

## Hover Animation

Desktop hover:

```text
Card
↓
Slight translateY
↓
Subtle shadow increase
↓
Image slight scale
↓
Secondary actions become visible
```

Do not use excessive movement.

Recommended:

```text
transform: translateY(-2px)
image scale: 1.03–1.05
```

Buttons should have:

* Hover background transition
* Slight scale where appropriate
* Active press feedback
* Focus ring

Avoid:

```text
scale(1.15)
```

or exaggerated bounce animations.

---

# 14. Product Details Page

Route:

```text
/products/[slug]
```

This is one of the highest-priority pages.

Layout:

```text
Breadcrumb
↓
Product Gallery | Product Information
↓
Description
↓
Features
↓
Specifications
↓
Reviews
↓
Related Products
```

Product information should include:

* Product name
* Rating
* Review count
* Price
* Compare-at price
* Discount
* Availability
* Variant selectors
* Quantity selector
* Add to Cart
* Buy Now
* Shipping information
* Return information
* Product SKU

---

# 15. Product Gallery

Features:

* Main image
* Thumbnail gallery
* Previous/next controls
* Image zoom
* Keyboard navigation
* Mobile swipe
* Accessible alt text

Images must not cause layout shifts.

Use explicit dimensions/aspect ratios.

---

# 16. Add-to-Cart Interaction

When the user clicks:

```text
Add to Cart
```

Provide immediate feedback.

Example:

```text
Add to Cart
↓
Button loading state
↓
Added ✓
↓
Cart count updates
↓
Optional cart drawer
```

Use Zustand for cart state.

Cart item:

```js
{
  productId,
  variantId,
  quantity,
  price
}
```

Do not store unnecessary duplicated product objects inside persistent state.

Prefer IDs and reconstruct display data from the product repository.

---

# 17. Cart

Route:

```text
/cart
```

Features:

* Product image
* Product name
* Variant
* Price
* Quantity control
* Remove
* Wishlist action
* Subtotal
* Shipping estimate
* Discount
* Total
* Checkout button

Empty cart:

```text
Illustration
"Your cart is empty"
Continue Shopping
```

---

# 18. Checkout

Route:

```text
/checkout
```

Version 1 is UI-only.

Sections:

```text
Contact
↓
Shipping Address
↓
Delivery Method
↓
Payment Method
↓
Order Summary
↓
Place Order
```

Use reusable form components.

Form requirements:

* Validation
* Accessible labels
* Error messages
* Loading states
* Disabled states
* Mobile-friendly inputs

Do not collect real payment information in the demo version.

---

# 19. Order Success

Route:

```text
/order/success
```

Display:

```text
Success icon
Order number
Order summary
Shipping information
Continue shopping
View order
```

Use a subtle success animation.

---

# 20. Search

Route:

```text
/search?q=
```

Features:

* Search input
* Search suggestions
* Product results
* Category results
* Empty state
* Recent searches
* Clear search

Search should be debounced.

For mock data, implement client-side search.

Later it can call the backend.

---

# 21. Category Pages

Route:

```text
/category/[slug]
```

Each category page should contain:

```text
Breadcrumb
Category heading
Description
Category banner
Product count
Filters
Sorting
Product grid
```

Category pages should also be SEO-friendly.

---

# 22. Wishlist

Wishlist functionality:

```text
Add
Remove
Toggle
```

Use Zustand.

Persist wishlist locally.

Provide:

```text
/wishlist
```

page.

---

# 23. Account

Demo account pages:

```text
/account
/account/orders
/account/orders/[id]
/account/wishlist
/account/addresses
/account/settings
```

Authentication is not real in Version 1.

Create a mock user session.

The architecture should make it possible to replace the mock session with real authentication later.

---

# 24. Admin Dashboard

Create a professional admin interface.

Route:

```text
/admin
```

Dashboard:

```text
Revenue
Orders
Customers
Products
Low Stock
Recent Orders
Sales Chart
Top Products
```

Additional pages:

```text
/admin/products
/admin/orders
/admin/customers
/admin/inventory
/admin/analytics
```

Use mock data.

Admin UI should use reusable:

* Tables
* Dialogs
* Forms
* Dropdown menus
* Tabs
* Charts
* Cards
* Status badges

---

# 25. Zustand Architecture

Create separate stores.

## Cart Store

Responsibilities:

```text
addItem()
removeItem()
updateQuantity()
clearCart()
getItemCount()
getSubtotal()
```

## Wishlist Store

Responsibilities:

```text
addToWishlist()
removeFromWishlist()
toggleWishlist()
isWishlisted()
```

## UI Store

Only for genuinely global UI state:

```text
mobile navigation
cart drawer
search overlay
```

Avoid creating one giant global store.

Do not use Zustand for:

```text
Static product catalog
SEO metadata
Server-renderable content
```

---

# 26. SEO Requirements

SEO is a first-class requirement.

Every product page must have unique SEO metadata.

Route:

```text
/products/[slug]
```

Use Next.js metadata APIs.

Each product page should generate:

```text
title
description
canonical URL
Open Graph metadata
Twitter metadata
robots metadata where appropriate
```

Example conceptual metadata:

```text
title:
"Wireless Portable Blender | Store Name"

description:
"Shop the Wireless Portable Blender with ..."
```

Do not use generic metadata for every product.

---

# 27. Product SEO Data

Each product should have:

```js
seo: {
  title,
  description,
  keywords,
  imageAlt
}
```

If custom SEO data does not exist, generate sensible metadata from product information.

Never generate:

```text
undefined
null
[object Object]
```

inside metadata.

---

# 28. Product Structured Data

Every indexable product page should generate valid Product structured data using JSON-LD.

Include appropriate fields such as:

```text
@context
@type
name
description
image
sku
brand
offers
price
priceCurrency
availability
aggregateRating
review
```

Only include information that actually exists in the product data.

Do not fabricate ratings, reviews, prices, availability, or brand information.

---

# 29. Canonical URLs

Every product page must have a canonical URL.

Preferred structure:

```text
/products/wireless-portable-blender
```

Avoid exposing unnecessary query parameters in canonical URLs.

Canonical URLs should be deterministic.

---

# 30. Sitemap

Generate a sitemap containing:

```text
Homepage
Product pages
Category pages
Important static pages
```

Product URLs should be generated from the product catalog.

Do not manually maintain hundreds of product URLs.

---

# 31. Robots

Create:

```text
robots.txt
```

Allow search engines to crawl:

```text
/
 /products/*
 /category/*
```

Do not index private/demo administrative areas:

```text
/admin/*
/account/*
/checkout/*
```

Review the final robots configuration before production deployment.

---

# 32. Open Graph

Product pages should generate product-specific Open Graph metadata.

Include:

```text
og:title
og:description
og:image
og:url
og:type
```

Use the primary product image.

Images should have appropriate dimensions and meaningful alt text.

---

# 33. Performance

Performance is a major requirement.

Use:

* Next.js Image
* Responsive image sizes
* Lazy loading where appropriate
* Code splitting
* Server Components
* Dynamic imports for heavy client-only components
* Minimal JavaScript
* Proper caching strategy
* Avoid unnecessary client components

Do not turn entire pages into:

```text
"use client"
```

unless required.

---

# 34. Core Web Vitals

Optimize for:

```text
LCP
CLS
INP
```

Avoid:

* Layout shifts
* Huge JavaScript bundles
* Blocking animations
* Unoptimized images
* Rendering unnecessary components
* Client-side fetching when server rendering is sufficient

---

# 35. Loading States

Every major asynchronous-looking interface should have a loading state.

Examples:

```text
Product skeleton
Product detail skeleton
Cart loading
Admin table skeleton
Search loading
```

Use shadcn Skeleton.

Avoid blank screens.

---

# 36. Error States

Create reusable error states.

Examples:

```text
Product not found
Category not found
Search failed
Cart error
Checkout error
```

Provide useful recovery actions.

Example:

```text
Product not found.

Browse all products
```

---

# 37. Empty States

Design intentional empty states for:

```text
Empty cart
Empty wishlist
No search results
No orders
No products
No notifications
```

Each empty state should explain what happened and provide an appropriate next action.

---

# 38. Accessibility

Target WCAG-conscious implementation.

Requirements:

* Semantic HTML
* Proper heading hierarchy
* Accessible buttons
* Accessible form labels
* Keyboard navigation
* Visible focus states
* Sufficient color contrast
* Alt text
* ARIA only when necessary
* Dialog focus management
* Escape key support
* Screen-reader-friendly status messages

Never use:

```jsx
<div onClick={...}>
```

when a semantic button is appropriate.

---

# 39. Responsive Design

Support:

```text
Mobile
Tablet
Laptop
Desktop
Large desktop
```

Design mobile-first.

Important breakpoints:

```text
~640px
~768px
~1024px
~1280px
~1536px
```

Do not optimize only for desktop.

---

# 40. Animation System

Create reusable animation patterns.

Examples:

```text
FadeIn
FadeUp
StaggerContainer
StaggerItem
ScaleIn
PageTransition
```

Use Framer Motion / Motion consistently.

Do not create random animation implementations throughout the application.

Animation rules:

```text
Micro interaction:
150–250ms

Component entrance:
300–500ms

Hero:
400–800ms

Page transition:
200–400ms
```

Use easing appropriate to the interaction.

Always support:

```text
prefers-reduced-motion
```

---

# 41. Magic UI Usage

Use Magic UI selectively for visually important areas.

Potential usage:

```text
Hero
Animated headline
Announcement section
Special promotion
Trust section
Decorative background
```

Do not use Magic UI components simply because they are available.

The application should remain cohesive.

---

# 42. shadcn/ui Usage

Prefer shadcn/ui for:

```text
Button
Input
Dialog
Sheet
Dropdown
Select
Tabs
Accordion
Badge
Card
Skeleton
Toast/Sonner
Tooltip
Form
Table
Pagination
```

Customize the components to match the design system.

Do not introduce another component library unless there is a strong reason.

---

# 43. Component Engineering

Components must follow single-responsibility principles.

Bad:

```text
ProductPage.jsx
```

containing:

```text
Gallery
Product information
Reviews
Related products
Cart logic
SEO
Analytics
```

Instead:

```text
ProductPage
├── ProductGallery
├── ProductInfo
├── ProductVariants
├── QuantitySelector
├── AddToCartButton
├── ProductShippingInfo
├── ProductDescription
├── ProductSpecifications
├── ProductReviews
└── RelatedProducts
```

---

# 44. Data Access Layer

Do not directly import mock arrays into every component.

Create repository functions.

Example:

```js
getProducts()
getProductBySlug(slug)
getProductsByCategory(slug)
getFeaturedProducts()
getRelatedProducts(productId)
searchProducts(query)
```

Components should consume these functions.

Future implementation:

```text
Mock repository
↓
Go API repository
```

The UI should remain mostly unchanged.

---

# 45. URL and Routing Rules

Use readable URLs.

Examples:

```text
/
 /products
 /products/wireless-portable-blender
 /category/kitchen
 /search?q=blender
 /cart
 /checkout
```

Avoid:

```text
/product?id=123
```

for the primary product URL.

Use stable slugs.

---

# 46. Product Images

Every image must have:

* Meaningful alt text
* Proper aspect ratio
* Optimized dimensions
* Responsive sizing

Example:

```text
Wireless Portable Blender — Front View
Wireless Portable Blender — Side View
Wireless Portable Blender — Charging Port
```

Do not use:

```text
image1
image2
img123
```

as final alt text.

---

# 47. Mock Data Requirements

Create realistic demo data.

Minimum:

```text
30 products
8 categories
20+ reviews
10 orders
5 customers
```

Products should contain different states:

```text
In stock
Low stock
Out of stock
Discounted
New
Featured
Best seller
Multiple variants
No variants
No reviews
Many reviews
```

This allows the UI to be tested against realistic scenarios.

---

# 48. Product Pricing

Support:

```text
price
compareAtPrice
discount percentage
currency
```

Example:

```text
$39.99
$59.99
33% OFF
```

Do not hard-code calculated discount values if they can be calculated safely.

Create a reusable price formatter.

---

# 49. Cart Persistence

Persist cart state using Zustand persistence.

On browser refresh:

```text
Cart remains available
```

Do not persist sensitive information.

---

# 50. UX Requirements

Every user action should have feedback.

Examples:

```text
Add to cart → success feedback
Remove item → immediate UI update
Wishlist → heart animation/state change
Form submission → loading state
Search → loading state
Invalid form → field error
```

Avoid interactions where the user cannot tell whether their action worked.

---

# 51. Security Considerations

Even though Version 1 uses demo data:

Do not:

* Put secrets in client code
* Put API keys in components
* Store passwords
* Simulate real payment credentials
* Expose future backend secrets
* Trust client-side authorization

Admin restrictions in Version 1 are UI/demo only.

Real authorization will be handled by the future backend.

---

# 52. Architecture for Future Go Backend

The frontend should eventually communicate with:

```text
Next.js
    ↓
Go REST API
    ↓
PostgreSQL
```

Potential future API:

```text
GET    /api/products
GET    /api/products/:slug
GET    /api/categories
GET    /api/categories/:slug

POST   /api/cart
GET    /api/cart

POST   /api/orders
GET    /api/orders
GET    /api/orders/:id

POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/logout

POST   /api/payments
POST   /api/payments/webhook
```

Do not implement these endpoints now.

The frontend should simply be structured so replacing the mock repository with API calls is straightforward.

---

# 53. SEO-Friendly Rendering Strategy

For public pages, prefer server rendering.

Especially:

```text
Homepage
Product pages
Category pages
```

Product information should be available to search engines without requiring client-side JavaScript execution.

Interactive components can remain client components.

Example architecture:

```text
Product Page
│
├── Server Component
│   ├── Product data
│   ├── SEO metadata
│   ├── JSON-LD
│   └── Product layout
│
└── Client Components
    ├── Gallery interactions
    ├── Variant selector
    ├── Quantity selector
    └── Add to cart
```

---

# 54. Error Boundaries

Create appropriate:

```text
error.jsx
not-found.jsx
loading.jsx
```

for important routes.

Product route:

```text
/products/[slug]
├── page.jsx
├── loading.jsx
└── not-found.jsx
```

Provide useful UI instead of raw errors.

---

# 55. Code Quality Rules

The coding agent MUST:

* Inspect existing code before modifying it.
* Reuse existing components.
* Avoid unnecessary dependencies.
* Avoid duplicate components.
* Keep components small.
* Keep business logic outside presentation components when practical.
* Use descriptive names.
* Avoid magic numbers.
* Avoid deeply nested components.
* Avoid unnecessary `useEffect`.
* Avoid unnecessary `useMemo`.
* Avoid unnecessary `useCallback`.
* Prefer server components.
* Keep client boundaries small.
* Preserve accessibility.
* Preserve responsive behavior.
* Test changed functionality before moving on.

---

# 56. Agentic Coding Rules

The coding agent must NOT attempt to build the entire application in one operation.

Work feature-by-feature.

Recommended implementation sequence:

```text
1. Project foundation
2. Design system
3. Layout
4. Header
5. Footer
6. Homepage
7. Product data model
8. Product card
9. Product listing
10. Product details
11. SEO
12. Search
13. Categories
14. Cart
15. Wishlist
16. Checkout
17. Account
18. Admin
19. Loading/error/empty states
20. Accessibility audit
21. Performance optimization
22. Final visual polish
```

After each major feature:

```text
Implement
↓
Run
↓
Test
↓
Review
↓
Fix
↓
Commit
↓
Continue
```

Never make unrelated changes while implementing a feature.

---

# 57. Git Strategy

Use meaningful commits.

Examples:

```text
feat: add storefront layout
feat: add product listing page
feat: add product detail page
feat: add product seo metadata
feat: add cart state with zustand
feat: add wishlist state
feat: add checkout ui
feat: add admin dashboard
perf: optimize product images
a11y: improve product card accessibility
refactor: extract product repository
```

Avoid:

```text
update
fix stuff
changes
final
done
```

---

# 58. Definition of Done

A feature is complete only when:

* Desktop works
* Mobile works
* Loading state exists
* Empty state exists where applicable
* Error state exists where applicable
* Keyboard interaction works
* Focus states work
* Animations work
* Reduced-motion behavior works
* No console errors
* No obvious hydration errors
* No broken links
* No layout shift caused by the feature
* Components are reusable
* No unnecessary duplicated logic
* SEO requirements are satisfied where applicable

---

# 59. Homepage Definition of Done

The homepage must contain:

* Responsive header
* Hero section
* First-view animation
* CTA
* Featured categories
* Featured products
* Product hover animations
* Promotional section
* Best sellers
* Trust section
* Reviews
* Newsletter
* Footer
* Mobile navigation
* Responsive layouts
* Loading/appropriate fallback states

---

# 60. Product Page Definition of Done

Every product page must contain:

* Unique URL
* Unique title
* Unique meta description
* Canonical URL
* Open Graph metadata
* Product image
* Product image alt text
* Product information
* Price
* Discount where applicable
* Rating where available
* Variants where applicable
* Quantity selector
* Add to Cart
* Wishlist
* Description
* Features
* Specifications
* Shipping information
* Related products
* Product JSON-LD
* Loading state
* Not-found state
* Responsive design
* Accessible interactions

---

# 61. Quality Bar

The final application should feel like a professionally designed e-commerce product.

It should NOT feel like:

```text
AI-generated template
```

or:

```text
tutorial project
```

The coding agent should prioritize:

```text
Consistency
+
Accessibility
+
Performance
+
SEO
+
Maintainability
+
UX
+
Visual polish
```

over simply producing a large amount of code.

---

# 62. Important Engineering Principle

The frontend must be designed around this future migration:

```text
CURRENT

Mock Data
   ↓
Repository
   ↓
Next.js UI


FUTURE

Go REST API
   ↓
API Repository
   ↓
Next.js UI
```

The presentation layer should not care whether the data comes from:

```text
mock data
```

or:

```text
PostgreSQL through Go
```

This separation is mandatory.

---

# 63. Final Instruction to Coding Agent

You are acting as a senior frontend engineer.

Before writing code:

1. Inspect the existing project.
2. Understand the current architecture.
3. Reuse existing dependencies and components.
4. Do not overwrite working code unnecessarily.
5. Plan the smallest coherent change.
6. Implement the feature.
7. Verify the implementation.
8. Fix errors.
9. Keep the code maintainable.
10. Only then move to the next feature.

Do not generate the entire application in one huge change.

Build incrementally.

Prioritize real production engineering practices over code volume.

The first release uses demo data only.

Do not implement backend business logic yet.

The eventual backend will be implemented using Go + PostgreSQL, so the frontend must remain API-ready from the beginning.
