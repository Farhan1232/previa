# Frontend reference audit — Kinnisvara24 → Previa

**Reference:** <https://kinnisvara24.ee/en>
**Audited:** 10 August 2026
**Method:** Automated headed Chromium session (1440×900), DOM capture, rendered
text extraction, viewport screenshot, and link-graph extraction from the
homepage markup.

This document records what was observed on the **public** reference site, which
UX Previa reproduces, what Previa deliberately improves, and what could not be
reached. Nothing was copied: no markup, CSS, imagery, copy or branding from the
reference is present in this repository. Only interaction patterns and
information architecture — which are not protected expression — informed the
Previa design.

---

## 1. Access summary

| Target | Result |
| --- | --- |
| Homepage `/en` | **Captured in full** (654 KB DOM, rendered text, screenshot) |
| Site link graph | **Captured** — complete public URL map extracted from homepage markup |
| `/en/kinnisvaraotsing` (search) | Blocked — Cloudflare bot verification |
| `/maakler/otsing` (brokers) | Blocked — Cloudflare bot verification |
| `/uusarendused` (developments) | Blocked — Cloudflare bot verification |
| `/lisa-kuulutus` (add advertisement) | Blocked — Cloudflare bot verification |
| `/hinnakiri` (pricing) | Blocked — Cloudflare bot verification |
| `/liitumine` (registration), `/blogi` (articles) | Blocked — Cloudflare bot verification |
| Property-detail page | Not reached (behind the blocked search page) |
| Any authenticated screen | **Not attempted** — no account was created and nothing was purchased |

The site is protected by Cloudflare Bot Management. The homepage passed
verification; subsequent navigations escalated to a full interstitial
("Verifying you are human"). **Probing was stopped at that point rather than
circumvented** — repeatedly working around a site's bot protection is not
appropriate, and the homepage plus the extracted link graph already provide the
information architecture needed.

Where a page could not be inspected, Previa follows the client requirements in
the project brief, which take priority over the reference in any case. Those
screens are marked **[assumed]** in section 5.

---

## 2. Information architecture (extracted from the homepage link graph)

The reference uses Estonian URL slugs even on the `/en` locale:

| Reference path | Purpose | Previa equivalent |
| --- | --- | --- |
| `/en/kinnisvaraotsing` | Property search | `/search` |
| `/uusarendused` | New developments | `/developments` |
| `/maakler/otsing` | Broker search | `/brokers` |
| `/blogi` | Articles / news | `/articles` |
| `/lisa-kuulutus` | Add advertisement | `/add-listing` |
| `/liitumine` | Registration | `/register` |
| `/password/reset` | Password reset | `/forgot-password` |
| `/hinnakiri` | Pricing | `/pricing` |
| `/statistika` | Market statistics | *(out of milestone scope)* |
| `/soovid`, `/soov/uus` | Buy/rent requests | *(noted, not in scope)* |
| `/kontakt`, `/reklaam` | Contact, advertising | `/help`, footer |
| `/kasutustingimused`, `/privaatsuspoliitika` | Terms, privacy | `/terms`, `/privacy` |
| `/sitemap` | Sitemap | `/sitemap.xml` placeholder |
| `/lang/en`, `/lang/et`, `/lang/ru` | Language switch | `/en/`, `/de/`, `/es/` prefixes |
| `/tehasemajad` | Prefabricated houses | **Dropped** — not a Previa requirement |
| `/kinnisvara-valismaal` | Property abroad | **Dropped** — client explicitly excluded "Abroad" |
| `/auth/google`, `/auth/facebook` | Social login | Mocked buttons on `/login` |

Previa uses English, human-readable, crawlable slugs throughout, with
language-prefixed URLs (`/en/`, `/de/`, `/es/`) as the client specified.

---

## 3. Components observed on the homepage

### 3.1 Header
Logo (left) · text nav: *Abroad · Articles · Find an agent · Developments ·
Prefabricated Houses* · `EN` language dropdown · *Log in* with a person icon ·
a large filled **ADD ADVERTISEMENT** button (right, dominant).

The header is compact, and the add-listing button is unambiguously the dominant
action — a pattern Previa keeps.

### 3.2 Persistent left filter rail (the reference's signature pattern)
The homepage itself carries the full search filter as a **left sidebar**, not a
hero search bar:

- **DEAL TYPE** — `Sell`, `Rent`, `Short rental` as selectable tiles with a check affordance
- **PROPERTY TYPE** — an 8-tile icon grid: Apartment, House, House part, Cottage, Abroad, Land, Commercial, Garage
- **REGION** — cascading `County` → `City` → `District` selects
- **Search for ad** — keyword input with a magnifier icon
- **ROOMS / AREA / PRICE / PRICE €/M²** — paired Min–Max inputs
- **ADVANCED SEARCH**, **SEARCH**, **CLEAR FIELDS** actions

### 3.3 Homepage content sections (in order)
1. Promoted new-development banner carousel — three visible slides with prev/next arrows, each showing `PRICE FROM`, a headline and a short line of copy
2. **Exclusive offers** — a full-width coloured section bar with a `VIEW ALL →` link
3. **New developments** — cards with room range, area range and price range
4. Broker/agency promotional block with explanatory copy and a search prompt
5. **Real estate news** — dated article cards with `Read more`
6. Market-statistics promo block
7. Newsletter subscription with a consent note
8. **SEO search-suggestion matrix** — four link columns (Apartments / Houses / Commercial / Popular) crossing property type against city
9. Footer — contact block with opening hours, link columns, social, legal row, copyright
10. A fixed bottom-left tab: **INSERT BUY / RENTAL REQUEST**

### 3.4 Property card anatomy
Heart favourite button (top-left of the image) · agency logo (top-right) ·
caption strip overlaying the image bottom · title · location line · fact row
(`6 rooms · 263 m² · 1389 m²` with a land-area icon) · price, with the previous
price struck through when reduced.

### 3.5 Broker card anatomy
Listing count · name · agency · certification level ("Real estate broker, level
6") · award badge ("Kinnisvara24 Top Broker 25") · email · phone · `VIEW PROFILE`.

---

## 4. What Previa reproduces

These reference behaviours are proven and are carried over — rebuilt from
scratch in Previa's own markup, CSS and branding:

- **Collapsible left filter sidebar** as the primary search surface, open by
  default on large screens, with results beside it rather than beneath it
- **Property-type icon tiles** instead of a plain multi-select
- **Paired Min–Max inputs** for rooms, area and price
- **Cascading region selects** (country → city → district)
- Section header bars carrying a **`VIEW ALL →`** link
- Card anatomy: favourite control, media badges, price prominence, compact fact row
- Struck-through previous price on reduced listings
- Promoted-development strip above the fold
- Broker cards showing listing count, agency, credentials and direct contact
- **SEO link matrix** in the footer — type × city internal links, fully crawlable
- Newsletter capture with an explicit consent line
- Dated article cards in a news section
- A persistent edge tab for a secondary conversion action

---

## 5. What Previa deliberately changes

| # | Reference behaviour | Previa decision | Why |
| --- | --- | --- | --- |
| 1 | "Abroad" nav item and property type | **Removed entirely** | Client requirement. Previa is global by default, so a separate "abroad" concept is incoherent. |
| 2 | "ADD ADVERTISEMENT" | **"Add listing"** | Client requirement. Not "List Your Property". |
| 3 | "Find an agent" | **"Brokers"** | Client requirement. |
| 4 | "Projects" / `uusarendused` | **"Developments"** | Client requirement. |
| 5 | "Prefabricated Houses" nav item | **Removed** | Estonia-specific vertical, not a Previa requirement. |
| 6 | Large top dropdown on mobile | **Right-side off-canvas drawer** | Client explicitly rejected the reference's mobile menu. Previa's drawer overlays (never shifts) the page, traps focus, closes on Escape/backdrop, and restores focus. |
| 7 | Sidebar is always present, no collapse | **Collapsible with state preservation** | Client requirement. Collapsed → results expand and a labelled rail button reopens it; below 1024px it becomes an overlay drawer with a fixed edge tab. Filter values survive both transitions. |
| 8 | No map on search results | **Two map modes**: Airbnb-style split (results left / map right) and full-screen map, with two-way card↔marker highlighting and clustering | Core Previa requirement; Google Maps is central to the product. |
| 9 | Address-only location entry when listing | **Google Maps address autocomplete *or* direct pin placement**, plus a public-precision control (exact / street / area) | Client asked for this to be improved rather than copied. Sellers can publish an approximate location. |
| 10 | Large stepper icons in the add-listing flow | **Compact right-side vertical stepper** with completed / current / error / incomplete states | Client asked for the vertical progression to be kept but the icons made smaller. |
| 11 | Dense red accent, heavy borders, small type | **Navy + warm off-white with a 10% gold accent**, 8px spacing system, larger type, soft elevation | Premium international positioning. Gold is restricted to featured/premium signals only. |
| 12 | Estonian slugs on the English locale | **English slugs, language-prefixed URLs** | Correctness for an international product; also better for SEO and hreflang. |
| 13 | Placeholder-only labels in several inputs | **Permanent visible labels on every field** | WCAG 2.1 AA. |
| 14 | Full page reload on filter change | **HTMX partial replacement** of the results region only | Preserves scroll position, sidebar state and map viewport. |
| 15 | Single currency | **Multi-currency** (EUR/GBP/CZK) with per-country defaults | Global marketplace requirement. |
| 16 | Single-country focus | **Country selector** driving banner, promoted brokers, map centre and default results | Client requirement; five seeded markets (EE, DE, ES, FI, PT). |

---

## 6. Workflows reproduced

- **Search → refine → view**: filter rail → HTMX result refresh → active filter chips → grid/list/split-map/full-map switch → property detail
- **Add listing**: sale-or-rent → category → location (map pin or address) → public-precision choice → details → rooms → features → description → media → price → contact → package → preview → publish
- **Draft continuation**: a wizard exited midway persists as a draft with a completion percentage and is resumable from *My listings → Drafts*
- **Favourites**: card-level toggle that updates optimistically and is reflected in the account area
- **Saved searches**: a filter set is nameable, storable and replayable, with alert frequency
- **Broker discovery**: browse/filter brokers → profile → their listings → contact
- **Package purchase**: pricing → package select → order summary → payment method (Stripe / PayPal / Paysera) → processing → success/failure/cancelled → invoice row in billing history

---

## 7. Assumptions made for pages that could not be inspected

Each of these follows the project brief, professional real-estate convention,
and the patterns visible on the reference homepage. They are **[assumed]**, not
observed:

1. **Search results page** — assumed to reuse the homepage's left filter rail with a results header carrying count, sort and view switches. Previa adds the map modes.
2. **Property detail** — assumed conventional: gallery, price/key facts, description, features, location map, broker card with phone/email reveal, similar properties. Previa adds Schedule Viewing, Share, Report Listing and a mobile sticky contact bar.
3. **Add-listing wizard** — the right-side vertical progression is a client-described detail of the reference. Step order follows the brief's suggested sequence.
4. **Pricing** — assumed tiered packages with duration, photo limits and bump counts. Previa seeds Basic / Standard / Premium / Agency.
5. **Broker and agency profiles** — assumed profile header, credentials, contact, and a paginated listing grid.
6. **Authentication** — assumed email + password with social buttons (the homepage exposes `/auth/google` and `/auth/facebook`). All mocked in Previa.
7. **User dashboard and admin panel** — not publicly reachable at all. Built entirely from the brief.

---

## 8. Compliance notes

- No account was registered on the reference site, and nothing was purchased.
- No reference markup, stylesheet, script, font, image or copy is present in
  this repository.
- Previa's logo, colour system, typography, component library, copy and mock
  data are original.
- The reference's bot protection was **not** circumvented; probing stopped when
  the site began challenging requests.
- Previa's photography comes from Unsplash, downloaded and stored locally in
  `web/static/img/` under the Unsplash License (free for commercial use, no
  permission required). Nothing is hotlinked at runtime, and nothing is taken
  from any competing portal. The broker portraits show real people who are not
  Previa staff, so they — like the property photography — should be replaced
  with the client's own licensed images before launch. See the README.
