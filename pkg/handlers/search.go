package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"previa/pkg/data"
	"previa/pkg/models"
	"previa/pkg/view"
)

// mapPoint is the marker payload handed to the client map component.
type mapPoint struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Price    string   `json:"price"`
	Full     string   `json:"full"`
	Lat      float64  `json:"lat"`
	Lng      float64  `json:"lng"`
	URL      string   `json:"url"`
	Type     string   `json:"type"`
	City     string   `json:"city"`
	Rooms    int      `json:"rooms"`
	Area     string   `json:"area"`
	Featured bool     `json:"featured"`
	Images   []string `json:"images"`
}

// buildMapConfig serialises markers and viewport for previaMap(). Marker data
// is produced server-side so the map works from the first paint.
func buildMapConfig(items []models.Property, c models.Country, key string) template.JS {
	pts := make([]mapPoint, 0, len(items))

	north, south := -90.0, 90.0
	east, west := -180.0, 180.0
	for _, p := range items {
		if p.Coords.Lat > north {
			north = p.Coords.Lat
		}
		if p.Coords.Lat < south {
			south = p.Coords.Lat
		}
		if p.Coords.Lng > east {
			east = p.Coords.Lng
		}
		if p.Coords.Lng < west {
			west = p.Coords.Lng
		}

		imgs := make([]string, 0, 4)
		for i, im := range p.Images {
			if i == 4 {
				break
			}
			imgs = append(imgs, im.URL)
		}

		pts = append(pts, mapPoint{
			ID: p.ID, Title: p.Title,
			Price: view.MoneyShort(p.Price), Full: view.Money(p.Price),
			Lat: p.Coords.Lat, Lng: p.Coords.Lng,
			URL:  view.PropertyURL(p), Type: data.TypeLabel(p.Type), City: p.City,
			Rooms: p.Rooms, Area: view.Area(p.Area),
			Featured: p.IsFeatured, Images: imgs,
		})
	}

	// Fall back to the active country when there is nothing to frame, and pad
	// the box so markers never sit exactly on the edge.
	if len(pts) == 0 {
		north, south = c.Lat+0.15, c.Lat-0.15
		east, west = c.Lng+0.25, c.Lng-0.25
	} else {
		padLat := (north-south)*0.12 + 0.02
		padLng := (east-west)*0.12 + 0.02
		north, south = north+padLat, south-padLat
		east, west = east+padLng, west-padLng
	}

	cfg := map[string]any{
		"points": pts,
		"apiKey": key,
		"lat":    (north + south) / 2,
		"lng":    (east + west) / 2,
		"zoom":   c.Zoom,
		"bounds": map[string]float64{"north": north, "south": south, "east": east, "west": west},
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return template.JS("{points:[],bounds:{north:1,south:0,east:1,west:0}}")
	}
	return template.JS(b)
}

// SearchData is the payload for the search screen and its HTMX fragment.
type SearchData struct {
	Result     data.SearchResult
	Filter     data.PropertyFilter
	Chips      []data.ActiveChip
	View       data.ViewMode
	Favourites map[string]bool
	Countries  []models.Country
	Cities     []string
	SortLabel  string
	TypeOpts   []Option
	CondOpts   []Option
	MapConfig  template.JS
	MapPoints  []models.Property
}

// Option is one checkbox choice in the filter panel. Building these in Go keeps
// the template free of type-conversion gymnastics.
type Option struct {
	Value   string
	Label   string
	Icon    string
	Checked bool
}

// typeOptions builds the property-type checkboxes with their current state.
func typeOptions(f data.PropertyFilter) []Option {
	defs := []struct{ v models.PropertyType; icon string }{
		{models.TypeApartment, "building"},
		{models.TypeHouse, "home"},
		{models.TypeVilla, "villa"},
		{models.TypeCommercial, "commercial"},
		{models.TypeLand, "land"},
		{models.TypeGarage, "garage"},
	}
	out := make([]Option, 0, len(defs))
	for _, d := range defs {
		on := false
		for _, t := range f.Types {
			if t == d.v {
				on = true
			}
		}
		out = append(out, Option{
			Value: string(d.v), Label: data.TypeLabel(d.v), Icon: d.icon, Checked: on,
		})
	}
	return out
}

// conditionOptions builds the condition checkboxes with their current state.
func conditionOptions(f data.PropertyFilter) []Option {
	defs := []models.Condition{
		models.ConditionNew, models.ConditionRenovated, models.ConditionGood,
		models.ConditionSatisfying, models.ConditionNeedsWork,
	}
	out := make([]Option, 0, len(defs))
	for _, c := range defs {
		on := false
		for _, x := range f.Conditions {
			if x == c {
				on = true
			}
		}
		out = append(out, Option{Value: string(c), Label: data.ConditionLabel(c), Checked: on})
	}
	return out
}

// Search renders the full search page.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	pd, sd := h.searchPayload(r)
	pd.Data = sd

	switch sd.View {
	case data.ViewFull:
		h.View.Render(w, http.StatusOK, "search-map-full", pd)
	case data.ViewMap:
		h.View.Render(w, http.StatusOK, "search-map-split", pd)
	default:
		h.View.Render(w, http.StatusOK, "search", pd)
	}
}

// SearchResults returns only the results region, for HTMX filter updates. The
// filter panel, map viewport and scroll position are all preserved because the
// rest of the page is never replaced.
func (h *Handler) SearchResults(w http.ResponseWriter, r *http.Request) {
	pd, sd := h.searchPayload(r)
	pd.Data = sd

	block := "results-region"
	if sd.View == data.ViewMap || sd.View == data.ViewFull {
		block = "map-results-region"
	}
	h.View.RenderPartial(w, http.StatusOK, "search", block, pd)
}

// searchPayload parses the query, runs the search and assembles page context.
func (h *Handler) searchPayload(r *http.Request) (PageData, SearchData) {
	ctx := r.Context()
	q := r.URL.Query()

	f := data.ParseFilter(q, multiFunc(q))

	// The map views plot every match, so they do not paginate the sidebar as
	// aggressively as the grid does.
	mode := data.ViewMode(q.Get("view"))
	switch mode {
	case data.ViewList, data.ViewMap, data.ViewFull:
	default:
		mode = data.ViewGrid
	}
	if mode == data.ViewMap || mode == data.ViewFull {
		f.PerPage = 30
	}

	// Resolve the country name so the active-filter chip is human-readable.
	if f.CountryCode != "" {
		if c, ok := h.Store.Catalog.Country(ctx, f.CountryCode); ok {
			f.CountryName = c.Name
		}
	}

	result, _ := h.Store.Properties.Search(ctx, f)

	favs := map[string]bool{}
	for _, p := range result.Items {
		favs[p.ID] = h.Store.Account.IsFavourite(ctx, p.ID)
	}

	title := "Property search"
	if f.Deal == models.DealSale {
		title = "Property for sale"
	} else if f.Deal == models.DealRent {
		title = "Property to rent"
	}
	if f.City != "" {
		title += " in " + f.City
	}

	pd := h.base(r, navFor(f.Deal), title+" — Previa",
		"Filter by location, price, size, rooms and features. View results as a grid, a list or on the map.")
	pd.NeedsMap = mode == data.ViewMap || mode == data.ViewFull

	// This page shows its own, shorter advertising strip above the results and
	// carries the market picker on it, so the header drops the picker here too.
	pd.MarketInBanner = true
	pd.Banner, pd.HasBanner = h.Store.Catalog.BannerFor(ctx, pd.Country.Code, "search")

	var cities []string
	for _, c := range h.Store.Catalog.Countries(ctx) {
		cities = append(cities, c.Cities...)
	}

	return pd, SearchData{
		Result:     result,
		Filter:     f,
		Chips:      f.Chips(),
		View:       mode,
		Favourites: favs,
		Countries:  h.Store.Catalog.Countries(ctx),
		Cities:     cities,
		SortLabel:  sortLabel(f.Sort),
		TypeOpts:   typeOptions(f),
		CondOpts:   conditionOptions(f),
		MapConfig:  buildMapConfig(result.MapPoints, pd.Country, h.Cfg.MapsKey),
		MapPoints:  result.MapPoints,
	}
}

func navFor(d models.DealType) string {
	switch d {
	case models.DealSale:
		return "buy"
	case models.DealRent:
		return "rent"
	}
	return "search"
}

func sortLabel(s data.SortOrder) string {
	switch s {
	case data.SortPriceAsc:
		return "Price: low to high"
	case data.SortPriceDesc:
		return "Price: high to low"
	case data.SortAreaDesc:
		return "Largest first"
	case data.SortPricePerM2:
		return "Price per m²"
	case data.SortPopular:
		return "Most viewed"
	}
	return "Newest first"
}
