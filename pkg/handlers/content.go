package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"previa/pkg/data"
	"previa/pkg/models"
)

// ListData is a generic listing payload for the index pages.
type ListData struct {
	Developments []models.Development
	Brokers      []models.Broker
	Agencies     []models.Agency
	Articles     []models.Article
	Packages     []models.Package
	Categories   []string
	Category     string
	Countries    []models.Country
	FilterQ      string
	FilterLang   string
	Languages    []LanguageOption

	// The broker directory's map search, which replaced its country and city
	// selects. FilterLocation is redisplayed in the Location box, FilterRadius
	// is the chosen distance, and LocationSuggestions backs the box's
	// autocomplete — the same control and the same catalogue the property
	// search uses, so a place found there is findable here.
	FilterLocation      string
	FilterRadius        int
	FilterLangs         []string
	LocationSuggestions []models.LocationSuggestion

	// ByMarket says the directory is showing the header market's paid strip
	// rather than a radius search, which is the state it opens in.
	//
	// The client gave the page two modes and made them exclusive: "on top in
	// the header is the choose your market button, and so whatever market is
	// chosen there, then these brokers what have bought ad in this market are
	// displayed there … and if the user in this search menu enters the location
	// and radius, then the 'choose your market' system is not active any more."
	// The page says which one it is running, because a list of four brokers
	// that does not explain itself reads as a directory with four people in it.
	ByMarket bool
}

// LanguageOption is one choice in the broker directory's language filter: the
// code that travels in the URL, the label a reader sees, and the country whose
// flag stands for it.
type LanguageOption struct {
	Code string
	Name string
	Flag string
}

// DefaultBrokerRadius is the client's figure, used when none is asked for —
// "from this location and with this radius (50 km) the brokers are displayed".
//
// The seven fixed distances that used to sit beside it are gone: the radius is
// typed now, stepped in tens by the field's own arrows, so any number is
// reachable and a list of choices would only get in the way.
const DefaultBrokerRadius = 50

// Developments renders the developments index.
func (h *Handler) Developments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd := h.base(r, "developments", "New developments — Previa",
		"Browse new-build projects across Europe with completion dates, unit availability and prices direct from the developer.")

	country := r.URL.Query().Get("country")
	if country == "" {
		country = pd.Country.Code
	}

	pd.Data = ListData{
		Developments: h.Store.Content.Developments(ctx, country),
		Countries:    h.Store.Catalog.Countries(ctx),
	}
	h.View.Render(w, http.StatusOK, "developments", pd)
}

// Brokers renders the broker directory.
func (h *Handler) Brokers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	pd := h.base(r, "brokers", "Find a broker — Previa",
		"Search real-estate brokers on the map by location, radius, language and specialisation.")

	// Every language on the catalogue, not only the ones a seeded broker
	// happens to speak: "in the language section enlist all the languages with
	// search menu and multiselect function." The picker is the market menu, so
	// it is searchable, and a language nobody speaks yet simply returns nothing
	// rather than being missing from the list.
	var langs []LanguageOption
	for _, l := range data.SpokenLanguages() {
		langs = append(langs, LanguageOption{Code: l.Code, Name: l.Name, Flag: l.Flag})
	}

	// The place, and how far from it to look.
	//
	// A label is resolved to a point through the same catalogue the property
	// search resolves against, so "Tallinn" means the same coordinates in both.
	// An unresolvable label keeps its text and the filter matches it against
	// the broker's city and country instead — see data.Mock.Filter.
	//
	// With nothing typed into it, the market chosen in the header decides
	// instead, and the page shows exactly what that market's homepage strip
	// shows: "whatever market is chosen there, then these brokers what have
	// bought ad in this market are displayed there — so the same as the
	// frontpage broker section." data.Mock.Filter is where the two modes are
	// held apart; the handler only has to hand it both.
	f := data.BrokerFilter{
		Location:   strings.TrimSpace(q.Get("location")),
		RadiusKm:   DefaultBrokerRadius,
		Languages:  languageCodes(q["language"]),
		Q:          strings.TrimSpace(q.Get("q")),
		MarketCode: pd.Country.Code,
	}
	if n, err := strconv.Atoi(q.Get("radius")); err == nil && n > 0 {
		f.RadiusKm = n
	}
	if f.Location != "" {
		if s, ok := h.Store.Catalog.ResolveLocation(f.Location); ok {
			f.Lat, f.Lng = s.Lat, s.Lng
		}
	}

	pd.Data = ListData{
		Brokers:             h.Store.Brokers.Filter(ctx, f),
		Countries:           h.Store.Catalog.Countries(ctx),
		FilterQ:             f.Q,
		FilterLocation:      f.Location,
		FilterRadius:        f.RadiusKm,
		FilterLangs:         f.Languages,
		Languages:           langs,
		LocationSuggestions: h.Store.Catalog.LocationSuggestions(ctx),
		ByMarket:            f.MarketCode != "" && f.IsBrowsing(),
	}
	h.View.Render(w, http.StatusOK, "brokers", pd)
}

// languageCodes normalises the repeated `language` parameter the multi-select
// submits.
//
// "The first selection is 'any'", and Any is the absence of a choice rather
// than a code of its own — it submits nothing, and an empty list matches every
// broker. Blanks are dropped so a stray `&language=` in a hand-edited URL
// cannot become a language nobody speaks.
func languageCodes(vals []string) []string {
	var out []string
	for _, v := range vals {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// Agencies renders the agency directory.
func (h *Handler) Agencies(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "brokers", "Real-estate agencies — Previa",
		"Browse the agencies listing property on Previa, with their teams, coverage and active listings.")
	pd.Data = ListData{
		Agencies:  h.Store.Brokers.Agencies(r.Context()),
		Countries: h.Store.Catalog.Countries(r.Context()),
	}
	h.View.Render(w, http.StatusOK, "agencies", pd)
}

// Articles renders the article index.
func (h *Handler) Articles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cat := r.URL.Query().Get("category")

	pd := h.base(r, "articles", "Property advice and market news — Previa",
		"Guides on buying, selling, renting, investing and renovating, written by working brokers.")

	seen := map[string]bool{}
	var cats []string
	for _, a := range h.Store.Content.Articles(ctx) {
		if !seen[a.Category] {
			seen[a.Category] = true
			cats = append(cats, a.Category)
		}
	}

	pd.Data = ListData{
		Articles:   h.Store.Content.ArticlesByCategory(ctx, cat),
		Categories: cats,
		Category:   cat,
	}
	h.View.Render(w, http.StatusOK, "articles", pd)
}

// Pricing renders the packages page.
func (h *Handler) Pricing(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Listing packages and pricing — Previa",
		"Compare Basic, Standard, Premium and Agency packages. Publish a property from €19.")
	pd.Data = ListData{Packages: h.Store.Catalog.Packages(r.Context())}
	h.View.Render(w, http.StatusOK, "pricing", pd)
}

// FAQ renders the frequently-asked-questions page.
//
// The accordion used to be the last band on /pricing; it is a page of its own
// now because the footer links to it directly.
func (h *Handler) FAQ(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "FAQ — Previa",
		"Answers to the questions Previa sellers, buyers and renters ask most often about listings, packages, promotion and contacting a seller.")
	h.View.Render(w, http.StatusOK, "faq", pd)
}

// Help renders the help and contact page.
func (h *Handler) Help(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Help and contact — Previa",
		"Get help with your Previa account, listings, payments and the buying or renting process.")
	h.View.Render(w, http.StatusOK, "help", pd)
}

// About renders the company page.
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "About Previa",
		"Previa is an international property marketplace covering eight European markets.")
	h.View.Render(w, http.StatusOK, "about", pd)
}

// Advertising renders the advertising information page.
func (h *Handler) Advertising(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Advertise on Previa",
		"Reach buyers, renters and sellers across eight European markets with banner and promoted-broker placements.")
	pd.Data = ListData{Countries: h.Store.Catalog.Countries(r.Context())}
	h.View.Render(w, http.StatusOK, "advertising", pd)
}

// LegalData carries which legal document to render.
type LegalData struct {
	Kind    string
	Heading string
	Updated string
}

// Legal renders the terms or privacy placeholder.
func (h *Handler) Legal(w http.ResponseWriter, r *http.Request) {
	kind := "terms"
	heading := "Terms of service"
	if strings.HasSuffix(r.URL.Path, "/privacy") {
		kind, heading = "privacy", "Privacy policy"
	}

	pd := h.base(r, "", heading+" — Previa",
		"The "+strings.ToLower(heading)+" governing the use of Previa.")
	pd.Meta.NoIndex = true
	pd.Data = LegalData{Kind: kind, Heading: heading, Updated: "1 July 2026"}
	h.View.Render(w, http.StatusOK, "legal", pd)
}

// Cookies renders the cookie-preference placeholder.
func (h *Handler) Cookies(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Cookie preferences — Previa",
		"Choose which cookies Previa may use.")
	pd.Meta.NoIndex = true
	h.View.Render(w, http.StatusOK, "cookies", pd)
}

// NotFound renders the 404 page.
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Page not found — Previa", "The page you were looking for does not exist.")
	pd.Meta.NoIndex = true
	h.View.Render(w, http.StatusNotFound, "404", pd)
}

// Robots serves a robots.txt placeholder.
func (h *Handler) Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("User-agent: *\n" +
		"Allow: /\n" +
		"Disallow: /admin\n" +
		"Disallow: /dashboard\n" +
		"Disallow: /billing\n" +
		"Disallow: /checkout\n" +
		"Disallow: /settings\n\n" +
		"Sitemap: " + strings.TrimSuffix(h.Cfg.BaseURL, "/") + "/sitemap.xml\n"))
}

// Sitemap serves a sitemap placeholder built from the mock catalogue.
func (h *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	base := strings.TrimSuffix(h.Cfg.BaseURL, "/")

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	add := func(loc, freq, pri string) {
		b.WriteString("  <url><loc>" + base + loc + "</loc>")
		b.WriteString("<changefreq>" + freq + "</changefreq><priority>" + pri + "</priority></url>\n")
	}

	for _, p := range []string{"/", "/search", "/developments", "/brokers", "/agencies",
		"/articles", "/pricing", "/faq", "/about", "/help"} {
		add(p, "daily", "0.9")
	}
	res, _ := h.Store.Properties.Search(ctx, parseAll())
	for _, p := range res.Items {
		add("/property/"+p.Slug, "weekly", "0.8")
	}
	for _, d := range h.Store.Content.Developments(ctx, "") {
		add("/development/"+d.Slug, "weekly", "0.7")
	}
	for _, br := range h.Store.Brokers.All(ctx) {
		add("/broker/"+br.Slug, "weekly", "0.6")
	}
	for _, a := range h.Store.Content.Articles(ctx) {
		add("/article/"+a.Slug, "monthly", "0.6")
	}
	b.WriteString("</urlset>\n")

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}
