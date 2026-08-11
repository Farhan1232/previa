// Package models holds the Previa domain types.
//
// These structs are the contract between the data layer and the templates.
// The current milestone serves them from an in-memory mock provider; a future
// MySQL-backed provider must return these same shapes so no template changes.
package models

import "time"

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

// ---------------------------------------------------------------------------
// Enumerations
// ---------------------------------------------------------------------------

// DealType separates the two primary search modes.
type DealType string

const (
	DealSale DealType = "sale"
	DealRent DealType = "rent"
)

// PropertyType is the physical category of a listing.
type PropertyType string

const (
	TypeApartment  PropertyType = "apartment"
	TypeHouse      PropertyType = "house"
	TypeVilla      PropertyType = "villa"
	TypeCommercial PropertyType = "commercial"
	TypeLand       PropertyType = "land"
	TypeGarage     PropertyType = "garage"
)

// ListingStatus is the moderation/lifecycle state of a listing.
type ListingStatus string

const (
	StatusActive   ListingStatus = "active"
	StatusPending  ListingStatus = "pending"
	StatusDraft    ListingStatus = "draft"
	StatusExpired  ListingStatus = "expired"
	StatusRejected ListingStatus = "rejected"
	StatusSold     ListingStatus = "sold"
)

// SellerKind distinguishes professional listings from private ones.
type SellerKind string

const (
	SellerBroker  SellerKind = "broker"
	SellerPrivate SellerKind = "private"
)

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
	ID        string
	Slug      string
	Title     string
	Deal      DealType
	Type      PropertyType
	Status    ListingStatus
	Price     Money
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
	Rooms      int
	Bedrooms   int
	Bathrooms  int
	Area       float64 // m² of interior space
	LandArea   float64 // m² of plot, 0 when not applicable
	Floor      int
	TotalFloors int
	BuildYear  int

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

	// Presentation
	Description string
	Highlights  []string
	Features    []Feature
	Images      []Image

	// Relations and flags
	SellerKind    SellerKind
	BrokerID      string
	AgencyID      string
	DevelopmentID string
	IsFeatured    bool
	IsNewDevelopment bool
	IsVerified    bool

	// Metrics
	Views     int
	Saves     int
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
}

// PrimaryImage returns the first image, or a placeholder when a listing has no
// photos yet (drafts often do not).
func (p Property) PrimaryImage() Image {
	if len(p.Images) > 0 {
		return p.Images[0]
	}
	return Image{URL: "/static/img/placeholder-property.svg", Alt: p.Title, Width: 800, Height: 600}
}

// IsRental reports whether the listing is a rental, used for price suffixes.
func (p Property) IsRental() bool { return p.Deal == DealRent }

// ---------------------------------------------------------------------------
// People and organisations
// ---------------------------------------------------------------------------

// Broker is an individual agent. "Broker" is the client's confirmed term; do
// not rename to "agent" in the UI.
type Broker struct {
	ID          string
	Slug        string
	Name        string
	Title       string
	AgencyID    string
	AgencyName  string
	Photo       string
	Phone       string
	Email       string
	CountryCode string
	City        string
	Languages   []string
	Specialties []string
	Bio         string
	Rating      float64
	Reviews     int
	ActiveListings int
	SoldCount   int
	YearsActive int
	IsPromoted  bool // shown in the country-specific promoted strip
	IsVerified  bool
}

// Agency groups brokers under one brand.
type Agency struct {
	ID          string
	Slug        string
	Name        string
	Logo        string
	Cover       string
	CountryCode string
	City        string
	Address     string
	Phone       string
	Email       string
	Website     string
	Description string
	BrokerCount int
	ListingCount int
	Founded     int
	IsVerified  bool
}

// User is the signed-in account. Mocked for this milestone.
type User struct {
	ID          string
	Name        string
	Email       string
	Phone       string
	Avatar      string
	Role        string // "user", "broker", "admin"
	CountryCode string
	Language    string
	MemberSince time.Time
	IsVerified  bool
}

// ---------------------------------------------------------------------------
// Content
// ---------------------------------------------------------------------------

// Article is an editorial piece in the advice section.
type Article struct {
	ID        string
	Slug      string
	Title     string
	Excerpt   string
	Body      []string // paragraphs; a real CMS would supply sanitised HTML
	Cover     Image
	Category  string
	Tags      []string
	Author    string
	AuthorRole string
	AuthorPhoto string
	ReadMinutes int
	PublishedAt time.Time
	IsFeatured  bool
}

// Development is a new-build project. The client's confirmed term is
// "Developments", not "Projects".
type Development struct {
	ID          string
	Slug        string
	Name        string
	Developer   string
	CountryCode string
	Country     string
	City        string
	District    string
	Address     string
	Coords      Coordinates
	Description string
	Cover       Image
	Images      []Image
	PriceFrom   Money
	AreaFrom    float64
	AreaTo      float64
	TotalUnits  int
	AvailableUnits int
	Floors      int
	CompletionQuarter string // e.g. "Q3 2027"
	IsCompleted bool
	EnergyRating string
	Features    []Feature
	PropertyIDs []string
}

// ---------------------------------------------------------------------------
// Commerce
// ---------------------------------------------------------------------------

// Package is a paid listing tier.
type Package struct {
	ID          string
	Name        string
	Tagline     string
	Price       Money
	DurationDays int
	Features    []string
	IsPopular   bool
	IsPremium   bool // renders with the gold accent
	PhotoLimit  int
	BumpCount   int
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
	ID          string
	InvoiceNo   string
	UserID      string
	PackageName string
	PropertyTitle string
	Amount      Money
	Method      string // "stripe", "paypal", "paysera"
	Status      PaymentStatus
	CreatedAt   time.Time
}

// ---------------------------------------------------------------------------
// Account activity
// ---------------------------------------------------------------------------

// SavedSearch stores a filter set the user asked to be notified about.
type SavedSearch struct {
	ID        string
	Name      string
	Query     string // encoded query string, replayed into the search page
	Summary   string // human-readable filter description
	Deal      DealType
	AlertsOn  bool
	Frequency string // "instant", "daily", "weekly"
	NewMatches int
	ResultCount int
	CreatedAt time.Time
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

// Draft is an unfinished add-listing wizard session.
type Draft struct {
	ID          string
	Title       string
	Deal        DealType
	Type        PropertyType
	City        string
	Step        int
	TotalSteps  int
	Completion  int // percent
	CoverImage  string
	UpdatedAt   time.Time
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
	Label  string
	Value  string
	Delta  string
	Trend  string // "up" | "down" | "flat"
	Icon   string
	Hint   string
}

// AdminStats aggregates dashboard data.
type AdminStats struct {
	Tiles          []AdminStat
	PendingCount   int
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
	Actor     string
	Action    string
	Target    string
	At        time.Time
	Kind      string // "create" | "approve" | "reject" | "delete" | "login"
}

// Language is a translation target in Settings → Languages.
type Language struct {
	Code        string
	Name        string
	NativeName  string
	Flag        string
	IsDefault   bool
	IsEnabled   bool
	TotalKeys   int
	Translated  int
	UpdatedAt   time.Time
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
	Code     string
	Name     string
	Reason   string
	AddedBy  string
	AddedAt  time.Time
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
