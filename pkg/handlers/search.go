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

// brokerPoint is a broker's marker payload: the rounded-rectangle photograph
// and name the client asked for on the map, plus everything the preview that
// opens from it shows.
//
//	"For this in the maps his rounded rectangle profile image will be displayed
//	 with his name - small icon, and if user clicks on it, then the bigger user
//	 profile will open, just like the real estate preview window, and if click
//	 there then will redirect to this broker's single page."
//
// It carries the profile's current values rather than anything stored with the
// placement, which is what makes the client's other rule true for free: "if he
// updates his profile by changing photo or phone, then this will be updated in
// this ad immediately."
type brokerPoint struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Title    string  `json:"title"`
	Agency   string  `json:"agency"`
	City     string  `json:"city"`
	Photo    string  `json:"photo"`
	URL      string  `json:"url"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	Listings int     `json:"listings"`
	Rating   float64 `json:"rating"`
	Distance string  `json:"distance"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

// brokerPoints serialises the brokers whose map placement is running.
func brokerPoints(brokers []models.Broker) []brokerPoint {
	out := make([]brokerPoint, 0, len(brokers))
	for _, b := range brokers {
		out = append(out, brokerPoint{
			ID: b.ID, Name: b.Name, Title: b.Title, Agency: b.AgencyName,
			City: b.City, Photo: b.Photo, URL: "/broker/" + b.Slug,
			Phone: b.Phone, Email: b.Email,
			Listings: b.ActiveListings, Rating: b.Rating,
			Distance: b.DistanceLabel(),
			Lat:      b.Office.Lat, Lng: b.Office.Lng,
		})
	}
	return out
}

// buildPinMap serialises a one-pin map: a place being chosen rather than a set
// of listings being browsed.
//
// It is the seller profile's location picker. `pin` tells previaMap to draw a
// plain marker and bind no popup, because there is nothing to say about a point
// beyond where it is, and a marker showing a price of zero is worse than no
// marker at all.
//
// The box is padded by roughly a kilometre either way, so the map opens close
// enough to see which street the pin is on — a country-sized view would make it
// impossible to tell whether the pin is right.
func buildPinMap(place models.MapPlace, c models.Country, key string) template.JS {
	lat, lng := place.Lat, place.Lng
	if !place.IsSet() {
		lat, lng = c.Lat, c.Lng
	}
	const pad = 0.012

	cfg := map[string]any{
		"points": []mapPoint{{ID: "pin", Title: place.Label, Lat: lat, Lng: lng}},
		"pin":    true,
		"apiKey": key,
		"lat":    lat,
		"lng":    lng,
		"zoom":   14,
		"bounds": map[string]float64{
			"north": lat + pad, "south": lat - pad,
			"east": lng + pad*2, "west": lng - pad*2,
		},
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return template.JS("{points:[],bounds:{north:1,south:0,east:1,west:0}}")
	}
	return template.JS(b)
}

// buildMapConfig serialises markers and viewport for previaMap(). Marker data
// is produced server-side so the map works from the first paint.
func buildMapConfig(items []models.Property, brokers []models.Broker, c models.Country, key string) template.JS {
	pts := make([]mapPoint, 0, len(items))

	north, south := -90.0, 90.0
	east, west := -180.0, 180.0
	// The box has to hold both kinds of marker. A map framed on the listings
	// alone would open with the broker pins outside the viewport, which reads
	// as the placement not having worked.
	grow := func(lat, lng float64) {
		if lat > north {
			north = lat
		}
		if lat < south {
			south = lat
		}
		if lng > east {
			east = lng
		}
		if lng < west {
			west = lng
		}
	}
	for _, p := range items {
		grow(p.Coords.Lat, p.Coords.Lng)

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
			URL: view.PropertyURL(p), Type: data.TypeLabel(p.Type), City: p.City,
			Rooms: p.Rooms, Area: view.Area(p.Area),
			Featured: p.IsFeatured, Images: imgs,
		})
	}

	bpts := brokerPoints(brokers)
	for _, b := range bpts {
		grow(b.Lat, b.Lng)
	}

	// Fall back to the active country when there is nothing to frame, and pad
	// the box so markers never sit exactly on the edge.
	if len(pts) == 0 && len(bpts) == 0 {
		north, south = c.Lat+0.15, c.Lat-0.15
		east, west = c.Lng+0.25, c.Lng-0.25
	} else {
		padLat := (north-south)*0.12 + 0.02
		padLng := (east-west)*0.12 + 0.02
		north, south = north+padLat, south-padLat
		east, west = east+padLng, west-padLng
	}

	cfg := map[string]any{
		"points":  pts,
		"brokers": bpts,
		"apiKey":  key,
		"lat":     (north + south) / 2,
		"lng":     (east + west) / 2,
		"zoom":    c.Zoom,
		"bounds":  map[string]float64{"north": north, "south": south, "east": east, "west": west},
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
	// LocationSuggestions backs the single Location field in the sidebar.
	LocationSuggestions []models.LocationSuggestion
	SortLabel           string
	TypeOpts            []Option
	CondOpts            []Option
	MapConfig           template.JS
	MapPoints           []models.Property
	// Brokers are the paid map placements shown alongside the listings when
	// the reader ticks Brokers in the filter panel — pins on the map, cards
	// among the results. Empty unless Filter.ShowBrokers is on.
	Brokers []models.Broker
}

// Option is one checkbox choice in the filter panel. Building these in Go keeps
// the template free of type-conversion gymnastics.
type Option struct {
	Value string
	Label string
	Icon  string
	// Flag is an ISO 3166-1 alpha-2 code for options that are drawn with one —
	// today, the languages of communication. Empty everywhere else.
	Flag    string
	Checked bool
}

// typeOptions builds the property-type checkboxes with their current state.
// typeOptions builds the property-type checkboxes with their current state.
//
// Property type is a multiple-choice filter: the client asked for House,
// Modular house and Panelized house to be selectable together, matching any of
// them. So every catalogue entry reports its own checked state and several can
// be on at once.
func typeOptions(f data.PropertyFilter) []Option {
	on := make(map[models.PropertyType]bool, len(f.Types))
	for _, t := range f.Types {
		on[t] = true
	}
	out := make([]Option, 0, len(models.PropertyTypes))
	for _, t := range models.PropertyTypes {
		out = append(out, Option{
			Value: string(t.Value), Label: t.Label, Icon: t.Icon, Checked: on[t.Value],
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

	// Tell the browser to show /search, not this fragment endpoint.
	//
	// The filter form pushes the URL it requested, which is /search/results —
	// so after any filter change the address bar pointed at a fragment, and
	// reloading or sharing that link returned bare markup with no layout. The
	// header names the page the query belongs to; the parameters, including
	// every repeated property_type, are carried across untouched.
	if isHTMX(r) {
		push := "/search"
		if q := r.URL.RawQuery; q != "" {
			push += "?" + q
		}
		w.Header().Set("HX-Push-Url", push)
	}

	// Which fragment answers depends on what the page around it looks like.
	//
	// The grid and list views keep their tag bar, count and Sort control inside
	// #results, so replacing that one element updates all of them.
	//
	// The split map view cannot: its header sits above the scrolling results
	// column, outside the swap target. So it gets a fragment that carries the
	// header alongside the list as an out-of-band swap — without it the tags a
	// visitor had just ticked in the filter drawer never appeared.
	//
	// The full-screen map has no results list and no header of that shape, so it
	// stays on the plain region.
	block := "results-region"
	switch sd.View {
	case data.ViewMap:
		block = "map-results-fragment"
	case data.ViewFull:
		block = "map-results-region"
	}
	h.View.RenderPartial(w, http.StatusOK, "search", block, pd)
}

// searchPayload parses the query, runs the search and assembles page context.
func (h *Handler) searchPayload(r *http.Request) (PageData, SearchData) {
	ctx := r.Context()
	q := r.URL.Query()

	f := data.ParseFilter(q, multiFunc(q))

	// The Location box submits one label; turn it into the country/city/
	// district/address the matcher works with. Done here rather than inside
	// ParseFilter because it needs the store's suggestion list.
	data.ApplyLocation(&f, h.Store.Catalog.ResolveLocation)

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

	// Brokers, when they were asked for. The client's rule is that they behave
	// like the listings do — "the system with the brokers is the same as with
	// the real estate" — so they are narrowed by the same place the properties
	// were, rather than by a search of their own.
	var brokers []models.Broker
	if f.ShowBrokers {
		brokers = h.Store.Brokers.OnMap(ctx, brokerSearch(f))
	}

	favs := map[string]bool{}
	for _, p := range result.Items {
		favs[p.ID] = h.Store.Account.IsFavourite(ctx, p.ID)
	}

	// The page title names the deal type only when exactly one is selected. A
	// search across several is "Property search" — the tag bar above the results
	// is what spells out which ones, and it can hold any combination.
	title := "Property search"
	if len(f.Deals) == 1 {
		switch f.Deals[0] {
		case models.DealSale:
			title = "Property for sale"
		case models.DealRent:
			title = "Property to rent"
		case models.DealShortRent:
			title = "Property for short rent"
		}
	}
	if f.City != "" {
		title += " in " + f.City
	}

	pd := h.base(r, navFor(f.Deals, mode), title+" — Previa",
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
		Result:              result,
		Filter:              f,
		Chips:               f.Chips(),
		View:                mode,
		Favourites:          favs,
		Countries:           h.Store.Catalog.Countries(ctx),
		Cities:              cities,
		LocationSuggestions: h.Store.Catalog.LocationSuggestions(ctx),
		SortLabel:           sortLabel(f.Sort),
		TypeOpts:            typeOptions(f),
		CondOpts:            conditionOptions(f),
		// Built from every match rather than the current page, so paging to
		// page two does not change which languages the panel offers.
		MapConfig: buildMapConfig(result.MapPoints, brokers, pd.Country, h.Cfg.MapsKey),
		MapPoints: result.MapPoints,
		Brokers:   brokers,
	}
}

// brokerSearch turns the property filter's place into the broker directory's.
//
// A resolved city, district or street is a point, and brokers near it are the
// ones within the directory's own default radius. A whole country is not: a
// 50 km circle around the geographic centre of Germany is a field near Kassel,
// not Germany, so the country is matched as a market instead and the radius
// dropped. Anything that did not resolve at all travels as text, which is what
// the directory already does with a label it could not place.
func brokerSearch(f data.PropertyFilter) data.BrokerFilter {
	bf := data.BrokerFilter{Languages: f.Languages}
	switch {
	case f.HasPoint() && (f.City != "" || f.District != "" || f.Address != ""):
		bf.Location = f.LocationLabel
		bf.Lat, bf.Lng = f.Lat, f.Lng
		bf.RadiusKm = DefaultBrokerRadius
	case f.CountryCode != "":
		bf.CountryCode = f.CountryCode
	case f.LocationLabel != "":
		bf.Location = f.LocationLabel
	}
	return bf
}

// navFor picks the header's active nav item. A mixed selection highlights
// neither Buy nor Rent — no single item would be telling the truth.
//
// The map views mark Map rather than Listings: the client asked for the two to
// be separate entries in the header — "in the header after Listings add section
// 'Map' so this is the map page" — and an entry that never lights up when you
// are on the page it names is worse than no entry at all.
func navFor(deals []models.DealType, mode data.ViewMode) string {
	if mode == data.ViewMap || mode == data.ViewFull {
		return "map"
	}
	if len(deals) != 1 {
		return "search"
	}
	switch deals[0] {
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
