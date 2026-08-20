package handlers

import (
	"net/http"

	"previa/pkg/models"
)

// AdminData is the payload shared by the administration screens.
//
// Every mutating control in the admin panel is a demonstration. Nothing in this
// package restarts a process, clears a cache, touches the filesystem, connects
// to a database or deletes anything.
type AdminData struct {
	Section    string
	Stats      models.AdminStats
	Listings   []models.Property
	Users      []models.User
	Brokers    []models.Broker
	Agencies   []models.Agency
	Devs       []models.Development
	Articles   []models.Article
	Packages   []models.Package
	Payments   []models.Payment
	Languages  []models.Language
	Strings    []models.TranslationString
	SEO        []models.SEOEntry
	Restricted []models.RestrictedCountry
	Backups    []models.Backup
	Files      []models.FileEntry
	Tables     []models.DBTable
	System     models.SystemInfo
	Countries  []models.Country
	Banners    []models.Banner
	// The two paid broker placements as products: the per-market daily price
	// list the homepage strip bills from — "in the backend there is option to
	// set the price per day for each country" — and the map placement's tiers.
	BrokerAdPlan    models.BrokerAdPlan
	BrokerMapAdPlan models.BrokerMapAdPlan
	// AllCountries is every market the price list can cover, not only the
	// eight that carry stock: a broker can advertise anywhere Previa sells.
	AllCountries []models.Country
	Status       string
	Lang         string
	MaxSignups   float64
	MaxRevenue   float64
	TypeTotal    int
}

func (h *Handler) adminBase(r *http.Request, section, title string) (PageData, AdminData) {
	pd := h.base(r, "", title+" — Previa admin", "Previa administration.")
	pd.Meta.NoIndex = true
	return pd, AdminData{Section: section, Countries: h.Store.Catalog.Countries(r.Context())}
}

// AdminDashboard renders the admin overview.
func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd, ad := h.adminBase(r, "dashboard", "Dashboard")

	ad.Stats = h.Store.Admin.Stats(ctx)
	for _, p := range ad.Stats.SignupsByMonth {
		if p.Value > ad.MaxSignups {
			ad.MaxSignups = p.Value
		}
	}
	for _, p := range ad.Stats.RevenueByMonth {
		if p.Value > ad.MaxRevenue {
			ad.MaxRevenue = p.Value
		}
	}
	for _, s := range ad.Stats.ListingsByType {
		ad.TypeTotal += s.Value
	}

	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/dashboard", pd)
}

// AdminListings renders listing moderation, filtered by status tab.
func (h *Handler) AdminListings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd, ad := h.adminBase(r, "listings", "Property listings")

	status := r.URL.Query().Get("status")
	if status == "" {
		res, _ := h.Store.Properties.Search(ctx, parseAll())
		ad.Listings = res.Items
	} else {
		ad.Listings = h.Store.Properties.ByStatus(ctx, models.ListingStatus(status))
		// The seed set is mostly active; show a representative slice for the
		// other tabs so moderation screens are never empty.
		if len(ad.Listings) == 0 {
			res, _ := h.Store.Properties.Search(ctx, parseAll())
			ad.Listings = take(res.Items, 6)
			for i := range ad.Listings {
				ad.Listings[i].Status = models.ListingStatus(status)
			}
		}
	}
	ad.Status = status
	ad.Stats = h.Store.Admin.Stats(ctx)

	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/listings", pd)
}

// AdminUsers renders the user table.
func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "users", "Users")
	ad.Users = h.Store.Admin.Users(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/users", pd)
}

// AdminBrokers renders broker administration.
func (h *Handler) AdminBrokers(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "brokers", "Brokers")
	ad.Brokers = h.Store.Brokers.All(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/brokers", pd)
}

// AdminAgencies renders agency administration.
func (h *Handler) AdminAgencies(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "agencies", "Agencies")
	ad.Agencies = h.Store.Brokers.Agencies(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/agencies", pd)
}

// AdminDevelopments renders development administration.
func (h *Handler) AdminDevelopments(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "developments", "Developments")
	ad.Devs = h.Store.Content.Developments(r.Context(), "")
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/developments", pd)
}

// AdminArticles renders article administration.
func (h *Handler) AdminArticles(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "articles", "Articles")
	ad.Articles = h.Store.Content.Articles(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/articles", pd)
}

// AdminBanners renders advertising banners and promoted brokers per country.
func (h *Handler) AdminBanners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd, ad := h.adminBase(r, "banners", "Advertising banners")

	for _, c := range h.Store.Catalog.Countries(ctx) {
		if b, ok := h.Store.Catalog.Banner(ctx, c.Code); ok {
			ad.Banners = append(ad.Banners, b)
		}
	}
	ad.Brokers = h.Store.Brokers.Promoted(ctx, "", 20)

	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/banners", pd)
}

// AdminPackages renders package administration.
func (h *Handler) AdminPackages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd, ad := h.adminBase(r, "packages", "Price packages")
	ad.Packages = h.Store.Catalog.Packages(ctx)
	ad.BrokerAdPlan = h.Store.Catalog.BrokerAdPlan(ctx)
	ad.BrokerMapAdPlan = h.Store.Catalog.BrokerMapAdPlan(ctx)
	ad.AllCountries = h.Store.Catalog.AllCountries(ctx)
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/packages", pd)
}

// AdminPayments renders the payment ledger.
func (h *Handler) AdminPayments(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "payments", "Payments")
	ad.Payments = h.Store.Account.Payments(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/payments", pd)
}

// AdminLanguages renders Settings → Languages.
func (h *Handler) AdminLanguages(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "languages", "Languages")
	ad.Languages = h.Store.Catalog.Languages(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/languages", pd)
}

// AdminStrings renders Settings → Options → Strings.
func (h *Handler) AdminStrings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd, ad := h.adminBase(r, "strings", "Translation strings")

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "de"
	}
	ad.Lang = lang
	ad.Languages = h.Store.Catalog.Languages(ctx)
	ad.Strings = h.Store.Admin.Translations(ctx, lang)

	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/strings", pd)
}

// AdminSEO renders Settings → SEO.
func (h *Handler) AdminSEO(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd, ad := h.adminBase(r, "seo", "SEO")
	ad.SEO = h.Store.Admin.SEOEntries(ctx)
	ad.Languages = h.Store.Catalog.Languages(ctx)
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/seo", pd)
}

// AdminMaps renders Google Maps settings.
func (h *Handler) AdminMaps(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "maps", "Google Maps settings")
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/maps", pd)
}

// AdminRestricted renders the restricted-country manager.
func (h *Handler) AdminRestricted(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "restricted", "Restricted countries")
	pd.NeedsMap = true
	ad.Restricted = h.Store.Catalog.RestrictedCountries(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/restricted", pd)
}

// AdminSettings renders general settings.
func (h *Handler) AdminSettings(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "settings", "General settings")
	ad.Countries = h.Store.Catalog.Countries(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/settings", pd)
}

// AdminBackups renders backup management. No backup is ever created or
// restored by this build.
func (h *Handler) AdminBackups(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "backups", "Backups")
	ad.Backups = h.Store.Admin.Backups(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/backups", pd)
}

// AdminFiles renders the file-manager mockup. The listing is synthetic and the
// real filesystem is never read.
func (h *Handler) AdminFiles(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "files", "File manager")
	ad.Files = h.Store.Admin.Files(r.Context(), r.URL.Query().Get("path"))
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/files", pd)
}

// AdminDatabase renders the MySQL-manager mockup. No database connection is
// opened and no query is executed.
func (h *Handler) AdminDatabase(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "database", "MySQL manager")
	ad.Tables = h.Store.Admin.Tables(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/database", pd)
}

// AdminSystem renders cache and restart controls. Both are simulations.
func (h *Handler) AdminSystem(w http.ResponseWriter, r *http.Request) {
	pd, ad := h.adminBase(r, "system", "System")
	ad.System = h.Store.Admin.SystemInfo(r.Context())
	pd.Data = ad
	h.View.Render(w, http.StatusOK, "admin/system", pd)
}

// AdminMockAction returns simulated feedback for a destructive-looking admin
// control. It performs no work whatsoever — it exists so the frontend can
// demonstrate confirmation and result states safely.
func (h *Handler) AdminMockAction(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	action := r.FormValue("action")

	messages := map[string]struct{ tone, text string }{
		"restart":      {"success", "Simulation only — the web application was not restarted."},
		"cache":        {"success", "Simulation only — 18 420 cache entries would be cleared."},
		"backup":       {"success", "Simulation only — no backup was created."},
		"restore":      {"warning", "Simulation only — no backup was restored."},
		"delete-file":  {"warning", "Simulation only — no file was deleted."},
		"optimize":     {"success", "Simulation only — no table was modified."},
		"delete-draft": {"success", "Draft removed from this demonstration view."},
		"approve":      {"success", "Listing approved in this demonstration view."},
		"reject":       {"warning", "Listing rejected in this demonstration view."},
	}

	m, ok := messages[action]
	if !ok {
		m = struct{ tone, text string }{"neutral", "Simulation only — no change was made."}
	}

	h.View.RenderPartial(w, http.StatusOK, "admin/system", "mock-result",
		map[string]any{"Tone": m.tone, "Text": m.text})
}
