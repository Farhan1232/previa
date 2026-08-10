# Previa — frontend milestone

A global real-estate marketplace: property for sale and rent across eight
European markets, with map-first search, verified broker profiles, a guided
add-listing flow, a user dashboard and a full administration panel.

This repository is the **frontend milestone**. Every screen is real,
server-rendered HTML driven by a centralised mock data layer. There is no
database, no real authentication and no payment processing — see
[Frontend-only limitations](#frontend-only-limitations).

- **Domain:** previa.estate
- **Stack:** Go · `html/template` · HTMX · Alpine.js · hand-written CSS
- **Reference audit:** [`docs/frontend-reference-audit.md`](docs/frontend-reference-audit.md)
- **Backend handover:** [`docs/backend-integration-points.md`](docs/backend-integration-points.md)

---

## Running it

Requires Go 1.22 or newer. Nothing else — no Node, no build step, no network.

```bash
git clone <this repo> previa && cd previa
go run ./cmd/previa
# → http://localhost:8080
```

During development, recompile templates on every request:

```bash
PREVIA_DEV=1 go run ./cmd/previa
```

Production build:

```bash
go build -o previa ./cmd/previa
./previa
```

Templates are compiled into the binary with `go:embed`, so it runs from
anywhere. Only the static assets are read from disk — point
`PREVIA_STATIC_DIR` at them if you run the binary outside the repo.

### Environment variables

| Variable | Default | Purpose |
| --- | --- | --- |
| `PREVIA_ADDR` | `:8080` | Listen address |
| `PREVIA_BASE_URL` | `https://previa.estate` | Canonical/hreflang base |
| `PREVIA_DEV` | `false` | Recompile templates per request; show template errors |
| `PREVIA_MAPS_API_KEY` | *(empty)* | Google Maps JavaScript API key |
| `PREVIA_DEFAULT_COUNTRY` | `EE` | Market shown before a visitor chooses |
| `PREVIA_TEMPLATE_DIR` | `web/templates` | Template root |
| `PREVIA_STATIC_DIR` | `public/static` | Static asset root |

**No secret is committed.** The Maps key is read from the environment only, and
never rendered into a template or stored in the data layer.

---

## Architecture

```
cmd/previa/main.go          entry point; wires config → data → view → handlers
internal/
  config/                   environment configuration
  models/                   domain types (Property, Broker, Article, …)
  data/                     repository interfaces + in-memory mock provider
    repository.go           the contract the backend must satisfy
    mock.go                 in-memory implementation
    seed_*.go               seed data (properties, content, admin)
  view/                     template engine + template function map
  handlers/                 HTTP handlers, one file per area
web/
  templates/
    layouts/base.html       document shell, theme bootstrap, meta
    partials/               header, footer, account nav, admin nav
    components/             logo, icons, cards, filters, map, results
    pages/                  one file per route (+ account/, admin/, auth/)
public/
  static/
    css/                    tokens → base → components → layout → pages
    js/previa.js            Alpine factories + the little glue that needs JS
    vendor/                 pinned HTMX and Alpine (no CDN)
    fonts/                  self-hosted DM Serif Display + Manrope (woff2)
    img/                    photography + generated WebP variants
docs/                       reference audit, backend integration points
```

### Data flow

```
request → handler → data.Store (interfaces) → PageData{…, Data: <page payload>}
                                            → view.Engine → html/template → HTML
```

Handlers never touch the mock directly; they depend on the interfaces in
`internal/data/repository.go`. Swapping in MySQL is one line in `main.go`.

### Template engine

Go's `html/template` has no inheritance, so each page is compiled as its own
set: `layouts/*` + `components/*` + `partials/*` + that one page file. Sets are
cached at boot (and recompiled per request when `PREVIA_DEV=1`, so template
edits show up without a restart).

Rendering goes through a buffer first — a template error mid-write can never
emit half a page on top of an already-sent `200`. In production a broken
template fails at **boot**, not on a visitor's request.

Components take named arguments through the `dict` helper:

```gotemplate
{{ template "property-card" dict "Property" $p "Variant" "compact" "Favourite" true }}
```

---

## Mock data

Everything comes from `data.NewStore(time.Now())`:

- **36 properties** across Estonia, Germany, Spain, Finland, Portugal, the
  Netherlands, Austria and Czechia — for sale and to rent, spanning apartments,
  houses, villas, commercial units, land and garages, with realistic titles,
  addresses, prices, descriptions and coordinates
- **14 brokers** across **8 agencies**, with credentials, languages and specialisms
- **8 developments**, **10 articles**, **4 packages**
- Favourites, saved searches, notifications, drafts, payments
- Country-specific banners and promoted brokers for every market
- Admin statistics, users, translations, SEO entries, backups, tables, system info

No Lorem Ipsum anywhere.

### Photography

`public/static/img/` holds **182 real photographs**, downloaded from Unsplash and
stored locally (nothing is hotlinked at runtime):

| Set | Count | Subject |
| --- | --- | --- |
| `properties/p001–p060` | 60 | Building exteriors — apartment blocks, houses, villas, European facades |
| `properties/p061–p130` | 70 | Interiors — living rooms, kitchens, bedrooms, bathrooms, offices |
| `properties/p131–p140` | 10 | Land and open plots |
| `brokers/br-01…br-14` | 14 | Professional portraits, face-cropped, gender-matched to each broker's name (7 women, 7 men) |
| `banners/city-*` | 8 | The actual cities — Tallinn, Berlin, Barcelona, Helsinki, Lisbon, Amsterdam, Vienna, Prague |
| `developments/`, `articles/`, `hero*` | 20 | New-build exteriors, key/handover shots, luxury hero images |

Galleries are assembled by subject, not by hand-kept indices — see
`buildGallery` in `internal/data/seed_properties.go`. Every listing leads with
an exterior, rooms follow, and a plot of land shows land rather than a bathroom.
The generator only needs the pool layout above to stay true.

**Licensing:** Unsplash photos are free to use commercially and
non-commercially without permission under the [Unsplash
License](https://unsplash.com/license); attribution is appreciated but not
required. They are fine for demonstrating the build. Before launch, replace
them with the client's own licensed photography — particularly the broker
portraits, which currently show real people who are not Previa staff. Nothing
in the templates changes: they consume `models.Image{URL, Alt, Width, Height}`.

---

## Design system

### Tokens — `public/static/css/tokens.css`

Every colour, size, space, radius, shadow and duration is a custom property.
Components never hardcode a raw value.

| Role | Light | Dark |
| --- | --- | --- |
| Primary navy | `#0C2D48` | inverts to `#E9EFF3` for buttons |
| Interactive teal | `#2F6F68` | `#5A9D94` |
| Gold accent | `#C39A5B` | `#D1AA68` |
| Page background | `#F6F3EE` | `#081722` |
| Surface | `#FFFFFF` | `#102635` |
| Text | `#172B3A` | `#E8EDF1` |
| Border | `#DDE3E7` | `#294353` |

Roughly 60% warm off-white / white, 30% navy and neutrals, 10% accent. Teal
carries selected and interactive states; gold is restricted to featured
badges, premium packages and small accents — never large fills or long text.

### Typography

- **DM Serif Display** — marketing and page headings only
- **Manrope** — everything else: navigation, body, prices, forms, tables, admin

Both self-hosted as latin/latin-ext `woff2` subsets (SIL Open Font License).
No runtime dependency on Google Fonts. Sizes are fluid via `clamp()`.

DM Serif Display ships a single weight, so serif headings are always 400 —
requesting a bold would make the browser synthesise one.

### Spacing and layout

8px system, 1280px content width, 32/24/16px gutters, 24px card gap. Sections
breathe at 80–120px on desktop and 48–64px on mobile. Content is centred; map
and search screens may use the full viewport.

### Themes

The theme is applied by an inline script in `<head>` **before first paint**, so
there is no flash. The choice persists in `localStorage` and defaults to the
system preference. Every screen is designed in both.

---

## HTMX, Alpine.js and JavaScript

Local pinned copies only — nothing loads from a CDN.

**HTMX** handles server-rendered fragments: filter updates, sorting, favourite
toggles, saved searches, contact forms, contact reveal, wizard autosave and the
admin mock actions. A filter change swaps only `#results`, so the filter panel,
scroll position and map viewport all survive.

**Alpine.js** handles browser-side state: dropdowns, modals, the mobile drawer,
the collapsible filter panel, gallery and lightbox, tabs, theme switching,
map marker/card synchronisation and wizard autosave state.

**Load order matters.** Alpine's build calls `start()` in a microtask as soon as
it executes, so `previa.js` (component factories) and the focus/collapse plugins
must be loaded *before* it. Getting this wrong makes every `x-data` expression
silently resolve against `window`.

**Plain JavaScript** is confined to `public/static/js/previa.js` and used only
where nothing else fits: theme persistence, focus and scroll bookkeeping around
overlays, geolocation, and the Google Maps integration with its offline
fallback.

---

## Google Maps

Two modes, both fully built:

- **Split view** (`/search?view=map`) — results left, map right, collapsible
  panel with an edge button to reopen, two-way card↔marker highlighting
- **Full-screen** (`/search?view=full`) — map fills the content area with a
  floating toolbar

Marker data is serialised server-side by `buildMapConfig`, so the map is
populated on first paint rather than after an XHR.

### With no API key

If `PREVIA_MAPS_API_KEY` is unset, Previa renders its own mock map: land, water,
parks, a road network, and the same navy price markers. Clustering (above 40
markers), the preview popup with its photo gallery, marker↔card synchronisation
and all the loading/error/restricted states behave identically. The whole
interaction can be reviewed without a billable key, and switching to live tiles
changes nothing else on the page.

---

## Responsive behaviour

Tested at 1440×900, 1280×800, 1024×768, 768×1024, 430×932, 390×844 and 360×800.

**Header** — compact and sticky. Below 1080px the desktop nav is replaced by a
right-side off-canvas drawer that overlays the page (never shifts it), traps
focus, closes on Escape, backdrop click or the close button, and restores focus
to the trigger.

**Filter sidebar** — the client's key requirement:

| Viewport | Behaviour |
| --- | --- |
| > 1024px | Docked left, open by default, results beside it. Collapsing widens the results and leaves a labelled rail button to reopen. The collapsed state persists. |
| ≤ 1024px | The same panel becomes an off-canvas drawer, closed by default, opened from a fixed left-edge tab that stays reachable while scrolling. |

There is exactly **one** filter form in the DOM at any width, so no filter value
is ever lost when the layout changes.

**Grid density** — homepage sections run 3 → 2 → 1. Search results are
auto-filling and reach 5–6 columns on very wide screens without dropping below a
readable card width. Map-side results use compact horizontal cards.

---

## Accessibility

Targets WCAG 2.1 AA.

- Semantic landmarks, correct heading order, skip link
- Visible keyboard focus everywhere; focus is trapped in drawers, modals and the
  lightbox, and restored on close
- Permanent visible labels on every field — placeholders are never the only label
- Errors are announced with `aria-invalid` + `aria-describedby`, and status is
  carried by text and icon, not colour alone
- Icon-only controls all carry accessible names
- `prefers-reduced-motion` disables animation and smooth scrolling
- Live regions on result counts and async feedback

---

## Routes

**Public** — `/` · `/search` (grid, list, split map, full map) · `/property/{slug}` ·
`/developments` · `/development/{slug}` · `/brokers` · `/broker/{slug}` ·
`/agencies` · `/agency/{slug}` · `/articles` · `/article/{slug}` · `/pricing` ·
`/help` · `/about` · `/advertising` · `/terms` · `/privacy` · `/cookies` ·
`/robots.txt` · `/sitemap.xml` · 404 · error

**Auth (mocked)** — `/login` · `/register` · `/forgot-password` · `/logout`

**Account** — `/dashboard` · `/my-listings` · `/drafts` · `/favourites` ·
`/saved-searches` · `/notifications` · `/settings` · `/billing` · `/checkout`
(select / processing / success / failed / cancelled)

**Add listing** — `/add-listing?step=1…14`

**Admin** — `/admin` plus listings, users, brokers, agencies, developments,
articles, banners, packages, payments, languages, strings, seo, maps,
restricted, settings, backups, files, database, system

All 73 routes plus every generated detail page return the expected status.

---

## UI states

Built and reachable, not just the happy path: loading, skeleton, empty, error,
success, disabled, active, selected, validation error, permission denied, no
search results, map unavailable, restricted country, missing translation,
payment failed, payment cancelled, draft saving, draft saved, network error,
404, and confirmation dialogs.

---

## Frontend-only limitations

Deliberately **not** implemented in this milestone:

- MySQL — all data is in memory and resets on restart
- Real authentication — any valid-looking email and an 8-character password
  "signs in"; the session cookie carries no identity and grants nothing
- Real payments — no provider is contacted and no card details are collected
- Email and SMS — nothing is sent
- File uploads — the uploader is a mock; no file is written
- Backups, file manager, MySQL manager, cache clearing, application restart —
  every one of these is a frontend demonstration. `AdminMockAction` performs no
  work and returns a fixed "simulation only" response.

`/admin` has **no access control** and must be gated before deployment.

---

## Deployment

The application runs two ways from the same wiring in `internal/app`:

- **`cmd/previa`** — a long-lived HTTP server.
- **`api/index.go`** — a Vercel serverless function. Vercel serves everything
  in `public/` from its CDN and `vercel.json` rewrites only what the filesystem
  does not resolve, so static assets never invoke the function.

Templates are embedded with `go:embed`, and the image manifest
(`internal/assets`) is generated Go source rather than a directory scan,
because a serverless bundle ships the binary without the repository beside it.

After adding or regenerating imagery, refresh the manifest:

```bash
go run ./internal/assets/cmd/genmanifest
```

---

## Verification

```bash
go build ./...     # compiles clean
go vet ./...       # no findings
```

Checked in a real browser across the seven breakpoints above, in both themes:
no horizontal overflow, no broken images, no console errors, and no layout
breakage on any captured screen.
