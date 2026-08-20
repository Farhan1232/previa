package app_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Tests for the client's 18 August notes.
//
// Four notes, four sections, each named after what was asked rather than after
// the code that answers it — the convention the five earlier rounds set.
// Anything the server renders is asserted against a real response; what only
// exists once a browser has laid the page out is asserted at the level of the
// rule that produces it, and was checked in a real headless Chrome as well.

// ---------------------------------------------------------------------------
// 41. "logo is already better, now make it 50 % bigger, there is enough room,
//     then the house on the green background, this background make #7CFC00, the
//     red background make #FF0000 and the blue make #FF00FF."
// ---------------------------------------------------------------------------

func TestLogoIsHalfAsBigAgain(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	// The whole lockup, not the mark alone: the reference the client sent shows
	// both together at the proportion they already had, so keeping that ratio
	// is what "50 % bigger" means. 40→60 and 25→37 were each ×1.5.
	//
	// The client asked for a further 20% the same afternoon, so the exact
	// figures now belong to TestLogoIsAFifthBiggerAgain in
	// feedback_18aug_pm_test.go. What this note still owns is the floor: no
	// variant may go back below what it granted here.
	for want, floor := range map[string]int{
		`.logo__mark { flex: none; width: (\d+)px`:     60,
		`.logo--sm .logo__mark { width: (\d+)px`:       51,
		`.logo--lg .logo__mark { width: (\d+)px`:       78,
		`\n  font-size: (\d+)px;\n  font-weight: 400;`: 37,
	} {
		m := regexp.MustCompile(want).FindStringSubmatch(pages)
		if m == nil {
			t.Fatalf("no size found for %q; the lockup rules have been renamed", want)
			continue
		}
		got, _ := strconv.Atoi(m[1])
		if got < floor {
			t.Errorf("%q is %dpx, below the %dpx this note asked for", want, got, floor)
		}
	}

	// The old sizes must be gone, or one variant is left behind at the old
	// scale and the header and the auth screens disagree about the mark.
	for _, gone := range []string{
		".logo__mark { flex: none; width: 40px; height: 40px; }",
		".logo--sm .logo__mark { width: 34px; height: 34px; }",
		".logo--lg .logo__mark { width: 52px; height: 52px; }",
	} {
		mustNotContain(t, pages, gone, "no logo variant may keep its old size")
	}

	// It has to fit: the header is 72px, and a mark whose *ink* were taller
	// than that would push the bar open. The drawing occupies about 70% of its
	// viewBox, so the 72px mark measured ~50px of ink in Chrome.
	if !strings.Contains(asset(t, cssDir+"/tokens.css"), "--header-h: 72px") {
		t.Error("the header height this size was chosen against has changed; re-check the fit")
	}
}

func TestPropertyDiscsUseTheColoursTheClientNamed(t *testing.T) {
	logo := asset(t, "../../web/templates/components/logo.html")
	favicon := asset(t, imgDir+"/favicon.svg")
	gen := asset(t, "../../docs/logo_gen.py")

	// Named exactly, so they are asserted exactly. Green is the upper-left disc
	// over the Atlantic, red the one over the Arabian Sea, magenta the one on
	// Africa — the same three the client identified by the colour on each.
	for name, svg := range map[string]string{"logo": logo, "favicon": favicon} {
		for _, want := range []string{"#7CFC00", "#FF0000", "#FF00FF"} {
			mustContain(t, svg, `fill="`+want+`"`, name+" must carry "+want)
		}
		for _, gone := range []string{"#2E8B4F", "#CE3B32", "#1D5FB0"} {
			mustNotContain(t, svg, gone, name+" must not keep the colour "+gone+" replaced")
		}
	}

	// The generator is the source of the drawing — regenerating it must not
	// undo the change, which is what would happen if only the output were
	// edited.
	mustContain(t, gen, `DISC_FILLS = ["#7CFC00", "#FF0000", "#FF00FF"]`,
		"the generator must produce the colours, not only the checked-in files")
}

// ---------------------------------------------------------------------------
// 42. "So with the brokers we have two locations now. The broker under his
//     profile can enter the country where he is active and the googlemaps
//     location. In the frontpage the broker section, there the broker can buy an
//     ad that under this country (when the user has chosen his market on the
//     frontpage banner) his ad is active. Then in the header is the 'Brokers'
//     button, in this page can search the brokers by googlemaps location."
// ---------------------------------------------------------------------------

// The two locations answer two different questions, and each drives its own
// screen. This is the whole note in one test.
func TestBrokerHasTwoLocationsDoingTwoJobs(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")

	// Both are set on the profile: the markets they work in, and the pin.
	mustContain(t, settings, `name="active_countries"`, "the markets the broker works in")
	mustContain(t, settings, `name="office_location"`, "and the point on the map")
	mustContain(t, settings, `name="office_lat"`, "…which is a real coordinate")

	// The country drives the homepage strip.
	mustContain(t, settings, `name="broker_ad_countries"`,
		"the ad is aimed at markets, which is the first location")

	// The pin drives the directory.
	brokers := mustGet(t, h, "/brokers")
	mustContain(t, brokers, `id="b-location"`, "the directory searches the pin, which is the second")
	mustContain(t, brokers, `id="b-radius"`, "…by distance from it")
}

func TestHomepageBrokerStripFollowsTheChosenMarket(t *testing.T) {
	h := newServer(t)

	strip := func(country string) []string {
		body := getWithCountry(t, h, "/", country)
		i := strings.Index(body, `class="broker-strip"`)
		if i < 0 {
			return nil
		}
		region := body[i:]
		if j := strings.Index(region, "</section>"); j > 0 {
			region = region[:j]
		}
		seen := map[string]bool{}
		var out []string
		for _, m := range regexp.MustCompile(`/broker/([a-z-]+)`).FindAllStringSubmatch(region, -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				out = append(out, m[1])
			}
		}
		return out
	}

	ee, de := strip("EE"), strip("DE")
	if len(ee) == 0 || len(de) == 0 {
		t.Fatalf("both seeded markets must have advertisers: EE=%v DE=%v", ee, de)
	}
	for _, name := range ee {
		if contains(de, name) {
			t.Errorf("%q appears in both the Estonian and German strips; an ad is "+
				"bought per market and must not run in one it was not bought for", name)
		}
	}

	// A broker who bought two markets appears in both — the cross-border case
	// the client's two-location model exists for. Jonas Weber advertises in
	// Germany, Austria and Czechia.
	if !contains(strip("AT"), "jonas-weber") || !contains(de, "jonas-weber") {
		t.Error("a broker who bought two markets must appear in both strips")
	}
}

// A market nobody bought shows nothing, rather than being filled out with
// brokers from elsewhere: they did not pay to reach this reader.
func TestAMarketWithNoAdvertisersHasNoStrip(t *testing.T) {
	h := newServer(t)
	body := getWithCountry(t, h, "/", "NL")

	mustNotContain(t, body, `class="broker-strip"`,
		"a market with no advertisers must not show a paid strip at all")
	mustNotContain(t, body, "Brokers to work with in Netherlands",
		"…nor its heading")

	// And the page is otherwise intact — the section is dropped, not the page.
	mustContain(t, body, `class="site-footer"`, "the rest of the homepage still renders")
}

func TestBrokerCanBuyTheHomepageAd(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/settings")

	mustContain(t, body, "Advertise on the Previa homepage", "the placement is offered")
	mustContain(t, body, `name="broker_ad"`, "…as something switched on")
	mustContain(t, body, `name="broker_ad_days"`, "…bought for a number of days")
	mustContain(t, body, `name="broker_ad_countries"`, "…in chosen markets")

	// Priced per day, per market. The tier ladder this used to check went on
	// 19 August, when the client replaced it with a rate table: "in the backend
	// there is option to set the price per day for each country."
	mustContain(t, body, "per day", "the placement must be priced per day")

	// A live ad reports itself before it offers to sell another.
	mustContain(t, body, "Running in", "an ad already running must say so")
	mustContain(t, body, "days left", "…and how much of it is left")

	// The markets offered are the ones the seller works in, so nobody has to
	// say where they work twice.
	ad := section(t, body, "Advertise on the Previa homepage", "Save changes", "the ad block")
	mustContain(t, ad, "previaTagPicker()", "markets are chosen with the picker used for countries")

	// The block is a dialog behind a button since 19 August; the button is the
	// violet one the client named, beside the Country label. See
	// TestTheAdvertiseButtonSitsAfterTheCountryTitle in the 19 August file.
	mustContain(t, body, "btn--advertise", "the placement is opened from its own button")

	// And it is reachable from where the client described buying it — the
	// homepage broker section itself, not only from the profile.
	home := getWithCountry(t, h, "/", "EE")
	sell := section(t, home, `class="broker-strip__sell"`, "</p>", "the advertise line")
	mustContain(t, sell, `href="/settings"`, "the strip must say where a placement is bought")
	// What a day in *this* market costs, since the price is now per country:
	// Estonia is seeded at €1 a day.
	mustContain(t, sell, "€1 per day", "…and what a day in this market costs")
}

// ---------------------------------------------------------------------------
// 43. "in the articles page move the articles more up, there is an empty gap,
//     and the current articles preview icons are too big, make them smaller, so
//     that on the wide screen 5 are in one row."
// ---------------------------------------------------------------------------

func TestArticleIndexIsFiveAcrossAndTightUnderTheHead(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/articles")

	// Up: the same 24px step after a page head the other pages use.
	mustContain(t, body, `class="container section section--after-head"`,
		"the articles must sit close under the page head")

	// Five across, which is also what makes the previews smaller — the card
	// takes its width from the column.
	mustContain(t, body, `class="grid grid--articles"`, "the index has its own grid")
	mustNotContain(t, body, `class="grid grid--3"`, "three across was the gap complaint")

	layout := asset(t, cssDir+"/layout.css")
	rule := section(t, layout, ".grid--articles {", "}", "the article grid")
	mustContain(t, rule, "repeat(5, minmax(0, 1fr))", "five in one row on a wide screen")

	// And it still breaks down on narrower screens rather than showing five
	// 180px cards on a tablet: four, three, two, one.
	for _, cols := range []string{"repeat(4", "repeat(3", "repeat(2", "1fr; }"} {
		mustContain(t, layout, ".grid--articles { grid-template-columns: "+cols,
			"the grid must step down to "+cols)
	}
	if n := strings.Count(layout, ".grid--articles"); n != 5 {
		t.Errorf("the article grid has %d rules; want the five-step ladder", n)
	}
}

// ---------------------------------------------------------------------------
// 44. "the article page, make the layout like in other page. At first the
//     navigation bar: home - articles - international - Buying in Spain align
//     left up like the navigation bar is elsewhere. Then the 'international'
//     section must be in the navigation bar (at the moment it is separately
//     below). Then under the header create this 'second header' like all the
//     other pages have. And the article title, date and author put there. The
//     author image is rectangle with rounded corners like it is everywhere."
// ---------------------------------------------------------------------------

func TestArticlePageUsesTheSiteLayout(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/article/buying-in-spain-as-a-non-resident")

	// The page head every other page has, and the breadcrumb inside it at the
	// page gutter rather than indented into a reading column.
	mustContain(t, body, `<div class="page-head">`, "the article needs the second header")
	head := section(t, body, `<div class="page-head">`, `class="article-cover"`, "the page head")
	mustContain(t, head, `<div class="container">`,
		"the breadcrumb must sit at the page gutter, not in the narrow column")
	mustNotContain(t, head, "container--narrow",
		"the head is full width; only the article body is a reading column")

	// Four steps, with the category among them and linking to its own index.
	crumbs := section(t, body, `<ol class="breadcrumbs">`, "</ol>", "the breadcrumb")
	for _, want := range []string{
		`href="/"`, `href="/articles"`, `href="/articles?category=International"`,
	} {
		mustContain(t, crumbs, want, "the breadcrumb must carry this step")
	}
	if n := strings.Count(crumbs, "<li>") + strings.Count(crumbs, `<li aria-current="page">`); n != 4 {
		t.Errorf("the breadcrumb has %d steps; the client asked for four — "+
			"home, articles, category, title", n)
	}

	// The category is in the breadcrumb and nowhere else: it used to be an
	// overline above the title, which is the "separately below" the note names.
	mustNotContain(t, body, `<p class="overline">International</p>`,
		"the category must not also sit on its own above the title")

	// Title, date and author, all in the head.
	pageHead := section(t, body, `<div class="page-head">`, `class="article-cover"`, "the head block")
	mustContain(t, pageHead, `class="page-head__title"`, "the title belongs in the head")
	mustContain(t, pageHead, "min read", "…with the reading time")
	mustContain(t, pageHead, "Marc Puig", "…and the author")
	if !regexp.MustCompile(`\d{1,2} \w{3} \d{4}`).MatchString(pageHead) {
		t.Error("the publication date must be in the head")
	}

	// The picture is the site's rounded rectangle, not a circle.
	mustContain(t, pageHead, "avatar-tile", "the author photo must be the shared rounded rectangle")
	mustNotContain(t, pageHead, `class="avatar"`, "a circular avatar is what the client asked to change")
}

// The reading column is still a reading column — the layout change moved the
// furniture above it, not the measure of the prose.
func TestArticleBodyStaysANarrowColumn(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/article/buying-in-spain-as-a-non-resident")

	prose := section(t, body, `class="prose prose--article"`, "Share this article", "the article body")
	if len(prose) == 0 {
		t.Fatal("the article body did not render")
	}
	mustContain(t, body, "container container--narrow", "the prose keeps its narrow measure")

	// And the container that holds it does not use the padding shorthand, which
	// would cancel its gutters — the bug behind three of the 17 August notes.
	mustNotContain(t, body, `class="container container--narrow" style="padding:`,
		"padding-block only; the shorthand zeroes the container's gutters")
}

// ---------------------------------------------------------------------------
// 45. "In FAQ page this title and nav-bar align to left like everywhere. And
//     this remove: The questions sellers, buyers and renters ask us most often…
//     Then make this 'second header' area smaller and the menu block below move
//     more up. Below this text remove: Still stuck? Help & contact reaches a
//     person, not a queue…"
// ---------------------------------------------------------------------------

func TestFAQPageIsLeftAlignedAndTight(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/faq")

	// Left, like every other page. Centring was this page's own arrangement and
	// nothing else on the site did it.
	head := section(t, body, `<div class="page-head`, `class="container section`, "the page head")
	if strings.Contains(head, "text-align:center") || strings.Contains(head, "align-items:center") {
		t.Error("the FAQ head must not centre its title and breadcrumb")
	}
	mustNotContain(t, body, `class="lede" style="margin-inline:auto`,
		"the centred lede is gone with the centring")

	// Smaller: the account screens' own compact head, which exists for the same
	// complaint about a head that opens with too much air.
	mustContain(t, body, `class="page-head page-head--compact"`,
		"the second header must be the compact one")
	mustContain(t, body, "section--after-head",
		"and the block below must take the 24px step rather than a full section")

	// Both removals, asserted by their text rather than their markup — the
	// client quoted the sentences.
	mustNotContain(t, body, "The questions sellers, buyers and renters ask us most often",
		"the lede was removed at the client's request")
	mustNotContain(t, body, "Still stuck?", "…as was the note under the questions")
	mustNotContain(t, body, "reaches a person, not a queue", "…all of it")

	// The questions themselves are untouched.
	mustContain(t, body, "How long does it take to publish a listing?", "the FAQ still answers")
	if n := strings.Count(body, `class="faq"`); n != 8 {
		t.Errorf("the accordion has %d questions; the eight seeded ones must survive", n)
	}
}

// "Align to left like everywhere" is about the page, not the heading alone: a
// centred column under a left-aligned title would have left the title hanging
// off one side of it.
func TestFAQQuestionsStartAtThePageGutter(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/faq")

	// The full container, so the block starts where the title does…
	mustContain(t, body, `class="container section section--after-head"`,
		"the questions sit in the page's own container")
	mustNotContain(t, body, "container container--narrow section",
		"a centred narrow column is what put the questions out of line")

	// …with the reading measure kept on the list itself rather than by
	// centring it.
	pages := asset(t, cssDir+"/pages.css")
	mustContain(t, pages, ".faq-list { max-width: var(--container-narrow); }",
		"the list keeps a reading measure, taken from the left edge")
	mustContain(t, body, `class="stack faq-list"`, "…and the list carries it")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// getWithCountry fetches a path with the market cookie set, which is how the
// homepage banner's picker travels.
func getWithCountry(t *testing.T, h http.Handler, path, code string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(&http.Cookie{Name: "previa_country", Value: code})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s (market %s): status %d", path, code, rec.Code)
	}
	return rec.Body.String()
}
