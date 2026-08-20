package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"previa/pkg/data"
	"previa/pkg/models"
)

// parseAll returns a filter that matches everything (used by the sitemap).
func parseAll() data.PropertyFilter {
	return data.PropertyFilter{Page: 1, PerPage: 200, Sort: data.SortNewest}
}

// AccountData is the payload shared by the account screens.
type AccountData struct {
	Section       string
	Listings      []models.Property
	Favourites    []models.Property
	Saved         []models.SavedSearch
	Notifications []models.Notification
	Drafts        []models.Draft
	Payments      []models.Payment
	Packages      []models.Package
	Promotions    []models.Promotion
	Counts        map[string]int
	Status        string
	Unread        int
	// Stats holds the per-day visitor series for each listing on the page,
	// keyed by property id. Built up front rather than fetched when the panel
	// opens: the whole set for a page of listings is a few hundred integers,
	// and it means opening the statistics is instant and works with no
	// round trip — which matters, because the panel is the thing a seller
	// flicks between listings to compare.
	Stats map[string]models.ListingStats

	// The settings screen's map location field: the autocomplete catalogue and
	// the map centred on wherever the seller's pin currently is.
	LocationSuggestions []models.LocationSuggestion
	OfficeMap           template.JS

	// The homepage broker strip as a purchasable placement, and whatever the
	// seller currently has running in it.
	BrokerAdPlan models.BrokerAdPlan
	BrokerAd     models.BrokerAd

	// And the placement that runs against the pin instead of against the
	// markets: this seller on the search map.
	BrokerMapAdPlan models.BrokerMapAdPlan
	BrokerMapAd     models.BrokerMapAd
}

// statsDays is how far back the statistics panel charts. Two weeks reads as a
// shape — a decay, a weekend rhythm, a spike after a promotion — where seven
// days is mostly noise and a month does not fit legibly in a dialog.
const statsDays = 14

// listingStats gathers the per-day series for a set of listings.
func (h *Handler) listingStats(r *http.Request, listings []models.Property) map[string]models.ListingStats {
	ctx := r.Context()
	out := make(map[string]models.ListingStats, len(listings))
	for _, p := range listings {
		if s, ok := h.Store.Account.ListingStats(ctx, p.ID, statsDays); ok {
			out[p.ID] = s
		}
	}
	return out
}

// listingCounts tallies the user's listings by state for the tab bar.
func (h *Handler) listingCounts(r *http.Request) map[string]int {
	ctx := r.Context()
	all := h.Store.Account.MyListings(ctx, "")
	c := map[string]int{"all": len(all)}
	for _, p := range all {
		c[string(p.Status)]++
	}
	c["drafts"] = len(h.Store.Account.Drafts(ctx))
	c["favourites"] = len(h.Store.Account.Favourites(ctx))
	c["saved"] = len(h.Store.Account.SavedSearches(ctx))
	return c
}

// Dashboard renders the account overview.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd := h.base(r, "", "Dashboard — Previa", "Your Previa account overview.")
	pd.Meta.NoIndex = true

	pd.Data = AccountData{
		Section:       "dashboard",
		Listings:      take(h.Store.Account.MyListings(ctx, ""), 3),
		Favourites:    take(h.Store.Account.Favourites(ctx), 3),
		Saved:         take(h.Store.Account.SavedSearches(ctx), 3),
		Notifications: take(h.Store.Account.Notifications(ctx), 4),
		Drafts:        h.Store.Account.Drafts(ctx),
		Counts:        h.listingCounts(r),
		Unread:        h.Store.Account.UnreadCount(ctx),
	}
	h.View.Render(w, http.StatusOK, "account/dashboard", pd)
}

// MyListings renders the user's listings, filtered by status tab.
func (h *Handler) MyListings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := models.ListingStatus(r.URL.Query().Get("status"))

	pd := h.base(r, "", "My listings — Previa", "Manage the properties you have published on Previa.")
	pd.Meta.NoIndex = true

	listings := h.Store.Account.MyListings(ctx, status)
	pd.Data = AccountData{
		Section:  "listings",
		Listings: listings,
		// Promotion is buyable from here as well as at publish time, so the
		// same add-on catalogue the wizard uses is needed on this page.
		Promotions: h.Store.Catalog.Promotions(ctx),
		Counts:     h.listingCounts(r),
		Status:     string(status),
		Stats:      h.listingStats(r, listings),
	}
	h.View.Render(w, http.StatusOK, "account/my-listings", pd)
}

// CloneListing duplicates one of the user's listings.
//
// Answers the client's "there add 'clone' so can duplicate the ad". The copy is
// always a draft — a duplicate has not been paid for, and draft is the only way
// into the lifecycle.
//
// Nothing is written: this build has no store behind it. The response reports
// what would have been created, names it, and points at the drafts the copy
// would join, so the flow is testable end to end without pretending the row
// exists. A backend implementation swaps the mock's CloneListing for a real
// INSERT and this handler is unchanged.
func (h *Handler) CloneListing(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/listing/clone/")
	copy, ok := h.Store.Account.CloneListing(r.Context(), id)
	if !ok {
		h.View.RenderPartial(w, http.StatusNotFound, "account/my-listings", "clone-failed", nil)
		return
	}
	h.View.RenderPartial(w, http.StatusOK, "account/my-listings", "clone-created", copy)
}

// Drafts renders unfinished add-listing sessions.
func (h *Handler) Drafts(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Draft listings — Previa", "Continue a listing you started earlier.")
	pd.Meta.NoIndex = true
	pd.Data = AccountData{
		Section: "drafts",
		Drafts:  h.Store.Account.Drafts(r.Context()),
		Counts:  h.listingCounts(r),
	}
	h.View.Render(w, http.StatusOK, "account/drafts", pd)
}

// Favourites renders saved properties.
func (h *Handler) Favourites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd := h.base(r, "", "Favourites — Previa", "Properties you have saved on Previa.")
	pd.Meta.NoIndex = true
	pd.Data = AccountData{
		Section:    "favourites",
		Favourites: h.Store.Account.Favourites(ctx),
		Counts:     h.listingCounts(r),
	}
	h.View.Render(w, http.StatusOK, "account/favourites", pd)
}

// SavedSearches renders stored searches and their alert settings.
func (h *Handler) SavedSearches(w http.ResponseWriter, r *http.Request) {
	pd := h.base(r, "", "Saved searches — Previa", "Your saved property searches and alerts.")
	pd.Meta.NoIndex = true
	pd.Data = AccountData{
		Section: "saved",
		Saved:   h.Store.Account.SavedSearches(r.Context()),
		Counts:  h.listingCounts(r),
	}
	h.View.Render(w, http.StatusOK, "account/saved-searches", pd)
}

// Notifications renders the notification centre.
func (h *Handler) Notifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd := h.base(r, "", "Notifications — Previa", "Alerts, messages and account activity.")
	pd.Meta.NoIndex = true
	pd.Data = AccountData{
		Section:       "notifications",
		Notifications: h.Store.Account.Notifications(ctx),
		Counts:        h.listingCounts(r),
		Unread:        h.Store.Account.UnreadCount(ctx),
	}
	h.View.Render(w, http.StatusOK, "account/notifications", pd)
}

// Settings renders profile, account and security settings.
func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd := h.base(r, "", "Settings — Previa", "Manage your Previa profile, account and security settings.")
	pd.Meta.NoIndex = true
	pd.NeedsMap = true // the profile's location picker draws one

	// The profile's map location, and the search box that moves its pin.
	//
	// Centred on the pin the seller already dropped, or on their market if they
	// have not dropped one — an empty map of the whole continent is no use for
	// placing an office.
	pd.Data = AccountData{
		Section:             "settings",
		Counts:              h.listingCounts(r),
		LocationSuggestions: h.Store.Catalog.LocationSuggestions(ctx),
		OfficeMap:           buildPinMap(pd.User.Office, pd.Country, h.Cfg.MapsKey),
		BrokerAdPlan:        h.Store.Catalog.BrokerAdPlan(ctx),
		BrokerAd:            pd.User.Ad,
		BrokerMapAdPlan:     h.Store.Catalog.BrokerMapAdPlan(ctx),
		BrokerMapAd:         pd.User.MapAd,
	}
	h.View.Render(w, http.StatusOK, "account/settings", pd)
}

// Billing renders payment history and billing details.
func (h *Handler) Billing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pd := h.base(r, "", "Billing and payments — Previa", "Your Previa invoices, payment methods and billing details.")
	pd.Meta.NoIndex = true
	pd.Data = AccountData{
		Section:  "billing",
		Payments: h.Store.Account.Payments(ctx),
		Packages: h.Store.Catalog.Packages(ctx),
		Counts:   h.listingCounts(r),
	}
	h.View.Render(w, http.StatusOK, "account/billing", pd)
}

// PayMethod is one payment choice on the checkout screen.
type PayMethod struct {
	Value    string
	Label    string
	Hint     string
	Logo     string // key into the "pay-logo" template
	Checked  bool
	Disabled bool
}

// payMethods is the list the client asked for, in his order.
//
// "Credit card" and "Stripe" are separate on purpose. He listed both, and they
// are two different flows: the first collects the card on Previa with Stripe
// processing behind it, the second hands the payer to Stripe's hosted checkout.
// Neither is dropped without him saying which he meant.
//
// Nothing here contacts a provider in this milestone — see CheckoutProcess.
//
// The lines under the names are the client's, 19 August: "The text under credit
// card write 'processed via Stripe'. Under Paypal and Stripe remove the text as
// everyone knows anyway what they are." So the card keeps a short line — it is
// the one row whose name does not say who processes it — PayPal and Stripe get
// none, and the two nobody can be expected to recognise keep their sentence.
var payMethods = []PayMethod{
	{Value: "card", Label: "Credit card", Logo: "card", Checked: true,
		Hint: "Processed via Stripe."},
	{Value: "paypal", Label: "PayPal", Logo: "paypal"},
	{Value: "stripe", Label: "Stripe", Logo: "stripe"},
	// Both of these were narrowed by the original copy and widened at the
	// client's request: Paysera carries bank links across Europe rather than
	// the Baltics alone, and NOWPayments settles far more than three coins.
	{Value: "paysera", Label: "Paysera bank links", Logo: "paysera",
		Hint: "Pay directly with many bank links around Europe."},
	{Value: "crypto", Label: "Crypto", Logo: "crypto",
		Hint: "Bitcoin, Ethereum, stablecoins and many others through NOWPayments."},
}

// CheckoutData drives the demonstration payment flow.
type CheckoutData struct {
	Package  models.Package
	Packages []models.Package
	State    string // "select" | "processing" | "success" | "failed" | "cancelled"
	Method   string
	Methods  []PayMethod
}

// Checkout renders the package checkout screen.
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	pkgs := h.Store.Catalog.Packages(ctx)
	selected := pkgs[1] // Standard is the default selection
	if id := q.Get("package"); id != "" {
		for _, p := range pkgs {
			if p.ID == id {
				selected = p
			}
		}
	}

	state := q.Get("state")
	if state == "" {
		state = "select"
	}

	pd := h.base(r, "", "Checkout — Previa", "Complete your Previa listing package purchase.")
	pd.Meta.NoIndex = true
	// Re-check whichever method the visitor last chose, so a failed payment
	// returns to the same row rather than resetting to the default.
	methods := make([]PayMethod, len(payMethods))
	copy(methods, payMethods)
	if m := q.Get("method"); m != "" {
		for i := range methods {
			methods[i].Checked = methods[i].Value == m
		}
	}

	pd.Data = CheckoutData{
		Package: selected, Packages: pkgs, State: state,
		Method: q.Get("method"), Methods: methods,
	}
	h.View.Render(w, http.StatusOK, "account/checkout", pd)
}

// CheckoutProcess simulates a payment result. No payment provider is contacted
// and no credentials exist in this build.
func (h *Handler) CheckoutProcess(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	method := r.FormValue("method")
	pkg := r.FormValue("package")
	action := r.FormValue("action")

	// Each outcome is reachable from the UI, deterministically, so the client
	// can review all three without any provider being involved. Paysera stands
	// in for a declined payment and the Cancel button for an abandoned one;
	// everything else succeeds.
	//
	// No provider is contacted, no credentials exist and nothing is stored.
	state := "success"
	switch {
	case action == "cancel":
		state = "cancelled"
	case method == "paysera":
		state = "failed"
	}

	http.Redirect(w, r,
		"/checkout?package="+pkg+"&method="+method+"&state="+state, http.StatusSeeOther)
}

// ContactBroker acknowledges a mock enquiry.
func (h *Handler) ContactBroker(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = "there"
	}
	h.View.RenderPartial(w, http.StatusOK, "property-detail", "contact-success",
		map[string]any{"Name": name})
}

// SaveSearch stores the search that was posted and confirms it.
//
// "Then add there button 'save search', if hit that then this search will be
// saved under user's 'saved searches' menu." So this reads the filter form the
// button posted rather than acknowledging a click: the filters become a row on
// /saved-searches, described in the same words the tag bar above the results
// uses, and replayable from the query it was saved with.
//
// Two callers, one route: the panel's footer button asks for the compact
// confirmation (compact=1, a line inside a 40px footer) and the tag bar's
// "Save this search" gets the full alert.
func (h *Handler) SaveSearch(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	form := r.PostForm

	f := data.ParseFilter(form, func(key string) []string { return form[key] })
	data.ApplyLocation(&f, h.Store.Catalog.ResolveLocation)
	if f.CountryCode != "" {
		if c, ok := h.Store.Catalog.Country(r.Context(), f.CountryCode); ok {
			f.CountryName = c.Name
		}
	}

	// How many listings it matches today, which is what the row on
	// /saved-searches reports beside it.
	result, _ := h.Store.Properties.Search(r.Context(), f)

	saved := h.Store.Account.AddSavedSearch(r.Context(), models.SavedSearch{
		Name:    savedSearchName(f),
		Query:   savedSearchQuery(form),
		Summary: savedSearchSummary(f),
		Deal:    savedSearchDeal(f),
		// Alerts on and instant, because saving a search is asking to hear
		// about it. Both are editable on the saved-searches screen.
		AlertsOn:    true,
		Frequency:   "instant",
		ResultCount: result.Total,
	})

	block := "save-search-success"
	if form.Get("compact") == "1" {
		block = "save-search-line"
	}
	h.View.RenderPartial(w, http.StatusOK, "search", block,
		map[string]any{"Saved": saved})
}

// savedSearchName is the row's title: the filters that identify it, which is
// the deal type and the place, falling back to what there is.
func savedSearchName(f data.PropertyFilter) string {
	chips := f.Chips()
	if len(chips) == 0 {
		return "All properties"
	}
	var parts []string
	for _, c := range chips {
		parts = append(parts, c.Label)
		if len(parts) == 2 {
			break
		}
	}
	return strings.Join(parts, " · ")
}

// savedSearchSummary is the whole filter in words — the same labels the tag bar
// above the results shows, so a saved search reads as what was on screen when
// it was saved.
func savedSearchSummary(f data.PropertyFilter) string {
	var parts []string
	for _, c := range f.Chips() {
		parts = append(parts, c.Label)
	}
	if len(parts) == 0 {
		return "Every listing, unfiltered"
	}
	return strings.Join(parts, " · ")
}

// savedSearchDeal is the deal type the row is filed under. A search across
// several is filed under the first, which is the order the panel offers them in.
func savedSearchDeal(f data.PropertyFilter) models.DealType {
	if len(f.Deals) == 0 {
		return models.DealSale
	}
	return f.Deals[0]
}

// savedSearchQuery is the query string that replays the search.
//
// The posted form minus the two fields that are about this request rather than
// about the search: `compact` chooses which confirmation comes back, and the
// blank values every unset control submits would otherwise be stored as filters
// that are set to nothing.
func savedSearchQuery(form url.Values) string {
	q := url.Values{}
	for key, values := range form {
		if key == "compact" {
			continue
		}
		for _, v := range values {
			if strings.TrimSpace(v) == "" {
				continue
			}
			q.Add(key, v)
		}
	}
	return q.Encode()
}

// RevealContact returns the broker's real contact details, mimicking the
// click-to-reveal pattern that limits scraping.
func (h *Handler) RevealContact(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := r.FormValue("broker")
	b, ok := h.Store.Brokers.ByID(r.Context(), id)
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	h.View.RenderPartial(w, http.StatusOK, "property-detail", "revealed-contact", b)
}
