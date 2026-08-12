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
	Deal        models.DealType
	CountryCode string
	// CountryName is filled in by the handler so the active-filter chip can
	// show "Germany" rather than "DE". It is never used for matching.
	CountryName string
	// LocationLabel is what the single Location field showed. It is the value
	// round-tripped in the `location` query parameter and redisplayed in the
	// field; the City/District/Address fields below are what actually match,
	// filled in by ResolveLocationInto from the chosen suggestion.
	LocationLabel  string
	City           string
	District       string
	Address        string
	Types          []models.PropertyType
	PriceMin       *float64
	PriceMax       *float64
	Currency       string
	Rooms          int
	Bedrooms       int
	Bathrooms      int
	AreaMin        *float64
	AreaMax        *float64
	LandAreaMin    *float64
	YearMin        int
	YearMax        int
	Conditions     []models.Condition
	Furnished      bool
	Parking        bool
	Balcony        bool
	Garden         bool
	Elevator       bool
	EnergyRating   string
	SellerKind     models.SellerKind
	NewDevelopment bool
	FeaturedOnly   bool
	Keyword        string
	Sort           SortOrder
	Page           int
	PerPage        int
}

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

// BrokerRepository serves brokers and agencies.
type BrokerRepository interface {
	All(ctx context.Context) []models.Broker
	ByID(ctx context.Context, id string) (models.Broker, bool)
	BySlug(ctx context.Context, slug string) (models.Broker, bool)
	Promoted(ctx context.Context, countryCode string, limit int) []models.Broker
	Filter(ctx context.Context, country, city, language, q string) []models.Broker
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
	Notifications(ctx context.Context) []models.Notification
	UnreadCount(ctx context.Context) int
	MyListings(ctx context.Context, status models.ListingStatus) []models.Property
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
		Deal:          models.DealType(strings.ToLower(v.Get("deal"))),
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

	// Deal type always has one of the three selected — the client asked for
	// Sell to be the state a visitor arrives in. There is no "any deal type":
	// with the Any option removed, an unset or unrecognised value falls back
	// to Sell rather than quietly widening the search to everything.
	switch f.Deal {
	case models.DealSale, models.DealRent, models.DealShortRent:
	default:
		f.Deal = models.DealSale
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

	f.Furnished = isOn(v.Get("furnished"))
	f.Parking = isOn(v.Get("parking"))
	f.Balcony = isOn(v.Get("balcony"))
	f.Garden = isOn(v.Get("garden"))
	f.Elevator = isOn(v.Get("elevator"))
	f.NewDevelopment = isOn(v.Get("new_development"))
	f.FeaturedOnly = isOn(v.Get("featured"))

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

	switch f.Deal {
	case models.DealSale:
		add("Sell", "deal", "")
	case models.DealRent:
		add("Rent", "deal", "")
	case models.DealShortRent:
		add("Short rent", "deal", "")
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
	if f.Keyword != "" {
		add("“"+f.Keyword+"”", "q", "")
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
