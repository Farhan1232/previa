// Package data defines the storage contracts Previa's handlers depend on and
// ships an in-memory implementation for the frontend milestone.
//
// Handlers and templates only ever see these interfaces. Swapping the mock for
// a MySQL-backed implementation is a wiring change in cmd/previa/main.go — no
// template or handler edits required. See docs/backend-integration-points.md.
package data

import (
	"context"
	"strconv"
	"strings"

	"previa/pkg/models"
)

// SortOrder controls result ordering.
type SortOrder string

const (
	SortNewest     SortOrder = "newest"
	SortPriceAsc   SortOrder = "price_asc"
	SortPriceDesc  SortOrder = "price_desc"
	SortAreaDesc   SortOrder = "area_desc"
	SortPricePerM2 SortOrder = "price_m2"
	SortPopular    SortOrder = "popular"
)

// ViewMode is how search results are presented.
type ViewMode string

const (
	ViewGrid ViewMode = "grid"
	ViewList ViewMode = "list"
	ViewMap  ViewMode = "map"  // split: results left, map right
	ViewFull ViewMode = "full" // full-screen map
)

// PropertyFilter carries every supported search parameter. A SQL provider
// translates these to WHERE clauses; the mock applies them in memory.
//
// Pointer fields distinguish "not set" from "set to zero", which matters for
// price and area bounds.
type PropertyFilter struct {
	// Deals is a multiple choice, like Types below: the client asked to be able
	// to look at Rent and Short rent together, and for each chosen deal type to
	// appear as its own removable tag. A listing matches when its deal is any
	// one of them.
	//
	// Never empty after ParseFilter — Sell is the arrival state.
	Deals       []models.DealType
	CountryCode string
	// CountryName is filled in by the handler so the active-filter chip can
	// show "Germany" rather than "DE". It is never used for matching.
	CountryName string
	// LocationLabel is what the single Location field showed. It is the value
	// round-tripped in the `location` query parameter and redisplayed in the
	// field; the City/District/Address fields below are what actually match,
	// filled in by ResolveLocationInto from the chosen suggestion.
	LocationLabel string
	City          string
	District      string
	Address       string
	// Lat, Lng is where LocationLabel resolved to, when it resolved at all.
	//
	// Nothing is *matched* on it: the client was explicit that entering a
	// location must not reorder the results — "the order of the real-estate ads
	// here is not automatically according to nearest distance. The order stays
	// like it is set up in the 'sort by' menu." It is carried so every card can
	// say how far away it is, and so the brokers shown alongside the listings
	// can be measured from the same point the properties are.
	Lat, Lng    float64
	Types       []models.PropertyType
	PriceMin    *float64
	PriceMax    *float64
	Currency    string
	Rooms       int
	Bedrooms    int
	Bathrooms   int
	AreaMin     *float64
	AreaMax     *float64
	LandAreaMin *float64
	YearMin     int
	YearMax     int
	Conditions  []models.Condition
	Furnished   bool
	Parking     bool
	Balcony     bool
	Garden      bool
	Elevator    bool
	// Amenities are the ticks from models.AmenityGroups that are not one of the
	// five booleans above, as amenity keys. Repeated in the URL —
	// amenity=sauna&amenity=dishwasher — and matched with AND: a buyer who
	// ticks a sauna and a dishwasher wants both, which is what ticking two
	// boxes on a filter has always meant.
	//
	// The five above keep their own parameters rather than joining this list,
	// so every bookmark, saved search and shared URL written before the full
	// catalogue existed still resolves to the same search.
	Amenities      []string
	EnergyRating   string
	SellerKind     models.SellerKind
	NewDevelopment bool
	FeaturedOnly   bool
	// Languages narrows to listings sold in any one of these, as ISO 639-1
	// codes. Multiple choice, and optional — the client's "this option is not
	// obligatory" — so an empty list is not a constraint and removes nothing.
	Languages []string
	Keyword   string
	// ShowBrokers turns the brokers who bought the map placement on, both as
	// pins on the map and as cards among the results.
	//
	// The client's 18 August note put the control in a fixed place: "in the
	// main search menu, where is 'deal type' sell and buy, there add 'brokers'
	// so the user can choose if he wants the brokers will be displayed or not.
	// Then the system with the brokers is the same as with real estate."
	//
	// Off by default. A reader who came to look at property gets property until
	// they ask for anything else.
	ShowBrokers bool
	Sort        SortOrder
	Page        int
	PerPage     int
}

// HasPoint reports whether the typed location resolved to somewhere on the map,
// which is what makes a distance sayable.
func (f PropertyFilter) HasPoint() bool { return f.Lat != 0 || f.Lng != 0 }

// SearchResult is one page of matches plus the metadata the results bar and
// map need.
type SearchResult struct {
	Items      []models.Property
	Total      int
	Page       int
	PerPage    int
	TotalPages int
	// MapPoints holds every match (not just the current page) so the map can
	// plot the full result set while the list paginates.
	MapPoints []models.Property
}

// HasPrev reports whether a previous page exists.
func (r SearchResult) HasPrev() bool { return r.Page > 1 }

// HasNext reports whether a further page exists.
func (r SearchResult) HasNext() bool { return r.Page < r.TotalPages }

// PrevPage is the page number before the current one, floored at 1.
func (r SearchResult) PrevPage() int {
	if r.Page <= 1 {
		return 1
	}
	return r.Page - 1
}

// NextPage is the page number after the current one, capped at TotalPages.
func (r SearchResult) NextPage() int {
	if r.Page >= r.TotalPages {
		return r.TotalPages
	}
	return r.Page + 1
}

// From is the 1-based index of the first item on this page.
func (r SearchResult) From() int {
	if r.Total == 0 {
		return 0
	}
	return (r.Page-1)*r.PerPage + 1
}

// To is the 1-based index of the last item on this page.
func (r SearchResult) To() int {
	end := r.Page * r.PerPage
	if end > r.Total {
		return r.Total
	}
	return end
}

// ---------------------------------------------------------------------------
// Repository interfaces
// ---------------------------------------------------------------------------

// PropertyRepository serves listings.
type PropertyRepository interface {
	Search(ctx context.Context, f PropertyFilter) (SearchResult, error)
	ByID(ctx context.Context, id string) (models.Property, bool)
	BySlug(ctx context.Context, slug string) (models.Property, bool)
	Featured(ctx context.Context, countryCode string, limit int) []models.Property
	Recent(ctx context.Context, countryCode string, limit int) []models.Property
	Similar(ctx context.Context, p models.Property, limit int) []models.Property
	ByBroker(ctx context.Context, brokerID string, limit int) []models.Property
	ByDevelopment(ctx context.Context, developmentID string) []models.Property
	ByStatus(ctx context.Context, status models.ListingStatus) []models.Property
	CountByType(ctx context.Context, countryCode string) map[models.PropertyType]int
	PopularLocations(ctx context.Context, countryCode string, limit int) []LocationCount
}

// LocationCount powers the "popular locations" tiles.
type LocationCount struct {
	City        string
	CountryCode string
	Country     string
	Count       int
	Image       string
	Coords      models.Coordinates
}

// BrokerFilter is what the broker directory searches on.
//
// The client replaced the directory's country and city selects with one
// Google-Maps-style Location box and a radius: "the broker search we relate to
// googlemaps … the user enters the location and radius, so from this location
// and with this radius (50 km) the brokers are displayed."
//
// So there is no country or city field here on purpose. A place is a point and
// a distance from it, which is the only thing that can answer "brokers near
// me" honestly — a city filter would miss the broker ten minutes over the
// boundary and keep the one at the far end of a large city.
type BrokerFilter struct {
	// Location is the label the reader typed or picked; Lat, Lng is where it
	// resolved to. A label that resolves to nothing leaves the point unset, and
	// the filter then falls back to matching the label as text rather than
	// silently returning every broker on the site.
	Location string
	Lat, Lng float64
	// RadiusKm applies only when the location resolved. Zero means no radius.
	RadiusKm int
	// Languages are language codes, any of which is a match — a buyer who
	// speaks German and English wants brokers reachable in either, not brokers
	// who speak both. Empty matches everyone, which is what "Any" selects.
	Languages []string
	// Q is the free-text name-or-speciality box.
	Q string

	// MarketCode is the market chosen in the header, and it is what the
	// directory falls back to when nothing has been typed into it.
	//
	// The client gave the page two modes, 18 August: "On top in the header is
	// the choose your market button, and so whatever market is chosen there,
	// then these brokers what have bought ad in this market are displayed there
	// — so the same as the frontpage broker section. And if the user in this
	// search menu enters the location and radius, then the 'choose your market'
	// system is not active any more and now the system displays there only
	// these brokers what are in this range and in this order, who is closer."
	//
	// So the two modes are exclusive by construction: a location, typed or
	// resolved, switches this off. Empty means no market restriction at all,
	// which is what the search modes use.
	MarketCode string

	// CountryCode narrows to brokers working in one market — their home
	// country or any of the countries they list themselves as active in.
	//
	// It is where the broker *is*, which is a different question from
	// MarketCode above (where they *bought* an ad). The search page uses it
	// when the reader typed a whole country into the location box: a 50 km
	// circle around a country's centre is a field, not a country.
	CountryCode string

	// MapAdOnly narrows to brokers who bought the map placement and dropped a
	// pin — the set the search page draws, as pins on the map and as cards
	// beside them.
	MapAdOnly bool
}

// HasPoint reports whether a radius search is possible: a place was entered and
// it resolved to somewhere on the map.
func (f BrokerFilter) HasPoint() bool { return f.Lat != 0 || f.Lng != 0 }

// IsBrowsing reports whether nothing at all has been searched for — which is
// when the header's market decides what the directory shows.
//
// The client named the location field when they described the switch: "if the
// user in this search menu enters the location and radius, then the 'choose
// your market' system is not active any more." The rule underneath it is
// browsing versus searching, so every input in that menu ends the market mode,
// not only the first one. A reader who has asked for brokers who speak Czech
// has searched, and answering with "nobody in your market speaks Czech" when
// three of them do elsewhere would be the market quietly overruling the
// question they actually asked.
func (f BrokerFilter) IsBrowsing() bool {
	return f.Location == "" && len(f.Languages) == 0 && f.Q == ""
}

// BrokerRepository serves brokers and agencies.
type BrokerRepository interface {
	All(ctx context.Context) []models.Broker
	ByID(ctx context.Context, id string) (models.Broker, bool)
	BySlug(ctx context.Context, slug string) (models.Broker, bool)
	Promoted(ctx context.Context, countryCode string, limit int) []models.Broker
	Filter(ctx context.Context, f BrokerFilter) []models.Broker
	// OnMap serves the search page: the brokers whose map placement is paid up
	// and who have a pin to place, narrowed by the same location and radius the
	// property search used. See BrokerFilter.MapAdOnly.
	OnMap(ctx context.Context, f BrokerFilter) []models.Broker
	Agencies(ctx context.Context) []models.Agency
	AgencyBySlug(ctx context.Context, slug string) (models.Agency, bool)
	BrokersByAgency(ctx context.Context, agencyID string) []models.Broker
}

// ContentRepository serves editorial content and developments.
type ContentRepository interface {
	Articles(ctx context.Context) []models.Article
	ArticleBySlug(ctx context.Context, slug string) (models.Article, bool)
	ArticlesByCategory(ctx context.Context, category string) []models.Article
	RelatedArticles(ctx context.Context, a models.Article, limit int) []models.Article
	Developments(ctx context.Context, countryCode string) []models.Development
	DevelopmentBySlug(ctx context.Context, slug string) (models.Development, bool)
}

// AccountRepository serves the signed-in user's data.
type AccountRepository interface {
	CurrentUser(ctx context.Context) models.User
	Favourites(ctx context.Context) []models.Property
	IsFavourite(ctx context.Context, propertyID string) bool
	ToggleFavourite(ctx context.Context, propertyID string) bool
	SavedSearches(ctx context.Context) []models.SavedSearch
	// AddSavedSearch stores a search the reader asked to keep and returns it
	// with the id and timestamp the store gave it.
	//
	// The client's note is about what happens after the click — "then add there
	// button 'save search', if hit that then this search will be saved under
	// user's 'saved searches' menu" — so the panel's button has to leave a row
	// behind rather than only say it did. In the mock the row lives for as long
	// as the process does; the MySQL implementation writes it to a table and
	// nothing above this line changes.
	AddSavedSearch(ctx context.Context, s models.SavedSearch) models.SavedSearch
	Notifications(ctx context.Context) []models.Notification
	UnreadCount(ctx context.Context) int
	MyListings(ctx context.Context, status models.ListingStatus) []models.Property
	// ListingStats is the per-day visitor series behind the statistics panel on
	// My listings. days is how far back to go; the result runs oldest first.
	// The second return is false for a listing the user does not own, so a
	// hand-edited id cannot read another seller's numbers.
	ListingStats(ctx context.Context, propertyID string, days int) (models.ListingStats, bool)
	// CloneListing duplicates a listing as a new draft and returns it. The
	// copy is never active: a duplicate has not been paid for.
	CloneListing(ctx context.Context, propertyID string) (models.Property, bool)
	Drafts(ctx context.Context) []models.Draft
	Payments(ctx context.Context) []models.Payment
}

// CatalogRepository serves reference data used across the site.
type CatalogRepository interface {
	Countries(ctx context.Context) []models.Country
	// AllCountries is every selectable market — the seeded ones plus the rest
	// of the ISO 3166-1 list, which is what the market selector offers.
	AllCountries(ctx context.Context) []models.Country
	// OtherCountries is AllCountries without the seeded markets, so a selector
	// that heads the seeded ones separately does not list them twice.
	OtherCountries(ctx context.Context) []models.Country
	// CountryHasListings separates markets with seeded stock from the rest.
	CountryHasListings(ctx context.Context, code string) bool
	// LocationSuggestions backs the single Location field used by the homepage
	// search, the filter sidebar, the map filter and the add-listing form.
	LocationSuggestions(ctx context.Context) []models.LocationSuggestion
	// ResolveLocation turns a typed or selected label into a structured place.
	ResolveLocation(label string) (models.LocationSuggestion, bool)
	// ReverseGeocode turns a map click into a structured address. Mocked in
	// this milestone; a Geocoding API call replaces it.
	ReverseGeocode(lat, lng float64) models.LocationSuggestion
	Country(ctx context.Context, code string) (models.Country, bool)
	Banner(ctx context.Context, countryCode string) (models.Banner, bool)
	// BannerFor returns the banner for a market and placement ("home" or
	// "search"). Inactive slots are reported with ok=false so callers render
	// nothing, while the row itself stays editable in admin.
	BannerFor(ctx context.Context, countryCode, placement string) (models.Banner, bool)
	// BannersAll lists every configured slot for the admin screen, including
	// the ones currently switched off.
	BannersAll(ctx context.Context) []models.Banner
	Packages(ctx context.Context) []models.Package
	Promotions(ctx context.Context) []models.Promotion
	// BrokerAdPlan is the homepage broker strip as a purchasable placement,
	// sold per market — see models.BrokerAd.
	BrokerAdPlan(ctx context.Context) models.BrokerAdPlan
	// BrokerMapAdPlan is the other placement a broker buys — their pin on the
	// search map. Bought once rather than per market; see models.BrokerMapAd.
	BrokerMapAdPlan(ctx context.Context) models.BrokerMapAdPlan
	Testimonials(ctx context.Context) []models.Testimonial
	Languages(ctx context.Context) []models.Language
	RestrictedCountries(ctx context.Context) []models.RestrictedCountry
}

// AdminRepository serves the administration screens.
type AdminRepository interface {
	Stats(ctx context.Context) models.AdminStats
	Users(ctx context.Context) []models.User
	Translations(ctx context.Context, lang string) []models.TranslationString
	SEOEntries(ctx context.Context) []models.SEOEntry
	Backups(ctx context.Context) []models.Backup
	Files(ctx context.Context, path string) []models.FileEntry
	Tables(ctx context.Context) []models.DBTable
	SystemInfo(ctx context.Context) models.SystemInfo
	ActivityLog(ctx context.Context) []models.ActivityEntry
}

// Store bundles every repository so handlers take one dependency.
type Store struct {
	Properties PropertyRepository
	Brokers    BrokerRepository
	Content    ContentRepository
	Account    AccountRepository
	Catalog    CatalogRepository
	Admin      AdminRepository
}

// ---------------------------------------------------------------------------
// Filter parsing
// ---------------------------------------------------------------------------

// Values is the subset of url.Values the parser needs, so this package does not
// depend on net/http.
type Values interface {
	Get(key string) string
}

// ParseFilter builds a PropertyFilter from request query parameters. Unknown or
// malformed values are ignored rather than erroring, so a hand-edited URL
// degrades to a broader search instead of a failure page.
func ParseFilter(v Values, multi func(string) []string) PropertyFilter {
	f := PropertyFilter{
		CountryCode:   strings.ToUpper(v.Get("country")),
		LocationLabel: strings.TrimSpace(v.Get("location")),
		City:          v.Get("city"),
		District:      v.Get("district"),
		Address:       v.Get("address"),
		Currency:      v.Get("currency"),
		Keyword:       strings.TrimSpace(v.Get("q")),
		EnergyRating:  v.Get("energy"),
		Sort:          SortOrder(v.Get("sort")),
		Page:          atoiDefault(v.Get("page"), 1),
		PerPage:       atoiDefault(v.Get("per_page"), 12),
	}

	// Deal type is a multiple choice — the parameter repeats, deal=rent&deal=short_rent.
	//
	// There is still no "any deal type": with the Any option removed, an empty
	// or entirely unrecognised selection falls back to Sell, which the client
	// asked to be the state a visitor arrives in, rather than quietly widening
	// the search to everything.
	seenDeal := map[models.DealType]bool{}
	for _, raw := range multi("deal") {
		d := models.DealType(strings.ToLower(strings.TrimSpace(raw)))
		switch d {
		case models.DealSale, models.DealRent, models.DealShortRent:
			if !seenDeal[d] {
				seenDeal[d] = true
				f.Deals = append(f.Deals, d)
			}
		}
	}
	if len(f.Deals) == 0 {
		f.Deals = []models.DealType{models.DealSale}
	}
	if f.Sort == "" {
		f.Sort = SortNewest
	}
	if f.PerPage <= 0 || f.PerPage > 60 {
		f.PerPage = 12
	}
	if f.Page < 1 {
		f.Page = 1
	}

	// Property type is a multiple choice: the client asked for House, Modular
	// house and Panelized house to be selectable together, so every value is
	// kept and a listing matches when its type is any one of them (see
	// containsType in the matcher — OR, not AND).
	//
	// The parameter is repeated: property_type=house&property_type=cottage.
	// `type` is still read so links and bookmarks from the single-select
	// version keep working; both feed the same field.
	seenType := map[string]bool{}
	for _, raw := range append(multi("property_type"), multi("type")...) {
		v := strings.ToLower(strings.TrimSpace(raw))
		// An unknown value would filter on nothing and silently return an
		// empty page, so a hand-edited URL is ignored rather than obeyed.
		if v == "" || seenType[v] || !models.IsPropertyType(v) {
			continue
		}
		seenType[v] = true
		f.Types = append(f.Types, models.PropertyType(v))
	}
	for _, c := range multi("condition") {
		if c != "" {
			f.Conditions = append(f.Conditions, models.Condition(c))
		}
	}

	// Language of communication, the client's 17 August addition. Repeated,
	// like deal and property type: language=de&language=en. An unknown code is
	// dropped rather than obeyed — a hand-edited URL would otherwise filter on
	// something that exists nowhere and return an empty page with no
	// explanation for it.
	seenLang := map[string]bool{}
	for _, raw := range multi("language") {
		v := strings.ToLower(strings.TrimSpace(raw))
		if v == "" || seenLang[v] || !IsSpokenLanguage(v) {
			continue
		}
		seenLang[v] = true
		f.Languages = append(f.Languages, v)
	}

	f.PriceMin = parseFloat(v.Get("price_min"))
	f.PriceMax = parseFloat(v.Get("price_max"))
	f.AreaMin = parseFloat(v.Get("area_min"))
	f.AreaMax = parseFloat(v.Get("area_max"))
	f.LandAreaMin = parseFloat(v.Get("land_min"))

	f.Rooms = atoiDefault(v.Get("rooms"), 0)
	f.Bedrooms = atoiDefault(v.Get("bedrooms"), 0)
	f.Bathrooms = atoiDefault(v.Get("bathrooms"), 0)
	f.YearMin = atoiDefault(v.Get("year_min"), 0)
	f.YearMax = atoiDefault(v.Get("year_max"), 0)

	// The rest of the "Features and amenities" catalogue, added to the search
	// menu on the client's 19 August note. Unknown keys are dropped for the
	// same reason an unknown property type is: a hand-edited URL would
	// otherwise filter on something that exists nowhere and return an empty
	// page with no explanation for it.
	seenAmenity := map[string]bool{}
	for _, raw := range multi("amenity") {
		key := strings.ToLower(strings.TrimSpace(raw))
		if key == "" || seenAmenity[key] || !models.IsAmenity(key) {
			continue
		}
		seenAmenity[key] = true
		f.Amenities = append(f.Amenities, key)
	}

	f.Furnished = isOn(v.Get("furnished"))
	f.Parking = isOn(v.Get("parking"))
	f.Balcony = isOn(v.Get("balcony"))
	f.Garden = isOn(v.Get("garden"))
	f.Elevator = isOn(v.Get("elevator"))
	f.NewDevelopment = isOn(v.Get("new_development"))
	f.FeaturedOnly = isOn(v.Get("featured"))
	f.ShowBrokers = isOn(v.Get("brokers"))

	switch strings.ToLower(v.Get("seller")) {
	case "broker":
		f.SellerKind = models.SellerBroker
	case "private":
		f.SellerKind = models.SellerPrivate
	}

	return f
}

// ActiveChip is one removable filter shown above the results.
type ActiveChip struct {
	Label string
	Key   string // query parameter to drop when the chip is dismissed
	Value string // specific value to drop for repeatable parameters
}

// Chips renders the filter as a list of removable chips.
func (f PropertyFilter) Chips() []ActiveChip {
	var out []ActiveChip
	add := func(label, key, value string) {
		out = append(out, ActiveChip{Label: label, Key: key, Value: value})
	}

	// One chip per selected deal type, each removable on its own — so a search
	// across Rent and Short rent can drop back to either without clearing both.
	//
	// The last remaining deal chip is not removable: dropping it would leave no
	// deal type at all, which the parser reads as "Sell", so the chip would
	// appear to do nothing. Chips() marks it by leaving Key empty and the
	// template renders it without a remove button.
	for _, d := range f.Deals {
		key := "deal"
		if len(f.Deals) == 1 {
			key = ""
		}
		out = append(out, ActiveChip{Label: DealLabel(d), Key: key, Value: string(d)})
	}
	if f.CountryCode != "" {
		label := f.CountryName
		if label == "" {
			label = f.CountryCode
		}
		add(label, "country", "")
	}
	if f.City != "" {
		add(f.City, "city", "")
	}
	if f.District != "" {
		add(f.District, "district", "")
	}
	// One chip per selected property type, so any single one can be removed
	// without clearing the rest of the selection.
	for _, t := range f.Types {
		add(TypeLabel(t), "property_type", string(t))
	}
	switch {
	case f.PriceMin != nil && f.PriceMax != nil:
		add("€"+formatThousands(*f.PriceMin)+" – €"+formatThousands(*f.PriceMax), "price", "")
	case f.PriceMin != nil:
		add("From €"+formatThousands(*f.PriceMin), "price_min", "")
	case f.PriceMax != nil:
		add("Up to €"+formatThousands(*f.PriceMax), "price_max", "")
	}
	switch {
	case f.AreaMin != nil && f.AreaMax != nil:
		add(formatThousands(*f.AreaMin)+" – "+formatThousands(*f.AreaMax)+" m²", "area", "")
	case f.AreaMin != nil:
		add("From "+formatThousands(*f.AreaMin)+" m²", "area_min", "")
	case f.AreaMax != nil:
		add("Up to "+formatThousands(*f.AreaMax)+" m²", "area_max", "")
	}
	if f.Rooms > 0 {
		add(strconv.Itoa(f.Rooms)+"+ rooms", "rooms", "")
	}
	if f.Bedrooms > 0 {
		add(strconv.Itoa(f.Bedrooms)+"+ bedrooms", "bedrooms", "")
	}
	if f.Bathrooms > 0 {
		add(strconv.Itoa(f.Bathrooms)+"+ bathrooms", "bathrooms", "")
	}
	for _, c := range f.Conditions {
		add(ConditionLabel(c), "condition", string(c))
	}
	if f.Furnished {
		add("Furnished", "furnished", "")
	}
	if f.Parking {
		add("Parking", "parking", "")
	}
	if f.Balcony {
		add("Balcony", "balcony", "")
	}
	if f.Garden {
		add("Garden", "garden", "")
	}
	if f.Elevator {
		add("Elevator", "elevator", "")
	}
	// One chip per ticked amenity, each removable on its own, so a search
	// narrowed by six of them can drop back one at a time.
	for _, key := range f.Amenities {
		add(models.AmenityLabel(key), "amenity", key)
	}
	if f.EnergyRating != "" {
		add("Energy "+f.EnergyRating, "energy", "")
	}
	if f.SellerKind == models.SellerBroker {
		add("Broker listings", "seller", "")
	} else if f.SellerKind == models.SellerPrivate {
		add("Private sellers", "seller", "")
	}
	if f.NewDevelopment {
		add("New development", "new_development", "")
	}
	if f.FeaturedOnly {
		add("Featured", "featured", "")
	}
	if f.YearMin > 0 {
		add("Built after "+strconv.Itoa(f.YearMin), "year_min", "")
	}
	// One chip per language, each removable on its own, like the deal and
	// property types above — a search across German and English can drop back
	// to either without clearing both.
	for _, l := range f.Languages {
		add(LanguageName(l), "language", l)
	}
	if f.Keyword != "" {
		add("“"+f.Keyword+"”", "q", "")
	}
	// Brokers last, because it does not narrow the property search — it adds a
	// second kind of result to it. Removable like any other tag, which is how
	// the map's tag bar can switch them off again without opening the drawer.
	if f.ShowBrokers {
		add("Brokers", "brokers", "")
	}
	return out
}

// ActiveCount is how many filter chips are currently applied.
func (f PropertyFilter) ActiveCount() int { return len(f.Chips()) }

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

func parseFloat(s string) *float64 {
	s = strings.TrimSpace(strings.ReplaceAll(s, " ", ""))
	if s == "" {
		return nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &n
}

func isOn(s string) bool {
	switch strings.ToLower(s) {
	case "1", "on", "true", "yes":
		return true
	}
	return false
}

func formatThousands(f float64) string {
	n := int64(f)
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// DealLabel is the human-readable name of a deal type, from the single
// catalogue in package models — so a chip reads "Short rent", exactly as the
// filter control that set it does.
//
// Distinct from view.DealLabel, which renders the badge wording on a card
// ("For sale"). A filter tag names the choice, not the listing.
func DealLabel(d models.DealType) string {
	for _, dt := range models.DealTypes {
		if dt.Value == d {
			return dt.Label
		}
	}
	return string(d)
}

// TypeLabel is the human-readable name of a property category, from the single
// catalogue in package models.
func TypeLabel(t models.PropertyType) string {
	for _, pt := range models.PropertyTypes {
		if pt.Value == t {
			return pt.Label
		}
	}
	return string(t)
}

// ConditionLabel is the human-readable name of a building condition.
func ConditionLabel(c models.Condition) string {
	switch c {
	case models.ConditionNew:
		return "Newly built"
	case models.ConditionRenovated:
		return "Renovated"
	case models.ConditionGood:
		return "Good"
	case models.ConditionSatisfying:
		return "Satisfactory"
	case models.ConditionNeedsWork:
		return "Needs work"
	}
	return string(c)
}
