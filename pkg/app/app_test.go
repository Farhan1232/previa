package app_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"previa/pkg/app"
	"previa/pkg/config"
)

// Tests for the client's 12 August corrections.
//
// These exercise what the server is responsible for: the routes, the filter
// state that has to survive being carried in a URL, and the wording changes.
// The browser-side behaviour those steps also cover — the card click zones, the
// dot and arrow paging, the market country search, the List/Grid switch, the
// waypoint navigation and the media reordering — is verified separately against
// a real browser, because it only exists once JavaScript has run.

func newServer(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Load()
	// Templates and static files are read relative to the repository root.
	cfg.TemplateDir = "../../web/templates"
	cfg.StaticDir = "../../public/static"
	h, err := app.New(cfg)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return h
}

// get performs a GET and returns the status and body.
func get(t *testing.T, h http.Handler, path string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

// mustGet fails the test unless the request returns 200.
func mustGet(t *testing.T, h http.Handler, path string) string {
	t.Helper()
	code, body := get(t, h, path)
	if code != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", path, code)
	}
	return body
}

// ---------------------------------------------------------------------------
// Route sweep
// ---------------------------------------------------------------------------

func TestRoutesRespond(t *testing.T) {
	h := newServer(t)

	paths := []string{
		"/", "/search", "/search?view=list", "/search?view=map", "/search?view=full",
		"/search/results", "/developments", "/brokers", "/agencies", "/articles",
		"/pricing", "/help", "/about", "/advertising", "/terms", "/privacy", "/cookies",
		"/login", "/register", "/forgot-password",
		"/add-listing", "/dashboard", "/my-listings", "/drafts", "/favourites",
		"/saved-searches", "/notifications", "/settings", "/billing", "/checkout",
		"/admin", "/admin/listings", "/admin/users", "/admin/packages", "/admin/payments",
		"/admin/languages", "/admin/banners", "/admin/maps", "/admin/system",
		"/robots.txt", "/sitemap.xml",
	}
	for _, p := range paths {
		if code, _ := get(t, h, p); code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", p, code)
		}
	}
}

// ---------------------------------------------------------------------------
// Step 7 — homepage search carries its three values to every destination
// ---------------------------------------------------------------------------

func TestHomeSearchFieldsAreDealLocationType(t *testing.T) {
	body := mustGet(t, newServer(t), "/")

	for _, want := range []string{`id="hero-deal"`, `id="hero-location"`, `id="hero-type"`} {
		if !strings.Contains(body, want) {
			t.Errorf("homepage search is missing %s", want)
		}
	}
	// Price and bedrooms were moved to the sidebar's advanced filters.
	for _, gone := range []string{`id="hero-price-min"`, `id="hero-bedrooms"`} {
		if strings.Contains(body, gone) {
			t.Errorf("homepage search still carries %s", gone)
		}
	}
	// The three continuation actions all submit the same form.
	for _, want := range []string{
		`name="filters" value="open"`,
		`name="view" value="map"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("homepage search is missing the continuation button %s", want)
		}
	}
}

func TestSearchFiltersTransferFromHomepage(t *testing.T) {
	h := newServer(t)

	// The URL each of the three homepage actions produces.
	cases := map[string]string{
		"search properties": "/search?deal=rent&location=Tallinn%2C+Estonia&property_type=apartment",
		"advanced filters":  "/search?deal=rent&location=Tallinn%2C+Estonia&property_type=apartment&filters=open",
		"search on map":     "/search?deal=rent&location=Tallinn%2C+Estonia&property_type=apartment&view=map",
	}

	for name, path := range cases {
		body := mustGet(t, h, path)

		// Deal type: the sidebar radio for "rent" comes back checked.
		if !regexp.MustCompile(`name="deal" value="rent"[^>]*checked`).MatchString(body) {
			t.Errorf("%s: sidebar did not prefill deal=rent", name)
		}
		// Location: the label is redisplayed in the single Location field.
		if !strings.Contains(body, `value="Tallinn, Estonia"`) {
			t.Errorf("%s: sidebar did not prefill the location label", name)
		}
		// Property type: the apartment checkbox comes back checked. The
		// attributes wrap across lines in the template, so match on the
		// squashed markup rather than assuming they sit together.
		if !regexp.MustCompile(`value="apartment"\s+checked`).MatchString(squash(body)) {
			t.Errorf("%s: sidebar did not prefill property_type=apartment", name)
		}
	}
}

// A location that resolves must actually narrow the result set, not just be
// echoed back into the field.
func TestLocationFilterNarrowsResults(t *testing.T) {
	h := newServer(t)

	all := mustGet(t, h, "/search")
	tallinn := mustGet(t, h, "/search?location=Tallinn%2C+Estonia")

	countAll := strings.Count(all, `class="pcard`)
	countTallinn := strings.Count(tallinn, `class="pcard`)

	if countTallinn == 0 {
		t.Fatal("filtering by Tallinn returned no properties")
	}
	if countTallinn >= countAll {
		t.Errorf("filtering by Tallinn did not narrow results: %d of %d", countTallinn, countAll)
	}
	if !strings.Contains(tallinn, "Tallinn") {
		t.Error("Tallinn results do not mention Tallinn")
	}
}

// ---------------------------------------------------------------------------
// Steps 8, 15 — Deal type wording, everywhere it appears
// ---------------------------------------------------------------------------

func TestDealTypeWordingIsConsistent(t *testing.T) {
	h := newServer(t)

	for _, path := range []string{"/search", "/search?view=map", "/add-listing"} {
		body := mustGet(t, h, path)
		for _, want := range []string{"Sell", "Rent", "Short rent"} {
			if !strings.Contains(body, want) {
				t.Errorf("%s does not offer %q", path, want)
			}
		}
		for _, gone := range []string{"I want to", "Sale or rent"} {
			if strings.Contains(body, gone) {
				t.Errorf("%s still uses the old wording %q", path, gone)
			}
		}
	}
}

func TestShortRentIsSearchable(t *testing.T) {
	h := newServer(t)

	body := mustGet(t, h, "/search?deal=short_rent")
	if strings.Contains(body, "No properties match these filters") {
		t.Error("no seeded listings are short rentals, so the filter looks broken")
	}
	if !strings.Contains(body, "Short rent") {
		t.Error("short-rent results do not carry the Short rent badge")
	}
}

// ---------------------------------------------------------------------------
// Step 5 — the market selector offers the whole world, searchably
// ---------------------------------------------------------------------------

func TestMarketSelectorCarriesEveryCountry(t *testing.T) {
	body := mustGet(t, newServer(t), "/")

	countries := regexp.MustCompile(`data-market-name="`).FindAllString(body, -1)
	if len(countries) < 190 {
		t.Errorf("market selector lists %d countries, want ~192", len(countries))
	}

	// Both halves of the client's acceptance test: by name and by code.
	if !strings.Contains(body, `data-market-name="Germany" data-market-code="DE"`) {
		t.Error("Germany is not searchable by name and code")
	}
	// The search field and its empty state are present.
	for _, want := range []string{"market-menu__input", "market-menu__empty", "market-menu__clear"} {
		if !strings.Contains(body, want) {
			t.Errorf("market selector is missing %s", want)
		}
	}
}

func TestSelectingAMarketSetsTheCookie(t *testing.T) {
	h := newServer(t)

	req := httptest.NewRequest(http.MethodGet, "/set-country?code=DE&return=/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("set-country = %d, want 303", rec.Code)
	}
	var got string
	for _, c := range rec.Result().Cookies() {
		if c.Name == "previa_country" {
			got = c.Value
		}
	}
	if got != "DE" {
		t.Errorf("market cookie = %q, want DE", got)
	}
}

// ---------------------------------------------------------------------------
// Steps 4, 6, 14 — header wording and controls
// ---------------------------------------------------------------------------

func TestLanguageMenuShowsNamesOnly(t *testing.T) {
	body := mustGet(t, newServer(t), "/dashboard")

	menu := between(body, `class="menu lang-menu"`, `</div>`)
	if menu == "" {
		t.Fatal("language menu not found")
	}
	for _, want := range []string{"English", "German", "Spanish"} {
		if !strings.Contains(menu, want) {
			t.Errorf("language menu is missing %q", want)
		}
	}
	if regexp.MustCompile(`\d+%`).MatchString(menu) {
		t.Error("language menu still shows translation percentages")
	}
	if strings.Contains(menu, `class="lang-code"`) {
		t.Error("language menu still shows two-letter codes")
	}
}

func TestUserButtonIsRectangularWithoutChevron(t *testing.T) {
	body := mustGet(t, newServer(t), "/dashboard")

	if !strings.Contains(body, `class="user-btn"`) {
		t.Error("account button is not the rectangular user-btn")
	}
	if !strings.Contains(body, `class="user-btn__initials"`) {
		t.Error("account button does not show initials")
	}
	btn := between(body, `class="user-btn"`, `</button>`)
	if strings.Contains(btn, "chevron-down") || strings.Contains(btn, "m6 9 6 6 6-6") {
		t.Error("account button still draws a down arrow")
	}
}

func TestAddListingWordingHasNoPlus(t *testing.T) {
	h := newServer(t)

	for _, path := range []string{"/", "/my-listings", "/dashboard", "/pricing"} {
		body := mustGet(t, h, path)
		if strings.Contains(body, "+ add listing") || strings.Contains(body, "+ Add listing") {
			t.Errorf("%s still shows a plus before the add listing label", path)
		}
		if strings.Contains(body, ">Add listing<") {
			t.Errorf("%s still uses the capitalised Add listing label", path)
		}
	}
}

// ---------------------------------------------------------------------------
// Steps 9, 10 — one Location field, euro only
// ---------------------------------------------------------------------------

func TestSidebarHasOneLocationFieldAndEuroOnly(t *testing.T) {
	body := mustGet(t, newServer(t), "/search")

	// The three separate location inputs are gone.
	for _, gone := range []string{`id="f-country"`, `id="f-city"`, `id="f-address"`} {
		if strings.Contains(body, gone) {
			t.Errorf("sidebar still carries the separate location field %s", gone)
		}
	}
	if !strings.Contains(body, `id="f-location"`) {
		t.Error("sidebar is missing the single Location field")
	}

	currency := between(body, `id="f-currency"`, `</select>`)
	if strings.Contains(currency, "CZK") || strings.Contains(currency, "GBP") {
		t.Error("currency control still offers a currency other than the euro")
	}
	if !strings.Contains(currency, "EUR") {
		t.Error("currency control does not offer the euro")
	}
}

// Every price in the interface is quoted in euro, so the EUR-only control is
// not contradicted by a koruna price on a card.
func TestPricesAreAllEuro(t *testing.T) {
	body := mustGet(t, newServer(t), "/search?per_page=60")
	if strings.Contains(body, "Kč") {
		t.Error("a price is still quoted in Czech koruna")
	}
}

// ---------------------------------------------------------------------------
// Steps 16, 17, 18, 19 — the add-listing form
// ---------------------------------------------------------------------------

func TestAddListingIsOneContinuousForm(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")

	// Every section is on the page at once.
	sections := []string{
		"ls-deal", "ls-category", "ls-location", "ls-details", "ls-rooms",
		"ls-features", "ls-description", "ls-media", "ls-price", "ls-contact",
		"ls-preview", "ls-publish",
	}
	for _, id := range sections {
		if !strings.Contains(body, `id="`+id+`"`) {
			t.Errorf("add-listing is missing section %s", id)
		}
	}
	// The step buttons are gone.
	for _, gone := range []string{"wizard-card__foot", "add-listing?step="} {
		if strings.Contains(body, gone) {
			t.Errorf("add-listing still carries %q", gone)
		}
	}
	if strings.Contains(body, ">Continue ") || strings.Contains(body, "> Back<") {
		t.Error("add-listing still has Back/Continue navigation")
	}
}

func TestAddListingWaypointsRenumbered(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")

	// Waypoint 4 (the old public-location-display step) is gone, and so is
	// the separate package waypoint.
	for _, gone := range []string{"Public location display", "Package and promotion", "Address and map pin"} {
		if strings.Contains(body, gone) {
			t.Errorf("waypoint %q should have been removed", gone)
		}
	}
	if !strings.Contains(body, "Photos &amp; videos") && !strings.Contains(body, "Photos & videos") {
		t.Error("media waypoint is not named Photos & videos")
	}
	// Package selection moved into Publish.
	publish := between(body, `id="ls-publish"`, "</section>")
	if !strings.Contains(publish, `name="package"`) {
		t.Error("Publish does not carry the package choice")
	}
	if !strings.Contains(publish, "Choose package") {
		t.Error("Publish does not label the package choice")
	}
}

func TestAddListingLocationSection(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")
	loc := between(body, `id="ls-location"`, "</section>")

	// The structured fields the map fills in, all read-only.
	for _, id := range []string{
		"w-country", "w-country-code", "w-state", "w-city",
		"w-district", "w-street", "w-lat", "w-lng",
	} {
		field := between(loc, `id="`+id+`"`, ">")
		if field == "" {
			t.Errorf("location section is missing %s", id)
			continue
		}
		if !strings.Contains(field, "readonly") {
			t.Errorf("%s should be read-only — it comes from the map", id)
		}
	}

	// The one editable field, and its exact label.
	if !strings.Contains(loc, "Edit your location as you want other users to see it") {
		t.Error("the public display address field is missing its label")
	}
	pub := between(loc, `id="w-public-address"`, ">")
	if strings.Contains(pub, "readonly") {
		t.Error("the public display address must stay editable")
	}
}

func TestAddListingMediaSection(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")
	media := between(body, `id="ls-media"`, "</section>")

	if strings.Contains(media, "JPEG or PNG, up to 10 MB each") {
		t.Error("the JPEG/PNG 10 MB line should have been removed")
	}
	if !strings.Contains(media, "15 MB") {
		t.Error("the media section does not state the 15 MB video limit")
	}
	accept := between(media, `accept="`, `"`)
	for _, want := range []string{"image/jpeg", "image/png", "video/mp4"} {
		if !strings.Contains(accept, want) {
			t.Errorf("file picker does not accept %s", want)
		}
	}
	// Keyboard reordering, as the alternative to dragging.
	if !strings.Contains(media, "photo-thumb__move") {
		t.Error("media tiles have no keyboard reorder controls")
	}
}

// A deep link from the old wizard still lands on the right section.
func TestAddListingDeepLinksResolveToSections(t *testing.T) {
	h := newServer(t)

	for query, want := range map[string]string{
		"?step=1":        "'deal'",
		"?step=3":        "'location'",
		"?section=media": "'media'",
		"?step=12":       "'publish'",
	} {
		body := mustGet(t, h, "/add-listing"+query)
		if !strings.Contains(body, "previaListingForm(12, "+want+")") {
			t.Errorf("/add-listing%s did not resolve to section %s", query, want)
		}
	}
}

// ---------------------------------------------------------------------------
// Steps 20, 21 — checkout
// ---------------------------------------------------------------------------

func TestCheckoutWordingAndVAT(t *testing.T) {
	body := mustGet(t, newServer(t), "/checkout?package=pk-standard")

	if !strings.Contains(body, "Choose package") {
		t.Error(`checkout does not say "Choose package"`)
	}
	if strings.Contains(body, "Your package") {
		t.Error(`checkout still says "Your package"`)
	}

	summary := between(body, "checkout-summary", "</aside>")
	if strings.Contains(summary, "VAT") {
		t.Error("the order summary still carries a VAT row")
	}
	// The total is the package price, with nothing added.
	if !strings.Contains(summary, "€39") {
		t.Error("the order summary total is not the package price")
	}
}

func TestCheckoutOffersEveryRequestedPaymentMethod(t *testing.T) {
	body := mustGet(t, newServer(t), "/checkout?package=pk-standard")

	// The client listed Credit card and Stripe separately; both are kept.
	for value, label := range map[string]string{
		"card":    "Credit card",
		"paypal":  "PayPal",
		"stripe":  "Stripe",
		"paysera": "Paysera bank links",
		"crypto":  "Crypto",
	} {
		if !strings.Contains(body, `value="`+value+`"`) {
			t.Errorf("checkout does not offer the %s method", label)
		}
		if !strings.Contains(body, label) {
			t.Errorf("checkout does not label the %s method", label)
		}
	}
	if !strings.Contains(body, "NOWPayments") {
		t.Error("the crypto method does not name NOWPayments")
	}
	// Local logos only — nothing is fetched from a provider.
	if !strings.Contains(body, "pay-logo__svg") {
		t.Error("payment methods have no provider logos")
	}
	if strings.Contains(body, "https://js.stripe.com") || strings.Contains(body, "paypal.com/sdk") {
		t.Error("a payment provider script is being loaded")
	}
}

func TestCheckoutReachesEveryMockOutcome(t *testing.T) {
	h := newServer(t)

	cases := []struct {
		form, wantState string
	}{
		{"package=pk-standard&method=card&terms=on", "success"},
		{"package=pk-standard&method=paysera&terms=on", "failed"},
		{"package=pk-standard&method=card&action=cancel&terms=on", "cancelled"},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/checkout/process", strings.NewReader(c.form))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Fatalf("checkout/process %q = %d, want 303", c.form, rec.Code)
		}
		if loc := rec.Header().Get("Location"); !strings.Contains(loc, "state="+c.wantState) {
			t.Errorf("checkout/process %q redirected to %q, want state=%s", c.form, loc, c.wantState)
		}
	}
}

func TestAdminPricePackagesScreen(t *testing.T) {
	body := mustGet(t, newServer(t), "/admin/packages")

	if !strings.Contains(body, "Price packages") {
		t.Error("the admin screen is not called Price packages")
	}
	for _, want := range []string{
		"Create package", "Edit package", "Included features", "Duration (days)", "Price (EUR)",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("admin packages screen is missing %q", want)
		}
	}
	// Enable/disable per package.
	if !strings.Contains(body, "the Standard package") {
		t.Error("admin packages screen has no per-package enable/disable control")
	}
}

// ---------------------------------------------------------------------------
// Step 17 — mock reverse geocoding
// ---------------------------------------------------------------------------

func TestReverseGeocodeReturnsAStructuredPlace(t *testing.T) {
	code, body := get(t, newServer(t), "/mock/reverse-geocode?lat=59.437&lng=24.7536")
	if code != http.StatusOK {
		t.Fatalf("reverse-geocode = %d, want 200", code)
	}
	for _, want := range []string{`"CountryCode":"EE"`, `"City":"Tallinn"`, `"Lat":59.437`} {
		if !strings.Contains(body, want) {
			t.Errorf("reverse-geocode response is missing %s\ngot: %s", want, body)
		}
	}
}

// ---------------------------------------------------------------------------
// Step 11 — the listing-page banner sits inside the results column
// ---------------------------------------------------------------------------

func TestSearchBannerSitsInsideTheResultsColumn(t *testing.T) {
	body := mustGet(t, newServer(t), "/search")

	layout := strings.Index(body, `class="search-layout`)
	banner := strings.Index(body, `class="search-banner"`)
	sidebar := strings.Index(body, `class="filter-panel"`)
	results := strings.Index(body, `id="results"`)

	if layout < 0 || banner < 0 || sidebar < 0 || results < 0 {
		t.Fatal("search page is missing one of layout, banner, sidebar or results")
	}
	if banner < layout {
		t.Error("the banner is still above the whole layout rather than inside the results column")
	}
	if banner < sidebar {
		t.Error("the banner comes before the sidebar; it should sit to its right")
	}
	if results < banner {
		t.Error("the listings should begin beneath the banner")
	}
	// The market picker rides on the banner.
	if !strings.Contains(between(body, `class="search-banner"`, "</div>\n\n"), "market-picker") {
		t.Error("Choose your market is not on the banner")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// squash collapses runs of whitespace, so an assertion about a sentence is not
// defeated by wherever the template happened to wrap it.
func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

// between returns the text after the first occurrence of start, up to the next
// occurrence of end. It returns "" when either marker is absent.
func between(s, start, end string) string {
	i := strings.Index(s, start)
	if i < 0 {
		return ""
	}
	rest := s[i+len(start):]
	j := strings.Index(rest, end)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// ---------------------------------------------------------------------------
// Deal type offers exactly the client's three options
// ---------------------------------------------------------------------------

func TestDealTypeHasNoAnyOption(t *testing.T) {
	h := newServer(t)

	// Sidebar and map filter: three radios, and no empty-valued fourth.
	for _, path := range []string{"/search", "/search?view=map"} {
		body := mustGet(t, h, path)
		group := between(body, `class="segmented segmented--full segmented--deal"`, "</div>")
		if group == "" {
			t.Fatalf("%s: deal-type group not found", path)
		}
		if strings.Contains(group, `value=""`) {
			t.Errorf("%s: deal type still offers an empty 'Any' radio", path)
		}
		if strings.Contains(group, ">Any<") {
			t.Errorf("%s: deal type still shows an 'Any' label", path)
		}
		for _, want := range []string{`value="sale"`, `value="rent"`, `value="short_rent"`} {
			if !strings.Contains(group, want) {
				t.Errorf("%s: deal type is missing %s", path, want)
			}
		}
		if n := strings.Count(group, `type="radio"`); n != 3 {
			t.Errorf("%s: deal type has %d radios, want exactly 3", path, n)
		}
	}

	// Homepage: three options, Sell selected, and nothing empty-valued.
	home := mustGet(t, h, "/")
	sel := between(home, `id="hero-deal"`, "</select>")
	if strings.Contains(sel, `value=""`) {
		t.Error("homepage deal type still offers an empty option")
	}
	if !regexp.MustCompile(`value="sale"\s+selected`).MatchString(squash(sel)) {
		t.Error("homepage deal type does not default to Sell")
	}
	if n := strings.Count(sel, "<option"); n != 3 {
		t.Errorf("homepage deal type has %d options, want exactly 3", n)
	}
}

// Sell is the state a visitor arrives in, and Clear all returns to it.
func TestDealTypeDefaultsToSell(t *testing.T) {
	h := newServer(t)

	fresh := between(mustGet(t, h, "/search"),
		`class="segmented segmented--full segmented--deal"`, "</div>")
	if !regexp.MustCompile(`value="sale"[^>]*checked`).MatchString(fresh) {
		t.Error("Sell is not selected on a fresh search")
	}
	if n := strings.Count(fresh, "checked"); n != 1 {
		t.Errorf("%d deal types are checked on a fresh search, want exactly 1", n)
	}

	// An explicit choice still wins, and only that one is checked.
	picked := mustGet(t, h, "/search?deal=rent")
	pickedGroup := between(picked, `class="segmented segmented--full segmented--deal"`, "</div>")
	if !regexp.MustCompile(`value="rent"[^>]*checked`).MatchString(pickedGroup) {
		t.Error("deal=rent did not check the Rent radio")
	}
	if strings.Count(pickedGroup, "checked") != 1 {
		t.Error("more than one deal type is checked")
	}

	// Clear all links back to the bare view, which is Sell again.
	if !strings.Contains(picked, `href="/search?view=grid"`) {
		t.Error("Clear all does not link back to the default search")
	}
	cleared := between(mustGet(t, h, "/search?view=grid"),
		`class="segmented segmented--full segmented--deal"`, "</div>")
	if !regexp.MustCompile(`value="sale"[^>]*checked`).MatchString(cleared) {
		t.Error("Clear all did not return to Sell")
	}
}

// Credit card and Stripe are visually distinct flows, and neither collects
// anything: the card fields are disabled, unnamed, and no provider script is
// loaded. This is the boundary the second milestone moves, not this one.
func TestCardAndStripeAreDistinctAndInert(t *testing.T) {
	body := mustGet(t, newServer(t), "/checkout?package=pk-standard")

	if !strings.Contains(body, `x-show="method === 'card'"`) {
		t.Error("no embedded card-form panel for the Credit card method")
	}
	if !strings.Contains(body, `x-show="method === 'stripe'"`) {
		t.Error("no hosted-checkout panel for the Stripe method")
	}

	// From the card panel's opening to where the Stripe panel begins, so the
	// slice cannot end early on an inner </div>.
	card := between(body, `x-show="method === 'card'"`, `x-show="method === 'stripe'"`)
	for _, id := range []string{"ck-card", "ck-exp", "ck-cvc"} {
		field := between(card, `id="`+id+`"`, ">")
		if field == "" {
			t.Errorf("card panel is missing %s", id)
			continue
		}
		if !strings.Contains(field, "disabled") {
			t.Errorf("%s is not disabled — the mock must not accept card entry", id)
		}
		if strings.Contains(field, "name=") {
			t.Errorf("%s has a name and would be submitted", id)
		}
	}
	// The template wraps this sentence, so compare on collapsed whitespace.
	if !strings.Contains(squash(card), "no card details are collected or sent anywhere") {
		t.Error("the card panel does not say it collects nothing")
	}

	// No provider SDK anywhere.
	for _, script := range []string{"js.stripe.com", "paypal.com/sdk", "nowpayments", "paysera.com"} {
		if strings.Contains(body, script) {
			t.Errorf("checkout loads a provider script: %s", script)
		}
	}
}
