// Package models holds the Previa domain types.
//
// These structs are the contract between the data layer and the templates.
// The current milestone serves them from an in-memory mock provider; a future
// MySQL-backed provider must return these same shapes so no template changes.
package models

import (
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Geography and money
// ---------------------------------------------------------------------------

// Country is a market Previa operates in. The active country drives the
// homepage banner, promoted brokers, initial map centre and default currency.
type Country struct {
	Code     string // ISO 3166-1 alpha-2, e.g. "EE"
	Name     string
	Flag     string // emoji flag, rendered next to the name in the selector
	Currency string // ISO 4217, e.g. "EUR"
	Locale   string // default content language for the market
	Lat, Lng float64
	Zoom     int
	Cities   []string
}

// Money keeps an amount together with its currency so templates never have to
// guess how to format a price.
type Money struct {
	Amount   float64
	Currency string
}

// Coordinates is a map point.
type Coordinates struct {
	Lat float64
	Lng float64
}

// Place kinds a location suggestion can represent, broad to narrow. These are
// the five granularities the client asked the single Location field to accept.
const (
	PlaceCountry  = "country"
	PlaceCity     = "city"
	PlaceDistrict = "district"
	PlaceStreet   = "street"
	PlaceAddress  = "address"
)

// LocationSuggestion is one entry in the Location autocomplete, and also the
// shape a map click resolves to.
//
// It carries both the display label and the structured fields behind it, so a
// selection can fill a search filter or the add-listing form's read-only
// address fields without a second lookup. A Google Places result maps onto this
// struct field for field.
type LocationSuggestion struct {
	Kind        string // one of the Place* constants
	Label       string // what the field shows, e.g. "Kalamaja, Tallinn, Estonia"
	CountryCode string
	Country     string
	State       string
	City        string
	District    string
	Address     string
	PostalCode  string
	Lat, Lng    float64
}

// ---------------------------------------------------------------------------
// Enumerations
// ---------------------------------------------------------------------------

// DealType separates the primary search modes.
//
// The client's wording is Sell / Rent / Short rent, used identically in the
// homepage search, the sidebar filter, the map filter and the add-listing form.
// The stored values stay lowercase and stable so a URL keeps working when the
// labels are translated.
type DealType string

const (
	DealSale      DealType = "sale"
	DealRent      DealType = "rent"
	DealShortRent DealType = "short_rent"
)

// DealTypes is the canonical ordered list, so every control that offers a deal
// type offers the same three in the same order.
var DealTypes = []struct {
	Value DealType
	Label string
}{
	{DealSale, "Sell"},
	{DealRent, "Rent"},
	{DealShortRent, "Short rent"},
}

// IsRentalDeal reports whether a deal type is one of the renting modes, which
// is what decides whether a price carries a period at all.
func IsRentalDeal(d DealType) bool { return d == DealRent || d == DealShortRent }

// PricePeriod is the unit a price of this deal type is quoted in: a month for
// a rental, a day for a short rental, and nothing at all for a sale.
//
// The client's rule, stated about the add-listing form: "if the user has
// chosen deal type rent or short, then the price must be in case of rent
// 1000 € / month and in case of short rent 100 € / day." It lives here rather
// than in the form because the form is only where the number is typed — the
// cards, the listing page and the map popup all have to quote the same unit,
// and a short rental priced per month was what they said before this.
func PricePeriod(d DealType) string {
	switch d {
	case DealRent:
		return "month"
	case DealShortRent:
		return "day"
	default:
		return ""
	}
}

// PropertyType is the physical category of a listing.
//
// Villa is gone: the client treats it as the same thing as a house, so those
// listings were reclassified. The values are lowercase and hyphenated because
// they travel in the URL — `property_type=modular-house` — and must stay
// stable when the labels are translated.
type PropertyType string

const (
	TypeApartment      PropertyType = "apartment"
	TypeHouse          PropertyType = "house"
	TypeHousePart      PropertyType = "house-part"
	TypeCottage        PropertyType = "cottage"
	TypeModularHouse   PropertyType = "modular-house"
	TypePanelizedHouse PropertyType = "panelized-house"
	TypeTrailerHouse   PropertyType = "trailer-house"
	TypeSauna          PropertyType = "sauna"
	TypeCommercial     PropertyType = "commercial"
	TypeIndustrial     PropertyType = "industrial"
	TypeLand           PropertyType = "land"
	TypeGarage         PropertyType = "garage"
	TypeNewDevelopment PropertyType = "new-development"
)

// PropertyTypes is the catalogue, in the order the client listed it. Every
// control that offers a property type reads this, so the homepage picker, the
// filter sidebar, the map filter and the add-listing form cannot drift apart.
//
// Icon names refer to web/templates/components/icons.html.
var PropertyTypes = []struct {
	Value PropertyType
	Label string
	Icon  string
}{
	{TypeApartment, "Apartment", "building"},
	{TypeHouse, "House", "home"},
	{TypeHousePart, "House part", "house-part"},
	{TypeCottage, "Cottage", "cottage"},
	{TypeModularHouse, "Modular house", "modular-house"},
	{TypePanelizedHouse, "Panelized house", "panelized-house"},
	{TypeTrailerHouse, "Trailer house", "trailer-house"},
	{TypeSauna, "Sauna", "sauna"},
	{TypeCommercial, "Commercial", "commercial"},
	{TypeIndustrial, "Industrial property", "industrial"},
	{TypeLand, "Land", "land"},
	{TypeGarage, "Garage", "garage"},
	{TypeNewDevelopment, "New development", "development"},
}

// IsPropertyType reports whether a raw value names a real category. Used when
// parsing a URL so a hand-edited `property_type` cannot filter on nothing.
func IsPropertyType(v string) bool {
	for _, t := range PropertyTypes {
		if string(t.Value) == v {
			return true
		}
	}
	return false
}

// ListingStatus is the lifecycle state of a listing.
//
// Three states, and they describe payment rather than moderation. The client's
// 17 August correction: "the listings with labels 'pending review' and
// 'rejected' should be what? There is no pending review as noone will not look
// the ads before publishing, and so nothing is rejected."
//
//	Draft    entered but not activated — the seller has not paid for it yet.
//	         Not public, and it stays here indefinitely
//	Active   paid for and online, running until ExpiresAt
//	Expired  the paid period has ended. Functionally a draft again — "it is the
//	         same as draft, just info for the user that listing active period is
//	         expired. So the user can edit it, clone it, re-activate it, or
//	         delete it" — but a distinct label, because the seller needs to know
//	         which of the two it is
//
// Sold is not part of that cycle: it is a seller marking an outcome, and it is
// what the archive is built from.
type ListingStatus string

const (
	StatusActive  ListingStatus = "active"
	StatusDraft   ListingStatus = "draft"
	StatusExpired ListingStatus = "expired"
	StatusSold    ListingStatus = "sold"
)

// IsEditable reports whether a listing is off the market and can be worked on
// freely — edited, re-activated or deleted without taking anything down.
//
// Draft and expired are the same state as far as the seller's options go, which
// is exactly how the client described expired. Keeping the distinction in one
// predicate rather than repeating `or draft expired` through the templates is
// what stops the two drifting apart.
func (s ListingStatus) IsEditable() bool {
	return s == StatusDraft || s == StatusExpired
}

// SellerKind distinguishes professional listings from private ones.
type SellerKind string

const (
	SellerBroker  SellerKind = "broker"
	SellerPrivate SellerKind = "private"
)

// ---------------------------------------------------------------------------
// Messenger contact
// ---------------------------------------------------------------------------

// MessengerKind is one of the chat apps a seller can be reached on.
//
// The five the client asked for, in the order they appear beside the phone
// field in the add-listing form and on the listing itself.
type MessengerKind string

const (
	MessengerWhatsApp MessengerKind = "whatsapp"
	MessengerTelegram MessengerKind = "telegram"
	MessengerViber    MessengerKind = "viber"
	MessengerSignal   MessengerKind = "signal"
	MessengerTeams    MessengerKind = "teams"
)

// MessengerKinds lists the supported apps in display order. Templates range
// over this so the form, the listing and the card can never disagree about
// which apps exist or what order they come in.
var MessengerKinds = []MessengerKind{
	MessengerWhatsApp, MessengerTelegram, MessengerViber, MessengerSignal, MessengerTeams,
}

// MessengerLabel is the app's display name.
func MessengerLabel(k MessengerKind) string {
	switch k {
	case MessengerWhatsApp:
		return "WhatsApp"
	case MessengerTelegram:
		return "Telegram"
	case MessengerViber:
		return "Viber"
	case MessengerSignal:
		return "Signal"
	case MessengerTeams:
		return "Teams"
	}
	return string(k)
}

// Messenger is one enabled chat app for a listing.
//
// Handle is whatever the seller entered. Most apps are reached from the phone
// number, so Handle is normally empty and the listing's phone is used; Telegram
// and Signal have their own link fields in the form because a Telegram account
// is not necessarily findable by number, and Signal's is a share link rather
// than a number.
type Messenger struct {
	Kind   MessengerKind
	Handle string
}

// Label is the app's display name.
func (m Messenger) Label() string { return MessengerLabel(m.Kind) }

// Link builds the URL that opens a chat with this contact.
//
// phone is the listing's contact number, used by the apps that address people
// by number. An entry that cannot produce a usable link returns "", and the
// caller skips it rather than rendering a dead icon.
//
// Telegram deliberately accepts both forms the client described: a username
// link (t.me/name) and a phone link (t.me/+372…). Pasting a whole URL works
// too, for any of the apps, so a seller can copy their share link straight out
// of the app.
func (m Messenger) Link(phone string) string {
	h := strings.TrimSpace(m.Handle)
	if strings.HasPrefix(h, "http://") || strings.HasPrefix(h, "https://") {
		return h
	}

	// The number to address, digits only. Deep links reject spaces and the
	// plus sign has to be either dropped or escaped depending on the scheme.
	num := digitsOnly(h)
	if num == "" {
		num = digitsOnly(phone)
	}

	switch m.Kind {
	case MessengerWhatsApp:
		if num == "" {
			return ""
		}
		return "https://wa.me/" + num

	case MessengerTelegram:
		// A username, with or without its leading @.
		if name := strings.TrimPrefix(h, "@"); name != "" && digitsOnly(name) != name {
			return "https://t.me/" + name
		}
		if num == "" {
			return ""
		}
		return "https://t.me/+" + num

	case MessengerViber:
		if num == "" {
			return ""
		}
		return "viber://chat?number=%2B" + num

	case MessengerSignal:
		if num == "" {
			return ""
		}
		return "https://signal.me/#p/+" + num

	case MessengerTeams:
		// Teams addresses people by email, so a handle is required — there is
		// nothing sensible to fall back to.
		if h == "" || !strings.Contains(h, "@") {
			return ""
		}
		return "https://teams.microsoft.com/l/chat/0/0?users=" + url.QueryEscape(h)
	}
	return ""
}

// digitsOnly strips everything that is not a digit, so "+372 5566 7788" and
// "+372-5566-7788" produce the same deep link.
func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Condition describes the state of the building.
type Condition string

const (
	ConditionNew        Condition = "new"
	ConditionRenovated  Condition = "renovated"
	ConditionGood       Condition = "good"
	ConditionSatisfying Condition = "satisfying"
	ConditionNeedsWork  Condition = "needs_work"
)

// LocationPrecision controls how exactly a listing's position is shown to the
// public. Chosen by the seller in the add-listing wizard.
type LocationPrecision string

const (
	PrecisionExact  LocationPrecision = "exact"  // exact pin and street address
	PrecisionStreet LocationPrecision = "street" // street name, approximate pin
	PrecisionArea   LocationPrecision = "area"   // district circle only
)

// ---------------------------------------------------------------------------
// Property
// ---------------------------------------------------------------------------

// Image is one photo attached to a listing, development or article.
type Image struct {
	URL    string
	Alt    string
	Width  int
	Height int
}

// Feature is an amenity chip shown on the detail page.
type Feature struct {
	Key   string
	Label string
	Icon  string
}

// Property is a single listing. It carries every field the search filters can
// query so the mock provider and a future SQL provider filter identically.
type Property struct {
	ID         string
	Slug       string
	Title      string
	Deal       DealType
	Type       PropertyType
	Status     ListingStatus
	Price      Money
	PricePerM2 float64
	RentPeriod string // "month" for rentals, empty for sales

	// Location
	CountryCode string
	Country     string
	City        string
	District    string
	Address     string
	PostalCode  string
	Coords      Coordinates
	Precision   LocationPrecision

	// Dimensions
	Rooms       int
	Bedrooms    int
	Bathrooms   int
	Area        float64 // m² of interior space
	LandArea    float64 // m² of plot, 0 when not applicable
	Floor       int
	TotalFloors int
	BuildYear   int

	// Qualities
	Condition    Condition
	EnergyRating string // A..G
	Furnished    bool
	Parking      bool
	Balcony      bool
	Garden       bool
	Elevator     bool
	Terrace      bool
	Sauna        bool
	SeaView      bool

	// Amenities is every tick from the "Features and amenities" catalogue this
	// listing carries, as AmenityGroups keys. The eight booleans above are the
	// ones the card and the detail page draw as icons and are mirrored in here;
	// the rest exist only as keys, because they are search terms rather than
	// things the listing page has a row for.
	Amenities []string

	// Presentation
	Description string
	Highlights  []string
	Features    []Feature
	Images      []Image

	// Contact
	ContactPhone string
	// Messengers the seller enabled, in display order. Empty for a listing
	// that is only reachable by phone, email or the Previa message form.
	Messengers []Messenger

	// Languages the property is sold in, as ISO 639-1 codes.
	//
	// The client's words for the search filter are "the languages in what the
	// property is sold", which is a property of the listing rather than a
	// lookup through to whoever happens to own it today. It is copied from the
	// seller's own "languages of communication" when the listing is created —
	// the broker's for an agency listing, the account's for a private one — so
	// a broker changing which languages they deal in does not silently rewrite
	// the terms of listings already published.
	//
	// Empty means the seller has not said, and the filter treats that as
	// "unknown" rather than "none": see LanguageMatch below.
	Languages []string

	// Relations and flags
	SellerKind    SellerKind
	BrokerID      string
	AgencyID      string
	DevelopmentID string
	IsFeatured    bool
	// DirectFromOwner marks a listing the owner of the property is selling
	// themselves, and it earns a label of its own wherever the listing appears.
	//
	// Not the same question as SellerKind. A private seller is someone without
	// an agency behind them, which is usually but not always the owner — an
	// heir, a landlord's relative or a company officer can all list privately
	// without owning anything. And a broker can be selling their own flat. The
	// client's reason for wanting it stated outright: "sometimes it is important
	// to note that the property owner itself is selling it."
	//
	// So it is a claim the seller makes on their profile rather than something
	// inferred, and it travels from there onto their listings.
	DirectFromOwner  bool
	IsNewDevelopment bool
	IsVerified       bool

	// Metrics
	Views     int
	Saves     int
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time

	// Distance is how far this listing is from the place the reader typed into
	// the search, in kilometres, and DistanceSet says whether it was measured
	// at all.
	//
	// Two fields rather than treating zero as "no distance": a listing sitting
	// exactly on the searched point really is 0.00 km away, and the card has to
	// say so. The first version of this used zero as the sentinel and dropped
	// the line from precisely the listing the reader had searched for.
	//
	// The client's rule about it is precise: "if the user has entered the
	// location to googlemaps then all the ads displayed have on bottom the
	// distance and distance number is red. The order of the real-estate ads
	// here is not automatically according to nearest distance. The order stays
	// like it is set up in the 'sort by' menu." So this is something a card
	// *says*, never something the result set is ordered by — which is why it
	// lives here rather than in SortOrder.
	Distance    float64
	DistanceSet bool
}

// PrimaryImage returns the first image, or a placeholder when a listing has no
// photos yet (drafts often do not).
func (p Property) PrimaryImage() Image {
	if len(p.Images) > 0 {
		return p.Images[0]
	}
	return Image{URL: "/static/img/placeholder-property.svg", Alt: p.Title, Width: 800, Height: 600}
}

// DistanceLabel is the card's "12.40 KM", to two decimals like the reference
// the client sent. Empty when no place was searched, so the line disappears
// rather than claiming a distance of zero from nowhere.
func (p Property) DistanceLabel() string {
	if !p.DistanceSet {
		return ""
	}
	return strconv.FormatFloat(p.Distance, 'f', 2, 64) + " KM"
}

// IsRental reports whether the listing is a rental, used for price suffixes.
// Short rentals count: they are quoted per period too.
func (p Property) IsRental() bool { return IsRentalDeal(p.Deal) }

// PricePeriod is the unit this listing's price is quoted in — "month", "day"
// or empty for a sale. Templates write "/{{ .PricePeriod }}" after the figure,
// so a short rental reads €100/day rather than the €100/month every rental
// used to claim.
func (p Property) PricePeriod() string { return PricePeriod(p.Deal) }

// LanguageMatch reports whether this listing is sold in any of the requested
// languages.
//
// Two deliberate rules, both of them because the filter is optional:
//
//   - asking for nothing matches everything. Selecting no language is not a
//     constraint, so it must not remove a single result
//   - a listing that has named no languages matches nothing once a language is
//     asked for. The alternative — letting unknowns through — would mean the
//     filter never actually narrowed anything, since most listings would carry
//     an empty list and sail past it
func (p Property) LanguageMatch(codes []string) bool {
	if len(codes) == 0 {
		return true
	}
	for _, want := range codes {
		for _, has := range p.Languages {
			if strings.EqualFold(has, want) {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// People and organisations
// ---------------------------------------------------------------------------

// Broker is an individual agent. "Broker" is the client's confirmed term; do
// not rename to "agent" in the UI.
type Broker struct {
	ID         string
	Slug       string
	Name       string
	Title      string
	AgencyID   string
	AgencyName string
	// CompanyLogo is the agency's mark, shown under the broker's contact
	// details on a listing and on their profile. Optional: a broker who has not
	// uploaded one simply does not get the block, and their agency's name still
	// reads beside their own.
	CompanyLogo string
	Photo       string
	Phone       string
	Email       string
	CountryCode string
	City        string
	// ActiveCountries are every market this broker works in, home market first.
	// A broker near a border can be active in two or three, which is the case
	// the single CountryCode above could not describe.
	ActiveCountries []string
	// Office is where this broker put their pin on the map, set on their own
	// profile: "each broker can specify under his profile his location on
	// googlemaps and then here in the broker search section the user enters the
	// location and radius, so from this location and with this radius (50 km)
	// the brokers are displayed."
	//
	// A broker who has not placed a pin has a zero Office, and BrokersNear
	// leaves them out of a radius search rather than treating (0,0) — a point
	// in the Atlantic — as their address.
	Office MapPlace
	// Languages are the languages the broker will deal in, as language codes.
	// The search filter matches against these.
	Languages   []string
	Specialties []string
	Bio         string
	// Messengers are the chat apps this broker can be reached on, exactly as a
	// listing and a user account carry them.
	//
	// "Under the broker / seller profile on the right where is his phone
	// number, if the user has marked any social apps, then these will be listed
	// there as well… and they are everywhere hyperlinks, if click on them they
	// will open with this seller's contact." The apps under a listing come from
	// the listing; these are the broker's own, so their profile can show them
	// beside the number they are reached on.
	Messengers     []Messenger
	Rating         float64
	Reviews        int
	ActiveListings int
	SoldCount      int
	YearsActive    int
	// Ad is the broker's paid placement in the market broker strips. Empty for
	// a broker who has not bought one, which is most of them.
	Ad BrokerAd
	// MapAd is the other paid service: this broker's pin on the search map and
	// their card among the search results. Empty for a broker who has not
	// bought it.
	MapAd      BrokerMapAd
	IsVerified bool

	// Distance is how far this broker's pin is from the place the reader
	// searched, in kilometres, and DistanceSet says whether it was measured at
	// all — which is why neither is part of what a broker *is*.
	//
	// Two fields for the same reason a listing has two: a broker whose office
	// is the place that was searched is 0.00 km away, not undisclosed.
	//
	// It exists because the client asked for it on the card: "on every broker
	// profile now come the 'distance' like it is in sexydate page." Computing
	// it in the template would mean handing every card the search point as well
	// as the broker, and the two could then disagree.
	Distance    float64
	DistanceSet bool
}

// IsPromoted reports whether this broker currently holds a paid placement
// anywhere. Templates use it for the "Promoted" badge, which is a statement
// about the broker rather than about one market, so it does not take a country;
// where a strip runs is BrokerAd.RunsIn.
func (b Broker) IsPromoted() bool { return b.Ad.IsLive(time.Now()) }

// IsOnMap reports whether this broker has bought the map placement and has a
// pin for it to place. Both halves are required: an unpaid pin is private, and
// a paid placement with nowhere to put it cannot be drawn.
func (b Broker) IsOnMap() bool { return b.MapAd.IsLive(time.Now()) && b.Office.IsSet() }

// DistanceLabel is the card's "75.78 KM", to two decimals like the reference
// the client sent. Empty when no place was searched, so the line disappears
// rather than claiming a distance of zero from nowhere.
func (b Broker) DistanceLabel() string {
	if !b.DistanceSet {
		return ""
	}
	return strconv.FormatFloat(b.Distance, 'f', 2, 64) + " KM"
}

// BrokerAd is a broker's paid placement in the market broker strips.
//
// The client's design, 18 August:
//
//	"So the broker can buy an ad so that his profile is displayed under
//	 Germany's market, if the user has chosen his market to be Germany. One
//	 broker can buy these ads in the frontpage under different countries, but
//	 he must pay separately for each country. At first he wants that his
//	 profile is displayed in the German market for 30 days, then he activates
//	 it with payment, and then he can activate his ad under France market as
//	 well for 30 days with new payment."
//
// So the unit that is bought is not the ad — it is **one market's run**. Each
// carries its own length, its own start and its own expiry, because each was
// paid for on its own day. A single Countries list with one EndsAt could not
// describe that: activating France a fortnight after Germany would either cut
// Germany short or hand France a free two weeks.
//
// It is deliberately a different thing from the map pin the directory searches
// on: a broker has two locations, and they answer two questions. The markets
// here are where they want to be advertised — chosen on their profile — while
// Office is the single point a buyer measures a radius from. The market strips
// read the first; /brokers reads the second when a place is typed into it.
//
// Nothing here stores a photograph, a phone number or a name: the placement
// points at the profile and the strip renders whatever the profile says today,
// which is the client's "if he updates his profile by changing photo or phone,
// then this will be updated in this ad immediately".
type BrokerAd struct {
	// Runs is one entry per market bought, newest purchase last.
	Runs []BrokerAdRun
}

// BrokerAdRun is one market's paid run: bought on its own, paid for on its own,
// and expiring on its own.
type BrokerAdRun struct {
	// Country is the market this run advertises in, as an ISO 3166-1 alpha-2
	// code.
	Country  string
	Days     int
	StartsAt time.Time
	EndsAt   time.Time
}

// IsLive reports whether this market's run is still going.
func (r BrokerAdRun) IsLive(now time.Time) bool {
	return r.Country != "" && !r.EndsAt.IsZero() && now.Before(r.EndsAt)
}

// DaysLeft is what the profile shows: whole days remaining, rounded up, so a
// run with six hours to go reads as "1 day left" rather than "0".
func (r BrokerAdRun) DaysLeft(now time.Time) int {
	if !r.IsLive(now) {
		return 0
	}
	return int(math.Ceil(r.EndsAt.Sub(now).Hours() / 24))
}

// IsLive reports whether any market's run is still going.
func (a BrokerAd) IsLive(now time.Time) bool {
	for _, r := range a.Runs {
		if r.IsLive(now) {
			return true
		}
	}
	return false
}

// RunsIn reports whether a market's run is live right now.
//
// It takes the clock rather than reading it, so a caller that has already
// decided what "now" means — the mock store holds one, and the tests set it —
// cannot disagree with this method about whether an ad has expired.
func (a BrokerAd) RunsIn(code string, now time.Time) bool {
	for _, r := range a.Runs {
		if r.IsLive(now) && strings.EqualFold(r.Country, code) {
			return true
		}
	}
	return false
}

// LiveRuns are the runs still going, in the order they were bought.
func (a BrokerAd) LiveRuns(now time.Time) []BrokerAdRun {
	var out []BrokerAdRun
	for _, r := range a.Runs {
		if r.IsLive(now) {
			out = append(out, r)
		}
	}
	return out
}

// Countries are the markets currently being advertised in. A method rather
// than a field: an expired run must stop counting as a market without anybody
// having to remember to delete it.
func (a BrokerAd) Countries(now time.Time) []string {
	var out []string
	for _, r := range a.LiveRuns(now) {
		out = append(out, r.Country)
	}
	return out
}

// BoughtIn is when this market's run was paid for, and zero when there is no
// live run in it.
//
// It is what the homepage strip orders on: "in the frontpage broker section are
// displayed all the new ads, if next ad will come then the last one will be
// pushed futher till it disappears from the frontpage." Newest purchase first,
// so a ninth advertiser pushes the oldest one off the two rows the homepage
// shows — and only off the homepage. The directory keeps every live run until
// its period ends, which is the other half of the same note: "in the broker
// page it stays till to the end of payd periode."
func (a BrokerAd) BoughtIn(code string, now time.Time) time.Time {
	var newest time.Time
	for _, r := range a.Runs {
		if r.IsLive(now) && strings.EqualFold(r.Country, code) && r.StartsAt.After(newest) {
			newest = r.StartsAt
		}
	}
	return newest
}

// BoughtAt is the most recent purchase across every live market, used when the
// list is not narrowed to one of them.
func (a BrokerAd) BoughtAt(now time.Time) time.Time {
	var newest time.Time
	for _, r := range a.LiveRuns(now) {
		if r.StartsAt.After(newest) {
			newest = r.StartsAt
		}
	}
	return newest
}

// EndsAt is the last moment any of this ad's runs is still up.
func (a BrokerAd) EndsAt(now time.Time) time.Time {
	var last time.Time
	for _, r := range a.LiveRuns(now) {
		if r.EndsAt.After(last) {
			last = r.EndsAt
		}
	}
	return last
}

// DaysLeft is the longest-running market's remaining days, which is how long
// the broker is still advertising somewhere.
func (a BrokerAd) DaysLeft(now time.Time) int {
	most := 0
	for _, r := range a.LiveRuns(now) {
		if d := r.DaysLeft(now); d > most {
			most = d
		}
	}
	return most
}

// BrokerMapAd is the second paid service the client described, 18 August:
//
//	"And the second thing what the broker can do, is to choose under his
//	 profile his googlemaps location like described earlier. And in this case,
//	 if he wants (this will be a paid service) he can activate that his broker
//	 profile is displayed in the googlemaps like the ads."
//
// So it buys one thing: the broker's pin on the search map, and with it their
// card among the search results, for as long as it is paid for. It is bought
// once rather than per market — the pin is a single point on one map, not a
// placement repeated in every country — which is the whole reason it is its own
// type and not another BrokerAdRun.
//
// A broker who has not placed a pin can still buy it; they simply do not appear
// until they drop one, and the profile says so.
type BrokerMapAd struct {
	Days     int
	StartsAt time.Time
	EndsAt   time.Time
}

// IsLive reports whether the map placement is paid up.
func (a BrokerMapAd) IsLive(now time.Time) bool {
	return !a.EndsAt.IsZero() && now.Before(a.EndsAt)
}

// DaysLeft rounds up for the same reason a run's does.
func (a BrokerMapAd) DaysLeft(now time.Time) int {
	if !a.IsLive(now) {
		return 0
	}
	return int(math.Ceil(a.EndsAt.Sub(now).Hours() / 24))
}

// MapPlace is a pin: a point, and the address the person picked it at.
//
// Shared by the broker's office and the seller's profile location, because
// they are the same thing entered on the same control — a Google-Maps-style
// search box over a map — and the broker directory reads both through it.
type MapPlace struct {
	Label string // what the picker showed, e.g. "Roseni 10, Tallinn, Estonia"

	// Public is what other people see, and the only part of a pin that is ever
	// shown to them: "make this location logo icon with the text what the
	// seller entered into the field 'edit your location like you want other
	// users to see it'".
	//
	// The same split the add-listing form already draws between a listing's
	// public display address and the geocoder output beneath it, applied to the
	// profile pin. Label and the coordinates stay internal — they place the
	// broker in a radius search; Public is the line on the listing page.
	Public string

	Lat, Lng float64
}

// Shown is the pin's public line, and empty when the owner has not written one.
// Deliberately no fall back to Label: Label is the street the pin sits on, and
// the whole point of the public line is that the street is not published unless
// its owner chooses to publish it.
func (p MapPlace) Shown() string { return p.Public }

// IsSet reports whether a pin was actually placed. Latitude and longitude are
// both zero only at a point off the coast of Ghana, so this is safe to treat as
// "empty" for the addresses this site deals in.
func (p MapPlace) IsSet() bool { return p.Lat != 0 || p.Lng != 0 }

// DistanceKm is the great-circle distance from this pin to a point, in
// kilometres — the measure behind the broker directory's radius.
//
// Haversine rather than a flat approximation: Previa's markets run from Lisbon
// to Helsinki, and at 60°N a degree of longitude is half the length it is at
// the equator. Treating the two as interchangeable would make a 50 km search in
// Helsinki reach twice as far east-west as one in Barcelona.
func (p MapPlace) DistanceKm(lat, lng float64) float64 {
	const earthKm = 6371.0
	rad := func(d float64) float64 { return d * math.Pi / 180 }

	dLat := rad(lat - p.Lat)
	dLng := rad(lng - p.Lng)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(p.Lat))*math.Cos(rad(lat))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthKm * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// SpeaksAny reports whether the broker deals in any of the given language
// codes. An empty request matches everyone: the filter is optional, so asking
// for no particular language must not exclude anybody.
func (b Broker) SpeaksAny(codes []string) bool {
	if len(codes) == 0 {
		return true
	}
	for _, want := range codes {
		for _, has := range b.Languages {
			if strings.EqualFold(has, want) {
				return true
			}
		}
	}
	return false
}

// Agency groups brokers under one brand.
type Agency struct {
	ID           string
	Slug         string
	Name         string
	Logo         string
	Cover        string
	CountryCode  string
	City         string
	Address      string
	Phone        string
	Email        string
	Website      string
	Description  string
	BrokerCount  int
	ListingCount int
	Founded      int
	IsVerified   bool
	// Office is the pin for Address above. A broker joining this agency starts
	// with it as their own map location, which is what makes the directory's
	// radius search work before anybody has touched their profile.
	Office MapPlace
}

// User is the signed-in account. Mocked for this milestone.
type User struct {
	ID    string
	Name  string
	Email string
	Phone string
	// Avatar is the profile picture. Optional for everyone — a private seller
	// is under no obligation to show a face.
	Avatar string
	Role   string // "user", "broker", "admin"
	// CountryCode is the account's home market, still a single value: it is
	// what the interface defaults to. Where the seller *works* is
	// ActiveCountries below, which is a different question and a different
	// answer for a broker near a border.
	CountryCode string
	Language    string
	MemberSince time.Time
	IsVerified  bool

	// Company is the seller's company or franchise — "Best House Ltd". Optional,
	// which is the point: a private seller leaves it empty and nothing about
	// their listings changes. Filled in, it is shown beside their name wherever
	// the name appears in the public interface.
	Company string
	// CompanyLogo is the company's own mark, shown at the foot of the seller
	// box on a listing. Separate from Avatar because a broker has both, and
	// they are not interchangeable: one is a person, the other a brand.
	CompanyLogo string

	// DirectFromOwner says this seller owns the property they are selling, and
	// puts a label to that effect on their listings. Set beside the name in
	// account settings; see Property.DirectFromOwner for why it is a claim
	// rather than something derived from SellerKind.
	DirectFromOwner bool

	// Bio is the seller's own description of themselves — the paragraph a
	// broker profile shows under the heading "About <name>".
	//
	// "Under broker name is 'About' but at the moment under user profile there
	// is no place the user can edit this text — create it." It is the same fact
	// Broker.Bio holds; this is where its owner writes it.
	Bio string

	// Languages are the languages this seller will deal in, as language codes.
	// They are offered as a filter on the search screen, so a buyer can look
	// for property they can actually negotiate for.
	Languages []string

	// ActiveCountries are the markets the seller works in, as ISO 3166-1
	// alpha-2 codes. More than one is normal for a broker working either side
	// of a border, which is why this is a list and CountryCode is not.
	ActiveCountries []string

	// Messengers are the chat apps the seller can be reached on, in the same
	// shape a listing carries. WhatsApp and Viber need no handle — both are
	// addressed by the phone number above — while Telegram and Signal each
	// carry their own link, and Teams an email address.
	Messengers []Messenger

	// Office is where this seller put their pin on the map: "under seller's
	// profile add more the googlemaps location place, where user can specify
	// his location on the googlemaps — so in the brokers section the users can
	// search the brokers on the googlemaps."
	//
	// It is the same MapPlace a Broker carries, because it is the same fact and
	// the directory's radius search reads it. Optional: a private seller who
	// never sets one is simply not found by a radius search, which is the
	// honest outcome — nobody knows where they are.
	Office MapPlace

	// Ad is this seller's paid placement in the homepage broker strip, bought
	// from their own profile. It runs against ActiveCountries above rather than
	// against Office: the strip is chosen by the market a visitor picked on the
	// homepage banner, which is a country, while the pin answers a radius
	// search. Two locations, two jobs — see BrokerAd.
	Ad BrokerAd

	// MapAd is the placement that runs against Office rather than against the
	// markets: this seller's pin on the search map, and their card among the
	// search results, for as long as it is paid for. See BrokerMapAd.
	MapAd BrokerMapAd
}

// IsAdmin reports whether the account may open the administration panel.
//
// The client's question on 18 August — "so this admin panel is the website
// backend? All user's do not have access there" — and its answer. The panel is
// the backend, and this is the single place that decides who reaches it: the
// account menu asks before drawing the link, and every /admin route asks before
// rendering anything (see requireAdmin in package handlers). A visitor whose
// account is not an administrator's is told the page does not exist, which is
// the correct answer to give a stranger about a back office.
func (u User) IsAdmin() bool { return u.Role == "admin" }

// Messenger returns the entry for one app and whether the seller enabled it,
// so a template can both tick the right boxes and refill the handle fields.
func (u User) Messenger(kind MessengerKind) (Messenger, bool) {
	for _, m := range u.Messengers {
		if m.Kind == kind {
			return m, true
		}
	}
	return Messenger{Kind: kind}, false
}

// ---------------------------------------------------------------------------
// Content
// ---------------------------------------------------------------------------

// Article is an editorial piece in the advice section.
type Article struct {
	ID          string
	Slug        string
	Title       string
	Excerpt     string
	Body        []string // paragraphs; a real CMS would supply sanitised HTML
	Cover       Image
	Category    string
	Tags        []string
	Author      string
	AuthorRole  string
	AuthorPhoto string
	ReadMinutes int
	PublishedAt time.Time
	IsFeatured  bool
}

// Development is a new-build project. The client's confirmed term is
// "Developments", not "Projects".
type Development struct {
	ID                string
	Slug              string
	Name              string
	Developer         string
	CountryCode       string
	Country           string
	City              string
	District          string
	Address           string
	Coords            Coordinates
	Description       string
	Cover             Image
	Images            []Image
	PriceFrom         Money
	AreaFrom          float64
	AreaTo            float64
	TotalUnits        int
	AvailableUnits    int
	Floors            int
	CompletionQuarter string // e.g. "Q3 2027"
	IsCompleted       bool
	EnergyRating      string
	Features          []Feature
	PropertyIDs       []string
}

// ---------------------------------------------------------------------------
// Commerce
// ---------------------------------------------------------------------------

// Package is a paid listing tier.
// ---------------------------------------------------------------------------
// Paid promotion
// ---------------------------------------------------------------------------

// PromotionKind identifies a paid add-on that runs for a chosen number of days.
type PromotionKind string

const (
	PromotionFeatured PromotionKind = "featured"
	PromotionBump     PromotionKind = "bump"
)

// PromotionTier is one "N days for X" option.
//
// Longer runs are cheaper per day, which is why the price is stored per tier
// rather than derived by multiplying a daily rate.
type PromotionTier struct {
	Days  int
	Price Money
}

// PerDay is the tier's effective daily rate, shown beside the longer options so
// the saving is visible rather than implied.
func (t PromotionTier) PerDay() Money {
	if t.Days <= 0 {
		return t.Price
	}
	return Money{Amount: t.Price.Amount / float64(t.Days), Currency: t.Price.Currency}
}

// Promotion is a paid add-on offered when a listing is published and, again,
// from the seller's own listing management afterwards.
//
// The client's rule: a promotion is bought for a number of days, independently
// of how long the listing itself runs, and can be topped up at any time while
// the listing is still online.
type Promotion struct {
	Kind PromotionKind
	// Name is the checkbox label.
	Name string
	// Info is the explanatory line revealed once the box is ticked.
	Info  string
	Tiers []PromotionTier
}

// DefaultTier is the option preselected when the promotion is switched on —
// the cheapest one, so ticking the box never quietly picks an expensive run.
func (p Promotion) DefaultTier() PromotionTier {
	if len(p.Tiers) == 0 {
		return PromotionTier{}
	}
	return p.Tiers[0]
}

// BrokerAdRate is one market's daily price for the homepage strip.
//
// The client set the shape of it on 19 August: "in the backend there is option
// to set the price per day for each country. For example in Germany 3 € per
// day, in Poland 1 € per day." So a market is a price, not a tier — how long
// the broker runs for is their own number, and the bill is the two multiplied.
type BrokerAdRate struct {
	// Country is an ISO 3166-1 alpha-2 code.
	Country string
	PerDay  Money
}

// BrokerAdPlan is the homepage broker strip sold as a product: a daily price
// per market, and a run length the broker chooses.
//
// Per market rather than per broker, because the strip is a different audience
// in each country — a visitor who picked Germany on the homepage banner never
// sees the Estonian strip — so a broker advertising in two markets is buying
// two placements, and each is priced on what that market is worth.
//
// The client's own worked example is the specification:
//
//	"So the broker choosess the Germany and Poland for 10 days. So the system
//	 will calculate a bill: Germany 3 € per day x 10 = 30 €; for Estonia 1 €
//	 per day x 10 = 10 €; total: 40 €."
//
// Which is why this carries rates rather than the PromotionTier ladder every
// other paid add-on on the site uses: a tier prices a run length, and here the
// run length is the same in every market while the price is not.
type BrokerAdPlan struct {
	Name string
	Info string
	// DefaultPerDay is charged in any market with no rate of its own. Previa
	// sells in every country in the world list and the administrator has not
	// priced them all, so there has to be a price for the rest.
	DefaultPerDay Money
	// Rates are the per-market prices set in the admin panel, one per market
	// that has been given one.
	Rates []BrokerAdRate
	// DayOptions are the run lengths offered as one-click choices. The field is
	// free — a broker can type any number of days — so these are shortcuts, not
	// a price list.
	DayOptions []int
}

// PerDay is what a day in this market costs.
func (p BrokerAdPlan) PerDay(code string) Money {
	for _, r := range p.Rates {
		if strings.EqualFold(r.Country, code) {
			return r.PerDay
		}
	}
	return p.DefaultPerDay
}

// CheapestPerDay is the lowest daily rate on offer, which is what the homepage
// strip quotes to the brokers reading it: "from €1 per day".
func (p BrokerAdPlan) CheapestPerDay() Money {
	low := p.DefaultPerDay
	for _, r := range p.Rates {
		if r.PerDay.Amount < low.Amount {
			low = r.PerDay
		}
	}
	return low
}

// DefaultDays is the run length the form opens on — the first shortcut, or ten
// days, which is the length the client's worked example uses.
func (p BrokerAdPlan) DefaultDays() int {
	if len(p.DayOptions) > 0 {
		return p.DayOptions[0]
	}
	return 10
}

// BrokerMapAdPlan is the map placement sold as a product.
//
// Bought once rather than per market, because what it buys is a single pin on
// a single map — "he can activate that his broker profile is displayed in the
// googlemaps like the ads" — so the multiplier the market strip carries has
// nothing to multiply here. Same tier ladder as every other paid add-on on the
// site: a number of days, cheaper per day the longer it runs.
type BrokerMapAdPlan struct {
	Name  string
	Info  string
	Tiers []PromotionTier
}

// DefaultTier is the cheapest run, preselected for the same reason a
// promotion's is.
func (p BrokerMapAdPlan) DefaultTier() PromotionTier {
	if len(p.Tiers) == 0 {
		return PromotionTier{}
	}
	return p.Tiers[0]
}

type Package struct {
	ID           string
	Name         string
	Tagline      string
	Price        Money
	DurationDays int
	Features     []string
	IsPopular    bool
	IsPremium    bool // renders with the gold accent
	PhotoLimit   int
	BumpCount    int
	// IsEnabled controls whether the package is offered at checkout. A
	// disabled tier stays in the admin table so its price and features are
	// not lost — switching it back on is one toggle.
	IsEnabled bool
}

// PaymentStatus is the outcome of a mock transaction.
type PaymentStatus string

const (
	PaymentPaid      PaymentStatus = "paid"
	PaymentPending   PaymentStatus = "pending"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
	PaymentCancelled PaymentStatus = "cancelled"
)

// Payment is one billing history row.
type Payment struct {
	ID            string
	InvoiceNo     string
	UserID        string
	PackageName   string
	PropertyTitle string
	Amount        Money
	Method        string // "stripe", "paypal", "paysera"
	Status        PaymentStatus
	CreatedAt     time.Time
}

// ---------------------------------------------------------------------------
// Account activity
// ---------------------------------------------------------------------------

// SavedSearch stores a filter set the user asked to be notified about.
type SavedSearch struct {
	ID          string
	Name        string
	Query       string // encoded query string, replayed into the search page
	Summary     string // human-readable filter description
	Deal        DealType
	AlertsOn    bool
	Frequency   string // "instant", "daily", "weekly"
	NewMatches  int
	ResultCount int
	CreatedAt   time.Time
}

// Notification is one row in the notification centre.
type Notification struct {
	ID        string
	Type      string // "match", "message", "system", "payment", "listing"
	Title     string
	Body      string
	Link      string
	Icon      string
	IsRead    bool
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------
// Listing statistics
// ---------------------------------------------------------------------------

// DayCount is one day's tally in a listing's statistics.
//
// Label is the axis caption ("14 Aug"); Date is the same day in full, for the
// tooltip and for anything that needs to sort or group. Keeping both means the
// chart never has to re-parse its own labels.
//
// The json tags matter: this struct is serialised straight into the statistics
// dialog's Alpine scope, and short lowercase keys keep a fourteen-day series
// for five listings down to a few hundred bytes of markup.
type DayCount struct {
	Date  time.Time `json:"-"`
	Label string    `json:"label"`
	Views int       `json:"views"`
}

// ListingStats is the per-listing performance panel the client asked for:
// "there somewhere has to be the statistics, what shows how many visitors have
// seen this ad per day, so statistic window."
//
// Deliberately a daily series rather than a running total. A total says a
// listing has been seen; a series says whether it still is, which is the thing
// a seller deciding whether to re-activate or promote actually needs to know.
type ListingStats struct {
	PropertyID string `json:"id"`
	Title      string `json:"title"`
	// Days runs oldest to newest, so the chart reads left to right.
	Days []DayCount `json:"days"`
	// Views and Saves are the lifetime totals, which is what the listings
	// table already shows — carried here so the panel and the table cannot
	// disagree.
	Views int `json:"views"`
	Saves int `json:"saves"`
}

// Total is the number of views across the charted period, which is not the
// same as lifetime Views and is labelled separately for that reason.
func (s ListingStats) Total() int {
	n := 0
	for _, d := range s.Days {
		n += d.Views
	}
	return n
}

// Peak is the busiest day's count, used to scale the bars. Never zero, so a
// listing with no views yet renders a flat axis rather than dividing by zero.
func (s ListingStats) Peak() int {
	max := 0
	for _, d := range s.Days {
		if d.Views > max {
			max = d.Views
		}
	}
	if max == 0 {
		return 1
	}
	return max
}

// PerDay is the mean over the charted period, rounded to one decimal.
func (s ListingStats) PerDay() float64 {
	if len(s.Days) == 0 {
		return 0
	}
	return float64(s.Total()) / float64(len(s.Days))
}

// Draft is an unfinished add-listing wizard session.
type Draft struct {
	ID         string
	Title      string
	Deal       DealType
	Type       PropertyType
	City       string
	Step       int
	TotalSteps int
	Completion int // percent
	CoverImage string
	UpdatedAt  time.Time
}

// ---------------------------------------------------------------------------
// Marketing
// ---------------------------------------------------------------------------

// Banner is the country-specific advertising strip above the homepage hero.
type Banner struct {
	ID          string
	CountryCode string
	Headline    string
	Body        string
	CTALabel    string
	CTAHref     string
	Sponsor     string
	Image       string
	Theme       string // "navy" | "gold" | "slate"

	// Placement separates the two advertising slots, which carry different
	// copy and different dimensions:
	//
	//   "home"   the strip laid over the homepage hero
	//   "search" the shorter strip above the search results
	//
	// A country can have one of each. An empty value means "home", so existing
	// rows keep working.
	Placement string

	// Active switches the slot off without deleting it. The client asked to
	// hide the homepage strip for now but keep it editable and re-activatable,
	// so dismissal is a flag rather than a removal.
	Active bool
}

// Slot returns the placement, defaulting to the homepage strip.
func (b Banner) Slot() string {
	if b.Placement == "" {
		return "home"
	}
	return b.Placement
}

// Testimonial supports the homepage trust section.
type Testimonial struct {
	Name  string
	Role  string
	City  string
	Photo string
	Quote string
	Stars int
}

// ---------------------------------------------------------------------------
// Administration
// ---------------------------------------------------------------------------

// AdminStat is one KPI tile on the admin dashboard.
type AdminStat struct {
	Label string
	Value string
	Delta string
	Trend string // "up" | "down" | "flat"
	Icon  string
	Hint  string
}

// AdminStats aggregates dashboard data.
type AdminStats struct {
	Tiles          []AdminStat
	ListingsByType []ChartSlice
	SignupsByMonth []ChartPoint
	RevenueByMonth []ChartPoint
	RecentActivity []ActivityEntry
}

// ChartSlice is one category in a proportion chart.
type ChartSlice struct {
	Label string
	Value int
	Color string
}

// ChartPoint is one value in a time series.
type ChartPoint struct {
	Label string
	Value float64
}

// ActivityEntry is an admin audit-log row.
type ActivityEntry struct {
	Actor  string
	Action string
	Target string
	At     time.Time
	Kind   string // "create" | "approve" | "reject" | "delete" | "login"
}

// Language is a translation target in Settings → Languages.
type Language struct {
	Code       string
	Name       string
	NativeName string
	Flag       string
	IsDefault  bool
	IsEnabled  bool
	TotalKeys  int
	Translated int
	UpdatedAt  time.Time
}

// Progress returns the completion percentage for the language table.
func (l Language) Progress() int {
	if l.TotalKeys == 0 {
		return 0
	}
	return l.Translated * 100 / l.TotalKeys
}

// Missing returns how many strings still need a translation.
func (l Language) Missing() int { return l.TotalKeys - l.Translated }

// TranslationString is one row in Settings → Options → Strings.
type TranslationString struct {
	Key       string
	Group     string
	English   string
	Value     string // translation in the selected language, empty when missing
	UpdatedAt time.Time
}

// IsMissing reports whether the selected language falls back to English.
func (t TranslationString) IsMissing() bool { return t.Value == "" }

// SEOEntry is per-page, per-language SEO metadata.
type SEOEntry struct {
	Path        string
	Language    string
	Title       string
	Description string
	OGImage     string
	NoIndex     bool
	UpdatedAt   time.Time
}

// RestrictedCountry blocks map and listing coverage for a market. Nothing is
// hardcoded — every entry is admin-managed.
type RestrictedCountry struct {
	Code    string
	Name    string
	Reason  string
	AddedBy string
	AddedAt time.Time
}

// BackupKind separates site backups from database backups.
type BackupKind string

const (
	BackupSite  BackupKind = "site"
	BackupMySQL BackupKind = "mysql"
)

// Backup is one row of mock backup history.
type Backup struct {
	ID          string
	Name        string
	Kind        BackupKind
	Size        string
	Destination string // "local" | "gdrive"
	Status      string // "complete" | "running" | "failed"
	CreatedAt   time.Time
}

// FileEntry is one row in the admin file-manager mockup.
type FileEntry struct {
	Name     string
	Path     string
	IsDir    bool
	Size     string
	Modified time.Time
	Perms    string
	Owner    string
}

// DBTable is one row in the admin MySQL-manager mockup.
type DBTable struct {
	Name      string
	Engine    string
	Rows      int
	Size      string
	Collation string
	UpdatedAt time.Time
}

// SystemInfo backs the restart/cache panel.
type SystemInfo struct {
	BinaryBuiltAt time.Time
	Version       string
	GoVersion     string
	Uptime        string
	Environment   string
	CacheEntries  int
	CacheSize     string
	MemoryUsage   string
}
