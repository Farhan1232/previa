package handlers

import (
	"net/http"
	"strings"
)

// Routes builds the application mux.
//
// Go 1.22's pattern router is used directly — no third-party router is needed
// for this route table.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Static assets. Long-lived caching is safe because filenames are stable
	// for the milestone; the backend can add fingerprinting later.
	fs := http.FileServer(http.Dir(h.Cfg.StaticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", cacheStatic(fs)))

	// --- Public -------------------------------------------------------------
	mux.HandleFunc("GET /{$}", h.Home)
	mux.HandleFunc("GET /search", h.Search)
	mux.HandleFunc("GET /search/results", h.SearchResults) // HTMX partial
	mux.HandleFunc("GET /property/{slug}", h.PropertyDetail)
	mux.HandleFunc("GET /developments", h.Developments)
	mux.HandleFunc("GET /development/{slug}", h.DevelopmentDetail)
	mux.HandleFunc("GET /brokers", h.Brokers)
	mux.HandleFunc("GET /broker/{slug}", h.BrokerProfile)
	mux.HandleFunc("GET /agencies", h.Agencies)
	mux.HandleFunc("GET /agency/{slug}", h.AgencyProfile)
	mux.HandleFunc("GET /articles", h.Articles)
	mux.HandleFunc("GET /article/{slug}", h.ArticleDetail)
	mux.HandleFunc("GET /pricing", h.Pricing)
	mux.HandleFunc("GET /help", h.Help)
	mux.HandleFunc("GET /about", h.About)
	mux.HandleFunc("GET /advertising", h.Advertising)
	mux.HandleFunc("GET /terms", h.Legal)
	mux.HandleFunc("GET /privacy", h.Legal)
	mux.HandleFunc("GET /cookies", h.Cookies)

	// --- Auth (mocked) ------------------------------------------------------
	mux.HandleFunc("GET /login", h.Login)
	mux.HandleFunc("POST /login", h.LoginSubmit)
	mux.HandleFunc("GET /register", h.Register)
	mux.HandleFunc("POST /register", h.RegisterSubmit)
	mux.HandleFunc("GET /forgot-password", h.ForgotPassword)
	mux.HandleFunc("POST /forgot-password", h.ForgotPasswordSubmit)
	mux.HandleFunc("GET /logout", h.Logout)

	// --- Add listing --------------------------------------------------------
	mux.HandleFunc("GET /add-listing", h.AddListing)
	mux.HandleFunc("POST /add-listing/save", h.AddListingSave)

	// --- Account ------------------------------------------------------------
	mux.HandleFunc("GET /dashboard", h.Dashboard)
	mux.HandleFunc("GET /my-listings", h.MyListings)
	mux.HandleFunc("GET /drafts", h.Drafts)
	mux.HandleFunc("GET /favourites", h.Favourites)
	mux.HandleFunc("GET /saved-searches", h.SavedSearches)
	mux.HandleFunc("GET /notifications", h.Notifications)
	mux.HandleFunc("GET /settings", h.Settings)
	mux.HandleFunc("GET /billing", h.Billing)
	mux.HandleFunc("GET /checkout", h.Checkout)
	mux.HandleFunc("POST /checkout/process", h.CheckoutProcess)

	// --- Mock interactions --------------------------------------------------
	mux.HandleFunc("POST /favourite/{id}", h.ToggleFavourite)
	mux.HandleFunc("GET /set-country", h.SetCountry)
	mux.HandleFunc("POST /contact-broker", h.ContactBroker)
	mux.HandleFunc("POST /save-search", h.SaveSearch)
	mux.HandleFunc("POST /reveal-contact", h.RevealContact)
	// Mock reverse geocoding for the add-listing map. Reads seeded data and
	// writes nothing; a Geocoding API call replaces the handler body.
	mux.HandleFunc("GET /mock/reverse-geocode", h.ReverseGeocode)

	// --- Admin --------------------------------------------------------------
	mux.HandleFunc("GET /admin", h.AdminDashboard)
	mux.HandleFunc("GET /admin/listings", h.AdminListings)
	mux.HandleFunc("GET /admin/users", h.AdminUsers)
	mux.HandleFunc("GET /admin/brokers", h.AdminBrokers)
	mux.HandleFunc("GET /admin/agencies", h.AdminAgencies)
	mux.HandleFunc("GET /admin/developments", h.AdminDevelopments)
	mux.HandleFunc("GET /admin/articles", h.AdminArticles)
	mux.HandleFunc("GET /admin/banners", h.AdminBanners)
	mux.HandleFunc("GET /admin/packages", h.AdminPackages)
	mux.HandleFunc("GET /admin/payments", h.AdminPayments)
	mux.HandleFunc("GET /admin/languages", h.AdminLanguages)
	mux.HandleFunc("GET /admin/strings", h.AdminStrings)
	mux.HandleFunc("GET /admin/seo", h.AdminSEO)
	mux.HandleFunc("GET /admin/maps", h.AdminMaps)
	mux.HandleFunc("GET /admin/restricted", h.AdminRestricted)
	mux.HandleFunc("GET /admin/settings", h.AdminSettings)
	mux.HandleFunc("GET /admin/backups", h.AdminBackups)
	mux.HandleFunc("GET /admin/files", h.AdminFiles)
	mux.HandleFunc("GET /admin/database", h.AdminDatabase)
	mux.HandleFunc("GET /admin/system", h.AdminSystem)
	mux.HandleFunc("POST /admin/mock-action", h.AdminMockAction)

	// --- SEO ----------------------------------------------------------------
	mux.HandleFunc("GET /robots.txt", h.Robots)
	mux.HandleFunc("GET /sitemap.xml", h.Sitemap)

	// Anything unmatched falls through to the 404 page.
	mux.HandleFunc("/", h.Home)

	return securityHeaders(mux)
}

// cacheStatic adds long-lived caching to static assets.
func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".woff2"):
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case strings.HasSuffix(r.URL.Path, ".jpg"), strings.HasSuffix(r.URL.Path, ".svg"),
			strings.HasSuffix(r.URL.Path, ".png"):
			w.Header().Set("Cache-Control", "public, max-age=604800")
		default:
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		next.ServeHTTP(w, r)
	})
}

// securityHeaders applies baseline response headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		next.ServeHTTP(w, r)
	})
}
