package data

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"previa/pkg/assets"
	"previa/pkg/models"
)

// Mock is the in-memory implementation of every repository interface. It is
// the only object that needs replacing when the MySQL backend lands.
//
// It is safe for concurrent use: reads take a read lock, and the few mutating
// operations (favourite toggling) take a write lock.
type Mock struct {
	mu sync.RWMutex

	now          time.Time
	properties   []models.Property
	brokers      []models.Broker
	agencies     []models.Agency
	articles     []models.Article
	developments []models.Development
	favourites   map[string]bool
	user         models.User
	// locations is the Location autocomplete list, derived from the seeded
	// listings once at construction — see seed_locations.go.
	locations []models.LocationSuggestion
}

// NewMock builds the seeded dataset. photoPool is the list of available
// property image URLs, discovered from disk so the seed table does not have to
// track how many images exist.
func NewMock(now time.Time) *Mock {
	pool := discoverPhotos()

	m := &Mock{
		now:          now,
		brokers:      buildBrokers(),
		agencies:     agencies,
		articles:     buildArticles(now),
		developments: buildDevelopments(now),
		favourites:   map[string]bool{},
		user: models.User{
			ID: "us-01", Name: "Anna Lehtinen", Email: "anna.lehtinen@example.com",
			Phone: "+372 5566 7788", Role: "user", CountryCode: "EE", Language: "en",
			MemberSince: now.AddDate(-2, -4, 0), IsVerified: true,
			Avatar: "",
		},
	}
	m.properties = buildProperties(now, pool)
	m.linkDevelopments()
	m.locations = buildLocationSuggestions(m.properties)

	// A handful of listings start favourited so the account screens have content.
	for _, id := range []string{"pr-002", "pr-007", "pr-013", "pr-019"} {
		m.favourites[id] = true
	}
	return m
}

// discoverPhotos returns the property photograph pool.
//
// The list comes from the generated manifest rather than a directory scan, so
// it is identical whether the app runs beside its files or as a serverless
// function that ships only the compiled binary.
func discoverPhotos() []string {
	return assets.PropertyPhotos()
}

// linkDevelopments attaches nearby new-build listings to their development so
// the development detail page has units to show.
func (m *Mock) linkDevelopments() {
	for di, d := range m.developments {
		var ids []string
		for pi, p := range m.properties {
			if p.CountryCode == d.CountryCode && p.City == d.City && p.IsNewDevelopment {
				ids = append(ids, p.ID)
				if m.properties[pi].DevelopmentID == "" {
					m.properties[pi].DevelopmentID = d.ID
				}
			}
		}
		m.developments[di].PropertyIDs = ids
	}
}

// ---------------------------------------------------------------------------
// PropertyRepository
// ---------------------------------------------------------------------------

// Search applies the filter, sorts, and paginates. MapPoints carries the whole
// match set so the map can plot results the current page does not show.
func (m *Mock) Search(ctx context.Context, f PropertyFilter) (SearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []models.Property
	for _, p := range m.properties {
		if p.Status != models.StatusActive {
			continue
		}
		if matches(p, f) {
			matched = append(matched, p)
		}
	}

	sortProperties(matched, f.Sort)

	total := len(matched)
	perPage := f.PerPage
	if perPage <= 0 {
		perPage = 12
	}
	pages := (total + perPage - 1) / perPage
	if pages == 0 {
		pages = 1
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	if page > pages {
		page = pages
	}

	start := (page - 1) * perPage
	end := start + perPage
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return SearchResult{
		Items:      matched[start:end],
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: pages,
		MapPoints:  matched,
	}, nil
}

// matches reports whether a listing satisfies every set filter field.
func matches(p models.Property, f PropertyFilter) bool {
	if f.Deal != "" && p.Deal != f.Deal {
		return false
	}
	if f.CountryCode != "" && p.CountryCode != f.CountryCode {
		return false
	}
	if f.City != "" && !strings.EqualFold(p.City, f.City) {
		return false
	}
	if f.District != "" && !strings.EqualFold(p.District, f.District) {
		return false
	}
	if f.Address != "" && !containsFold(p.Address, f.Address) {
		return false
	}
	if len(f.Types) > 0 && !containsType(f.Types, p.Type) {
		return false
	}
	if f.PriceMin != nil && p.Price.Amount < *f.PriceMin {
		return false
	}
	if f.PriceMax != nil && p.Price.Amount > *f.PriceMax {
		return false
	}
	if f.AreaMin != nil && p.Area < *f.AreaMin {
		return false
	}
	if f.AreaMax != nil && p.Area > *f.AreaMax {
		return false
	}
	if f.LandAreaMin != nil && p.LandArea < *f.LandAreaMin {
		return false
	}
	if f.Rooms > 0 && p.Rooms < f.Rooms {
		return false
	}
	if f.Bedrooms > 0 && p.Bedrooms < f.Bedrooms {
		return false
	}
	if f.Bathrooms > 0 && p.Bathrooms < f.Bathrooms {
		return false
	}
	if f.YearMin > 0 && p.BuildYear < f.YearMin {
		return false
	}
	if f.YearMax > 0 && p.BuildYear > f.YearMax {
		return false
	}
	if len(f.Conditions) > 0 && !containsCondition(f.Conditions, p.Condition) {
		return false
	}
	if f.Furnished && !p.Furnished {
		return false
	}
	if f.Parking && !p.Parking {
		return false
	}
	if f.Balcony && !p.Balcony {
		return false
	}
	if f.Garden && !p.Garden {
		return false
	}
	if f.Elevator && !p.Elevator {
		return false
	}
	if f.EnergyRating != "" && p.EnergyRating != f.EnergyRating {
		return false
	}
	if f.SellerKind != "" && p.SellerKind != f.SellerKind {
		return false
	}
	if f.NewDevelopment && !p.IsNewDevelopment {
		return false
	}
	if f.FeaturedOnly && !p.IsFeatured {
		return false
	}
	if f.Keyword != "" {
		hay := strings.ToLower(strings.Join([]string{
			p.Title, p.Description, p.City, p.District, p.Address, p.Country,
		}, " "))
		for _, term := range strings.Fields(strings.ToLower(f.Keyword)) {
			if !strings.Contains(hay, term) {
				return false
			}
		}
	}
	return true
}

func containsFold(hay, needle string) bool {
	return strings.Contains(strings.ToLower(hay), strings.ToLower(needle))
}

func containsType(list []models.PropertyType, v models.PropertyType) bool {
	for _, t := range list {
		if t == v {
			return true
		}
	}
	return false
}

func containsCondition(list []models.Condition, v models.Condition) bool {
	for _, c := range list {
		if c == v {
			return true
		}
	}
	return false
}

func sortProperties(items []models.Property, order SortOrder) {
	switch order {
	case SortPriceAsc:
		sort.SliceStable(items, func(i, j int) bool { return items[i].Price.Amount < items[j].Price.Amount })
	case SortPriceDesc:
		sort.SliceStable(items, func(i, j int) bool { return items[i].Price.Amount > items[j].Price.Amount })
	case SortAreaDesc:
		sort.SliceStable(items, func(i, j int) bool { return items[i].Area > items[j].Area })
	case SortPricePerM2:
		sort.SliceStable(items, func(i, j int) bool { return items[i].PricePerM2 < items[j].PricePerM2 })
	case SortPopular:
		sort.SliceStable(items, func(i, j int) bool { return items[i].Views > items[j].Views })
	default: // SortNewest
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	}
}

// ByID looks a listing up by identifier.
func (m *Mock) ByID(ctx context.Context, id string) (models.Property, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.properties {
		if p.ID == id {
			return p, true
		}
	}
	return models.Property{}, false
}

// BySlug looks a listing up by its URL slug.
func (m *Mock) BySlug(ctx context.Context, slug string) (models.Property, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.properties {
		if p.Slug == slug {
			return p, true
		}
	}
	return models.Property{}, false
}

// Featured returns promoted listings, preferring the active country but
// falling back to other markets so the homepage is never short.
func (m *Mock) Featured(ctx context.Context, country string, limit int) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var local, other []models.Property
	for _, p := range m.properties {
		if p.Status != models.StatusActive || !p.IsFeatured {
			continue
		}
		if country == "" || p.CountryCode == country {
			local = append(local, p)
		} else {
			other = append(other, p)
		}
	}
	sortProperties(local, SortPopular)
	sortProperties(other, SortPopular)
	return take(append(local, other...), limit)
}

// Recent returns the newest listings for the active country first.
func (m *Mock) Recent(ctx context.Context, country string, limit int) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var local, other []models.Property
	for _, p := range m.properties {
		if p.Status != models.StatusActive {
			continue
		}
		if country == "" || p.CountryCode == country {
			local = append(local, p)
		} else {
			other = append(other, p)
		}
	}
	sortProperties(local, SortNewest)
	sortProperties(other, SortNewest)
	return take(append(local, other...), limit)
}

// Similar scores other listings against the given one by city, type, deal and
// price proximity.
func (m *Mock) Similar(ctx context.Context, p models.Property, limit int) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type scored struct {
		p models.Property
		s int
	}
	var list []scored
	for _, o := range m.properties {
		if o.ID == p.ID || o.Status != models.StatusActive {
			continue
		}
		s := 0
		if o.City == p.City {
			s += 4
		}
		if o.CountryCode == p.CountryCode {
			s += 2
		}
		if o.Type == p.Type {
			s += 3
		}
		if o.Deal == p.Deal {
			s += 3
		}
		if p.Price.Amount > 0 {
			ratio := o.Price.Amount / p.Price.Amount
			if ratio > 0.7 && ratio < 1.4 {
				s += 2
			}
		}
		if s >= 6 {
			list = append(list, scored{o, s})
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].s > list[j].s })

	out := make([]models.Property, 0, limit)
	for _, sc := range list {
		if len(out) == limit {
			break
		}
		out = append(out, sc.p)
	}
	return out
}

// ByBroker returns a broker's active listings.
func (m *Mock) ByBroker(ctx context.Context, brokerID string, limit int) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []models.Property
	for _, p := range m.properties {
		if p.BrokerID == brokerID && p.Status == models.StatusActive {
			out = append(out, p)
		}
	}
	sortProperties(out, SortNewest)
	return take(out, limit)
}

// ByDevelopment returns the units linked to a development.
func (m *Mock) ByDevelopment(ctx context.Context, devID string) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []models.Property
	for _, p := range m.properties {
		if p.DevelopmentID == devID {
			out = append(out, p)
		}
	}
	return out
}

// ByStatus returns listings in a given lifecycle state (used by admin).
func (m *Mock) ByStatus(ctx context.Context, status models.ListingStatus) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []models.Property
	for _, p := range m.properties {
		if p.Status == status {
			out = append(out, p)
		}
	}
	return out
}

// CountByType powers the "browse by property type" tiles.
func (m *Mock) CountByType(ctx context.Context, country string) map[models.PropertyType]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := map[models.PropertyType]int{}
	for _, p := range m.properties {
		if p.Status != models.StatusActive {
			continue
		}
		if country != "" && p.CountryCode != country {
			continue
		}
		out[p.Type]++
	}
	return out
}

// PopularLocations aggregates active listings by city.
func (m *Mock) PopularLocations(ctx context.Context, country string, limit int) []LocationCount {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agg := map[string]*LocationCount{}
	for _, p := range m.properties {
		if p.Status != models.StatusActive {
			continue
		}
		key := p.CountryCode + "|" + p.City
		if agg[key] == nil {
			agg[key] = &LocationCount{
				City: p.City, CountryCode: p.CountryCode, Country: p.Country,
				Image:  fmt.Sprintf("/static/img/banners/city-%s.jpg", strings.ToLower(p.City)),
				Coords: p.Coords,
			}
		}
		agg[key].Count++
	}

	out := make([]LocationCount, 0, len(agg))
	for _, v := range agg {
		out = append(out, *v)
	}
	// Active country first, then by listing count.
	sort.SliceStable(out, func(i, j int) bool {
		li := out[i].CountryCode == country
		lj := out[j].CountryCode == country
		if li != lj {
			return li
		}
		return out[i].Count > out[j].Count
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func take[T any](items []T, limit int) []T {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

// ---------------------------------------------------------------------------
// BrokerRepository
// ---------------------------------------------------------------------------

// All returns every broker.
func (m *Mock) All(ctx context.Context) []models.Broker { return m.brokers }

// ByID finds a broker by identifier.
func (m *Mock) ByIDBroker(ctx context.Context, id string) (models.Broker, bool) {
	for _, b := range m.brokers {
		if b.ID == id {
			return b, true
		}
	}
	return models.Broker{}, false
}

// BySlug finds a broker by URL slug.
func (m *Mock) BySlugBroker(ctx context.Context, slug string) (models.Broker, bool) {
	for _, b := range m.brokers {
		if b.Slug == slug {
			return b, true
		}
	}
	return models.Broker{}, false
}

// Promoted returns brokers advertised for the active country.
func (m *Mock) Promoted(ctx context.Context, country string, limit int) []models.Broker {
	var local, other []models.Broker
	for _, b := range m.brokers {
		if !b.IsPromoted {
			continue
		}
		if country == "" || b.CountryCode == country {
			local = append(local, b)
		} else {
			other = append(other, b)
		}
	}
	return take(append(local, other...), limit)
}

// Filter narrows brokers by country, city, language and free text.
func (m *Mock) Filter(ctx context.Context, country, city, language, q string) []models.Broker {
	var out []models.Broker
	for _, b := range m.brokers {
		if country != "" && b.CountryCode != country {
			continue
		}
		if city != "" && !strings.EqualFold(b.City, city) {
			continue
		}
		if language != "" && !hasString(b.Languages, language) {
			continue
		}
		if q != "" {
			hay := strings.ToLower(b.Name + " " + b.AgencyName + " " + strings.Join(b.Specialties, " "))
			if !strings.Contains(hay, strings.ToLower(q)) {
				continue
			}
		}
		out = append(out, b)
	}
	return out
}

func hasString(list []string, v string) bool {
	for _, s := range list {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}

// Agencies returns every agency.
func (m *Mock) Agencies(ctx context.Context) []models.Agency { return m.agencies }

// AgencyBySlug finds an agency by URL slug.
func (m *Mock) AgencyBySlug(ctx context.Context, slug string) (models.Agency, bool) {
	for _, a := range m.agencies {
		if a.Slug == slug {
			return a, true
		}
	}
	return models.Agency{}, false
}

// BrokersByAgency lists the brokers working for an agency.
func (m *Mock) BrokersByAgency(ctx context.Context, agencyID string) []models.Broker {
	var out []models.Broker
	for _, b := range m.brokers {
		if b.AgencyID == agencyID {
			out = append(out, b)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// ContentRepository
// ---------------------------------------------------------------------------

// Articles returns all articles, newest first.
func (m *Mock) Articles(ctx context.Context) []models.Article {
	out := append([]models.Article(nil), m.articles...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PublishedAt.After(out[j].PublishedAt) })
	return out
}

// ArticleBySlug finds an article by slug.
func (m *Mock) ArticleBySlug(ctx context.Context, slug string) (models.Article, bool) {
	for _, a := range m.articles {
		if a.Slug == slug {
			return a, true
		}
	}
	return models.Article{}, false
}

// ArticlesByCategory filters articles by category.
func (m *Mock) ArticlesByCategory(ctx context.Context, category string) []models.Article {
	if category == "" {
		return m.Articles(ctx)
	}
	var out []models.Article
	for _, a := range m.Articles(ctx) {
		if strings.EqualFold(a.Category, category) {
			out = append(out, a)
		}
	}
	return out
}

// RelatedArticles returns other articles, preferring the same category.
func (m *Mock) RelatedArticles(ctx context.Context, a models.Article, limit int) []models.Article {
	var same, rest []models.Article
	for _, o := range m.Articles(ctx) {
		if o.ID == a.ID {
			continue
		}
		if o.Category == a.Category {
			same = append(same, o)
		} else {
			rest = append(rest, o)
		}
	}
	return take(append(same, rest...), limit)
}

// Developments returns developments, active country first.
func (m *Mock) Developments(ctx context.Context, country string) []models.Development {
	var local, other []models.Development
	for _, d := range m.developments {
		if country == "" || d.CountryCode == country {
			local = append(local, d)
		} else {
			other = append(other, d)
		}
	}
	return append(local, other...)
}

// DevelopmentBySlug finds a development by slug.
func (m *Mock) DevelopmentBySlug(ctx context.Context, slug string) (models.Development, bool) {
	for _, d := range m.developments {
		if d.Slug == slug {
			return d, true
		}
	}
	return models.Development{}, false
}

// ---------------------------------------------------------------------------
// AccountRepository
// ---------------------------------------------------------------------------

// CurrentUser returns the mock signed-in account.
func (m *Mock) CurrentUser(ctx context.Context) models.User { return m.user }

// Favourites returns the user's saved listings.
func (m *Mock) Favourites(ctx context.Context) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []models.Property
	for _, p := range m.properties {
		if m.favourites[p.ID] {
			out = append(out, p)
		}
	}
	return out
}

// IsFavourite reports whether a listing is saved.
func (m *Mock) IsFavourite(ctx context.Context, id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.favourites[id]
}

// ToggleFavourite flips the saved state and returns the new value.
func (m *Mock) ToggleFavourite(ctx context.Context, id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.favourites[id] = !m.favourites[id]
	return m.favourites[id]
}

// SavedSearches returns the user's stored searches.
func (m *Mock) SavedSearches(ctx context.Context) []models.SavedSearch {
	return buildSavedSearches(m.now)
}

// Notifications returns the notification centre rows.
func (m *Mock) Notifications(ctx context.Context) []models.Notification {
	return buildNotifications(m.now)
}

// UnreadCount is the header badge value.
func (m *Mock) UnreadCount(ctx context.Context) int {
	n := 0
	for _, x := range buildNotifications(m.now) {
		if !x.IsRead {
			n++
		}
	}
	return n
}

// MyListings returns the user's own listings in a given state. The mock
// attributes a slice of the seeded catalogue to the signed-in user.
func (m *Mock) MyListings(ctx context.Context, status models.ListingStatus) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()

	owned := map[string]models.ListingStatus{
		"pr-003": models.StatusActive,
		"pr-006": models.StatusActive,
		"pr-014": models.StatusPending,
		"pr-027": models.StatusExpired,
		"pr-031": models.StatusRejected,
	}
	var out []models.Property
	for _, p := range m.properties {
		st, ok := owned[p.ID]
		if !ok {
			continue
		}
		p.Status = st
		if status == "" || st == status {
			out = append(out, p)
		}
	}
	return out
}

// Drafts returns unfinished add-listing sessions.
func (m *Mock) Drafts(ctx context.Context) []models.Draft { return buildDrafts(m.now) }

// Payments returns billing history.
func (m *Mock) Payments(ctx context.Context) []models.Payment { return buildPayments(m.now) }

// ---------------------------------------------------------------------------
// CatalogRepository
// ---------------------------------------------------------------------------

// Countries lists the markets that carry seeded stock.
//
// This is the short, curated list — the one the mobile drawer's country pills
// and the city datalist are built from. The selector uses AllCountries.
func (m *Mock) Countries(ctx context.Context) []models.Country { return countries }

// AllCountries lists every selectable market: the seeded eight followed by the
// rest of the world. The market selector offers this list so the control is
// exercised at the size the production catalogue will be.
func (m *Mock) AllCountries(ctx context.Context) []models.Country { return allCountries }

// LocationSuggestions is the Location autocomplete list: countries, cities,
// districts, streets and exact addresses drawn from the seeded stock.
func (m *Mock) LocationSuggestions(ctx context.Context) []models.LocationSuggestion {
	return m.locations
}

// OtherCountries is AllCountries minus the seeded markets. The selector heads
// the seeded ones separately, so it needs the tail without them.
func (m *Mock) OtherCountries(ctx context.Context) []models.Country { return otherCountries }

// CountryHasListings reports whether a market carries seeded stock, so the
// selector can group the live markets above the rest.
func (m *Mock) CountryHasListings(ctx context.Context, code string) bool {
	return hasListings(strings.ToUpper(code))
}

// Country finds one market by ISO code, across the full list — a visitor who
// picks a market with no seeded stock still gets a real country back, and the
// pages render their ordinary empty state rather than bouncing to the default.
func (m *Mock) Country(ctx context.Context, code string) (models.Country, bool) {
	for _, c := range allCountries {
		if strings.EqualFold(c.Code, code) {
			return c, true
		}
	}
	return models.Country{}, false
}

// Banner returns the active homepage banner for a market.
//
// Retained for callers that predate placements; it is BannerFor with the
// placement fixed to "home".
func (m *Mock) Banner(ctx context.Context, code string) (models.Banner, bool) {
	return m.BannerFor(ctx, code, "home")
}

// BannerFor returns the banner for a market and placement.
//
// A slot that exists but is switched off returns ok=false: the caller renders
// nothing, while admin still lists the row so it can be turned back on.
func (m *Mock) BannerFor(ctx context.Context, code, placement string) (models.Banner, bool) {
	if placement == "" {
		placement = "home"
	}
	for _, b := range banners {
		if strings.EqualFold(b.CountryCode, code) && b.Slot() == placement {
			if !b.Active {
				return b, false
			}
			return b, true
		}
	}
	return models.Banner{}, false
}

// BannersAll lists every configured slot, active or not.
func (m *Mock) BannersAll(ctx context.Context) []models.Banner { return banners }

// Packages lists the paid listing tiers.
func (m *Mock) Packages(ctx context.Context) []models.Package { return packages }

// Testimonials backs the homepage trust section.
func (m *Mock) Testimonials(ctx context.Context) []models.Testimonial { return testimonials }

// Languages lists translation targets.
func (m *Mock) Languages(ctx context.Context) []models.Language { return buildLanguages(m.now) }

// RestrictedCountries lists admin-configured map restrictions.
func (m *Mock) RestrictedCountries(ctx context.Context) []models.RestrictedCountry {
	return buildRestricted(m.now)
}

// ---------------------------------------------------------------------------
// Adapters
//
// Go interfaces cannot carry two methods with the same name, and the property
// and broker repositories both want ByID/BySlug. These thin wrappers expose the
// one Mock under both interfaces.
// ---------------------------------------------------------------------------

// BrokerAdapter presents Mock as a BrokerRepository.
type BrokerAdapter struct{ *Mock }

// ByID finds a broker by identifier.
func (a BrokerAdapter) ByID(ctx context.Context, id string) (models.Broker, bool) {
	return a.Mock.ByIDBroker(ctx, id)
}

// BySlug finds a broker by slug.
func (a BrokerAdapter) BySlug(ctx context.Context, slug string) (models.Broker, bool) {
	return a.Mock.BySlugBroker(ctx, slug)
}

// NewStore wires the mock into the repository bundle handlers depend on.
func NewStore(now time.Time) *Store {
	m := NewMock(now)
	return &Store{
		Properties: m,
		Brokers:    BrokerAdapter{m},
		Content:    m,
		Account:    m,
		Catalog:    m,
		Admin:      m,
	}
}
