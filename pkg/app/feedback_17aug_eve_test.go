package app_test

import (
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// Tests for the client's 17 August evening corrections.
//
// Nine notes, nine sections, each named after what was asked rather than after
// the code that answers it. Anything the server renders is asserted against a
// real response; what only exists once a browser has laid the page out is
// asserted at the level of the rule that produces it, and was checked in a real
// headless Chrome as well.

// ---------------------------------------------------------------------------
// 18. "in the single ad page this address in the upper part 'Avenida de Sintra
//     310, Cascais, Lisbon, Portugal' this make blue (no underline) as it is
//     link, and if click on it, then a googlemaps will open — not our
//     googlemaps. Rather a googlemaps for navigation."
// ---------------------------------------------------------------------------

func TestListingAddressIsANavigationLink(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, firstListingPath(t, h))

	// The address is inside the listing's own header block, above the price.
	head := section(t, body, `<h1 class="display display--h2">`, `class="fact-grid"`,
		"the listing header")
	mustContain(t, head, `class="address-link"`, "the address itself must be the link")

	href := hrefOf(t, head, "address-link")
	if !strings.HasPrefix(href, "https://www.google.com/maps/dir/?api=1&destination=") {
		t.Fatalf("the address must open Google Maps navigation, got %q", href)
	}

	// A destination, not a place to look at, and the destination is the address
	// the page prints: the client's "this address is there as final
	// destination".
	dest := strings.TrimPrefix(href, "https://www.google.com/maps/dir/?api=1&destination=")
	got, err := url.QueryUnescape(dest)
	if err != nil {
		t.Fatalf("destination is not a decodable value: %v", err)
	}
	printed := strings.Join(strings.Fields(
		regexp.MustCompile(`<[^>]*>`).ReplaceAllString(
			section(t, head, `class="address-link"`, "</a>", "the address link"), " ")), " ")
	printed = strings.TrimSpace(strings.TrimPrefix(printed, `href=`))
	if !strings.HasSuffix(printed, got) {
		t.Errorf("the link points at %q but the page prints %q", got, printed)
	}
	if strings.Contains(head, "maps/search") {
		t.Error("the address must not fall back to a look-at-this-pin link")
	}

	// A new tab, so a visitor who came to read the listing still has it.
	mustContain(t, head, `target="_blank"`, "Maps opens beside the listing, not over it")
	mustContain(t, head, `rel="noopener"`, "…and without handing the new tab this one")
}

func TestAddressLinkIsBlueWithoutAnUnderline(t *testing.T) {
	components := asset(t, cssDir+"/components.css")

	// The blue is #0000FF since the client named it on 18 August — see
	// TestTheAddressLinkIsPureBlue. It was a per-theme token until then, which
	// is why the two assertions about --link that used to be here are gone:
	// the client asked for one value, in those words, for this one link.
	rule := section(t, components, ".address-link {", "}", "the address link")
	mustContain(t, rule, "color: #0000FF", "the client asked for blue")
	mustContain(t, rule, "text-decoration: none", "…and for no underline")

	hover := section(t, components, ".address-link:hover {", "}", "the hover state")
	mustContain(t, hover, "text-decoration: none", "the underline must not come back on hover")

	// The tokens themselves stay: other links on the site still use them.
	tokens := asset(t, cssDir+"/tokens.css")
	mustContain(t, tokens, "--link:", "the shared link colour must still exist for everything else")
}

// ---------------------------------------------------------------------------
// 19. "below this map window is big 'open in googlemaps' what opens the
//     googlemaps for navigation, like it is now — so this is the same thing as
//     the blue link in upper part I was just telling about"
// ---------------------------------------------------------------------------

func TestOpenInGoogleMapsIsTheSameNavigationLink(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, firstListingPath(t, h))

	// Previa's own map stays where it is, above the button.
	location := section(t, body, `id="location"`, "Report listing", "the location section")
	mustContain(t, location, `class="detail-map"`, "our own map must still be on the page")
	mustContain(t, location, "Open in Google Maps", "…with the button underneath it")

	// One URL, two places. The client's note is explicit that the button and
	// the address line are "the same thing".
	head := section(t, body, `<h1 class="display display--h2">`, `class="fact-grid"`,
		"the listing header")
	if got, want := hrefOf(t, location, "btn btn--quiet btn--sm"), hrefOf(t, head, "address-link"); got != want {
		t.Errorf("the button opens %q, the address opens %q — they must match", got, want)
	}
}

func TestNoLinkStillOpensAPlainMapSearch(t *testing.T) {
	// The old helper produced a maps/search URL, which opens Maps looking at a
	// pin with no route. Nothing may reach for it again.
	funcs := asset(t, "../view/funcs.go")
	mustNotContain(t, funcs, "maps/search", "every Maps link on the site is now a directions link")
	mustContain(t, funcs, "https://www.google.com/maps/dir/?api=1&destination=",
		"…built from one place")
}

// ---------------------------------------------------------------------------
// 20. "this contact email make info@previa.estate and this text 'Monday to
//     Friday, 9:00–17:00 CET Typical reply time: under one working day.' remove"
// ---------------------------------------------------------------------------

func TestHelpPageContactDetails(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/help")

	card := section(t, body, "Contact details", "For brokers and agencies", "the contact card")
	mustContain(t, card, "info@previa.estate", "the address the client asked for")
	mustContain(t, card, `href="mailto:info@previa.estate"`, "…and it must be clickable")
	mustNotContain(t, card, "hello@previa.estate", "the old address must be gone")
	mustNotContain(t, card, "Monday to Friday", "the office hours were removed at the client's request")
	mustNotContain(t, card, "Typical reply time", "…as was the reply-time promise")
}

func TestTheOldContactAddressIsGoneEverywhere(t *testing.T) {
	h := newServer(t)
	// Including the admin screen that configures it: two different support
	// addresses on one site is the bug this note is really about.
	mustContain(t, mustGet(t, h, "/admin/settings"), "info@previa.estate",
		"the admin support address must match the one the site prints")
	for _, path := range []string{"/help", "/admin/settings", "/faq", "/about"} {
		mustNotContain(t, mustGet(t, h, path), "hello@previa.estate", "on "+path)
	}
}

// ---------------------------------------------------------------------------
// 21. "The footer cut off where the page content ends, so make it the same wide
//     as the page content, and footer upper corners are rounded"
// ---------------------------------------------------------------------------

func TestFooterIsAsWideAsThePageContent(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")
	rule := section(t, layout, ".site-footer {", "}", "the footer")

	// The content box of the container: --container less its two gutters. That
	// is the box the cards and headings above the footer sit in, so the
	// footer's edges line up with theirs.
	mustContain(t, rule, "width: min(100% - var(--gutter) * 2, calc(var(--container) - var(--gutter) * 2))",
		"the footer must stop where the page content stops")
	mustContain(t, rule, "margin-inline: auto", "…centred on the page")

	// Pushed to the bottom of a short page by the ground it now stands on
	// rather than by the footer itself — see the 17 August evening note about
	// a closing band running behind the footer, which is what .page__foot is.
	mustContain(t, section(t, layout, ".page__foot {", "}", "the footer's ground"),
		"margin-top: auto", "the footer must still sit at the bottom of a short page")

	// Upper corners only: the lower two are at the end of the document.
	mustContain(t, rule, "border-radius: var(--r-xl) var(--r-xl) 0 0",
		"the client asked for rounded upper corners")
}

// ---------------------------------------------------------------------------
// 22. "in the header before 'Articles' make button 'Listings' which redirects to
//     the main listings page where the big filter-menu is"
// ---------------------------------------------------------------------------

func TestHeaderHasListingsBeforeArticles(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/")

	nav := section(t, body, `<nav class="header__nav"`, "</nav>", "the header nav")
	mustContain(t, nav, `href="/search"`, "Listings must go to the search screen")
	mustContain(t, nav, ">Listings<", "…and be labelled Listings")

	// Order: before Articles, as asked.
	if i, j := strings.Index(nav, ">Listings<"), strings.Index(nav, ">Articles<"); i < 0 || j < 0 || i > j {
		t.Errorf("Listings must come before Articles in the header, got %q", nav)
	}
}

func TestListingsPointsAtThePageWithTheFilterMenu(t *testing.T) {
	h := newServer(t)
	// The destination is the results screen with the filter sidebar on it —
	// "the main listings page where the big filter-menu is".
	search := mustGet(t, h, "/search")
	mustContain(t, search, `class="filter-panel"`, "the target page must carry the filter menu")
	// Lower case since the 19 August note that also turned this button green
	// and narrowed it: "'Apply filters' make in small letters and the area of
	// the button make smaller". What this test is about is that the page the
	// header points at is the one with the filter menu on it, so it follows the
	// label rather than pinning the old one.
	mustContain(t, search, "apply filters", "…and its apply control")

	// And it marks itself as the current section while a visitor is there,
	// including on a listing opened from it and on a deal-filtered search.
	for _, path := range []string{"/search", "/search?deal=sale", "/search?deal=rent"} {
		nav := section(t, mustGet(t, h, path), `<nav class="header__nav"`, "</nav>", "the header nav")
		item := section(t, nav, `href="/search"`, "</a>", "the Listings item")
		mustContain(t, item, `aria-current="page"`, "Listings must be current on "+path)
	}
}

func TestListingsAlsoReachableOnAPhone(t *testing.T) {
	// The header bar is hidden below 1080px, so the drawer is where a phone
	// reaches Listings. Same position as in the bar: first, above Articles.
	h := newServer(t)
	body := mustGet(t, h, "/")
	browse := section(t, body, `aria-label="Browse"`, "</nav>", "the drawer's browse group")
	if i, j := strings.Index(browse, "Listings"), strings.Index(browse, "Articles"); i < 0 || j < 0 || i > j {
		t.Errorf("the drawer needs Listings above Articles, got %q", browse)
	}
	mustContain(t, browse, `href="/search"`, "…pointing at the same page as the header's")
}

// ---------------------------------------------------------------------------
// 23. "in listing single page the broker/seller on the right, if click to the
//     seller image or name, then it redirects to this broker page, so the same
//     as 'View all listings from this broker'"
// ---------------------------------------------------------------------------

func TestSellerPictureAndNameLinkToTheBrokerPage(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, firstListingPath(t, h))

	aside := section(t, body, `<aside class="detail-aside">`, `id="enquiry"`, "the seller box")
	identity := section(t, aside, `class="seller-identity"`, "</a>", "the seller identity link")

	// Picture and name inside one anchor, so they cannot disagree about where
	// they lead.
	mustContain(t, identity, "avatar-tile", "the picture belongs inside the link")
	mustContain(t, identity, "seller-identity__name", "…and so does the name")

	// The same destination as the link at the foot of the box — the client's
	// "so the same as 'View all listings from this broker'".
	href := hrefOf(t, aside, "seller-identity")
	if !strings.HasPrefix(href, "/broker/") {
		t.Fatalf("the seller identity must link to a broker page, got %q", href)
	}
	viewAll := regexp.MustCompile(`href="([^"]+)"[^>]*>\s*View all listings from this broker`).
		FindStringSubmatch(aside)
	if viewAll == nil {
		t.Fatal("the view-all link is missing from the seller box")
	}
	if viewAll[1] != href {
		t.Errorf("picture and name lead to %q, View all listings leads to %q", href, viewAll[1])
	}
}

func TestSellerIdentityLooksClickable(t *testing.T) {
	components := asset(t, cssDir+"/components.css")
	mustContain(t, components, ".seller-identity:hover .seller-identity__name { color: var(--link); }",
		"the name must answer the pointer")
	rule := section(t, components, ".seller-identity {", "}", "the seller identity")
	mustContain(t, rule, "display: flex", "the row keeps the layout it had as a plain div")
}

// ---------------------------------------------------------------------------
// 24. "this big title move more up close to header, these is big gap at the
//     moment, lets reduce it. The navigation bar home - developments move quite
//     up next to the header, it is always there in the same place, like in the
//     main listings page."
// ---------------------------------------------------------------------------

func TestPageHeadOpensCloseToTheHeader(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")
	rule := section(t, layout, ".page-head {", "}", "the page head")

	// 16px, which is the figure the search page already uses above its
	// breadcrumb — that is what "like in the main listings page" means.
	mustContain(t, rule, "padding: var(--sp-4) 0", "the gap above the breadcrumb must shrink")
	if strings.Contains(rule, "var(--sp-9)") {
		t.Error("the 48px opening gap the client complained about is still there")
	}

	search := asset(t, "../../web/templates/pages/search.html")
	mustContain(t, search, "padding-top:var(--sp-4)",
		"the search page is the reference for that figure — keep the two in step")
}

func TestEveryPageHeadCarriesTheBreadcrumb(t *testing.T) {
	h := newServer(t)
	// The client's "it is always there in the same place" needs it to be there
	// at all: seven of these pages opened with the title and no breadcrumb.
	for _, path := range []string{
		"/about", "/advertising", "/cookies", "/faq", "/help", "/terms", "/privacy",
		"/pricing", "/brokers", "/developments", "/articles", "/agencies",
	} {
		body := mustGet(t, h, path)
		head := section(t, body, `class="page-head`, "</div>\n</div>", "the page head of "+path)
		mustContain(t, head, `aria-label="Breadcrumb"`, "the breadcrumb is missing on "+path)
		mustContain(t, head, `<a href="/">Home</a>`, "…and must start at Home on "+path)
	}
}

// ---------------------------------------------------------------------------
// 25. "in the main listing page where the header and menu block under it are
//     different colors, then remove this separator line under header"
// ---------------------------------------------------------------------------

func TestNoLineUnderTheHeaderWhenTheBlockBelowIsADifferentColour(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")
	mustContain(t, layout, "body:not(:has(.page-head)) .site-header { border-bottom: 0; }",
		"a page that opens on the page background needs no separator")

	// The search page — the client's "main listing page" — is one of them: it
	// opens on its breadcrumb over --bg, with no page head.
	h := newServer(t)
	search := mustGet(t, h, "/search")
	mustNotContain(t, search, `class="page-head`, "the search page has no page head, so the rule applies to it")
	mustContain(t, search, `class="search-crumb"`, "…its breadcrumb sits straight on the page")
}

// ---------------------------------------------------------------------------
// 26. "in these pages where the menu block under header is the same color as
//     header, the separator line stays. But remove the separator line under this
//     block, where the next menu block is different color"
// ---------------------------------------------------------------------------

func TestTheHeaderKeepsItsLineAboveAPageHead(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")
	header := section(t, layout, ".site-header {", "}", "the header")
	mustContain(t, header, "border-bottom: 1px solid var(--border)",
		"the header keeps its line where the block below shares its colour")

	// The page head is that block: it paints the header's own surface.
	head := section(t, layout, ".page-head {", "}", "the page head")
	mustContain(t, head, "background: var(--surface)", "the page head shares the header's colour")
}

func TestNoLineUnderThePageHead(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")
	head := section(t, layout, ".page-head {", "}", "the page head")
	if strings.Contains(head, "border-bottom") {
		t.Error("the line under the page head was removed at the client's request")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// firstListingPath finds a real listing URL in the search results, so these
// tests never hard-code a slug the seed data could rename. It picks a listing
// that has a broker, because two of the notes are about the seller box.
func firstListingPath(t *testing.T, h http.Handler) string {
	t.Helper()
	body := mustGet(t, h, "/search")
	re := regexp.MustCompile(`href="(/property/[a-z0-9\-]+)"`)
	seen := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(body, -1) {
		if seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		if strings.Contains(mustGet(t, h, m[1]), `class="seller-identity"`) {
			return m[1]
		}
	}
	t.Fatal("no listing with a broker found in the search results")
	return ""
}

// hrefOf returns the href of the first element in body carrying class, with the
// HTML entities the template escaper writes into an attribute decoded again —
// a browser reads &amp; as &, and so should an assertion about a URL.
func hrefOf(t *testing.T, body, class string) string {
	t.Helper()
	// The class may sit before or after the href on the element.
	re := regexp.MustCompile(`<a[^>]*class="` + regexp.QuoteMeta(class) + `"[^>]*href="([^"]+)"|<a[^>]*href="([^"]+)"[^>]*class="` + regexp.QuoteMeta(class) + `"`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no <a class=%q> with an href found", class)
	}
	if m[1] != "" {
		return html.UnescapeString(m[1])
	}
	return html.UnescapeString(m[2])
}
