package handlers

import (
	"net/http"
	"strings"

	"previa/pkg/data"
	"previa/pkg/models"
)

// HomeData is the homepage payload.
type HomeData struct {
	Featured     []models.Property
	Recent       []models.Property
	Developments []models.Development
	Brokers      []models.Broker
	Articles     []models.Article
	Locations    []data.LocationCount
	Types        []TypeTile
	Testimonials []models.Testimonial
	TotalCount   int
	// LocationSuggestions backs the hero's single Location field, the same
	// component and the same list the search sidebar uses.
	LocationSuggestions []models.LocationSuggestion
	// TypeOpts backs the hero's property-type picker — the same component and
	// the same catalogue the filter sidebar renders.
	TypeOpts []Option
	// Favourites is keyed by property ID, the same shape the search page uses.
	// Without it every card on the homepage drew an empty heart even for a
	// listing already saved, and the first press appeared to do nothing.
	Favourites map[string]bool
}

// TypeTile backs the "browse by property type" section.
type TypeTile struct {
	Type  models.PropertyType
	Label string
	Icon  string
	Count int
}

// Home renders the homepage.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	// The router sends every unmatched path here, so anything that is not
	// exactly "/" is a genuine 404.
	if r.URL.Path != "/" {
		h.NotFound(w, r)
		return
	}

	ctx := r.Context()
	pd := h.base(r, "home",
		"Previa — Property for sale and rent worldwide",
		"Search apartments, houses, cottages, commercial property and land across Europe. Verified brokers, accurate maps and new listings every day.")

	// The hero carries the market picker on this page, so the header drops it.
	pd.MarketInBanner = true

	// Browse-by-type tiles, straight from the shared catalogue so a category
	// added there appears here without a second edit.
	counts := h.Store.Properties.CountByType(ctx, "")
	tiles := make([]TypeTile, 0, len(models.PropertyTypes))
	for _, t := range models.PropertyTypes {
		tiles = append(tiles, TypeTile{t.Value, t.Label, t.Icon, counts[t.Value]})
	}

	all, _ := h.Store.Properties.Search(ctx, data.PropertyFilter{PerPage: 1})

	featured := h.Store.Properties.Featured(ctx, pd.Country.Code, 10)
	recent := h.Store.Properties.Recent(ctx, pd.Country.Code, 8)

	favs := map[string]bool{}
	for _, p := range append(append([]models.Property{}, featured...), recent...) {
		favs[p.ID] = h.Store.Account.IsFavourite(ctx, p.ID)
	}

	pd.Data = HomeData{
		// Ten fills two complete rows at the five-across desktop density.
		Featured:            featured,
		Recent:              recent,
		Developments:        take(h.Store.Content.Developments(ctx, pd.Country.Code), 3),
		Brokers:             h.Store.Brokers.Promoted(ctx, pd.Country.Code, 4),
		Articles:            take(h.Store.Content.Articles(ctx), 3),
		Locations:           h.Store.Properties.PopularLocations(ctx, pd.Country.Code, 8),
		Types:               tiles,
		Testimonials:        h.Store.Catalog.Testimonials(ctx),
		TotalCount:          all.Total,
		Favourites:          favs,
		LocationSuggestions: h.Store.Catalog.LocationSuggestions(ctx),
		// Nothing preselected: the homepage always opens on "Any type".
		TypeOpts: typeOptions(data.PropertyFilter{}),
	}

	h.View.Render(w, http.StatusOK, "home", pd)
}

func take[T any](items []T, n int) []T {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

// SetCountry stores the market choice in a cookie and returns the visitor to
// where they were. A manual choice is authoritative and is never overwritten by
// geolocation on a later visit.
func (h *Handler) SetCountry(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	code := strings.ToUpper(q.Get("code"))

	// A geolocation callback supplies coordinates instead of a code.
	if code == "" && q.Get("lat") != "" {
		code = h.nearestCountry(r)
	}
	if _, ok := h.Store.Catalog.Country(r.Context(), code); !ok {
		code = h.Cfg.DefaultCountry
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "previa_country",
		Value:    code,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		HttpOnly: false, // the client reads it to keep the Alpine store in sync
		SameSite: http.SameSiteLaxMode,
	})

	dest := q.Get("return")
	if dest == "" || !strings.HasPrefix(dest, "/") || strings.HasPrefix(dest, "//") {
		dest = "/"
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

// nearestCountry maps supplied coordinates onto the closest seeded market.
func (h *Handler) nearestCountry(r *http.Request) string {
	lat := parseFloatOr(r.URL.Query().Get("lat"), 0)
	lng := parseFloatOr(r.URL.Query().Get("lng"), 0)
	if lat == 0 && lng == 0 {
		return h.Cfg.DefaultCountry
	}

	best, bestDist := h.Cfg.DefaultCountry, 1e18
	for _, c := range h.Store.Catalog.Countries(r.Context()) {
		dLat, dLng := c.Lat-lat, c.Lng-lng
		d := dLat*dLat + dLng*dLng
		if d < bestDist {
			best, bestDist = c.Code, d
		}
	}
	return best
}

// ToggleFavourite flips the saved state and returns the refreshed button.
func (h *Handler) ToggleFavourite(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/favourite/")
	p, ok := h.Store.Properties.ByID(r.Context(), id)
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	state := h.Store.Account.ToggleFavourite(r.Context(), id)
	h.View.RenderPartial(w, http.StatusOK, "home", "favourite-button", map[string]any{
		"ID":        p.ID,
		"Title":     p.Title,
		"Favourite": state,
	})
}

func parseFloatOr(s string, def float64) float64 {
	var f float64
	var neg bool
	if s == "" {
		return def
	}
	i := 0
	if s[0] == '-' {
		neg, i = true, 1
	}
	seenDot, frac := false, 0.1
	for ; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			if seenDot {
				f += float64(c-'0') * frac
				frac /= 10
			} else {
				f = f*10 + float64(c-'0')
			}
		case c == '.' && !seenDot:
			seenDot = true
		default:
			return def
		}
	}
	if neg {
		f = -f
	}
	return f
}
