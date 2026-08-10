package data

import (
	"context"
	"runtime"
	"time"

	"previa/internal/models"
)

// ---------------------------------------------------------------------------
// Languages, translations, SEO
// ---------------------------------------------------------------------------

func buildLanguages(now time.Time) []models.Language {
	return []models.Language{
		{Code: "en", Name: "English", NativeName: "English", Flag: "🇬🇧", IsDefault: true, IsEnabled: true,
			TotalKeys: 1284, Translated: 1284, UpdatedAt: now.AddDate(0, 0, -2)},
		{Code: "de", Name: "German", NativeName: "Deutsch", Flag: "🇩🇪", IsEnabled: true,
			TotalKeys: 1284, Translated: 1211, UpdatedAt: now.AddDate(0, 0, -5)},
		{Code: "es", Name: "Spanish", NativeName: "Español", Flag: "🇪🇸", IsEnabled: true,
			TotalKeys: 1284, Translated: 1168, UpdatedAt: now.AddDate(0, 0, -8)},
		{Code: "et", Name: "Estonian", NativeName: "Eesti", Flag: "🇪🇪", IsEnabled: true,
			TotalKeys: 1284, Translated: 1284, UpdatedAt: now.AddDate(0, 0, -3)},
		{Code: "fi", Name: "Finnish", NativeName: "Suomi", Flag: "🇫🇮", IsEnabled: true,
			TotalKeys: 1284, Translated: 946, UpdatedAt: now.AddDate(0, 0, -14)},
		{Code: "pt", Name: "Portuguese", NativeName: "Português", Flag: "🇵🇹", IsEnabled: true,
			TotalKeys: 1284, Translated: 903, UpdatedAt: now.AddDate(0, 0, -19)},
		{Code: "nl", Name: "Dutch", NativeName: "Nederlands", Flag: "🇳🇱", IsEnabled: false,
			TotalKeys: 1284, Translated: 412, UpdatedAt: now.AddDate(0, -1, -6)},
		{Code: "cs", Name: "Czech", NativeName: "Čeština", Flag: "🇨🇿", IsEnabled: false,
			TotalKeys: 1284, Translated: 287, UpdatedAt: now.AddDate(0, -2, -1)},
	}
}

// translationSeeds are representative rows for Settings → Options → Strings.
// A missing Value means the UI falls back to the English source string.
var translationSeeds = []struct {
	key, group, en string
	de, es         string
}{
	{"nav.articles", "Navigation", "Articles", "Artikel", "Artículos"},
	{"nav.brokers", "Navigation", "Brokers", "Makler", "Agentes"},
	{"nav.developments", "Navigation", "Developments", "Neubauprojekte", "Promociones"},
	{"nav.add_listing", "Navigation", "Add listing", "Anzeige aufgeben", "Publicar anuncio"},
	{"nav.login", "Navigation", "Log in", "Anmelden", "Iniciar sesión"},
	{"search.buy", "Search", "Buy", "Kaufen", "Comprar"},
	{"search.rent", "Search", "Rent", "Mieten", "Alquilar"},
	{"search.location", "Search", "Location", "Ort", "Ubicación"},
	{"search.property_type", "Search", "Property type", "Immobilientyp", "Tipo de propiedad"},
	{"search.price_range", "Search", "Price range", "Preisspanne", "Rango de precio"},
	{"search.bedrooms", "Search", "Bedrooms", "Schlafzimmer", ""},
	{"search.submit", "Search", "Search Properties", "Immobilien suchen", "Buscar propiedades"},
	{"search.advanced", "Search", "Advanced filters", "Erweiterte Filter", ""},
	{"search.clear", "Search", "Clear all", "Alle löschen", "Borrar todo"},
	{"search.results_count", "Search", "%d properties found", "%d Immobilien gefunden", "%d propiedades encontradas"},
	{"search.no_results", "Search", "No properties match your filters", "Keine Immobilien gefunden", ""},
	{"card.for_sale", "Property card", "For sale", "Zu verkaufen", "En venta"},
	{"card.for_rent", "Property card", "For rent", "Zu vermieten", "En alquiler"},
	{"card.featured", "Property card", "Featured", "Empfohlen", "Destacado"},
	{"card.view_property", "Property card", "View Property", "Objekt ansehen", "Ver propiedad"},
	{"card.save", "Property card", "Save to favourites", "Zu Favoriten hinzufügen", ""},
	{"detail.schedule_viewing", "Property detail", "Schedule a Viewing", "Besichtigung vereinbaren", "Programar una visita"},
	{"detail.contact_broker", "Property detail", "Contact Broker", "Makler kontaktieren", "Contactar agente"},
	{"detail.show_phone", "Property detail", "Show phone number", "Telefonnummer anzeigen", ""},
	{"detail.report", "Property detail", "Report listing", "Anzeige melden", "Denunciar anuncio"},
	{"detail.similar", "Property detail", "Similar properties", "Ähnliche Immobilien", "Propiedades similares"},
	{"wizard.step_deal", "Add listing", "Sale or rent", "Verkauf oder Vermietung", "Venta o alquiler"},
	{"wizard.step_location", "Add listing", "Location", "Lage", "Ubicación"},
	{"wizard.saving", "Add listing", "Saving…", "Wird gespeichert…", "Guardando…"},
	{"wizard.saved", "Add listing", "Saved", "Gespeichert", "Guardado"},
	{"wizard.publish", "Add listing", "Publish listing", "Anzeige veröffentlichen", ""},
	{"account.favourites", "Account", "Favourites", "Favoriten", "Favoritos"},
	{"account.saved_searches", "Account", "Saved searches", "Gespeicherte Suchen", "Búsquedas guardadas"},
	{"account.my_listings", "Account", "My listings", "Meine Anzeigen", "Mis anuncios"},
	{"account.billing", "Account", "Billing", "Abrechnung", ""},
	{"payment.processing", "Payments", "Processing your payment", "Zahlung wird verarbeitet", ""},
	{"payment.success", "Payments", "Payment successful", "Zahlung erfolgreich", "Pago realizado"},
	{"payment.failed", "Payments", "Payment failed", "Zahlung fehlgeschlagen", "Pago fallido"},
	{"error.404_title", "Errors", "We couldn't find that page", "Seite nicht gefunden", "Página no encontrada"},
	{"error.map_unavailable", "Errors", "The map is temporarily unavailable", "Karte nicht verfügbar", ""},
}

// Translations returns the string table for a language, leaving Value empty
// where a translation is missing so the UI can show the English fallback.
func (m *Mock) Translations(ctx context.Context, lang string) []models.TranslationString {
	out := make([]models.TranslationString, 0, len(translationSeeds))
	for _, s := range translationSeeds {
		v := ""
		switch lang {
		case "de":
			v = s.de
		case "es":
			v = s.es
		case "en", "":
			v = s.en
		}
		out = append(out, models.TranslationString{
			Key: s.key, Group: s.group, English: s.en, Value: v,
			UpdatedAt: m.now.AddDate(0, 0, -len(s.key)%30),
		})
	}
	return out
}

// SEOEntries returns per-page, per-language SEO metadata.
func (m *Mock) SEOEntries(ctx context.Context) []models.SEOEntry {
	rows := []struct{ path, lang, title, desc string }{
		{"/", "en", "Previa — Property for sale and rent worldwide",
			"Search apartments, houses, villas, commercial property and land across Europe. Verified brokers, accurate maps and daily new listings."},
		{"/", "de", "Previa — Immobilien kaufen und mieten",
			"Wohnungen, Häuser, Villen, Gewerbeimmobilien und Grundstücke in ganz Europa. Geprüfte Makler und täglich neue Angebote."},
		{"/", "es", "Previa — Propiedades en venta y alquiler",
			"Busca pisos, casas, villas, locales comerciales y terrenos en toda Europa. Agentes verificados y anuncios nuevos cada día."},
		{"/search", "en", "Property search — Previa",
			"Filter by location, price, size, rooms and features. View results as a grid, a list or on the map."},
		{"/search", "de", "Immobiliensuche — Previa",
			"Filtern Sie nach Ort, Preis, Größe, Zimmern und Ausstattung. Ergebnisse als Raster, Liste oder Karte."},
		{"/developments", "en", "New developments — Previa",
			"Browse new-build projects across Europe with completion dates, unit availability and prices from the developer."},
		{"/brokers", "en", "Find a broker — Previa",
			"Search verified real-estate brokers by country, city, language and specialisation."},
		{"/articles", "en", "Property advice and market news — Previa",
			"Guides on buying, selling, renting, investing and renovating, written by working brokers."},
		{"/pricing", "en", "Listing packages and pricing — Previa",
			"Compare Basic, Standard, Premium and Agency packages. Publish a property from €19."},
		{"/login", "en", "Log in — Previa", "Access your saved properties, searches and listings."},
	}
	out := make([]models.SEOEntry, 0, len(rows))
	for i, r := range rows {
		out = append(out, models.SEOEntry{
			Path: r.path, Language: r.lang, Title: r.title, Description: r.desc,
			OGImage: "/static/img/hero.jpg", UpdatedAt: m.now.AddDate(0, 0, -i*3),
		})
	}
	return out
}

// buildRestricted lists admin-managed map/listing restrictions. Nothing is
// hardcoded in the application — every entry is editable in the admin UI.
func buildRestricted(now time.Time) []models.RestrictedCountry {
	return []models.RestrictedCountry{
		{Code: "RU", Name: "Russia", Reason: "Sanctions compliance — listings and map coverage disabled",
			AddedBy: "system.admin", AddedAt: now.AddDate(0, -8, -12)},
		{Code: "BY", Name: "Belarus", Reason: "Sanctions compliance — listings and map coverage disabled",
			AddedBy: "system.admin", AddedAt: now.AddDate(0, -8, -12)},
		{Code: "KP", Name: "North Korea", Reason: "Service not offered in this market",
			AddedBy: "system.admin", AddedAt: now.AddDate(0, -8, -12)},
	}
}

// ---------------------------------------------------------------------------
// Admin dashboard
// ---------------------------------------------------------------------------

// Stats builds the admin dashboard payload.
func (m *Mock) Stats(ctx context.Context) models.AdminStats {
	return models.AdminStats{
		Tiles: []models.AdminStat{
			{Label: "Active listings", Value: "1 486", Delta: "+4.2%", Trend: "up", Icon: "home", Hint: "vs. previous 30 days"},
			{Label: "Pending approval", Value: "23", Delta: "+6", Trend: "up", Icon: "clock", Hint: "awaiting moderation"},
			{Label: "Registered users", Value: "18 240", Delta: "+2.8%", Trend: "up", Icon: "users", Hint: "vs. previous 30 days"},
			{Label: "Revenue (30 days)", Value: "€42 180", Delta: "+11.4%", Trend: "up", Icon: "card", Hint: "package sales"},
			{Label: "Active brokers", Value: "312", Delta: "+9", Trend: "up", Icon: "badge", Hint: "with a live listing"},
			{Label: "Enquiries sent", Value: "6 704", Delta: "-1.9%", Trend: "down", Icon: "mail", Hint: "vs. previous 30 days"},
		},
		PendingCount: 23,
		ListingsByType: []models.ChartSlice{
			{Label: "Apartments", Value: 812, Color: "var(--navy)"},
			{Label: "Houses", Value: 341, Color: "var(--slate)"},
			{Label: "Commercial", Value: 168, Color: "var(--gold)"},
			{Label: "Land", Value: 122, Color: "#8AA0B4"},
			{Label: "Villas", Value: 43, Color: "#B9C6D2"},
		},
		SignupsByMonth: []models.ChartPoint{
			{Label: "Mar", Value: 980}, {Label: "Apr", Value: 1120}, {Label: "May", Value: 1340},
			{Label: "Jun", Value: 1290}, {Label: "Jul", Value: 1510}, {Label: "Aug", Value: 1680},
		},
		RevenueByMonth: []models.ChartPoint{
			{Label: "Mar", Value: 28400}, {Label: "Apr", Value: 31200}, {Label: "May", Value: 34800},
			{Label: "Jun", Value: 33100}, {Label: "Jul", Value: 37900}, {Label: "Aug", Value: 42180},
		},
		RecentActivity: []models.ActivityEntry{
			{Actor: "moderator.liis", Action: "approved listing", Target: "Sea-view penthouse in Kalamaja", Kind: "approve", At: m.now.Add(-18 * time.Minute)},
			{Actor: "moderator.jonas", Action: "rejected listing", Target: "Untitled listing #4821", Kind: "reject", At: m.now.Add(-52 * time.Minute)},
			{Actor: "admin.previa", Action: "updated package", Target: "Premium — €89", Kind: "create", At: m.now.Add(-3 * time.Hour)},
			{Actor: "moderator.liis", Action: "verified broker", Target: "Petra Novák", Kind: "approve", At: m.now.Add(-6 * time.Hour)},
			{Actor: "admin.previa", Action: "added restricted country", Target: "KP — North Korea", Kind: "create", At: m.now.AddDate(0, 0, -1)},
			{Actor: "moderator.marc", Action: "deleted listing", Target: "Duplicate — Carrer de Mallorca 214", Kind: "delete", At: m.now.AddDate(0, 0, -1)},
			{Actor: "admin.previa", Action: "signed in", Target: "83.145.22.14", Kind: "login", At: m.now.AddDate(0, 0, -2)},
		},
	}
}

// Users returns the admin user table.
func (m *Mock) Users(ctx context.Context) []models.User {
	rows := []struct {
		id, name, email, role, country string
		days                           int
		verified                       bool
	}{
		{"us-01", "Anna Lehtinen", "anna.lehtinen@example.com", "user", "EE", 870, true},
		{"us-02", "Kadri Tamm", "kadri.tamm@kadaka.example", "broker", "EE", 1620, true},
		{"us-03", "Jonas Weber", "j.weber@hauptstadt.example", "broker", "DE", 1240, true},
		{"us-04", "Marc Puig", "marc.puig@mediterrania.example", "broker", "ES", 980, true},
		{"us-05", "Tobias Lang", "tobias.lang@example.com", "user", "DE", 412, true},
		{"us-06", "Clara Ferrer", "clara.ferrer@example.com", "user", "ES", 96, false},
		{"us-07", "Aino Virtanen", "aino.virtanen@pohjolakoti.example", "broker", "FI", 760, true},
		{"us-08", "Rui Almeida", "rui.almeida@tejo.example", "broker", "PT", 640, true},
		{"us-09", "Helena Mägi", "helena.magi@example.com", "user", "EE", 288, true},
		{"us-10", "System Administrator", "admin@previa.estate", "admin", "EE", 1900, true},
		{"us-11", "Daan Visser", "daan.visser@grachten.example", "broker", "NL", 520, true},
		{"us-12", "Petra Novák", "petra.novak@vltava.example", "broker", "CZ", 210, false},
	}
	out := make([]models.User, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.User{
			ID: r.id, Name: r.name, Email: r.email, Role: r.role,
			CountryCode: r.country, Language: "en", IsVerified: r.verified,
			MemberSince: m.now.AddDate(0, 0, -r.days),
		})
	}
	return out
}

// Backups returns mock backup history. No real backup is ever performed.
func (m *Mock) Backups(ctx context.Context) []models.Backup {
	return []models.Backup{
		{ID: "bk-01", Name: "previa-full-2026-08-10-0200", Kind: models.BackupSite, Size: "4.2 GB",
			Destination: "gdrive", Status: "complete", CreatedAt: m.now.AddDate(0, 0, 0).Add(-20 * time.Hour)},
		{ID: "bk-02", Name: "previa-db-2026-08-10-0200", Kind: models.BackupMySQL, Size: "812 MB",
			Destination: "gdrive", Status: "complete", CreatedAt: m.now.AddDate(0, 0, 0).Add(-20 * time.Hour)},
		{ID: "bk-03", Name: "previa-db-2026-08-09-0200", Kind: models.BackupMySQL, Size: "808 MB",
			Destination: "local", Status: "complete", CreatedAt: m.now.AddDate(0, 0, -1)},
		{ID: "bk-04", Name: "previa-full-2026-08-08-0200", Kind: models.BackupSite, Size: "4.1 GB",
			Destination: "local", Status: "complete", CreatedAt: m.now.AddDate(0, 0, -2)},
		{ID: "bk-05", Name: "previa-db-2026-08-07-0200", Kind: models.BackupMySQL, Size: "801 MB",
			Destination: "gdrive", Status: "failed", CreatedAt: m.now.AddDate(0, 0, -3)},
		{ID: "bk-06", Name: "previa-full-2026-08-06-0200", Kind: models.BackupSite, Size: "4.0 GB",
			Destination: "local", Status: "complete", CreatedAt: m.now.AddDate(0, 0, -4)},
	}
}

// Files returns a mock directory listing for the file-manager screen. It never
// touches the real filesystem.
func (m *Mock) Files(ctx context.Context, path string) []models.FileEntry {
	base := []models.FileEntry{
		{Name: "web", Path: "/web", IsDir: true, Size: "—", Perms: "drwxr-xr-x", Owner: "previa", Modified: m.now.AddDate(0, 0, -2)},
		{Name: "internal", Path: "/internal", IsDir: true, Size: "—", Perms: "drwxr-xr-x", Owner: "previa", Modified: m.now.AddDate(0, 0, -2)},
		{Name: "cmd", Path: "/cmd", IsDir: true, Size: "—", Perms: "drwxr-xr-x", Owner: "previa", Modified: m.now.AddDate(0, 0, -6)},
		{Name: "docs", Path: "/docs", IsDir: true, Size: "—", Perms: "drwxr-xr-x", Owner: "previa", Modified: m.now.AddDate(0, 0, -1)},
		{Name: "uploads", Path: "/uploads", IsDir: true, Size: "—", Perms: "drwxrwxr-x", Owner: "previa", Modified: m.now.Add(-4 * time.Hour)},
		{Name: "previa", Path: "/previa", IsDir: false, Size: "24.8 MB", Perms: "-rwxr-xr-x", Owner: "previa", Modified: m.now.Add(-20 * time.Hour)},
		{Name: "go.mod", Path: "/go.mod", IsDir: false, Size: "218 B", Perms: "-rw-r--r--", Owner: "previa", Modified: m.now.AddDate(0, 0, -9)},
		{Name: "README.md", Path: "/README.md", IsDir: false, Size: "18.4 KB", Perms: "-rw-r--r--", Owner: "previa", Modified: m.now.AddDate(0, 0, -1)},
		{Name: ".env.example", Path: "/.env.example", IsDir: false, Size: "412 B", Perms: "-rw-r--r--", Owner: "previa", Modified: m.now.AddDate(0, 0, -4)},
	}
	return base
}

// Tables returns a mock schema listing for the MySQL-manager screen. No
// database connection is opened.
func (m *Mock) Tables(ctx context.Context) []models.DBTable {
	rows := []struct {
		name   string
		n      int
		size   string
		engine string
	}{
		{"properties", 1486, "218.4 MB", "InnoDB"},
		{"property_images", 11284, "94.2 MB", "InnoDB"},
		{"users", 18240, "12.8 MB", "InnoDB"},
		{"brokers", 312, "1.4 MB", "InnoDB"},
		{"agencies", 84, "412 KB", "InnoDB"},
		{"developments", 96, "820 KB", "InnoDB"},
		{"articles", 214, "6.2 MB", "InnoDB"},
		{"favourites", 42180, "3.1 MB", "InnoDB"},
		{"saved_searches", 8940, "2.4 MB", "InnoDB"},
		{"notifications", 128400, "22.6 MB", "InnoDB"},
		{"payments", 6218, "1.8 MB", "InnoDB"},
		{"listing_drafts", 1902, "8.4 MB", "InnoDB"},
		{"translations", 10272, "2.2 MB", "InnoDB"},
		{"seo_entries", 640, "284 KB", "InnoDB"},
		{"sessions", 4210, "1.1 MB", "InnoDB"},
		{"audit_log", 96420, "31.2 MB", "InnoDB"},
	}
	out := make([]models.DBTable, 0, len(rows))
	for i, r := range rows {
		out = append(out, models.DBTable{
			Name: r.name, Engine: r.engine, Rows: r.n, Size: r.size,
			Collation: "utf8mb4_unicode_ci", UpdatedAt: m.now.Add(-time.Duration(i*37) * time.Minute),
		})
	}
	return out
}

// SystemInfo backs the restart and cache panel.
func (m *Mock) SystemInfo(ctx context.Context) models.SystemInfo {
	return models.SystemInfo{
		BinaryBuiltAt: buildTime,
		Version:       "0.9.0-frontend",
		GoVersion:     runtime.Version(),
		Uptime:        "4d 6h 18m",
		Environment:   "staging",
		CacheEntries:  18420,
		CacheSize:     "184 MB",
		MemoryUsage:   "412 MB / 2 GB",
	}
}

// ActivityLog returns the admin audit trail.
func (m *Mock) ActivityLog(ctx context.Context) []models.ActivityEntry {
	return m.Stats(ctx).RecentActivity
}

// buildTime is stamped when the process starts; the admin panel shows it as
// the running binary's build timestamp.
var buildTime = time.Now()
