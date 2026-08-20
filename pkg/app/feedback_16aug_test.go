package app_test

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Tests for the client's 16 August corrections.
//
// Eight notes, eight sections, each named after what was asked rather than
// after the code that answers it — the same split the earlier rounds use.
// Anything the server renders is asserted against a real response; the parts
// that only exist once a browser has laid the page out or run JavaScript are
// asserted at the level of the rule or the component that produces them, and
// were checked in a real browser as well. Where a number was measured there,
// it is quoted in the comment so a later change that moves it is obvious.

// ---------------------------------------------------------------------------
// 1. "in this search menu page the smaller banner, thos choose your cpuntry
//    button, make the same colors as in frontpage"
// ---------------------------------------------------------------------------

func TestSearchBannerPickerUsesTheHomepageColours(t *testing.T) {
	h := newServer(t)

	// The homepage asks for the "light" tone. With a banner running, the search
	// page must now ask for exactly the same one — not a variant of its own.
	home := mustGet(t, h, "/")
	mustContain(t, home, `class="market-picker market-picker--light"`,
		"the homepage picker is the reference for this tone")

	search := mustGet(t, h, "/search")
	banner := section(t, search, `class="search-banner"`, `id="results"`, "search banner")
	mustContain(t, banner, `class="market-picker market-picker--light"`,
		"the search banner picker must use the homepage's colours")
	mustNotContain(t, banner, "market-picker--banner",
		"the banner-only tone was dropped in favour of the homepage's")

	// And the rules behind it are gone, so there is one treatment to maintain.
	pages := asset(t, cssDir+"/pages.css")
	mustNotContain(t, pages, ".market-picker--banner .market-picker__btn",
		"the dropped tone must not linger as dead CSS")
	mustContain(t, pages, ".market-picker--light .market-picker__btn",
		"the shared tone must still be defined")
}

// ---------------------------------------------------------------------------
// 2. "when the country menu opens then make it to overlay the other things -
//    at the moment it stays behind some elements"
// ---------------------------------------------------------------------------

// The menu opened behind the result cards because a card is not a stacking
// context: its badges (2) and its heart (4) were competing directly with the
// picker's own wrapper, which sat at 2. Two rules fix it — the cards isolate,
// and the wrapper sits an order of magnitude above them.
//
// Verified in Chrome: with the menu open on /search, elementFromPoint over the
// panel returns the menu's own rows, and no card badge or heart is drawn over
// the panel.
func TestCountryMenuOpensOverTheResults(t *testing.T) {
	components := asset(t, cssDir+"/components.css")
	pages := asset(t, cssDir+"/pages.css")

	card := section(t, components, ".pcard {", "}", "the card rule")
	mustContain(t, card, "isolation: isolate",
		"a card must contain its own z-indexes so nothing on it can reach over a menu")

	aside := section(t, pages, ".search-banner__aside {", "}", "the picker wrapper")
	z := regexp.MustCompile(`z-index:\s*(\d+)`).FindStringSubmatch(aside)
	if z == nil {
		t.Fatal("the picker wrapper must carry an explicit z-index")
	}
	n, _ := strconv.Atoi(z[1])
	if n < 10 {
		t.Errorf("picker wrapper z-index is %d; the card layers reach 4, so this "+
			"needs clear headroom above them", n)
	}
}

// ---------------------------------------------------------------------------
// 3. "if the ads are in list view then as well make these green dots there and
//    can swith the preview images"
// ---------------------------------------------------------------------------

func TestListViewCardsPageTheirPhotos(t *testing.T) {
	h := newServer(t)
	list := mustGet(t, h, "/search?view=list")

	card := section(t, list, `class="pcard pcard--row`, "</article>", "the first list card")
	for _, want := range []string{
		"previaCardGallery(",   // the same component the grid cards use
		`class="pcard__track"`, // the sliding strip of photographs
		`class="pcard__pager"`, // the capsule
		`class="pcard__dot"`,   // the dots themselves
		"pcard__zone--prev",    // and the click zones either side
		"pcard__zone--next",
	} {
		mustContain(t, card, want, "a list card must carry the grid card's carousel")
	}

	// The dots are the same green as the grid's, because it is one rule for both.
	components := asset(t, cssDir+"/components.css")
	mustContain(t, components, ".pcard__dot.is-on { background: var(--success-on-dark)",
		"the active dot must stay green")

	// Map cards are still the one variant that opts out: they live inside the
	// map's own Alpine scope, where a nested x-data would shadow it.
	mapView := mustGet(t, h, "/search?view=map")
	mapCard := section(t, mapView, `class="pcard pcard--map`, "</article>", "a map card")
	mustNotContain(t, mapCard, "previaCardGallery(",
		"map cards must not nest a second Alpine component")
}

// ---------------------------------------------------------------------------
// 4. "these social app images are totally wrong! Here are the correct svg ones
//    for day and nigh, use them"
// ---------------------------------------------------------------------------

// The client supplied ten files — a day and a night drawing per app. They are
// identical but for the ink of the glyph, so the artwork ships once with the
// ink as currentColor and the theme switches it. These assertions pin the tile
// colours and one path fragment from each supplied file, so a redraw that
// wanders off the artwork fails here.
func TestMessengerMarksAreTheSuppliedArtwork(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/")

	for _, want := range []struct{ app, tile, fragment string }{
		{"whatsapp", "#25d366", "M123 393l14-65a138 138 0 1150 47z"},
		{"telegram", "#37aee2", "M204 319l135 99c14 9 26 4 30-14l55-258"},
		{"viber", "#7b519d", "M421.915 345.457c-12.198-9.82-25.233-18.634"},
		{"signal", "#3a76f0", "M256 100a156 156 0 00-132 239l-15 64 64-15"},
		{"teams", "#6264a7", "M20.625 8.127q-.55 0-1.025-.205"},
	} {
		mustContain(t, body, want.tile, want.app+"'s tile must keep the supplied colour")
		mustContain(t, body, want.fragment, want.app+"'s glyph must be the supplied path")
	}

	// The marks the earlier screenshots produced are gone.
	for _, old := range []string{"#66D072", "#5CACDD", "#755399", "#4975E8", "#6364A3"} {
		mustNotContain(t, body, old, "the rejected tile colours must not survive")
	}

	// Day is white ink, night is black — the only difference between the two
	// sets of files the client sent.
	components := asset(t, cssDir+"/components.css")
	marks := section(t, components, ".msg-mark {", "/* The day plane", "the mark rules")
	mustContain(t, marks, "color: #fff", "the day drawings use white ink")
	mustContain(t, marks, "[data-theme='dark'] .msg-mark { color: #000; }",
		"the night drawings use black ink")
}

// ---------------------------------------------------------------------------
// 5. "on the map if opened this preview then then can not close it if click to
//    the corss on the top-right, so make it to work. And make this cross in a
//    rounded rectangle backbroud, nad if move the mouse on the cross, then the
//    background color changes"
// ---------------------------------------------------------------------------

// Why it did not close: the invisible link over the photograph carries
// z-index 1 and the close button carried none, so the link took the click and
// opened the listing instead.
//
// Why it did not look like a button: Leaflet's own rule for it is
// `.leaflet-container a.leaflet-popup-close-button`, one part more specific
// than the two-class rule here, and leaflet.css loads last — so its transparent
// background and 24px box quietly won.
//
// Verified in Chrome on /search?view=map: elementFromPoint over the cross now
// returns the cross, clicking it takes the popup count from 1 to 0 without
// navigating, and its background goes from rgba(255,255,255,.92) to
// rgb(12,45,72) on hover.
func TestMapPopupCloseButtonWorksAndLooksLikeAButton(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	const sel = ".leaflet-container .map-popup-shell a.leaflet-popup-close-button"
	rule := section(t, pages, sel+" {", "}", "the close button")

	// Above the photo link at 1 and the pager at 2.
	z := regexp.MustCompile(`z-index:\s*(\d+)`).FindStringSubmatch(rule)
	if z == nil {
		t.Fatal("the close button must carry a z-index or the photo link keeps taking its clicks")
	}
	if n, _ := strconv.Atoi(z[1]); n < 3 {
		t.Errorf("close button z-index is %d, want at least 3 — the photo link sits at 1 "+
			"and the pager at 2", n)
	}

	// A rounded rectangle, not the disc it used to ask for.
	mustContain(t, rule, "border-radius: var(--r-md)", "the cross sits in a rounded rectangle")
	mustNotContain(t, rule, "border-radius: 50%", "it is no longer a circle")

	// And a hover state, so it reads as pressable.
	mustContain(t, pages, sel+":hover,", "the cross must change colour under the pointer")

	// The old two-class selector could never win against the vendor rule.
	mustNotContain(t, pages, ".map-popup-shell .leaflet-popup-close-button {",
		"the rule must outrank leaflet.css, which loads after it")
}

// ---------------------------------------------------------------------------
// 6. "in the frontpate menu this deal type make multiselect as well, like the
//    property type"
// ---------------------------------------------------------------------------

func TestHomepageDealTypeIsMultiSelect(t *testing.T) {
	h := newServer(t)
	home := mustGet(t, h, "/")

	form := section(t, home, `<form class="searchbox"`, "</form>", "homepage search form")
	deal := section(t, form, `class="field type-picker deal-picker"`,
		`location-field`, "the deal field")

	if n := strings.Count(deal, `type="checkbox"`); n != 3 {
		t.Errorf("deal type offers %d checkboxes, want the three deal types", n)
	}
	mustNotContain(t, deal, "<select", "the single-choice select must be gone")
	mustContain(t, deal, `x-data="previaDealPicker()"`, "the dropdown needs its component")

	// The trigger reports the selection, the way the property type beside it does.
	mustContain(t, deal, `x-text="summary"`, "the trigger must caption itself")

	js := asset(t, jsDir+"/previa.js")
	mustContain(t, js, "window.previaDealPicker", "the component must exist")
	mustContain(t, js, "' deal types'", "two or more choices report a count")

	// Several deal types selected on the homepage survive the hand-off, because
	// the parameter simply repeats — which the results page already accepted.
	both := mustGet(t, h, "/search?deal=rent&deal=short_rent")
	group := between(both, `class="segmented segmented--full segmented--deal"`, "</div>")
	for _, want := range []string{`value="rent"`, `value="short_rent"`} {
		if !regexp.MustCompile(regexp.QuoteMeta(want) + `[^>]*checked`).MatchString(group) {
			t.Errorf("a homepage multi-selection did not survive: %s came back unticked", want)
		}
	}
	if regexp.MustCompile(`value="sale"[^>]*checked`).MatchString(group) {
		t.Error("Sell came back ticked when only Rent and Short rent were asked for")
	}
}

// ---------------------------------------------------------------------------
// 7. "if the menu opens there then it overlays everything, at the moment it
//    stays hidden"
// ---------------------------------------------------------------------------

// The hero is a stacking context (isolation: isolate, for the photograph behind
// it at -2). With z-index: auto it painted in document order alongside the
// positioned cards below, so every card in the Featured row drew over an open
// panel. Verified in Chrome: with either hero panel open, the panel covers the
// Featured row completely.
func TestHomepageMenusOpenOverTheSectionsBelow(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	hero := section(t, pages, ".hero { ", "}", "the hero rule")
	mustContain(t, hero, "isolation: isolate", "the hero keeps its own layers inside it")
	z := regexp.MustCompile(`z-index:\s*(\d+)`).FindStringSubmatch(hero)
	if z == nil {
		t.Fatal("the hero needs a z-index, or its dropdowns open behind the sections below")
	}
	n, _ := strconv.Atoi(z[1])
	if n < 1 {
		t.Errorf("hero z-index is %d, which does not lift it above later sections", n)
	}
	// And still under the header, which is the one thing that should cover it.
	if n >= 100 {
		t.Errorf("hero z-index is %d, which would cover the header at 100", n)
	}
}

// ---------------------------------------------------------------------------
// 8. "logo a bit bigger, and so that two houses would not be inside of the
//    circle rather a bit out of it ... the same yellow sea and green
//    contintens"
// ---------------------------------------------------------------------------

func TestLogoIsBiggerAndTwoHousesOverhangTheRim(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	// Bigger: 34px was the header mark before this round, 40 after it, 60 after
	// the client's 18 August morning "make it 50 % bigger", and 72 after that
	// afternoon's "make this logo more 20 % bigger". The current figure is
	// asserted by TestLogoIsAFifthBiggerAgain in feedback_18aug_pm_test.go;
	// what this note still owns is that it never goes back.
	mustContain(t, pages, ".logo__mark { flex: none; width: 72px; height: 72px; }",
		"the mark must be drawn larger than the 34px it was")

	logo := asset(t, "../../web/templates/components/logo.html")
	favicon := asset(t, imgDir+"/favicon.svg")

	// Yellow sea, green continents — the palette the client pointed at. Both are
	// single gradients, because the note ends "then we start to test with
	// different colors", and that test should be one edit per file.
	for name, svg := range map[string]string{"logo": logo, "favicon": favicon} {
		mustContain(t, svg, `stop-color="#F8CE55"`, name+" must carry the yellow sea")
		mustContain(t, svg, `stop-color="#48A863"`, name+" must carry the green continents")
		mustNotContain(t, svg, `stop-color="#2E7C7A"`, name+" must not keep the teal sphere")
	}

	// Two of the three discs cross the rim; the third stays inside it. The globe
	// and the discs are all plain circles, so this is arithmetic on the drawing
	// itself rather than a promise made in a comment.
	circle := regexp.MustCompile(`<circle cx="([\d.]+)" cy="([\d.]+)" r="([\d.]+)"`)
	var globe [3]float64
	overhanging, inside := 0, 0
	badges := section(t, logo, `<g class="logo__badges"`, "</g>", "the disc group")

	if m := circle.FindStringSubmatch(logo); m != nil {
		for i := 1; i <= 3; i++ {
			globe[i-1], _ = strconv.ParseFloat(m[i], 64)
		}
	} else {
		t.Fatal("the globe circle was not found in the mark")
	}

	for _, m := range circle.FindAllStringSubmatch(badges, -1) {
		cx, _ := strconv.ParseFloat(m[1], 64)
		cy, _ := strconv.ParseFloat(m[2], 64)
		r, _ := strconv.ParseFloat(m[3], 64)
		d := math.Hypot(cx-globe[0], cy-globe[1])
		switch {
		case d+r > globe[2]+1: // reaches past the rim by more than a hairline
			overhanging++
			if d > globe[2] {
				t.Errorf("a disc at %.1f,%.1f has left the globe entirely; the client "+
					"asked for a bit out of it, not off it", cx, cy)
			}
		case d+r <= globe[2]:
			inside++
		}
	}
	if overhanging != 2 {
		t.Errorf("%d discs overhang the rim, want exactly 2", overhanging)
	}
	if inside != 1 {
		t.Errorf("%d discs sit inside the rim, want exactly 1", inside)
	}

	// One drawing, two files: the discs must be in the same places in both.
	for _, m := range circle.FindAllStringSubmatch(badges, -1) {
		mustContain(t, favicon, m[0], "the favicon must be the same drawing as the mark")
	}
}

// ---------------------------------------------------------------------------
// 9. "these hearts, when not chosen, then this white border make red, inside
//    stays gray. And if chosen, then all the heart is red"
//
//	Revised 17 August: "when the heart in the ad is not checked, then make it's
//	background transparent, the borderline red and background transparent. At
//	the momene the background is gray, but make it totally transparent. So that
//	only the red heart bordereline on the image and that's it."
//
//	So the grey body is gone as well as the white outline — an unsaved heart is
//	a red outline over the photograph and nothing more.
//
// ---------------------------------------------------------------------------

func TestHeartIsRedOutlinedUntilSaved(t *testing.T) {
	components := asset(t, cssDir+"/components.css")

	glyph := section(t, components, ".pcard__fav svg {", "}", "the heart glyph")
	mustContain(t, glyph, "fill: none", "an unsaved heart must have no body at all")
	// --fav since 18 August, when the client pinned the red to #FF0000 in both
	// themes; --error is themed and was the salmon they were looking at. The
	// point of this assertion is unchanged: the outline is red, not white.
	mustContain(t, glyph, "stroke: var(--fav)", "the outline must be red, not white")
	mustNotContain(t, glyph, "stroke: #fff", "the white outline is what the client asked to lose")
	mustNotContain(t, glyph, "fill: currentColor",
		"a filled body is the grey plate the client asked to make transparent")

	// Nothing may reintroduce a plate behind the glyph either.
	button := section(t, components, ".pcard__fav {", "}", "the heart button")
	mustContain(t, button, "background: none", "the button itself must stay transparent")
	mustNotContain(t, button, "rgba(0, 0, 0, 0.52)", "the grey body is gone")

	// Saved: the same red fills the body, so no second colour rings the glyph.
	mustContain(t, components, ".pcard__fav[aria-pressed='true'] svg { fill: var(--fav); }",
		"a saved heart is red throughout")
}

// ---------------------------------------------------------------------------
// 10. "the info text on them need to make a bit more compact ... The text lines
//     what belong together ... barely touch each other. The price move more up
//     closer to the image."
// ---------------------------------------------------------------------------

// Measured in Chrome on the homepage's Featured row: a card's text block is
// 34px shorter than before, the two lines of a wrapped title sit 19px apart
// rather than 22px, and the price starts 8px below the photograph rather
// than 20px.
func TestCardTextIsCompact(t *testing.T) {
	components := asset(t, cssDir+"/components.css")

	body := section(t, components, ".pcard__body {", "}", "the card text block")
	mustContain(t, body, "gap: var(--sp-2)",
		"the blocks tighten from 12px to 8px — reduced, not removed")
	mustContain(t, body, "padding: var(--sp-3) var(--sp-5) var(--sp-5)",
		"the price moves up towards the photograph")

	// The lines that belong together are the real change.
	title := section(t, components, ".pcard__title {", "}", "the title")
	lh := regexp.MustCompile(`line-height:\s*([\d.]+)`).FindStringSubmatch(title)
	if lh == nil {
		t.Fatal("the title must set its own line-height")
	}
	if v, _ := strconv.ParseFloat(lh[1], 64); v > 1.25 {
		t.Errorf("title line-height is %s; a wrapped title must set as one block", lh[1])
	}

	loc := section(t, components, ".pcard__location {", "}", "the location")
	lh = regexp.MustCompile(`line-height:\s*([\d.]+)`).FindStringSubmatch(loc)
	if lh == nil {
		t.Fatal("the location must set its own line-height")
	}
	if v, _ := strconv.ParseFloat(lh[1], 64); v > 1.3 {
		t.Errorf("location line-height is %s; a wrapped address must set as one block", lh[1])
	}

	// The rule above the facts row stays: it is what says which lines are which.
	facts := section(t, components, ".pcard__facts {", "}", "the facts row")
	mustContain(t, facts, "border-top: 1px solid var(--border)",
		"the separation between the copy and the numbers must survive the tightening")
}
