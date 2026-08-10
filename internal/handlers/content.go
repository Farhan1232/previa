package handlers

import (
	"net/http"
	"strings"

	"previa/internal/models"
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
	FilterCity   string
	FilterLang   string
	Languages    []string
}

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
		"Search verified real-estate brokers by country, city, language and specialisation.")

	langs := []string{"English", "German", "Spanish", "Estonian", "Finnish",
		"Portuguese", "Dutch", "Czech", "French", "Russian", "Swedish", "Catalan", "Italian"}

	pd.Data = ListData{
		Brokers: h.Store.Brokers.Filter(ctx,
			strings.ToUpper(q.Get("country")), q.Get("city"), q.Get("language"), q.Get("q")),
		Countries:  h.Store.Catalog.Countries(ctx),
		FilterQ:    q.Get("q"),
		FilterCity: q.Get("city"),
		FilterLang: q.Get("language"),
		Languages:  langs,
	}
	h.View.Render(w, http.StatusOK, "brokers", pd)
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
		"/articles", "/pricing", "/about", "/help"} {
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
