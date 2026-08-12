// Package handlers contains Previa's HTTP layer.
//
// Handlers read from the data.Store interfaces only, so replacing the mock
// provider with MySQL requires no changes here.
package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"previa/pkg/config"
	"previa/pkg/data"
	"previa/pkg/models"
	"previa/pkg/view"
)

// Handler bundles the dependencies every route needs.
type Handler struct {
	Store *data.Store
	View  *view.Engine
	Cfg   config.Config
}

// New builds a Handler.
func New(store *data.Store, v *view.Engine, cfg config.Config) *Handler {
	return &Handler{Store: store, View: v, Cfg: cfg}
}

// ---------------------------------------------------------------------------
// Page data
// ---------------------------------------------------------------------------

// Alternate is one hreflang entry.
type Alternate struct {
	Lang string
	Href string
}

// Meta carries per-page SEO values.
type Meta struct {
	Title       string
	Description string
	Canonical   string
	OGImage     string
	OGType      string
	NoIndex     bool
	Alternates  []Alternate
}

// PageData is the root object every template receives. Page-specific payloads
// hang off Data.
type PageData struct {
	Meta        Meta
	Nav         string
	Lang        string
	LangName    string
	CurrentPath string
	Query       url.Values

	Country   models.Country
	Countries []models.Country
	// AllCountries is every selectable market. OtherCountries is that list
	// without the seeded ones, which is what the selector renders below its
	// "With listings" group. Countries stays the short seeded set, still used
	// where a compact list is wanted (the drawer's country pills, the city
	// datalist).
	AllCountries   []models.Country
	OtherCountries []models.Country
	Languages      []models.Language
	Banner         models.Banner
	HasBanner      bool

	// MarketInBanner moves the country selector out of the header and onto the
	// page's banner. Set by pages that render one (home, search); everywhere
	// else the header keeps the control so a visitor is never stuck on the
	// wrong market with no way to change it.
	MarketInBanner bool

	// BannerDismissed reports that the sponsored strip is switched off for the
	// active market. The banner stays configured and re-activatable in admin.
	BannerDismissed bool

	User        models.User
	UnreadCount int

	MapsKey string
	HasMaps bool
	// NeedsMap loads Leaflet only on pages that actually render a map, so the
	// homepage and account screens do not pay for it.
	NeedsMap bool
	Now      time.Time

	Data any
}

// base assembles the shared page context.
func (h *Handler) base(r *http.Request, nav, title, desc string) PageData {
	ctx := r.Context()

	country := h.activeCountry(r)

	// The homepage slot is the default for PageData.Banner. It reports
	// hasBanner=false while switched off, which is the current state — the
	// strip stays configured but is not drawn over the hero.
	banner, hasBanner := h.Store.Catalog.BannerFor(ctx, country.Code, "home")

	path := r.URL.Path
	canonical := strings.TrimSuffix(h.Cfg.BaseURL, "/") + path

	langs := h.Store.Catalog.Languages(ctx)
	lang := h.activeLang(r, langs)

	var alts []Alternate
	for _, l := range langs {
		if l.IsEnabled {
			alts = append(alts, Alternate{
				Lang: l.Code,
				Href: strings.TrimSuffix(h.Cfg.BaseURL, "/") + "/" + l.Code + path,
			})
		}
	}

	return PageData{
		Meta: Meta{
			Title:       title,
			Description: desc,
			Canonical:   canonical,
			OGType:      "website",
			Alternates:  alts,
		},
		Nav:            nav,
		Lang:           lang,
		LangName:       langName(langs, lang),
		CurrentPath:    path,
		Query:          r.URL.Query(),
		Country:        country,
		Countries:      h.Store.Catalog.Countries(ctx),
		AllCountries:   h.Store.Catalog.AllCountries(ctx),
		OtherCountries: h.Store.Catalog.OtherCountries(ctx),
		Languages:      langs,
		Banner:         banner,
		HasBanner:      hasBanner,
		User:           h.Store.Account.CurrentUser(ctx),
		UnreadCount:    h.Store.Account.UnreadCount(ctx),
		MapsKey:        h.Cfg.MapsKey,
		HasMaps:        h.Cfg.MapsKey != "",
		Now:            time.Now(),
	}
}

// activeCountry resolves the market from the cookie, falling back to the
// configured default. A manual choice always wins; geolocation only ever
// writes this cookie after an explicit user action.
func (h *Handler) activeCountry(r *http.Request) models.Country {
	code := h.Cfg.DefaultCountry
	if c, err := r.Cookie("previa_country"); err == nil && c.Value != "" {
		code = c.Value
	}
	if c, ok := h.Store.Catalog.Country(r.Context(), code); ok {
		return c
	}
	all := h.Store.Catalog.Countries(r.Context())
	if len(all) > 0 {
		return all[0]
	}
	return models.Country{Code: "EE", Name: "Estonia", Currency: "EUR"}
}

// activeLang resolves the UI language from the URL prefix or cookie, defaulting
// to English. Missing translations fall back to English at render time.
func (h *Handler) activeLang(r *http.Request, langs []models.Language) string {
	if seg := firstSegment(r.URL.Path); seg != "" {
		for _, l := range langs {
			if l.IsEnabled && l.Code == seg {
				return l.Code
			}
		}
	}
	if c, err := r.Cookie("previa_lang"); err == nil && c.Value != "" {
		for _, l := range langs {
			if l.IsEnabled && l.Code == c.Value {
				return l.Code
			}
		}
	}
	return "en"
}

func firstSegment(p string) string {
	p = strings.TrimPrefix(p, "/")
	if i := strings.Index(p, "/"); i >= 0 {
		return p[:i]
	}
	return p
}

func langName(langs []models.Language, code string) string {
	for _, l := range langs {
		if l.Code == code {
			return l.Name
		}
	}
	return "English"
}

// isHTMX reports whether the request came from an HTMX swap.
func isHTMX(r *http.Request) bool { return r.Header.Get("HX-Request") == "true" }

// multiFunc adapts url.Values for data.ParseFilter's repeatable-key lookup.
func multiFunc(q url.Values) func(string) []string {
	return func(key string) []string { return q[key] }
}
