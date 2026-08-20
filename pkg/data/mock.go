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
	// savedSearches are the ones stored from the filter panel while the process
	// runs, oldest first. The seeded four are built on demand and are not in
	// here — see SavedSearches, which puts these in front of them.
	savedSearches []models.SavedSearch
	user          models.User
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
		brokers:      buildBrokers(now),
		agencies:     buildAgencies(),
		articles:     buildArticles(now),
		developments: buildDevelopments(now),
		favourites:   map[string]bool{},
		// The signed-in account is seeded as a broker rather than a bare buyer,
		// because the fields the client added on 17 August — the company name,
		// the company logo, the languages of communication and the markets the
		// seller is active in — only have anything to show on a selling
		// account. A private seller leaves every one of them empty, and both
		// states are reachable from the settings screen.
		//
		// The role is "admin" because this is the account the client walks the
		// demonstration in, and the administration panel is part of what they
		// are reviewing. It is a real permission, not a formality: the account
		// menu draws the Admin panel entry only for an administrator and every
		// /admin route answers everyone else with the 404 page, so seeding this
		// one as "user" instead is all it takes to see what an ordinary visitor
		// sees — which is no link, and no back office behind the URL either.
		user: models.User{
			ID: "us-01", Name: "Anna Lehtinen", Email: "anna.lehtinen@example.com",
			Phone: "+372 5566 7788", Role: "admin", CountryCode: "EE", Language: "en",
			MemberSince: now.AddDate(-2, -4, 0), IsVerified: true,
			Avatar:      "",
			Company:     "Best House Ltd",
			CompanyLogo: "",
			// Off by default in the seed: the label has to be worth something,
			// and the settings screen is where a seller turns it on. Ticking
			// the box there is what the client asked to be possible.
			DirectFromOwner: false,
			// The paragraph the public profile shows under "About Anna
			// Lehtinen", seeded so the settings field opens with something in
			// it and the client can see what editing it changes.
			Bio: "Anna sells apartments and houses across Tallinn and the north " +
				"Estonian coast, and works with Finnish buyers relocating across " +
				"the gulf. Twelve years in the market, most of them in Kalamaja " +
				"and Pirita.",
			// Estonian, English and Finnish: the Tallinn–Helsinki pairing the
			// seeded brokers work in too.
			Languages: []string{"et", "en", "fi"},
			// Active either side of the gulf, which is the client's example of
			// why one country was not enough.
			ActiveCountries: []string{"EE", "FI"},
			Messengers: []models.Messenger{
				// No handle: both are reached on the phone number above.
				{Kind: models.MessengerWhatsApp},
				{Kind: models.MessengerViber},
				// Telegram and Signal carry their own address, which is exactly
				// why the settings form gives each of them a field.
				{Kind: models.MessengerTelegram, Handle: "t.me/annalehtinen"},
			},
			// A pin already dropped, so the settings screen shows the map doing
			// its job rather than an empty rectangle. Rotermann, Tallinn.
			Office: models.MapPlace{
				Label: "Roseni 10, Tallinn, Estonia",
				// What this seller chose to publish, which is not their street:
				// the settings form opens with the field already filled so its
				// purpose is legible at a glance.
				Public: "Rotermanni quarter, Tallinn — visits by appointment",
				Lat:    59.4380, Lng: 24.7530,
			},
			// And two market ads already running — bought on different days
			// and for different lengths, which is the client's case exactly:
			// "at first he wants that his profile is displayed in the German
			// market for 30 days, then he activates it with payment, and then
			// he can activate his ad under France market as well for 30 days
			// with new payment." Estonia was bought a fortnight ago, Finland
			// the day before yesterday, and they expire two weeks apart.
			Ad: models.BrokerAd{
				Runs: []models.BrokerAdRun{
					{
						Country: "EE", Days: 30,
						StartsAt: now.AddDate(0, 0, -12), EndsAt: now.AddDate(0, 0, 18),
					},
					{
						Country: "FI", Days: 30,
						StartsAt: now.AddDate(0, 0, -2), EndsAt: now.AddDate(0, 0, 28),
					},
				},
			},
			// And the map placement, so the settings screen shows a live one
			// rather than only the form for buying it, and this seller's pin
			// is on the search map from the first page load.
			MapAd: models.BrokerMapAd{
				Days: 30, StartsAt: now.AddDate(0, 0, -6), EndsAt: now.AddDate(0, 0, 24),
			},
		},
	}
	m.properties = buildProperties(now, pool)
	m.countBrokerListings()
	m.linkDevelopments()
	m.locations = buildLocationSuggestions(m.properties)

	// A handful of listings start favourited so the account screens have content.
	for _, id := range []string{"pr-002", "pr-007", "pr-013", "pr-019"} {
		m.favourites[id] = true
	}
	return m
}

// countBrokerListings replaces each broker's seeded listing count with the
// number of live listings they actually have.
//
// It matters more than it used to: the client removed the rating, the completed
// sales and the years active from the broker profile — "there is no way we can
// verify it" — and active listings is the one figure left. The site can count
// that one, so it should, rather than keep a hand-written number that said 9
// beside a page showing 4.
func (m *Mock) countBrokerListings() {
	counts := map[string]int{}
	for _, p := range m.properties {
		if p.Status == models.StatusActive {
			counts[p.BrokerID]++
		}
	}
	for i := range m.brokers {
		m.brokers[i].ActiveListings = counts[m.brokers[i].ID]
	}
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

	// How far each match is from the place that was typed, once the order is
	// settled — and deliberately after sortProperties rather than instead of
	// it. The client's rule: "the order of the real-estate ads here is not
	// automatically according to nearest distance. The order stays like it is
	// set up in the 'sort by' menu." So the distance is something a card says,
	// never something the list is arranged by.
	//
	// Measured from the listing's own coordinates with the same haversine the
	// broker directory uses, so "12.40 KM" means the same thing on both pages.
	if f.HasPoint() {
		here := models.MapPlace{Lat: f.Lat, Lng: f.Lng}
		for i := range matched {
			matched[i].Distance = here.DistanceKm(matched[i].Coords.Lat, matched[i].Coords.Lng)
			matched[i].DistanceSet = true
		}
	}

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
	// Deal types are OR-ed, like Types below: a search across Rent and Short
	// rent matches a listing that is either.
	if len(f.Deals) > 0 && !containsDeal(f.Deals, p.Deal) {
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
	// Every ticked amenity, not any of them: two boxes ticked on a filter has
	// always meant "both", and the five booleans above are ANDed for the same
	// reason.
	for _, key := range f.Amenities {
		if !models.HasAmenity(p.Amenities, key) {
			return false
		}
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
	if !p.LanguageMatch(f.Languages) {
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

func containsDeal(list []models.DealType, v models.DealType) bool {
	for _, d := range list {
		if d == v {
			return true
		}
	}
	return false
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
				Image:  cityBanner(p.City),
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

// cityBanner returns the photograph for a city tile, or "" when the mock data
// set has no photograph for that city.
//
// The URL is derived from the city's name, and not every city that carries a
// listing has a picture — Tartu and Pärnu do not — so deriving it blindly put
// two `<img>` elements on the homepage pointing at files that 404. The
// generated asset manifest is the authority on what actually shipped; a city
// with no picture returns nothing and its tile renders without one.
func cityBanner(city string) string {
	base := fmt.Sprintf("/static/img/banners/city-%s", strings.ToLower(city))
	if !assets.HasVariant(base + "-400.webp") {
		return ""
	}
	return base + ".jpg"
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

// Promoted returns the brokers whose paid ad is running in this market.
//
// The client's rule for the homepage strip: "the broker can buy an ad that
// under this country (when the user has chosen his market on the frontpage
// banner) his ad is active." So a market shows the brokers who paid for that
// market and nobody else — an ad bought for Estonia does not quietly fill a
// gap on the German homepage, because a reader in Germany was not what the
// broker paid to reach.
//
// That makes an empty strip possible, and correct: a market nobody advertises
// in has nothing to show, and the homepage drops the section rather than
// padding it out with brokers who did not buy it.
//
// Ordered by run length remaining, longest first, so the strip is stable from
// one page load to the next rather than reshuffling under the reader.
func (m *Mock) Promoted(ctx context.Context, country string, limit int) []models.Broker {
	var out []models.Broker
	for _, b := range m.brokers {
		// An empty country is "any market", which is what the admin
		// advertising screen asks for when it lists every running placement.
		// A market code narrows to the brokers who bought *that* market.
		if country == "" {
			if b.Ad.IsLive(m.now) {
				out = append(out, b)
			}
			continue
		}
		if b.Ad.RunsIn(country, m.now) {
			out = append(out, b)
		}
	}
	// Newest purchase first, which is what makes the homepage strip a queue:
	// "in the frontpage broker section are displayed all the new ads, if next
	// ad will come then the last one will be pushed futher till it disappears
	// from the frontpage."
	//
	// Ordered on when the run was bought rather than on when it expires. Those
	// are different orders — a thirty-day run bought a fortnight ago outlives a
	// seven-day one bought this morning — and only the first of them makes a new
	// advertiser appear at the front, which is what the client described.
	//
	// The caller's limit is what actually pushes anyone off; nothing is dropped
	// here. /brokers passes no limit and shows every live run to the end of its
	// paid period, which is the rest of the same note.
	bought := func(b models.Broker) time.Time {
		if country == "" {
			return b.Ad.BoughtAt(m.now)
		}
		return b.Ad.BoughtIn(country, m.now)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return bought(out[i]).After(bought(out[j]))
	})
	return take(out, limit)
}

// Filter narrows brokers by place, radius, languages spoken and free text.
//
// The directory has two modes, and which one is running is decided by whether
// a place was entered. The client, 18 August:
//
//	"On top in the header is the choose your market button, and so whatever
//	 market is chosen there, then these brokers what have bought ad in this
//	 market are displayed there — so the same as the frontpage broker section.
//	 And if the user in this search menu enters the location and radius, then
//	 the 'choose your market' system is not active any more and now the system
//	 displays there only these brokers what are in this range and in this
//	 order, who is closer."
//
// So with an empty Location the market decides and the list is the market's
// paid strip; with a place typed in, the market is ignored and the radius
// decides, ordered nearest first — the whole point of a radius search is "who
// is close", so the order has to say so.
func (m *Mock) Filter(ctx context.Context, f BrokerFilter) []models.Broker {
	// A place that did not resolve is still a search. It matches as text
	// against the broker's city and country instead of being dropped, so a
	// typed "Tallinn" that missed the suggestion list finds Tallinn brokers
	// rather than everybody.
	loose := ""
	if f.Location != "" && !f.HasPoint() {
		loose = strings.ToLower(f.Location)
	}

	// The market only decides while nothing has been searched for. That is the
	// client's "not active any more", stated once here rather than at each of
	// the call sites that could forget it — see BrokerFilter.IsBrowsing.
	byMarket := f.MarketCode != "" && f.IsBrowsing()

	var out []models.Broker
	for _, b := range m.brokers {
		if byMarket && !b.Ad.RunsIn(f.MarketCode, m.now) {
			continue
		}
		// The map placement is a different purchase from the market strip, so
		// it is asked about separately: a broker advertising in Germany is not
		// thereby on the map, and one on the map has not thereby bought
		// Germany.
		if f.MapAdOnly && !(b.MapAd.IsLive(m.now) && b.Office.IsSet()) {
			continue
		}
		// Where the broker works, as opposed to where they bought an ad. A
		// cross-border broker counts as being in every market they list.
		if f.CountryCode != "" &&
			!strings.EqualFold(b.CountryCode, f.CountryCode) &&
			!hasString(b.ActiveCountries, f.CountryCode) {
			continue
		}
		if f.HasPoint() && f.RadiusKm > 0 {
			// A broker with no pin cannot be placed on the map, so a radius
			// search cannot honestly claim they are inside it.
			if !b.Office.IsSet() {
				continue
			}
			if b.Office.DistanceKm(f.Lat, f.Lng) > float64(f.RadiusKm) {
				continue
			}
		}
		if loose != "" {
			hay := strings.ToLower(b.City + " " + b.CountryCode + " " + CountryName(b.CountryCode) + " " + b.Office.Label)
			if !strings.Contains(hay, loose) {
				continue
			}
		}
		if !b.SpeaksAny(f.Languages) {
			continue
		}
		if f.Q != "" {
			hay := strings.ToLower(b.Name + " " + b.AgencyName + " " + strings.Join(b.Specialties, " "))
			if !strings.Contains(hay, strings.ToLower(f.Q)) {
				continue
			}
		}
		out = append(out, b)
	}

	// How far each one is, on the card. "So on every broker profile now come
	// the 'distance' like it is in sexydate page" — the reference the client
	// sent shows the word, then the number in red, then KM.
	//
	// Set before the sort so the two cannot disagree about which broker is
	// nearest: the order and the number a card prints come from the same
	// measurement rather than from two calls that could drift.
	if f.HasPoint() {
		for i := range out {
			out[i].Distance = out[i].Office.DistanceKm(f.Lat, f.Lng)
			out[i].DistanceSet = true
		}
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Distance < out[j].Distance
		})
	}
	return out
}

// OnMap serves the search page: the brokers who bought the map placement,
// narrowed by the same place and radius the property search ran with.
//
// The client's rule, 18 August: "the system with the brokers is the same as
// with the real estate. If chosen to see the brokers then in the maps like the
// real estate the small broker images are displayed. And in the listings page
// the broker profile-ads are displayed like the real-estate ads." Same search,
// two kinds of result — so this takes the property search's own point and
// radius rather than a second search of its own.
func (m *Mock) OnMap(ctx context.Context, f BrokerFilter) []models.Broker {
	f.MapAdOnly = true
	// Never the market fallback: on the search page a reader who has typed
	// nothing is looking at every market's map, not at their own market's paid
	// strip. Filter() reads MarketCode only when no place was entered, and
	// clearing it here says that outright.
	f.MarketCode = ""
	return m.Filter(ctx, f)
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

// SavedSearches returns the user's stored searches, newest first: anything
// saved from the filter panel this session, then the seeded four.
func (m *Mock) SavedSearches(ctx context.Context) []models.SavedSearch {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]models.SavedSearch, 0, len(m.savedSearches)+4)
	for i := len(m.savedSearches) - 1; i >= 0; i-- {
		out = append(out, m.savedSearches[i])
	}
	return append(out, buildSavedSearches(m.now)...)
}

// AddSavedSearch stores one and hands it back with its id.
//
// Kept in memory, like the favourites above: this milestone has no database,
// and a button that claims to have saved something while /saved-searches shows
// nothing new is the thing the client would find first.
func (m *Mock) AddSavedSearch(ctx context.Context, s models.SavedSearch) models.SavedSearch {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s.ID == "" {
		s.ID = fmt.Sprintf("ss-u%02d", len(m.savedSearches)+1)
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = m.now
	}
	m.savedSearches = append(m.savedSearches, s)
	return s
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

// ownedListings attributes a slice of the seeded catalogue to the signed-in
// user, with the state each one is in.
//
// Three states, one of each plus a second active listing: draft (entered, not
// paid for), active (paid and online), expired (the paid period has run out).
// Pending review and rejected are gone — nothing moderates a listing before it
// goes live, so neither state could ever occur.
var ownedListings = map[string]models.ListingStatus{
	"pr-003": models.StatusActive,
	"pr-006": models.StatusActive,
	"pr-014": models.StatusDraft,
	"pr-027": models.StatusExpired,
	"pr-031": models.StatusDraft,
}

// MyListings returns the user's own listings in a given state.
func (m *Mock) MyListings(ctx context.Context, status models.ListingStatus) []models.Property {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []models.Property
	for _, p := range m.properties {
		st, ok := ownedListings[p.ID]
		if !ok {
			continue
		}
		p.Status = st
		// A draft has never been paid for, so it has no expiry to show. The
		// seeded catalogue gives every property one; leaving it in place would
		// have the listings table promise a draft goes offline on a date it was
		// never online for.
		if st == models.StatusDraft {
			p.ExpiresAt = time.Time{}
		}
		if status == "" || st == status {
			out = append(out, p)
		}
	}
	return out
}

// ListingStats builds the per-day visitor series for one of the user's own
// listings.
//
// The shape is invented, but not arbitrarily: it is derived from the listing's
// own lifetime view count so a busy listing charts busier than a quiet one, and
// from its id so the same listing draws the same chart on every request —
// nothing here may move between two page loads or the panel looks broken.
//
// Two things are modelled because they are what a real series looks like and
// what makes the panel worth reading: traffic decays as a listing ages, and
// weekends are quieter than weekdays. A backend replacing this reads from a
// page-view table and can drop the shaping entirely.
func (m *Mock) ListingStats(ctx context.Context, propertyID string, days int) (models.ListingStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := ownedListings[propertyID]; !ok {
		return models.ListingStats{}, false
	}
	if days <= 0 {
		days = 14
	}

	var p models.Property
	found := false
	for _, cand := range m.properties {
		if cand.ID == propertyID {
			p, found = cand, true
			break
		}
	}
	if !found {
		return models.ListingStats{}, false
	}

	// A small integer hash of the id, so two listings with similar view counts
	// still draw visibly different charts.
	seed := 0
	for _, r := range propertyID {
		seed = seed*31 + int(r)
	}
	if seed < 0 {
		seed = -seed
	}

	// A daily baseline from the lifetime total. The +1 keeps a listing with no
	// views at all from flooring every day to zero and dividing by nothing.
	base := p.Views/(days*3) + 1

	stats := models.ListingStats{
		PropertyID: p.ID,
		Title:      p.Title,
		Views:      p.Views,
		Saves:      p.Saves,
		Days:       make([]models.DayCount, 0, days),
	}
	// Oldest first, so the chart reads left to right.
	start := m.now.AddDate(0, 0, -(days - 1))
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)

		// Decay: the oldest day of the window carries about 1.6× the newest.
		weight := 160 - (i*60)/days

		// Saturday and Sunday run about a third quieter.
		switch d.Weekday() {
		case time.Saturday, time.Sunday:
			weight = weight * 68 / 100
		}

		// Deterministic jitter, ±25%, from the id and the day index.
		jitter := 75 + (seed+i*37)%51

		v := base * weight / 100 * jitter / 100
		if v < 0 {
			v = 0
		}
		stats.Days = append(stats.Days, models.DayCount{
			Date:  d,
			Label: d.Format("2 Jan"),
			Views: v,
		})
	}
	return stats, true
}

// CloneListing duplicates one of the user's listings as a new draft.
//
// The copy is a draft whatever the original was, because a duplicate has not
// been paid for — the client's lifecycle has no other way in. Its metrics start
// at zero rather than inheriting the original's, which would otherwise report
// views a brand-new listing has not had, and its title is marked so the two are
// tellable apart in the table.
//
// Nothing is persisted: this build has no store to write to, so the caller
// reports what would have been created. The shape returned is what a real
// implementation would insert.
func (m *Mock) CloneListing(ctx context.Context, propertyID string) (models.Property, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := ownedListings[propertyID]; !ok {
		return models.Property{}, false
	}
	for _, p := range m.properties {
		if p.ID != propertyID {
			continue
		}
		c := p
		c.ID = p.ID + "-copy"
		c.Title = p.Title + " (copy)"
		c.Slug = p.Slug + "-copy"
		c.Status = models.StatusDraft
		c.Views, c.Saves = 0, 0
		c.IsFeatured = false
		c.CreatedAt, c.UpdatedAt = m.now, m.now
		c.ExpiresAt = time.Time{} // nothing expires until it is paid for
		return c, true
	}
	return models.Property{}, false
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

// Promotions lists the paid add-ons a seller can buy per day, both at publish
// time and later from their own listing management.
func (m *Mock) Promotions(ctx context.Context) []models.Promotion { return promotions }

// BrokerAdPlan returns the homepage placement's price list.
func (m *Mock) BrokerAdPlan(ctx context.Context) models.BrokerAdPlan { return brokerAdPlan }

// BrokerMapAdPlan returns the map placement's price list.
func (m *Mock) BrokerMapAdPlan(ctx context.Context) models.BrokerMapAdPlan {
	return brokerMapAdPlan
}

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
