package app_test

import (
	"html"
	"regexp"
	"strings"
	"testing"
)

// Tests for the client's 17 August late-evening corrections.
//
// Eleven notes, eleven sections, each named after what was asked rather than
// after the code that answers it — the convention the four earlier rounds set.
// Anything the server renders is asserted against a real response; what only
// exists once a browser has laid the page out is asserted at the level of the
// rule that produces it, and was checked in a real headless Chrome as well.
//
// Three of the notes — "make it narrower with the agencies", "make the page
// content narrower", "make the footer wider" — turned out to be one bug seen
// from three pages, so they share a test. See
// TestPageContentLinesUpWithTheFooterEverywhere.

// ---------------------------------------------------------------------------
// 27. "this text 'Every broker on Previa is verified…' remove, as we do not
//     deal with broker verification"
// ---------------------------------------------------------------------------

func TestBrokerDirectoryDoesNotClaimVerification(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/brokers")

	mustNotContain(t, body, "Every broker on Previa is verified",
		"Previa does not verify brokers, so the page must not say it does")
	// The rest of that sentence described a search by market and city the page
	// no longer offers either.
	mustNotContain(t, body, "Filter by market, city or specialisation",
		"the page no longer filters by market or city")
	mustContain(t, body, "Find a broker", "the title itself stays")
}

// ---------------------------------------------------------------------------
// 28. "the broker search we relate to googlemaps. Each broker can specify under
//     his profile his location on googlemaps and then here in the broker search
//     section the user enters the location and radius, so from this location and
//     with this radius (50 km) the brokers are displayed. So in this search
//     menu, where is country, there make 'Location' … The City remove."
// ---------------------------------------------------------------------------

func TestBrokerSearchIsALocationAndRadius(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/brokers")

	// The Location box is the property search's own control, so a place that
	// resolves there resolves here to the same point.
	mustContain(t, body, `id="b-location"`, "the directory searches by location")
	mustContain(t, body, `name="location"`, "…under the parameter the rest of the site uses")
	mustContain(t, body, "previaLocation()", "…with the same autocomplete behaviour")

	// Radius, defaulting to the client's own figure. It became a typed number
	// on 18 August — "manually can user set any number" — so the default now
	// arrives as the field's value rather than as a selected option; see
	// TestBrokerRadiusIsATypedNumberInKilometres.
	mustContain(t, body, `id="b-radius"`, "and by a radius")
	mustContain(t, body, `value="50"`, "50 km is the client's default")

	// The two controls it replaced are gone. Scoped to the search form: the
	// header's own market picker is a different control and still says "All
	// countries" over its list, quite correctly.
	form := section(t, body, `class="card broker-search"`, "</form>", "the broker search form")
	mustNotContain(t, form, `id="b-country"`, "the country select is replaced by the location box")
	mustNotContain(t, form, `id="b-city"`, "the city field was removed at the client's request")
	mustNotContain(t, form, "All countries", "…and so is its placeholder")
	mustNotContain(t, form, "Any city", "…and the city placeholder with it")
}

// The radius is a real distance from a real point, not a relabelled country
// filter. Tallinn and Helsinki are 82 km apart across the gulf, which is the
// case the client cares about and the one a country filter cannot express.
func TestBrokerRadiusMeasuresRealDistance(t *testing.T) {
	h := newServer(t)

	names := func(path string) []string {
		body := mustGet(t, h, path)
		found := regexp.MustCompile(`/broker/([a-z-]+)`).FindAllStringSubmatch(body, -1)
		seen := map[string]bool{}
		var out []string
		for _, m := range found {
			if !seen[m[1]] {
				seen[m[1]] = true
				out = append(out, m[1])
			}
		}
		return out
	}

	near := names("/brokers?location=Tallinn%2C+Estonia&radius=50")
	gulf := names("/brokers?location=Tallinn%2C+Estonia&radius=200")
	far := names("/brokers?location=Berlin%2C+Germany&radius=500")

	if len(near) == 0 {
		t.Fatal("a 50 km search from Tallinn must find the Tallinn brokers")
	}
	if len(gulf) <= len(near) {
		t.Errorf("200 km from Tallinn must reach Helsinki as well: got %v then %v", near, gulf)
	}
	for _, n := range near {
		if !contains(gulf, n) {
			t.Errorf("widening the radius must not drop %q", n)
		}
	}
	// Berlin's own brokers plus Prague at ~280 km, and nobody from the Baltic.
	if !contains(far, "petra-novak") {
		t.Error("500 km from Berlin must reach Prague")
	}
	if contains(far, "kadri-tamm") {
		t.Error("500 km from Berlin must not reach Tallinn, which is over 1000 km away")
	}

	// Nearest first: a radius search that does not order by distance is not
	// answering the question that was asked.
	ordered := names("/brokers?location=Helsinki%2C+Finland&radius=200")
	if len(ordered) < 2 || !strings.HasPrefix(ordered[0], "aino") && !strings.HasPrefix(ordered[0], "mikko") {
		t.Errorf("a radius search must list the nearest brokers first, got %v", ordered)
	}
}

// ---------------------------------------------------------------------------
// 29. "in the language section enlist all the languages with search menu and
//     multiselect function. The first selection is 'any'."
// ---------------------------------------------------------------------------

func TestBrokerLanguageIsAMultiSelect(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/brokers")

	// The market picker's menu: a search box over a flagged list, several at
	// once, each one a removable tag.
	mustContain(t, body, "previaTagPicker()", "the same control as the market picker")
	mustContain(t, body, "Search language", "…with its search box")
	mustContain(t, body, `name="language" value="cs"`, "every catalogue language is offered")
	mustContain(t, body, "tagpick__tag", "…and a choice becomes a removable tag")

	// "The first selection is 'any'": nothing chosen means no filter, shown as
	// a resting state rather than offered as an option to un-tick.
	mustContain(t, body, "Any language", "the resting state says so")
	mustNotContain(t, body, `value="any"`, "Any is the absence of a choice, not a code")

	// Every language, not only the ones a seeded broker happens to speak.
	// French is in the catalogue; no seeded broker lists it alone.
	mustContain(t, body, `value="fr"`, "all the languages must be there")

	// Several at once, OR-ed.
	one := mustGet(t, h, "/brokers?language=fi")
	two := mustGet(t, h, "/brokers?language=fi&language=cs")
	if brokerCount(t, one) >= brokerCount(t, two) {
		t.Error("a second language must widen the results, not narrow them")
	}

	// And a language still narrows — but no longer measured against the bare
	// page, which stopped meaning "every broker" in the 18 August evening round
	// (note 64): with nothing searched for, the directory shows the header
	// market's paid strip. The whole directory is now what a search that
	// excludes nobody returns, so that is what a language search is compared
	// with. A 5000 km radius from Tallinn reaches every seeded broker.
	all := mustGet(t, h, "/brokers?location=Tallinn%2C+Estonia&radius=5000")
	if brokerCount(t, all) <= brokerCount(t, one) {
		t.Error("choosing a language must narrow the directory")
	}

	// The other half of that note, and the reason the comparison moved: a
	// language search is not confined to the market. The reader asked for
	// brokers who speak a language, and answering with only the ones
	// advertising in Estonia would be the market overruling the question.
	//
	// Membership, not a count. Until 19 August the Estonian strip held two
	// brokers and any language search was bigger than it, so a count said this
	// on its own; the strip is nine deep now — the client asked the homepage
	// for two full rows of advertisers — and a search that returns three
	// German speakers from outside it is smaller than the strip while still
	// reaching well past it. The question is which brokers come back, so that
	// is what is asserted.
	strip := brokerNames(t, mustGet(t, h, "/brokers"))
	german := brokerNames(t, mustGet(t, h, "/brokers?language=de"))
	if len(german) == 0 {
		t.Fatal("a German-language search must find the German-speaking brokers")
	}
	var beyond int
	for _, name := range german {
		if !contains(strip, name) {
			beyond++
		}
	}
	if beyond == 0 {
		t.Errorf("a language search must reach past the market's own strip: %q are all in %q",
			german, strip)
	}
}

// ---------------------------------------------------------------------------
// 30. "and this broker menu bar move more up, there is too big gap"
// ---------------------------------------------------------------------------

func TestBrokerSearchSitsUnderTheTitle(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/brokers")

	// The same tightened step the account pages use after a page head: 24px
	// rather than the section default of 48–96.
	mustContain(t, body, `class="container section section--after-head"`,
		"the search bar must sit close under the title")

	layout := asset(t, cssDir+"/layout.css")
	mustContain(t, section(t, layout, ".section--after-head {", "}", "the tightened step"),
		"padding-top: var(--sp-6)", "…which is what that class is for")
}

// ---------------------------------------------------------------------------
// 31. "here remove the separator line as the menu section are different colour
//     anyway"
// ---------------------------------------------------------------------------

func TestBandsHaveNoSeparatorLine(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")

	// The colour moved onto a pseudo-element on 18 August, when bands became
	// inset panels — see TestColouredBlocksArePanelsAtTheFooterWidth. This
	// note's point is unchanged and is asserted against wherever the colour
	// lives now: a block that differs in colour needs no rule as well.
	band := section(t, layout, ".section--band::before { background:", "\n", "the band's colour")
	mustContain(t, band, "var(--surface-2)", "the block keeps its colour")

	rules := section(t, layout, ".section--band,\n.section--navy {", "}", "the band")
	mustNotContain(t, rules, "border-block", "a block that differs in colour needs no rule as well")
}

// ---------------------------------------------------------------------------
// 32/33/35. "here are two pages where the content has different width — the
//     content width is in every page the same. So make it narrower with the
//     agencies." · "in all the pages make that the footer is the same width as
//     the page content, so here make the page content narrower." · "in this page
//     make the footer wider as here is more content."
// ---------------------------------------------------------------------------

// Three notes, one bug. `.section` set its padding with the shorthand, so every
// element carrying both `container` and `section` — which is most content pages
// — had the container's 32px gutters zeroed out by `padding: … 0` and ran a
// gutter wider on each side than the footer. Agencies, help and the account
// screens are the three pages the client photographed; the fix is one property
// and it aligns all of them at once.
func TestPageContentLinesUpWithTheFooterEverywhere(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")

	for _, rule := range []string{".section {", ".section--lg {", ".section--tight {"} {
		body := section(t, layout, rule, "}", rule)
		mustContain(t, body, "padding-block:", "vertical rhythm only, so the gutters survive")
		if strings.Contains(body, "padding:") {
			t.Errorf("%s must not use the padding shorthand: it zeroes .container's gutters", rule)
		}
	}

	// The footer's width is unchanged — it was always the content box — so
	// with the gutters restored the two agree on every page.
	mustContain(t, section(t, layout, ".site-footer {", "}", "the footer"),
		"width: min(100% - var(--gutter) * 2, calc(var(--container) - var(--gutter) * 2))",
		"the footer still stops where the page content stops")

	// The pages the client named all use the ordinary container, so nothing
	// needs a per-page width any more.
	h := newServer(t)
	for _, path := range []string{"/agencies", "/help", "/dashboard", "/brokers"} {
		body := mustGet(t, h, path)
		mustContain(t, body, `class="container section`, path+" uses the shared container")
	}
}

// ---------------------------------------------------------------------------
// 34. "in one page the footer is nice, in the other page next to footer is
//     panel's end. Make that this panel with this background colour goes till to
//     the end of the page so the footer stays inside there."
//
//     Reversed by the client on 18 August, once the coloured blocks stopped
//     being full-bleed bands: "remove this block around the footer and footer
//     stays into the same background as in other pages." A band running to the
//     bottom of the window was the right answer while a band ran to both edges;
//     with every block now an inset panel it would have wrapped the footer in a
//     colour nothing above it shared. What survives of this note is the
//     .page__foot ground itself, which is still what stops the footer sitting
//     a few pixels below the last thing on the page.
// ---------------------------------------------------------------------------

func TestTheFooterStandsOnAGroundOfItsOwn(t *testing.T) {
	h := newServer(t)
	mustContain(t, mustGet(t, h, "/"), `class="page__foot"`,
		"the footer stands on a ground of its own")

	layout := asset(t, cssDir+"/layout.css")
	ground := section(t, layout, ".page__foot {", "}", "that ground")
	mustContain(t, ground, "background: var(--bg)", "which is the page colour on every page")
	mustContain(t, ground, "margin-top: auto", "and still pushes the footer down a short page")
}

// ---------------------------------------------------------------------------
// 36. "this active and draft make with small letters"
// ---------------------------------------------------------------------------

// Only the four listing states. The seller kind beside them on the admin table
// is drawn in the same badge and stays capitalised — "Broker" and "Private"
// name what someone is, where a status names what a listing is doing.
func TestListingStatusIsLowerCase(t *testing.T) {
	h := newServer(t)

	for _, path := range []string{"/dashboard", "/my-listings", "/admin/listings"} {
		body := mustGet(t, h, path)
		seen := false
		for _, state := range []string{"active", "draft", "expired", "sold"} {
			upper := strings.ToUpper(state[:1]) + state[1:]
			mustNotContain(t, body, ">"+upper+"</span>", path+": a status badge must be lower case")
			if strings.Contains(body, ">"+state+"</span>") {
				seen = true
			}
		}
		if !seen {
			t.Errorf("%s: no listing status badge found at all", path)
		}
	}
}

// ---------------------------------------------------------------------------
// 37. "these dropdown menus open in wrong place, make them to be opened in the
//     same place where you click them"
// ---------------------------------------------------------------------------

// The menu was placed with `top: calc(100% + 4px)`, and a percentage top is a
// percentage of the containing block's height — so the 100% measured the whole
// field, hint included, rather than the box the button sits in. On the settings
// screen that dropped the menu clear of the card.
func TestTagPickerMenuOpensUnderItsButton(t *testing.T) {
	picker := asset(t, "../../web/templates/components/tag-picker.html")
	mustContain(t, picker, `class="tagpick__anchor"`,
		"the box and its menu need a positioned wrapper of their own")

	components := asset(t, cssDir+"/components.css")
	mustContain(t, section(t, components, ".tagpick__anchor {", "}", "that wrapper"),
		"position: relative", "…so 100% means the height of the box")

	// Selected through the anchor, because .dropdown and .dropdown > .menu are
	// declared in pages.css — which loads after this file and would otherwise
	// win at equal specificity, leaving the element `relative`.
	menu := section(t, components, ".tagpick__anchor > .tagpick__dropdown {", "}", "the menu")
	mustContain(t, menu, "position: absolute", "the menu is positioned, not offset in flow")
	mustContain(t, menu, "top: 100%", "…to the bottom of the box")

	// And it flips above the box when what is below cannot hold it — the case
	// of the search sidebar, which scrolls.
	mustContain(t, components, `.tagpick__anchor.is-up > .tagpick__dropdown`,
		"a menu with no room below must open upward")
	js := asset(t, jsDir+"/previa.js")
	mustContain(t, js, "place: function ()", "…decided when the menu opens")
	mustContain(t, js, "this.$root.querySelector('.tagpick__anchor')",
		"$el is the clicked button inside an event handler; the root is what is wanted")
}

// ---------------------------------------------------------------------------
// 38. "under seller's profile add more the googlemaps location place, where user
//     can specify his location on the googlemaps — so in the brokers section the
//     users can search the brokers on the googlemaps"
// ---------------------------------------------------------------------------

func TestSellerProfileHasAMapLocation(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/settings")

	mustContain(t, body, "previaProfileLocation(", "the profile carries a map location")
	mustContain(t, body, `name="office_location"`, "…searched for by address")
	mustContain(t, body, `name="office_lat"`, "…and submitted as a point")
	mustContain(t, body, `name="office_lng"`, "…in both axes")
	mustContain(t, body, "previaMap(", "…with a map to place the pin on")

	// Leaflet is loaded only where a map is drawn, so the picker needs it asked
	// for — without this the map renders its "didn't load" state.
	mustContain(t, body, "/static/vendor/leaflet.js", "the settings page must load the map library")

	// The pin is a plain marker: a price bubble on a point with no price reads
	// as "0", which is what this screen showed before. The config travels in an
	// attribute, so its quotes arrive as entities.
	mustContain(t, html.UnescapeString(body), `"pin":true`, "a location pin is not a price marker")
	js := asset(t, jsDir+"/previa.js")
	mustContain(t, js, "if (config.pin)", "…which the map honours")

	// It is the same fact the directory measures against, so it is the same
	// shape: a broker's office and a seller's location are one type.
	mustContain(t, mustGet(t, h, "/brokers"), `id="b-location"`,
		"the directory searches the pins this screen sets")
}

// ---------------------------------------------------------------------------
// 39. "the up/down arrows need to make better, make the same style as in
//     sexydate, that they sit on the right side of the field"
// ---------------------------------------------------------------------------
//
// Asserted by TestNumberSteppersSitInTheCorner in feedback_17aug_pm_test.go,
// which is where the client's first pass at this note already lives. Keeping
// both halves of one complaint in one test is the point.

// ---------------------------------------------------------------------------
// 40. "in the search menu the language of communication needs to be the same
//     menu as the choose your market, with the search field and all the
//     languages must be there. So the user checks the one he needs and tags will
//     come to see, what can be deleted as well"
// ---------------------------------------------------------------------------

func TestSearchFilterLanguageIsTheMarketMenu(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search?language=de")

	panel := section(t, body, `class="filter-panel`, `filter-panel__footer`, "the filter panel")
	mustContain(t, panel, "Language of communication", "the client's wording stays")
	mustContain(t, panel, "previaTagPicker()", "it is the market picker's menu now")
	mustContain(t, panel, "Search language", "…with the search field")
	mustContain(t, panel, "tagpick__tag", "…and tags that can be deleted")
	mustContain(t, panel, "Any language", "…resting on Any")

	// Every language, where the panel used to offer only those some listing was
	// already sold in.
	for _, code := range []string{"cs", "fr", "sv", "ru"} {
		mustContain(t, panel, `value="`+code+`"`, "the whole catalogue must be offered")
	}

	// The column of checkboxes it replaced is gone, and so is the CSS that
	// laid it out.
	mustNotContain(t, panel, `class="lang-filter"`, "the old two-column checkbox list is replaced")
	mustNotContain(t, asset(t, cssDir+"/components.css"), ".lang-filter",
		"its rule must not linger as dead CSS")

	// Filtering still works, and still ORs.
	if brokerCount(t, mustGet(t, h, "/search?deal=sale")) <= brokerCount(t, mustGet(t, h, "/search?deal=sale&language=de")) {
		t.Error("choosing a language must still narrow the results")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// brokerCount reads the result count both directories print in the same markup.
func brokerCount(t *testing.T, body string) int {
	t.Helper()
	m := regexp.MustCompile(`<strong class="numeric">(\d+)</strong>`).FindStringSubmatch(body)
	if m == nil {
		t.Fatal("no result count in the response")
	}
	n := 0
	for _, c := range m[1] {
		n = n*10 + int(c-'0')
	}
	return n
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
