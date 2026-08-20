# Backend integration points

Everything in this milestone is served from an in-memory mock provider. This
document is the map from that mock to the production Go + MySQL backend.

**The rule that makes this cheap:** handlers and templates depend only on the
interfaces in `internal/data/repository.go`. Replacing the mock is a change to
one line in `cmd/previa/main.go`:

```go
// today
store := data.NewStore(time.Now())

// production
store := mysql.NewStore(db)   // must satisfy the same *data.Store interfaces
```

No template and no handler changes are required, provided the new provider
returns the same `models.*` shapes.

---

## 1. Repository interfaces to implement

All six live in `internal/data/repository.go`.

| Interface | Methods | Backing tables (suggested) |
| --- | --- | --- |
| `PropertyRepository` | `Search`, `ByID`, `BySlug`, `Featured`, `Recent`, `Similar`, `ByBroker`, `ByDevelopment`, `ByStatus`, `CountByType`, `PopularLocations` | `properties`, `property_images`, `property_features` |
| `BrokerRepository` | `All`, `ByID`, `BySlug`, `Promoted`, `Filter`, `Agencies`, `AgencyBySlug`, `BrokersByAgency` | `brokers`, `agencies`, `broker_languages` |
| `ContentRepository` | `Articles`, `ArticleBySlug`, `ArticlesByCategory`, `RelatedArticles`, `Developments`, `DevelopmentBySlug` | `articles`, `developments` |
| `AccountRepository` | `CurrentUser`, `Favourites`, `IsFavourite`, `ToggleFavourite`, `SavedSearches`, `AddSavedSearch`, `Notifications`, `UnreadCount`, `MyListings`, `ListingStats`, `CloneListing`, `Drafts`, `Payments` | `users`, `favourites`, `saved_searches`, `notifications`, `listing_drafts`, `payments` |
| `CatalogRepository` | `Countries`, `Country`, `Banner`, `Packages`, `Testimonials`, `Languages`, `RestrictedCountries` | `countries`, `banners`, `packages`, `languages`, `restricted_countries` |
| `AdminRepository` | `Stats`, `Users`, `Translations`, `SEOEntries`, `Backups`, `Files`, `Tables`, `SystemInfo`, `ActivityLog` | `audit_log`, `translations`, `seo_entries`, `backups` |

Note: Go interfaces cannot declare two methods with the same name, so the
property and broker repositories both wanting `ByID`/`BySlug` is resolved by
`data.BrokerAdapter`. A real implementation can simply provide separate types
and drop the adapter.

---

## 2. Search and filtering

`data.PropertyFilter` already carries every supported parameter, and
`data.ParseFilter` builds it from query values. The mock filters in memory;
SQL should translate the same struct.

Fields map to `WHERE` clauses directly. Pointer fields (`PriceMin`, `PriceMax`,
`AreaMin`, `AreaMax`, `LandAreaMin`) are `nil` when unset — do not emit a clause
for those.

```sql
SELECT ... FROM properties p
WHERE p.status = 'active'
  AND (? = ''  OR p.deal_type = ?)
  AND (? = ''  OR p.country_code = ?)
  AND (? = ''  OR p.city = ?)
  AND (? IS NULL OR p.price >= ?)
  AND (? IS NULL OR p.price <= ?)
  AND (? = 0   OR p.rooms >= ?)
  AND (? = 0   OR p.bedrooms >= ?)
  AND (? = false OR p.parking = 1)
  AND (? = '' OR MATCH(p.title, p.description) AGAINST (? IN BOOLEAN MODE))
ORDER BY /* see SortOrder */
LIMIT ? OFFSET ?
```

Indexes worth having from day one:

```sql
CREATE INDEX idx_prop_search  ON properties (status, deal_type, country_code, city, price);
CREATE INDEX idx_prop_geo     ON properties (lat, lng);
CREATE INDEX idx_prop_created ON properties (status, created_at DESC);
CREATE INDEX idx_prop_broker  ON properties (broker_id, status);
FULLTEXT INDEX ft_prop        ON properties (title, description);
```

`SearchResult.MapPoints` must hold **every** match, not just the current page —
the map plots the full set while the list paginates. For very large result sets,
cap it server-side (a few thousand) and cluster beyond that.

### Distance, and why it must not sort

When the Location box resolves to a point, `ApplyLocation` fills
`PropertyFilter.Lat/Lng` and the search stamps `Property.Distance` /
`DistanceSet` on every match so each card can print "distance 12.40 KM".

Three rules, all of them the client's:

- **Nothing is matched on the point.** The location still narrows by
  country/city/district/address exactly as it did; the coordinates are carried
  only so a distance can be stated.
- **Nothing is ordered by it.** "The order of the real-estate ads here is not
  automatically according to nearest distance. The order stays like it is set up
  in the 'sort by' menu." In the mock the measurement deliberately happens
  *after* `sortProperties`, so it is structurally incapable of affecting the
  order; a SQL implementation should likewise select the distance as a column
  and leave `ORDER BY` to `SortOrder`.
- **Zero is a real distance.** A listing on the searched point is 0.00 km away,
  which is why there is a separate `DistanceSet` flag rather than a zero
  sentinel. `ST_Distance_Sphere(p.coords, POINT(?, ?)) / 1000 AS distance_km`
  with a `distance_set` derived from whether a point was supplied.

The broker directory measures the same way (`models.MapPlace.DistanceKm`,
haversine) but *does* order by it — there, nearest-first is the point of the
search.

### The broker directory's two modes

`/brokers` answers with one of two lists, and `BrokerFilter.IsBrowsing()` is
what decides which:

| State | What is shown |
| --- | --- |
| Nothing searched | The brokers whose **market ad** covers the header's chosen country — the same list the homepage strip shows, in the same order |
| Anything searched (place, language, name) | The whole directory narrowed by it; the market no longer applies. With a resolved place, ordered nearest first with the distance on each card |

The client's wording named the location field — "if the user in this search menu
enters the location and radius, then the 'choose your market' system is not
active any more" — and the rule underneath it is browsing versus searching, so
every input in that menu ends the market mode. A language search confined to one
market would be the market overruling the question that was actually asked.

### Showing brokers among the results

`brokers=1` sets `PropertyFilter.ShowBrokers`, which is not a property filter at
all: it adds a second kind of result rather than narrowing the first. The
handler answers it with `BrokerRepository.OnMap`, narrowed by the same place the
listings were (`brokerSearch` in `pkg/handlers/search.go` — a resolved city
becomes a 50 km radius, a whole country becomes a market match, because a 50 km
circle around a country's centroid is a field, not a country).

### Language of communication

`PropertyFilter.Languages` is a repeated parameter (`language=de&language=en`)
holding ISO 639-1 codes from the catalogue in `data.SpokenLanguages()`. It is
**optional and OR-ed**: an empty list must not narrow anything, and a listing
matches if it is sold in *any* one of the requested languages.

`Property.Languages` carries the languages a listing is sold in. It is copied
from the seller's own "languages of communication" at publication time — the
broker's for an agency listing, the account's for a private one — deliberately
*not* joined through to the seller at query time, so a broker changing which
languages they deal in does not silently rewrite listings already published.

```sql
CREATE TABLE property_languages (
  property_id BIGINT UNSIGNED NOT NULL,
  language    CHAR(2) NOT NULL,
  PRIMARY KEY (property_id, language),
  KEY idx_lang (language)
) ENGINE=InnoDB;

-- ... AND (? = 0 OR EXISTS (
--       SELECT 1 FROM property_languages pl
--       WHERE pl.property_id = p.id AND pl.language IN (?)))
```

`handlers.languageOptions` builds the filter's checkboxes from the languages
present in the current result set rather than from the whole catalogue, so the
panel never offers a language that cannot return anything. Keep that behaviour:
a `SELECT DISTINCT language FROM property_languages` scoped to the same `WHERE`
clause is the SQL equivalent.

---

## 3. Authentication

Currently mocked: `internal/handlers/auth.go` validates field shape only, sets a
meaningless `previa_demo_session` cookie, and never checks a credential.

To productionise:

1. Hash passwords with bcrypt or argon2id.
2. Replace the demo cookie with a signed, `HttpOnly`, `Secure`, `SameSite=Lax` session.
3. Add middleware populating the request context with the user, and have
   `AccountRepository.CurrentUser` read from it instead of returning a fixture.
4. Protect `/dashboard`, `/my-listings`, `/drafts`, `/favourites`,
   `/saved-searches`, `/notifications`, `/settings`, `/billing`, `/checkout`,
   `/add-listing` and everything under `/admin`.
5. Add CSRF tokens to every mutating form. HTMX can send them via
   `hx-headers` set once on `<body>`.
6. Wire the social buttons on `/login` to real OAuth (`/auth/google`, `/auth/facebook`).

`/admin` currently has **no access control at all** and must be gated before any
deployment.

---

## 4. Property CRUD and the add-listing wizard

### The listing lifecycle

Three states, and they describe **payment, not moderation**. The client's rule:
"there is no pending review as noone will not look the ads before publishing,
and so nothing is rejected."

| Status | Meaning | Public? | Seller can |
| --- | --- | --- | --- |
| `draft` | entered, never paid for. `expires_at` is NULL | no | edit, clone, activate, delete |
| `active` | paid for, online until `expires_at` | yes | edit, clone, promote, view, delete |
| `expired` | the paid period ended | no | edit, clone, **re-activate**, delete |

`sold` sits outside the cycle — a seller marking an outcome, and what an archive
would be built from.

Draft and expired leave the seller with the same options, deliberately;
`models.ListingStatus.IsEditable()` is the single predicate for that, so the two
cannot drift apart in the templates. Activation and re-activation both go
through checkout — there is no other way into `active`.

Two consequences for the backend:

- **Nothing queues before publication.** No approval endpoint, no reviewer role
  in front of `active`, no rejection reason. Admin moderation is a *takedown*
  after the fact, from the listing row.
- **`expires_at` must be NULL for a draft.** The listings table renders an em
  dash for it. A date on something that was never online is a promise the
  listing cannot keep.

### Clone

`POST /listing/clone/{id}` → `handlers.CloneListing` →
`AccountRepository.CloneListing`. The copy is **always a draft** whatever the
original was, its metrics reset to zero, its `expires_at` NULL, and its title
suffixed so the two are tellable apart. Ownership is checked in the repository,
so a hand-edited id cannot duplicate somebody else's advertisement — keep that
check when the mock is replaced.


`internal/handlers/addlisting.go` renders 14 steps; state lives in
`localStorage` via `previaWizard()` in `web/static/js/previa.js`.

To productionise:

- `POST /add-listing/save` should upsert a `listing_drafts` row keyed by user +
  draft id, and return the same `autosave-indicator` fragment.
- Load an existing draft into the step templates on `GET /add-listing?draft=…`.
- On publish: validate server-side (the client's required-field marks are hints,
  not enforcement), create the `properties` row with `status='pending'`, and
  attach uploaded media.
- `Draft.Completion` should be computed from which required fields are filled.

**Location precision** (`models.LocationPrecision`) must be enforced when
serialising, not just in the UI:

| Value | What the public API may return |
| --- | --- |
| `exact` | Full address, exact lat/lng |
| `street` | Street name; lat/lng jittered within ~150 m |
| `area` | District name only; lat/lng snapped to district centroid |

Never send exact coordinates for `street`/`area` listings and hide them in the
client — the payload itself must be redacted.

---

## 5. Media uploads

Currently: `web/static/img/**` holds generated placeholder imagery, and the
uploader in wizard step 9 is a mock.

To productionise: accept multipart uploads, validate MIME type and dimensions,
strip EXIF (including GPS), generate derivatives (thumb 400×300, card 800×600,
full 1600×1200), store on S3-compatible object storage behind a CDN, and persist
rows in `property_images` with an explicit `sort_order`. The templates already
consume `models.Image{URL, Alt, Width, Height}`, so only the URLs change.

### Direct from the owner

`users.direct_from_owner` is a claim the seller makes in account settings, and
it is copied onto their listings as `properties.direct_from_owner` at
publication time — the same pattern as the languages of communication above, and
for the same reason.

It is **not** derivable from `seller_kind`. A private seller is someone without
an agency behind them, which is usually but not always the owner (an heir, a
landlord's relative, a company officer); and a broker can be selling their own
flat. The client asked for it precisely because the two questions differ:
"sometimes it is important to note that the property owner itself is selling
it." The listing page shows the two facts separately for that reason.

### The profile text, and a broker's chat apps

Two fields added on 19 August, both of them the same fact in two places:

- `users.bio` is the paragraph a profile shows under "About <name>", written in
  account settings (`name="about"`). `brokers.bio` already held it for the
  seeded brokers; this is where its owner edits it. The client's note: "under
  broker name is 'About' but at the moment under user profile there is no place
  the user can edit this text — create it."
- `brokers.messengers` mirrors `users.messengers` and
  `properties.messengers` — the chat apps this broker ticked, shown under the
  phone number on their profile and rendered by the same `messenger-links`
  component the listings use. Storage shape is unchanged: a kind plus an
  optional handle, per `models.Messenger`.

### Listing statistics

`AccountRepository.ListingStats(ctx, propertyID, days)` returns a per-day view
series, oldest first, for a listing **the signed-in user owns** — the ownership
check is part of the contract, not a detail of the mock.

The mock shapes a plausible series from the listing's own view count and id, so
it is deterministic between requests; a real implementation is a grouped count
over a page-view table:

```sql
SELECT DATE(viewed_at) AS day, COUNT(DISTINCT visitor_hash) AS views
FROM property_views
WHERE property_id = ? AND viewed_at >= CURDATE() - INTERVAL ? DAY
GROUP BY day ORDER BY day;
```

`COUNT(DISTINCT visitor_hash)` matters: the panel says "counted once per visitor
per day", and the lifetime `properties.views` counter it is shown beside does
not de-duplicate. If the real numbers cannot honour that sentence, change the
sentence rather than the query.

The whole series for a page of listings is embedded in the markup up front
rather than fetched when the dialog opens — a few hundred integers, and it makes
flicking between listings instant. Revisit that only if a seller can hold
hundreds of listings.

### Profile picture and company logo

Two separate uploads on Account → Settings → Profile, and they are not
interchangeable — one is a person, the other a brand. Both are optional.

| Field | Shown | Frame | Recommended source |
| --- | --- | --- | --- |
| `User.Avatar` | seller box, broker card, profile header | rounded rectangle, cropped `object-fit: cover` | 400 × 400 |
| `User.CompanyLogo` | foot of the seller box on a listing, agency header | rounded rectangle, **never cropped** — `object-fit: contain` on white | 400 × 120 |

The client's instruction on the accepted formats is that the interface should
not name any: "our system accepts all the image types, and all of them are
converted into .webp anyway". So validate by sniffing the decoded image rather
than by extension, accept anything the decoder accepts, and convert on ingest.
The copy under each control says only the recommended dimensions; keep it that
way if the conversion pipeline changes.

A logo is somebody else's artwork. Do not re-crop it, do not tint it, and do not
composite it onto a themed background — `.logo-tile` and `.seller-logo` both
paint white in either theme for that reason.

---

## 6. Google Maps

Config: `PREVIA_MAPS_API_KEY` (see `internal/config/config.go`). Never committed.

- **With a key:** `previaMap()` loads the official Maps JavaScript API and
  renders into `.map-canvas`.
- **Without a key:** the mock map in `web/templates/components/map.html` renders
  instead — same markers, same clustering threshold, same card↔marker
  synchronisation, same preview popup.

Marker data is built server-side by `buildMapConfig` in
`pkg/handlers/search.go`, so the map is populated on first paint rather than
after an XHR. It carries **two** marker sets: `points` (listings, drawn as price
bubbles) and `brokers` (paid map placements, drawn as a rounded-rectangle
photograph and a name). The second is empty unless the reader ticked Brokers in
the filter panel — see `PropertyFilter.ShowBrokers` and
`BrokerRepository.OnMap`. When moving to live tiles:

- Restrict the key by HTTP referrer and to the Maps JavaScript API only.
- Move clustering to `@googlemaps/markerclusterer`; the `> 40` threshold is
  already the mock's behaviour.
- Add viewport-bounded search (`?bounds=n,s,e,w`) so panning refetches.

---

## 7. Geolocation and country selection

`SetCountry` in `internal/handlers/public.go` writes a `previa_country` cookie.
Priority is: manual choice → geolocation (only after an explicit click) → the
`PREVIA_DEFAULT_COUNTRY` default. The browser is never prompted on page load,
and a denied permission falls back to manual selection.

For production, consider a GeoIP lookup for the *first* visit only, and keep the
manual choice authoritative thereafter.

---

## 8. Payments

Mocked in `CheckoutProcess` (`internal/handlers/account.go`), which maps the
chosen method onto a deterministic outcome so every state is reachable.

Real integration:

| Provider | Flow |
| --- | --- |
| Stripe | Create a PaymentIntent server-side, confirm with Stripe Elements, settle on `payment_intent.succeeded` webhook |
| PayPal | Create an order, redirect, capture on return, verify via webhook |
| Paysera | Signed request to the bank-link gateway, verify the signed callback |

Rules: never trust a client-reported success — only a verified webhook may mark
a payment paid and a listing live; make webhook handling idempotent; store
`provider_reference` per payment; generate invoice PDFs server-side.

The five UI states (`select`, `processing`, `success`, `failed`, `cancelled`) are
already built and only need the real outcome wired in.

### Broker placements

Two products, sold from the profile (`web/templates/pages/account/settings.html`)
and priced by `Catalog.BrokerAdPlan` / `Catalog.BrokerMapAdPlan`. They are
separate purchases and buying one does not grant the other.

| Placement | Model | Bought | Where it shows |
| --- | --- | --- | --- |
| Market strip | `models.BrokerAd` → `[]BrokerAdRun` | **Per market, per run** | The homepage broker section for that market, and `/brokers` while nothing has been searched |
| Search map | `models.BrokerMapAd` | Once, per run | The search map's pins and the broker cards beside the results, wherever a visitor ticks Brokers in the filters |

The market strip is a **list of runs, not a list of countries**. The client's
rule is that each market is activated and paid for on its own day — "at first he
wants that his profile is displayed in the German market for 30 days, then he
activates it with payment, and then he can activate his ad under France market
as well for 30 days with new payment" — so each run carries its own
`StartsAt`/`EndsAt` and expires independently. A schema that stored one row per
ad with a country list could not express two markets ending a fortnight apart.

Suggested tables:

```sql
broker_ad_run  (id, broker_id, country_code, days, starts_at, ends_at, payment_id)
broker_map_ad  (id, broker_id, days, starts_at, ends_at, payment_id)

-- The price list the strip bills from, edited in Admin → Price packages.
-- One row per market that has been given a rate; anything not in here is
-- charged the default below.
broker_ad_rate (country_code CHAR(2) PRIMARY KEY, per_day DECIMAL(8,2) NOT NULL)
```

### Pricing (19 August)

The market strip is **priced per market, per day**, not in run-length tiers:

> "In the backend there is option to set the price per day for each country. For
> example in Germany 3 € per day, in Poland 1 € per day. So the broker choosess
> the Germany and Poland for 10 days. So the system will calculate a bill:
> Germany 3 € per day x 10 = 30 €; for Estonia 1 € per day x 10 = 10 €; total:
> 40 €."

So the bill is `Σ rate(country) × days` — one line per market, each at its own
rate — and `models.BrokerAdPlan` carries `Rates`, `DefaultPerDay` and the
`DayOptions` shortcuts rather than a `Tiers` ladder. `BrokerAdPlan.PerDay(code)`
is the lookup; recompute the total server-side at checkout rather than trusting
the figure the dialog posted.

The **map placement is the other way round**: a ladder of fixed runs, which is
how the client priced it — "5 days - 1 €, 10 days - 2 €, 20 days - 3 € etc" — so
`BrokerMapAdPlan.Tiers` stays as it was. There is nothing to multiply: what is
bought is one pin on one map.

Every period is quoted to the broker in **UTC** (`view.UTC`), because the site
is global; store `starts_at`/`ends_at` in UTC and convert nowhere.

Rules:

- A purchase creates a **new row**; it never extends an existing one in place,
  or the history of what was paid for is lost.
- Liveness is `now < ends_at` and nothing else — expiry is not a job that has to
  run. `BrokerAd.RunsIn(code, now)` and `BrokerMapAd.IsLive(now)` are the only
  two predicates the templates use.
- Neither row stores a photograph, a name or a phone number. The placement
  points at the profile and the strip renders whatever the profile says now,
  which is what makes the client's "if he updates his profile by changing photo
  or phone, then this will be updated in this ad immediately" true without any
  propagation step. **Do not denormalise the profile into these tables.**
- The map placement additionally requires a pin (`Broker.Office.IsSet()`); see
  `Broker.IsOnMap`. A broker who buys it before dropping a pin simply does not
  appear, and the profile says so.
- The homepage shows **the eight newest purchases** in a market and no more, in
  `starts_at DESC` order — "in the frontpage there are only 2 rows broker ads …
  if next ad will come then the last one will be pushed futher till it
  disappears from the frontpage". `/brokers` applies no such cap: "in the broker
  page it stays till to the end of payd periode." Both read the same
  `Brokers.Promoted(ctx, country, limit)`; only the limit differs, so a backend
  must order on the purchase date rather than on the expiry.

The forms post `broker_ad`/`broker_ad_countries`/`broker_ad_days` and
`broker_map_ad`/`broker_map_ad_days`; the admin price list posts
`broker_ad_default_per_day` and `broker_ad_rate_<CC>` per market. Nothing is
persisted this milestone.

---

## 9. Notifications, favourites, saved searches

- `ToggleFavourite` is already an HTMX endpoint returning the refreshed button
  fragment — swap the in-memory map for a `favourites` upsert/delete.
- Saved searches store the raw query string, so a matcher job can re-run
  `ParseFilter` against new listings and enqueue notifications per
  `SavedSearch.Frequency`.
- `POST /save-search` **writes**: it parses the posted filter form, resolves the
  location, counts the matches and calls `AccountRepository.AddSavedSearch`,
  which is the one insert to implement (`saved_searches`). The handler builds
  the row's `Name` and `Summary` from `PropertyFilter.Chips()`, so a saved
  search reads exactly as the tag bar above the results did. `compact=1` on the
  request chooses the one-line confirmation the filter panel's footer shows;
  without it the endpoint returns the full alert the results header uses.
- Anything replaying a stored query must build the URL with `view.SearchURL`,
  not `href="/search?{{ .Query }}"` — html/template escapes an interpolated
  query as a single parameter, which made every saved search run unfiltered.
- `UnreadCount` drives the header badge; keep it cheap (a counter column or a
  cached count).

---

## 10. Languages, translations and SEO

Admin screens exist at `/admin/languages`, `/admin/strings` and `/admin/seo`.

- English is the default **and** the fallback: when a translation is missing the
  English source string renders. `models.TranslationString.IsMissing()` already
  expresses this.
- URLs are language-prefixed (`/en/`, `/de/`, `/es/`). `activeLang` in
  `handlers.go` reads the prefix, then a cookie, then defaults to English.
  **Not yet done:** the router does not register prefixed routes — add a
  language-stripping middleware in front of the mux.
- `Meta.Alternates` already emits hreflang for every enabled language.
- `SEOEntry` rows are per path *and* language, and should override the
  handler-supplied defaults when present.
- `/sitemap.xml` and `/robots.txt` are generated in `content.go` from the
  catalogue; keep them in step with published content.

---

## 11. Restricted countries

`/admin/restricted` manages `models.RestrictedCountry`. Nothing is hardcoded —
the seeded entries are data, not logic.

Enforcement to add server-side:

1. Exclude restricted countries from `Countries()` in the selector.
2. Drop their listings from every search result.
3. Return the `map-restricted` state for direct URL access.
4. Refuse new listings in those markets in the wizard.

---

## 12. Admin operations — deliberately not implemented

These screens are frontend demonstrations. `AdminMockAction` performs **no**
work and returns a fixed "simulation only" fragment. Before any of them do
something real, each needs authentication, authorisation, an audit-log entry,
rate limiting and a confirmation step.

| Screen | What production must add |
| --- | --- |
| Backups | Scheduled `mysqldump` + file archive, Google Drive upload, retention pruning, restore behind a second confirmation |
| File manager | A jailed root, path-traversal rejection, MIME allowlist, no execute permissions |
| MySQL manager | Read-only credentials, statement allowlist, hard `LIMIT`, query timeout, full audit |
| Cache | Real cache invalidation |
| Restart | A supervisor/systemd signal with connection draining — never `exec` from a request handler |

`SystemInfo.BinaryBuiltAt` is stamped at process start in
`internal/data/seed_admin.go`; production should inject a real build timestamp
via `-ldflags "-X main.buildTime=…"`.

---

## 13. Suggested schema sketch

```sql
CREATE TABLE properties (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  slug          VARCHAR(220) NOT NULL UNIQUE,
  title         VARCHAR(200) NOT NULL,
  deal_type     ENUM('sale','rent') NOT NULL,
  property_type ENUM('apartment','house','villa','commercial','land','garage') NOT NULL,
  status        ENUM('draft','active','expired','sold') NOT NULL DEFAULT 'draft',
  price         DECIMAL(12,2) NOT NULL,
  currency      CHAR(3) NOT NULL DEFAULT 'EUR',
  country_code  CHAR(2) NOT NULL,
  city          VARCHAR(120) NOT NULL,
  district      VARCHAR(120),
  address       VARCHAR(255),
  postal_code   VARCHAR(20),
  lat           DECIMAL(10,7),
  lng           DECIMAL(10,7),
  precision_    ENUM('exact','street','area') NOT NULL DEFAULT 'exact',
  rooms         SMALLINT, bedrooms SMALLINT, bathrooms SMALLINT,
  area          DECIMAL(10,2), land_area DECIMAL(12,2),
  floor         SMALLINT, total_floors SMALLINT, build_year SMALLINT,
  condition_    ENUM('new','renovated','good','satisfying','needs_work'),
  energy_rating CHAR(1),
  description   TEXT,
  seller_kind   ENUM('broker','private') NOT NULL DEFAULT 'broker',
  broker_id     BIGINT UNSIGNED NULL,
  agency_id     BIGINT UNSIGNED NULL,
  development_id BIGINT UNSIGNED NULL,
  is_featured   TINYINT(1) NOT NULL DEFAULT 0,
  direct_from_owner TINYINT(1) NOT NULL DEFAULT 0,
  views         INT UNSIGNED NOT NULL DEFAULT 0,
  created_at    DATETIME NOT NULL,
  updated_at    DATETIME NOT NULL,
  expires_at    DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

Boolean amenities (`furnished`, `parking`, `balcony`, `garden`, `elevator`,
`terrace`, `sauna`, `sea_view`) are best kept as columns rather than a join
table — they are all filterable, they are what the card and the detail page draw
as icons, and the set is small and stable.

The rest of the "Features and amenities" catalogue is not: it is thirty-four
ticks in five groups and the client adds to it, so it belongs in a join table.

```sql
CREATE TABLE listing_amenities (
  listing_id BIGINT UNSIGNED NOT NULL,
  amenity    VARCHAR(40) NOT NULL,   -- models.AmenityGroups keys, e.g. "coffee-maker"
  PRIMARY KEY (listing_id, amenity),
  KEY (amenity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

The catalogue itself lives in `models.AmenityGroups` and is rendered by both the
add-listing form and the search sidebar — one list, so a seller's tick and a
buyer's filter are guaranteed to be the same key. The eight columns above keep
their own query parameters (`furnished=1`, `parking=1`, …) so URLs written
before the full catalogue existed still resolve; everything else travels as
repeated `amenity=<key>` and is matched with AND.

The seller's own profile, added on 17 August:

```sql
ALTER TABLE users
  ADD company           VARCHAR(160) NULL,   -- "Best House Ltd"; shown beside the name
  ADD company_logo      VARCHAR(255) NULL,   -- see §5
  ADD avatar            VARCHAR(255) NULL,
  -- Copied onto each listing at publication time; see §4.
  ADD direct_from_owner TINYINT(1) NOT NULL DEFAULT 0;

-- Languages of communication. Codes from data.SpokenLanguages().
CREATE TABLE user_languages (
  user_id  BIGINT UNSIGNED NOT NULL,
  language CHAR(2) NOT NULL,
  PRIMARY KEY (user_id, language)
) ENGINE=InnoDB;

-- Markets the seller works in. A broker near a border is active in more than
-- one, which is why this is a table and users.country_code is still a column:
-- the column is the account's home market and drives the interface default,
-- these rows are where they actually trade and are shown as "Active in".
CREATE TABLE user_countries (
  user_id      BIGINT UNSIGNED NOT NULL,
  country_code CHAR(2) NOT NULL,
  sort_order   SMALLINT NOT NULL DEFAULT 0,  -- home market first
  PRIMARY KEY (user_id, country_code)
) ENGINE=InnoDB;

-- Per-day view counts behind the statistics panel. Hash the visitor rather than
-- storing an address: the panel promises "counted once per visitor per day",
-- and that is the only column needed to keep the promise.
CREATE TABLE property_views (
  property_id  BIGINT UNSIGNED NOT NULL,
  viewed_at    DATETIME NOT NULL,
  visitor_hash BINARY(16) NOT NULL,
  KEY idx_pv_day (property_id, viewed_at)
) ENGINE=InnoDB;

-- Chat apps, the same shape a listing already stores. WhatsApp and Viber are
-- addressed by users.phone and carry no handle; Telegram and Signal each hold
-- their own link, Teams an email address.
CREATE TABLE user_messengers (
  user_id BIGINT UNSIGNED NOT NULL,
  kind    ENUM('whatsapp','telegram','viber','signal','teams') NOT NULL,
  handle  VARCHAR(255) NULL,
  PRIMARY KEY (user_id, kind)
) ENGINE=InnoDB;
```

`models.Messenger.Link(phone)` builds every deep link from these two columns and
already handles a pasted full URL, a bare username and a bare number, so the
backend only has to store what the seller typed.

---

## 14. Deployment

- Build: `go build -o previa ./cmd/previa`. The binary needs `web/` beside it,
  or set `PREVIA_TEMPLATE_DIR` / `PREVIA_STATIC_DIR`. Embedding via `embed.FS`
  is a reasonable later step.
- Set `PREVIA_DEV=0` in production so templates compile once at boot and a
  broken template fails startup rather than a request.
- Environment: `PREVIA_ADDR`, `PREVIA_BASE_URL`, `PREVIA_DEFAULT_COUNTRY`,
  `PREVIA_MAPS_API_KEY`, plus the future `PREVIA_DSN`.
- Terminate TLS at a reverse proxy, add HSTS, and set a Content-Security-Policy.
  The app ships no inline event handlers, but the theme bootstrap in
  `layouts/base.html` is an inline script and needs a nonce or a hash.
- Static assets are served with long cache lifetimes; add content hashing to
  filenames before enabling immutable caching on CSS and JS.
