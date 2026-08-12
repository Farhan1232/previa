package handlers

import (
	"net/http"
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
	Counts        map[string]int
	Status        string
	Unread        int
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

	pd.Data = AccountData{
		Section:  "listings",
		Listings: h.Store.Account.MyListings(ctx, status),
		Counts:   h.listingCounts(r),
		Status:   string(status),
	}
	h.View.Render(w, http.StatusOK, "account/my-listings", pd)
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
	pd := h.base(r, "", "Profile and settings — Previa", "Manage your Previa profile, account and security settings.")
	pd.Meta.NoIndex = true
	pd.Data = AccountData{Section: "settings", Counts: h.listingCounts(r)}
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
var payMethods = []PayMethod{
	{Value: "card", Label: "Credit card", Logo: "card", Checked: true,
		Hint: "Visa, Mastercard or American Express. Processed through Stripe."},
	{Value: "paypal", Label: "PayPal", Logo: "paypal",
		Hint: "You'll be redirected to PayPal to approve the payment."},
	{Value: "stripe", Label: "Stripe", Logo: "stripe",
		Hint: "Pay on Stripe's own checkout — wallets, saved cards and local methods."},
	{Value: "paysera", Label: "Paysera bank links", Logo: "paysera",
		Hint: "Pay directly from an Estonian, Latvian or Lithuanian bank account."},
	{Value: "crypto", Label: "Crypto", Logo: "crypto",
		Hint: "Bitcoin, Ethereum and stablecoins through NOWPayments."},
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

// SaveSearch acknowledges a mock saved search.
func (h *Handler) SaveSearch(w http.ResponseWriter, r *http.Request) {
	h.View.RenderPartial(w, http.StatusOK, "search", "save-search-success", nil)
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
