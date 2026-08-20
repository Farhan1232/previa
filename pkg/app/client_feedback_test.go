package app_test

// Tests for the round of changes the client sent on 13 August 2026.
//
// One test per note, named after what the client asked for rather than after
// the code that implements it, so a future change that quietly undoes one of
// them fails with a sentence explaining what was wanted and why.
//
// Helpers (mustGet, mustContain, section, asset, cssDir, jsDir) live in
// app_test.go and update_test.go.

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 1. Footer: no CTA band, five links spread across the width, FAQ its own page
// ---------------------------------------------------------------------------

func TestFooterKeepsOnlyTheFiveNamedLinks(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/")

	// The closing call-to-action panel is gone from the homepage entirely.
	mustNotContain(t, body, "Have a property to sell",
		"the homepage call-to-action band was removed at the client's request")

	foot := section(t, body, `<footer class="site-footer">`, "</footer>", "footer")

	// The column headings went with the columns.
	for _, gone := range []string{"Sell &amp; rent out", "Company", "footer__col-title"} {
		mustNotContain(t, foot, gone, "the footer column headings were removed")
	}

	// Exactly the five links the client listed, in the link row.
	row := section(t, foot, `class="footer__main"`, "</nav>", "footer link row")
	for _, want := range []string{
		`href="/pricing">Listing packages`,
		`href="/faq">FAQ`,
		`href="/about">About Previa`,
		`href="/help">Help &amp; contact`,
		`href="/advertising">Advertise with us`,
	} {
		mustContain(t, row, want, "the footer must keep the links the client named")
	}
	if n := strings.Count(row, "<a "); n != 5 {
		t.Errorf("the footer link row has %d links, want exactly the 5 the client listed", n)
	}

	// Everything else the client asked to drop.
	for _, gone := range []string{"Add a listing", "For agencies", "Your dashboard",
		"Manage listings", "Billing &amp; invoices", "Terms of service", "Privacy policy",
		"Cookie preferences"} {
		mustNotContain(t, row, gone, "the footer link row must carry only the five named links")
	}

	// The legal strip below the separator is untouched, and is where Terms,
	// Privacy and Cookies already live — so they are not duplicated above.
	legal := section(t, foot, `class="footer__legal"`, "</nav>", "legal strip")
	for _, want := range []string{`href="/terms"`, `href="/privacy"`, `href="/cookies"`} {
		mustContain(t, legal, want, "the legal strip under the separator stays as it was")
	}

	// Spread across the width rather than stacked in a column, which is what
	// makes the footer shorter.
	rule := section(t, asset(t, cssDir+"/layout.css"), ".footer__main {", "}", "footer row rule")
	mustContain(t, rule, "grid-template-columns: repeat(5,",
		"the footer links must divide the width evenly instead of stacking")
}

// "Common questions" became FAQ and moved off the packages page onto one of
// its own, which is what the footer links to.
func TestFAQIsItsOwnPage(t *testing.T) {
	h := newServer(t)

	faq := mustGet(t, h, "/faq")
	mustContain(t, faq, `class="page-head__title">FAQ<`, "the FAQ page must be titled FAQ")
	mustContain(t, faq, "How long does it take to publish a listing?",
		"the FAQ page must carry the questions that used to be on /pricing")

	pricing := mustGet(t, h, "/pricing")
	mustNotContain(t, pricing, "Common questions",
		"the accordion moved to /faq; leaving a copy on /pricing means two sets of answers to keep current")
}

// ---------------------------------------------------------------------------
// 2. The sidebar drawer's day/dark row
// ---------------------------------------------------------------------------

func TestDrawerThemeRowUsesTheHeaderButton(t *testing.T) {
	body := mustGet(t, newServer(t), "/")

	row := section(t, body, `class="drawer__row" x-data @click="$store.theme.toggle()"`,
		"</button>", "drawer theme row")

	// The whole row is the control, so a click anywhere along it toggles.
	if !strings.HasPrefix(strings.TrimSpace(row), `x-data @click`) &&
		!strings.Contains(body, `<button type="button" class="drawer__row" x-data @click="$store.theme.toggle()"`) {
		t.Error("the day/dark row must be a button, so clicking anywhere in it switches the theme")
	}
	// The header's own button, not the green pill switch it replaced.
	mustContain(t, row, `class="icon-btn drawer__row-toggle"`,
		"the row must end in the same theme button the header uses")
	mustNotContain(t, row, `class="switch"`,
		"the pill switch was replaced by the header's theme button")

	// Hover feedback, like the language and notification rows either side.
	layout := asset(t, cssDir+"/layout.css")
	mustContain(t, layout, "button.drawer__row:hover { background: var(--surface-3); }",
		"the theme row must react to the pointer the way its neighbours do")
	mustContain(t, layout, "button.drawer__row:hover .drawer__row-toggle { color: var(--text); }",
		"the borrowed button must lift with the row rather than staying inert")
}

// ---------------------------------------------------------------------------
// 3–4. The account screens: compact head, collapsible menu
// ---------------------------------------------------------------------------

func TestAccountHeadIsCompact(t *testing.T) {
	body := mustGet(t, newServer(t), "/dashboard")

	mustContain(t, body, `class="page-head page-head--compact"`,
		"account screens must use the compact head, so the breadcrumb sits under the site header")
	mustNotContain(t, body, "Everything you're selling, saving and searching",
		"the dashboard subtitle was removed at the client's request")
	mustContain(t, body, `class="container section section--after-head"`,
		"the panels must come up into the gap the removed subtitle left")

	layout := asset(t, cssDir+"/layout.css")
	mustContain(t, layout, ".page-head--compact .page-head__title { font-size: var(--fs-h3); }",
		"the welcome heading must be smaller than a display title")
	mustContain(t, layout, ".section--after-head { padding-top: var(--sp-6); }",
		"a full section gap under the head is what the client asked to remove")
}

func TestAccountMenuCollapsesLikeTheSearchFilter(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/settings")

	// Same component as the search page's filter panel, so the two behave
	// identically — which is exactly what the client asked for.
	mustContain(t, body, `x-data="previaPanel('previa-account-nav')"`,
		"the account menu must use the shared collapsible-panel component")
	mustContain(t, body, `class="shell shell--nav"`,
		"the account shell must opt into the collapsible layout")

	// Both ways back in: the rail on a wide screen, the edge tab on a narrow one.
	mustContain(t, body, `class="filter-rail"`, "a collapsed menu must leave a rail to reopen it")
	mustContain(t, body, `class="filter-fab"`, "the overlay drawer needs a visible pull-out tab")
	mustContain(t, body, `class="filter-backdrop"`, "the overlay drawer needs a backdrop to dismiss it")

	// Renamed, matching the header menu and the mobile drawer.
	mustContain(t, body, `href="/settings"`, "the settings link must survive the rename")
	mustNotContain(t, body, "Profile and settings",
		`"Profile and settings" was shortened to "Settings"`)

	// Sections sit closer together.
	layout := asset(t, cssDir+"/layout.css")
	mustContain(t, layout, ".side-nav__group + .side-nav__group { margin-top: var(--sp-3); }",
		"the client asked for the menu sections to be more compact")

	// The admin shell shares .side-nav and must not have been dragged along.
	admin := mustGet(t, h, "/admin")
	mustNotContain(t, admin, "shell--nav",
		"only the account shell was asked to collapse; the admin one keeps its static sidebar")
}

// ---------------------------------------------------------------------------
// 5. Add listing: the living-area error clears, the progress rail slides away
// ---------------------------------------------------------------------------

// The living area no longer has an error to clear — the client made it optional
// on 19 August, and TestTheAreaFieldsAreOptional covers that. What this test
// still pins is the machinery underneath, which is what made the original bug
// fixable: the state is computed from the fields, never rendered into the page.
func TestRequiredFieldStateIsLiveNotHardCoded(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")

	mustNotContain(t, body, `aria-invalid="true"`,
		"a hard-coded invalid state cannot clear when the field is filled in")
	mustContain(t, body, `@input="touch(); recheck()"`,
		"typing must re-run the required-field check")
	mustContain(t, body, "data-required",
		"the check needs at least one field to check")

	// The waypoint rail follows the same state, so a "!" clears with it.
	js := asset(t, jsDir+"/previa.js")
	mustContain(t, js, "this.blank[key] === undefined ? base === 'error' : this.blank[key] > 0",
		"the waypoint marker must follow the live check, not the value the server rendered")
	mustContain(t, js, "if (String(el.value).trim() === '') next[key] += 1;",
		"…and the live check is a count of the empty required fields in each section")
}

func TestProgressRailSlidesOffTheRightEdge(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")

	mustContain(t, body, `x-data="previaPanel('previa-listing-rail')"`,
		"the progress rail must use the same collapsible component as the search filter")
	mustContain(t, body, `class="filter-rail filter-rail--right"`,
		"a collapsed rail must leave a way to reopen it")
	mustContain(t, body, `class="filter-fab filter-fab--right"`,
		"the client asked for the pull-out button to stay visible when the rail slides away")

	// Hinged on the right, not the left.
	pages := asset(t, cssDir+"/pages.css")
	rule := section(t, pages, ".wizard .stepper {", "}", "narrow-screen rail rule")
	mustContain(t, rule, "right: 0;", "the rail must sit against the right edge")
	mustContain(t, rule, "transform: translateX(100%);", "it must slide off to the right")
	mustContain(t, rule, "border-radius: var(--r-lg) 0 0 var(--r-lg);",
		"the corners that meet the page are the left pair on a right-hinged panel")

	// The stand-in <select> is gone with the rail it stood in for.
	mustNotContain(t, body, "listing-nav-compact",
		"the rail itself is reachable at every width now, so the substitute picker went")
}

// ---------------------------------------------------------------------------
// 6. The public-location box, and 8. the comparable-price note
// ---------------------------------------------------------------------------

func TestPublicLocationSitsUnderTheMap(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")

	mustContain(t, body,
		"This is the location shown on the public listing. The Google maps\n              data below is not displayed publicly.",
		"the client dictated this wording")
	mustNotContain(t, body, "The coordinates above stay as they are",
		"the old wording described the box's old position")

	// Above the geocoder output, not below it.
	editable := strings.Index(body, "field--editable-location")
	fromMap := strings.Index(body, "From the map")
	if editable < 0 || fromMap < 0 {
		t.Fatal("the location section is missing either the editable box or the read-only fields")
	}
	if editable > fromMap {
		t.Error("the public-location box must sit straight below the map, above the read-only fields")
	}
}

func TestComparablePriceNoteIsGone(t *testing.T) {
	mustNotContain(t, mustGet(t, newServer(t), "/add-listing"),
		"Comparable renovated apartments",
		"the comparable-prices note was removed at the client's request")
}

// ---------------------------------------------------------------------------
// 7. Photo tile controls in dark mode
// ---------------------------------------------------------------------------

func TestPhotoTileButtonsAreLegibleInDarkMode(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	// An explicit dark palette, rather than a translucent white disc that
	// disappears against a bright photograph.
	mustContain(t, pages, "[data-theme='dark'] .photo-thumb__remove,",
		"the tile controls need a dark-mode treatment of their own")
	mustContain(t, pages, ".photo-thumb__move button:hover:not(:disabled) {",
		"the client asked for the colour to change on hover")
	mustContain(t, pages, "[data-theme='dark'] .photo-thumb__move button:hover:not(:disabled) {",
		"the hover state has to work in dark mode too — that is where the problem was")
	mustContain(t, pages, ".photo-thumb__remove:hover { background: var(--error);",
		"the remove button needs a hover state as well")
}

// ---------------------------------------------------------------------------
// 9. Messenger apps
// ---------------------------------------------------------------------------

func TestMessengerAppsAreOfferedAndLinked(t *testing.T) {
	h := newServer(t)

	// The five apps, as togglable tiles beside the phone field.
	form := mustGet(t, h, "/add-listing")
	picker := section(t, form, `class="msg-picker"`, "</div>", "messenger picker")
	for _, app := range []string{"whatsapp", "telegram", "viber", "signal", "teams"} {
		mustContain(t, picker, `value="`+app+`"`, "the client named all five apps")
	}
	// Telegram and Signal carry their own link fields, as the client asked.
	mustContain(t, form, `id="w-telegram"`, "Telegram needs its own link field")
	mustContain(t, form, `id="w-signal"`, "Signal needs its own link field")

	// On the front page the icons are live links into a chat with that seller.
	home := mustGet(t, h, "/")
	mustContain(t, home, `href="https://wa.me/`, "WhatsApp icons must open a chat with the seller")
	mustContain(t, home, `class="msg-link"`, "the icons on a card must be links")

	// Both Telegram link forms are supported: a username and a phone number.
	search := mustGet(t, h, "/search")
	if !strings.Contains(search, `href="https://t.me/previa_seller_`) {
		t.Error("the username form of a Telegram link is missing")
	}
	if !strings.Contains(search, `href="https://t.me/&#43;`) {
		t.Error("the phone form of a Telegram link is missing")
	}
	for _, want := range []string{`href="viber://chat?number=`, `href="https://signal.me/#p/`,
		`href="https://teams.microsoft.com/l/chat/0/0?users=`} {
		if !strings.Contains(search, want) {
			t.Errorf("missing deep link form: %s", want)
		}
	}
}

// ---------------------------------------------------------------------------
// 10–12. Promotion is a paid add-on, priced by the day
// ---------------------------------------------------------------------------

func TestPromotionIsPricedByTheDay(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/add-listing")

	// The client's wording, and their price scale.
	mustContain(t, body, "Featured listing has golden frame around it",
		"the client dictated the featured-listing wording")
	mustContain(t, body, "Bump on top of search results",
		`"Bump to the top of search results weekly" was shortened`)
	mustContain(t, body, "By buying this option your listing will be bumped on top of searched listings.",
		"the client dictated the bump explanation")
	mustNotContain(t, body, "Highlight in the weekly market newsletter",
		"the newsletter option was removed for now")

	promo := section(t, body, `class="promo-picker"`, `class="publish-done"`, "promotion picker")
	for _, want := range []string{
		"1 day", "€3", "2 days", "€6", "3 days", "€8",
	} {
		mustContain(t, promo, want, "the client gave this price scale explicitly")
	}
	mustContain(t, promo, `name="promo_featured_days"`,
		"ticking the box must open a picker for how many days it runs")
	mustContain(t, promo, `name="promo_bump_days"`,
		"the bump option needs the same day picker")

	// And the same add-ons, buyable later from the seller's own listings.
	listings := mustGet(t, h, "/my-listings")
	mustContain(t, listings, `class="promo-picker"`,
		"the client asked to be able to buy more days later from the profile")
	mustContain(t, listings, "Promote this listing",
		"a listing that is already online needs a way into the promotion picker")
}

// ---------------------------------------------------------------------------
// 13–15. Checkout copy and spacing
// ---------------------------------------------------------------------------

func TestCheckoutCopyAndSpacing(t *testing.T) {
	body := mustGet(t, newServer(t), "/checkout")

	mustContain(t, body, "Pay directly with many bank links around Europe.",
		"Paysera is not limited to the Baltics")
	mustContain(t, body, "Bitcoin, Ethereum, stablecoins and many others through NOWPayments.",
		"NOWPayments settles more than three coins")
	mustNotContain(t, body, "Estonian, Latvian or Lithuanian bank account",
		"the narrower Paysera wording was replaced")

	// The note this pinned was about vertical space — the page starting a full
	// section below the header — and section--after-head is what answers it.
	// The reading-width container beside it went on 19 August, when the payment
	// methods were asked for in two columns and the column they sit in was too
	// narrow to hold two; see TestPaymentMethodsAreInTwoColumns.
	mustContain(t, body, `class="container section section--after-head"`,
		"the client asked for the checkout content to start closer to the header")
}

// ---------------------------------------------------------------------------
// 16. The homepage hero search
// ---------------------------------------------------------------------------

func TestHeroSearchIsNarrowerAndLowercase(t *testing.T) {
	body := mustGet(t, newServer(t), "/")
	form := section(t, body, `<form class="searchbox"`, "</form>", "homepage search form")

	mustContain(t, form, "advanced filters", "the client asked for a lowercase label")
	mustContain(t, form, "search properties</span>", "the client asked for a lowercase label")
	mustNotContain(t, form, "Advanced filters", "the label must not be capitalised")
	mustNotContain(t, form, "Search properties", "the label must not be capitalised")

	// The panel is capped by an absolute ceiling rather than by a percentage.
	//
	// It used to be `max-width: 70%` inside a min-width: 1200px query, which the
	// client caught out: below 1200px the rule stopped applying and the panel
	// sprang from 851px to the full content width, so narrowing the window made
	// it bigger. A ceiling can only ever let it shrink.
	pages := asset(t, cssDir+"/pages.css")
	mustContain(t, pages, "max-width: min(100%, var(--hero-search-max));",
		"the hero panel must be capped by an absolute ceiling, not a percentage")
	mustNotContain(t, pages, "max-width: 70%",
		"the percentage rule let the panel grow as the window narrowed")

	// The ceiling is 70% of the widest content box, so a full screen looks
	// exactly as it did before.
	mustContain(t, asset(t, cssDir+"/tokens.css"), "--hero-search-max: 851px;",
		"the ceiling must stay at 70% of the 1216px content box")
}

// ---------------------------------------------------------------------------
// 18. The filter drawer's corners
// ---------------------------------------------------------------------------

func TestFilterDrawerHasRoundedRightCorners(t *testing.T) {
	// The search page's drawer only exists below 1024px, so its rule lives
	// inside that media query rather than beside the docked-sidebar rule.
	narrow := section(t, asset(t, cssDir+"/layout.css"),
		"@media (max-width: 1024px) {\n  .search-layout {", "\n}", "narrow-screen search rules")
	mustContain(t, narrow, "border-radius: 0 var(--r-lg) var(--r-lg) 0;",
		"the search page's drawer must be rounded on the right, top and bottom")

	overlay := section(t, asset(t, cssDir+"/pages.css"),
		".filter-panel--overlay {", "}", "the map page's drawer")
	mustContain(t, overlay, "border-radius: 0 var(--r-lg) var(--r-lg) 0;",
		"the map page's drawer must be rounded on the right, top and bottom")
}
