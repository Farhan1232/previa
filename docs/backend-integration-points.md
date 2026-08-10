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
| `AccountRepository` | `CurrentUser`, `Favourites`, `IsFavourite`, `ToggleFavourite`, `SavedSearches`, `Notifications`, `UnreadCount`, `MyListings`, `Drafts`, `Payments` | `users`, `favourites`, `saved_searches`, `notifications`, `listing_drafts`, `payments` |
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

---

## 6. Google Maps

Config: `PREVIA_MAPS_API_KEY` (see `internal/config/config.go`). Never committed.

- **With a key:** `previaMap()` loads the official Maps JavaScript API and
  renders into `.map-canvas`.
- **Without a key:** the mock map in `web/templates/components/map.html` renders
  instead — same markers, same clustering threshold, same card↔marker
  synchronisation, same preview popup.

Marker data is built server-side by `buildMapConfig` in
`internal/handlers/search.go`, so the map is populated on first paint rather
than after an XHR. When moving to live tiles:

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

---

## 9. Notifications, favourites, saved searches

- `ToggleFavourite` is already an HTMX endpoint returning the refreshed button
  fragment — swap the in-memory map for a `favourites` upsert/delete.
- Saved searches store the raw query string, so a matcher job can re-run
  `ParseFilter` against new listings and enqueue notifications per
  `SavedSearch.Frequency`.
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
  status        ENUM('active','pending','draft','expired','rejected','sold') NOT NULL DEFAULT 'pending',
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
  views         INT UNSIGNED NOT NULL DEFAULT 0,
  created_at    DATETIME NOT NULL,
  updated_at    DATETIME NOT NULL,
  expires_at    DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

Boolean amenities (`furnished`, `parking`, `balcony`, `garden`, `elevator`,
`terrace`, `sauna`, `sea_view`) are best kept as columns rather than a join
table — they are all filterable and the set is small and stable.

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
