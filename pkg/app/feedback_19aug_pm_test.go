package app_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Tests for the client's 19 August afternoon notes (12:39–13:21).
//
// Seven messages, numbered 91–97 on from the ninety before them, each section
// named after what was asked rather than after the code that answers it. The
// round's first message (11:50, the floating menu on the right) is note 90 and
// was answered in the morning round — its test lives in feedback_19aug_am_test.go
// and is not repeated here.
//
// Two of these are removals, and a removal is the kind of note that is easy to
// answer in one place and miss in three, so those sections check every screen
// the thing appeared on rather than the one in the screenshot.

// postForm performs a form POST and returns the status and body.
func postForm(t *testing.T, h http.Handler, path, form string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

// ---------------------------------------------------------------------------
// 91. "In the add listing price section, if the user has chosen deal type rent
//     or short, then the price must be in case of rent 1000 € / month and in
//     case of short rent 100 € / day."
// ---------------------------------------------------------------------------

func TestThePriceIsQuotedInTheUnitTheDealTypeAsksFor(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/add-listing")

	price := section(t, body, `id="ls-price"`, "</section>", "the price section")

	// The label and the unit beside the field both follow the deal type, so a
	// seller filling in a rent is looking at "1000 € / month" while they type.
	mustContain(t, price, `x-text="priceLabel()"`, "the label must name the unit")
	mustContain(t, price, `class="price-field__unit"`, "the field must carry the unit beside it")
	mustContain(t, price, `x-text="pricePeriod()"`, "…and the period in it comes from the deal type")

	// The deal-type radios at the top of the form are what write it.
	deal := section(t, body, `id="ls-deal"`, "</section>", "the deal-type section")
	mustContain(t, deal, `name="deal" value="rent" x-model="deal"`,
		"the radios must publish the choice the price section reads")

	// The three answers, at the source: month for a rent, day for a short rent,
	// nothing for a sale.
	js := asset(t, jsDir+"/previa.js")
	period := section(t, js, "pricePeriod: function ()", "},", "the period helper")
	mustContain(t, period, "if (this.deal === 'rent') return 'month';", "rent is quoted per month")
	mustContain(t, period, "if (this.deal === 'short_rent') return 'day';", "short rent is quoted per day")
	mustContain(t, period, "return '';", "a sale has no period")

	// And the client's own example figures, so choosing Rent shows 1000 and
	// Short rent 100 rather than a sale price with a period stuck on it.
	mustContain(t, js, "demoPrices: { sale: '429000', rent: '1000', short_rent: '100' }",
		"the prefilled figures are the ones the note names")
	watch := section(t, js, "this.$watch('deal'", "el.value = self.demoPrices[next]", "the price watcher")
	mustContain(t, watch, "if (demo.indexOf(String(el.value).trim()) === -1) return;",
		"a figure the seller typed is never overwritten")
}

// The unit is not only a label on the form: a short rental has to read per day
// wherever its price is printed, or the form and the listing disagree.
func TestAShortRentalIsPricedPerDayEverywhere(t *testing.T) {
	h := newServer(t)

	for _, path := range []string{"/search?deal=short_rent", "/search?deal=short_rent&view=list"} {
		body := mustGet(t, h, path)
		mustContain(t, body, `<span class="pcard__price-period">/day</span>`,
			"a short-rent card must quote a day: "+path)
		mustNotContain(t, body, `<span class="pcard__price-period">/month</span>`,
			"…and never a month: "+path)
	}

	// A long let still reads per month, which is what it always said.
	rent := mustGet(t, h, "/search?deal=rent")
	mustContain(t, rent, `<span class="pcard__price-period">/month</span>`,
		"a rental is still quoted per month")

	// A sale carries no period at all.
	sale := mustGet(t, h, "/search?deal=sale")
	mustNotContain(t, sale, `class="pcard__price-period"`, "a sale price has no period")
}

// ---------------------------------------------------------------------------
// 92. "In the search menu on bottom. The 'Clear all' replace with the red
//     cross, the same as under filters. 'Apply filters' make in small letters
//     and the area of the button make smaller and the background of this button
//     make green #008000 — the same in day and dark mode. Then add there button
//     'save search', if hit that then this search will be saved under user's
//     'saved searches' menu."
// ---------------------------------------------------------------------------

func TestTheFilterPanelFooterIsACrossAGreenApplyAndSaveSearch(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search")

	footer := section(t, body, `class="filter-panel__footer"`, "</form>", "the panel footer")

	// Clear all is the same control as the one in the tag bar above the
	// results — the same class, so the two cannot drift apart — and it keeps
	// the name a bare cross cannot carry.
	mustContain(t, footer, `class="clear-filters"`, "clear all must be the red cross")
	mustContain(t, footer, `aria-label="Clear all filters"`, "…and keep its accessible name")
	mustNotContain(t, footer, ">Clear all<", "the words must be gone from the footer")

	// Apply, in small letters.
	mustContain(t, footer, "apply filters", "the apply button is lower case")
	mustNotContain(t, footer, "Apply filters", "…and no longer capitalised")
	mustContain(t, footer, `class="btn btn--sm btn--apply"`, "…and carries its own modifier")
	mustNotContain(t, footer, `style="flex:1"`, "…and no longer stretches across the footer")

	// Save search, which posts the form it sits in.
	mustContain(t, footer, "save search", "the save button is there, in small letters")
	mustContain(t, footer, `hx-post="/save-search"`, "…and posts to the endpoint that stores one")
	mustContain(t, footer, `hx-include="#filter-form"`,
		"…including the filters, or it would save an empty search")

	// The green: the client's literal value, declared once so it is the same in
	// both themes rather than a token that changes with them.
	components := cssCode(t, "components.css")
	apply := section(t, components, "\n.btn--apply {", "}", "the apply button")
	mustContain(t, apply, "background: #008000;", "the button must be the client's green")
	mustContain(t, apply, "color: #fff;", "…with white on it")
	mustNotContain(t, apply, "var(--success)", "…not the theme's success colour, which differs by theme")

	// Smaller: shorter than the 40px .btn--sm it would otherwise take.
	size := section(t, components, ".btn--sm.btn--apply {", "}", "the apply button's size")
	mustContain(t, size, "min-height: 34px;", "the button's area is smaller than a standard small button")
}

// The button does what it says: the search becomes a row under Saved searches.
func TestSavingASearchAddsItToSavedSearches(t *testing.T) {
	h := newServer(t)

	before := strings.Count(mustGet(t, h, "/saved-searches"), `class="card__body"`)

	code, body := postForm(t, h, "/save-search",
		"deal=rent&location=Tallinn&price_max=1500&compact=1")
	if code != http.StatusOK {
		t.Fatalf("POST /save-search = %d, want 200", code)
	}
	// The panel's footer is 40px tall, so its confirmation is a line, and it
	// points at where the search went.
	mustContain(t, body, `class="save-search-line"`, "the panel gets the compact confirmation")
	mustContain(t, body, `href="/saved-searches"`, "…which links to the menu the note names")

	after := mustGet(t, h, "/saved-searches")
	if n := strings.Count(after, `class="card__body"`); n != before+1 {
		t.Errorf("saved searches went from %d rows to %d, want one more", before, n)
	}

	// Described in the words the tag bar above the results uses, and replayable:
	// the row's link carries the filters that were posted.
	mustContain(t, after, "Rent · Estonia", "the row is named after the filters it holds")
	mustContain(t, after, "Rent · Estonia · Tallinn · Up to €1 500",
		"…and summarised by every one of them, in the tag bar's own words")
	mustContain(t, after, "location=Tallinn", "…and can be run again")

	// The full alert is what the tag bar's own button gets.
	_, full := postForm(t, h, "/save-search", "deal=sale&country=EE")
	mustContain(t, full, `class="alert alert--success"`, "the results header gets the full confirmation")
	mustContain(t, full, "Saved: Sell · Estonia", "…which names what it saved")
}

// A stored query has to survive being put in an href, which it did not: written
// straight into `/search?{{ .Query }}` html/template escaped it as one
// parameter, so every saved search ran unfiltered.
func TestASavedSearchLinkIsARealQuery(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/saved-searches")

	mustContain(t, body, `href="/search?bedrooms=2&amp;city=Tallinn&amp;country=EE&amp;deal=sale`,
		"a saved search must link to a query with separators in it")
	mustNotContain(t, body, "deal%3dsale", "the query must not be escaped as a single value")
}

// ---------------------------------------------------------------------------
// 93. "Under the seller's profile remove the stars and evaluations, and remove
//     the 'Promoted' label."
// ---------------------------------------------------------------------------

func TestBrokerCardsCarryNoRatingAndNoPromotedLabel(t *testing.T) {
	h := newServer(t)

	// The directory, the market strip on the homepage and the map's results
	// column all draw the same card, so all three are checked.
	for _, path := range []string{"/brokers", "/brokers?location=Tallinn", "/", "/search?brokers=1&view=map"} {
		body := mustGet(t, h, path)
		mustNotContain(t, body, "bcard__rating", "no rating on a broker card: "+path)
		mustNotContain(t, body, "out of 5 from", "…and no review count either: "+path)
		mustNotContain(t, body, "Promoted", "and no Promoted badge: "+path)
	}

	// The card itself, at the source — a rendered page cannot prove the row is
	// gone rather than merely empty for these brokers.
	card := asset(t, componentDir+"/property-card.html")
	mustNotContain(t, card, "bcard__meta", "the rating row is gone from the component")
	mustNotContain(t, card, "$b.Rating", "…and so is the figure it drew")
	mustNotContain(t, card, "$b.IsPromoted", "…and the paid-placement label with it")

	// What the card keeps: how many listings this broker has, which the site
	// counts rather than takes on trust.
	mustContain(t, card, "bcard__count", "the listing count stays")

	// And the same removal on the listing's own seller box.
	detail := mustGet(t, h, "/search")
	slug := section(t, detail, `class="pcard__title"`, "</a>", "the first card's link")
	href := between(slug, `href="`, `"`)
	listing := mustGet(t, h, href)
	seller := section(t, listing, `class="detail-aside"`, "Send an enquiry", "the seller box")
	mustNotContain(t, seller, "reviews)", "the seller box states no rating")
	mustContain(t, seller, "listings</span>", "…but still says how many listings the broker has")
}

// ---------------------------------------------------------------------------
// 94. "Under the broker / seller profile on the right where is his phone
//     number, if the user has marked any social apps, then these will be listed
//     there as well. The social apps are on the bottom of every ad, this is
//     good. Now add them under the seller's phone as well. And they are
//     everywhere hyperlinks, if click on them they will open with this seller's
//     contact."
// ---------------------------------------------------------------------------

func TestABrokersChatAppsSitUnderTheirPhoneNumber(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/broker/kadri-tamm")

	actions := section(t, body, `class="profile-head__actions"`, `class="container section`,
		"the profile head's contact block")

	// The phone is still the control above them.
	mustContain(t, actions, `href="tel:&#43;372%205123%204471"`, "the phone button stays")
	// And the apps this broker ticked are under it, as links that open a chat.
	mustContain(t, actions, `class="msg-links`, "the chat apps are in the same block")
	mustContain(t, actions, `href="https://wa.me/37251234471"`,
		"WhatsApp opens a chat with this broker's own number")
	mustContain(t, actions, `target="_blank"`, "…in a new tab, like every other messenger link")

	// The same component the listings use, so the marks and their order are
	// identical in both places.
	profile := asset(t, "../../web/templates/pages/broker-profile.html")
	mustContain(t, profile, `{{ template "messenger-links" dict`,
		"the profile reuses the listing's own component")

	// A broker who has ticked nothing gets no row rather than an empty one.
	quiet := mustGet(t, h, "/broker/jonas-weber")
	quietActions := section(t, quiet, `class="profile-head__actions"`, `class="container section`,
		"a profile with no apps")
	mustNotContain(t, quietActions, `class="msg-links`, "no row without apps to put in it")
}

// ---------------------------------------------------------------------------
// 95. "In the brokers profile remove the ratings, years active and completed
//     sales — as there is no way we can verify it. The active listings will
//     stay. And the title 'Current listings' there replace with 'Active
//     Listings'."
// ---------------------------------------------------------------------------

func TestTheBrokerProfileKeepsOnlyTheFigureItCanCheck(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/broker/liis-kask")

	aside := section(t, body, `class="detail-aside"`, "Active in", "the profile's side panel")
	mustNotContain(t, aside, "Rating", "the rating is gone")
	mustNotContain(t, aside, "Completed sales", "the sales history is gone")
	mustNotContain(t, aside, "Years active", "the career length is gone")
	mustContain(t, aside, "Active listings", "the one countable figure stays")
	mustNotContain(t, aside, "Track record", "…and the panel is named after what is left in it")

	// The promoted badge went with the rating on the profile too — same note as
	// the cards, one message earlier.
	head := section(t, body, `class="profile-head"`, "</div>\n  </div>", "the profile head")
	mustNotContain(t, head, "Promoted", "no paid-placement label on the profile either")

	// The listings section is titled the client's way.
	mustContain(t, body, `<h2 class="section__title">Active Listings</h2>`,
		"the listings section carries the client's title")
	mustNotContain(t, body, "Current listings", "the old title is gone")
}

// The figure that survived has to be true, which the seeded one was not: it
// said 9 beside a page showing 4.
func TestABrokersListingCountIsCounted(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/broker/liis-kask")

	// The panel and the section under it are two readings of the same number.
	panel := section(t, body, "Active listings", "</dl>", "the active-listings panel")
	count := strings.TrimSpace(between(panel, `<dd class="numeric">`, "</dd>"))
	mustContain(t, body, count+" properties on the market",
		"the panel's figure must be the number of listings the page shows")
}

// ---------------------------------------------------------------------------
// 96. "Under broker name is 'About' but at the moment under user profile there
//     is no place the user can edit this text — create it."
// ---------------------------------------------------------------------------

func TestTheAboutTextIsEditableOnTheProfile(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/settings")

	field := section(t, body, `for="p-about"`, "</div>", "the About field")
	mustContain(t, field, "About you", "the field is labelled for what it is")
	mustContain(t, field, `name="about"`, "…and posts under its own name")
	mustContain(t, field, "<textarea", "a paragraph needs a textarea, not a single-line input")

	// It opens with what the profile currently says, or it is a field that
	// cannot edit anything — only replace it.
	mustContain(t, field, "Anna sells apartments and houses",
		"the field must be filled with the text it edits")

	// And it says where the text goes, in the same shape as the other fields
	// on this screen that publish something.
	mustContain(t, field, "Shown on your public profile under",
		"the hint says where the text appears")
}

// ---------------------------------------------------------------------------
// 97. "Under user's account this 'delete account' make with small letter and
//     the background of this red button #FF0000."
// ---------------------------------------------------------------------------

func TestDeleteAccountIsLowerCaseAndPureRed(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/settings")

	zone := section(t, body, `class="card danger-zone"`, "</template>", "the delete-account block")
	mustContain(t, zone, ">delete account<", "the button is lower case")
	mustNotContain(t, zone, ">Delete account<", "…including the one in the confirmation dialog")

	pages := cssCode(t, "pages.css")
	rule := section(t, pages, ".danger-zone .btn--danger {", "}", "the delete button")
	mustContain(t, rule, "background: #FF0000;", "the client's literal red")
	mustContain(t, rule, "border-color: #FF0000;", "…on the border as well, or the button reads outlined")
	mustNotContain(t, rule, "var(--error)", "not the theme's error colour, which is a different red per theme")
}
