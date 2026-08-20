package app_test

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

// Tests for the client's second batch of 18 August notes (12:21–13:14).
//
// Fourteen notes, numbered 45–58 on from the morning's four, each section named
// after what was asked rather than after the code that answers it — the
// convention the six earlier rounds set. Anything the server renders is
// asserted against a real response; what only exists once a browser has laid
// the page out or run the JavaScript is asserted at the level of the rule or
// the function that produces it, and was checked in a real headless Chrome as
// well. Numbers measured in that pass are quoted in the comments, so a later
// change that moves them is obvious.

// ---------------------------------------------------------------------------
// 45. "Around the earth there is a borderlin, this remove completely, so the
//     earth stretches now till to the end (do not make it smaller). Then make
//     this logo more 20 % bigger. The header part do not make bigger any more,
//     the header size stays."
// ---------------------------------------------------------------------------

func TestNothingRingsThePlanetAnyMore(t *testing.T) {
	logo := asset(t, "../../web/templates/components/logo.html")
	favicon := asset(t, imgDir+"/favicon.svg")
	gen := asset(t, "../../docs/logo_gen.py")

	// The rim was a stroked circle just inside the radius; the limb darkening
	// was a second circle filled with a gradient that went opaque at the edge.
	// Between them they drew the ring the client is pointing at, so both go.
	for name, svg := range map[string]string{"logo": logo, "favicon": favicon} {
		mustNotContain(t, svg, `stroke="#7A4A08"`, name+" must not ring the globe with a stroke")
		mustNotContain(t, svg, "pvE", name+" must not keep the limb-darkening layer")
	}
	mustNotContain(t, gen, `stroke="#7A4A08"`, "the generator must not draw the rim back in")
	mustNotContain(t, gen, `id="pvE`, "the generator must not define the limb-darkening gradient")

	// "Do not make it smaller": the sea circle still runs to the full radius.
	mustContain(t, logo, `<circle cx="32" cy="32.4" r="22" fill="url(#pvS`,
		"the sea must still be drawn at the full radius")
}

func TestLogoIsAFifthBiggerAgain(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	// ×1.2 on every size, the lockup scaling together as it has each time.
	for _, want := range []string{
		".logo__mark { flex: none; width: 72px; height: 72px; }",
		"font-size: 44px;",
		".logo--sm .logo__mark { width: 61px; height: 61px; }",
		".logo--sm .logo__word { font-size: 37px; }",
		".logo--lg .logo__mark { width: 94px; height: 94px; }",
		".logo--lg .logo__word { font-size: 58px; }",
	} {
		mustContain(t, pages, want, "the lockup must be a fifth bigger again")
	}
	for _, gone := range []string{
		".logo__mark { flex: none; width: 60px; height: 60px; }",
		".logo--sm .logo__mark { width: 51px; height: 51px; }",
		".logo--lg .logo__mark { width: 78px; height: 78px; }",
	} {
		mustNotContain(t, pages, gone, "no logo variant may keep yesterday's size")
	}

	// "The header size stays." The mark is exactly the header's height, which
	// fits because the drawing occupies about 70% of its viewBox: measured in
	// Chrome at 72px, the ink is ~50px tall inside a 72px row.
	mustContain(t, asset(t, cssDir+"/tokens.css"), "--header-h: 72px",
		"the header height must not have grown with the logo")
}

// ---------------------------------------------------------------------------
// 46. "The favicon we do that this planet with the houses in a max size, what
//     can be, so in the favicon area the left and right borders of the planet
//     and up and down borders touch the favicon area borders … around the
//     favicon we draw rounded rectangle like it is now, and this background
//     make #cc00cc."
// ---------------------------------------------------------------------------

func TestFaviconFillsItsTileOnMagenta(t *testing.T) {
	favicon := asset(t, imgDir+"/favicon.svg")

	// #cc00cc was the first magenta; "the favicon backgroun lets try: #8B008B"
	// replaced it the same evening. The shape of the note is unchanged — one
	// flat colour on a rounded tile — so the assertion tracks the new value
	// rather than being duplicated in the later round's file.
	mustContain(t, favicon, `<rect width="64" height="64" rx="14" fill="#8B008B"/>`,
		"the tile must be a rounded rectangle in the colour the client named")
	// The lit gradient the tile used to carry is gone with it: one named
	// colour, not a shaded surface competing with the planet on top of it.
	mustNotContain(t, favicon, "pvTf", "the tile must not keep its own gradient")

	// The mark was inset at .88 to keep the overhanging discs off the corners.
	// It now fills the tile instead.
	mustNotContain(t, favicon, "scale(.88)", "the mark must not be inset in the tile any more")
	mustContain(t, favicon, `<g transform="translate(32 32) scale(1.4287)`,
		"the mark must be scaled up to the tile's own size")

	// The scale is computed, so regenerating cannot quietly undo it.
	mustContain(t, asset(t, "../../docs/logo_gen.py"), "def favicon_transform(",
		"the fit must come from the generator rather than being hand-typed")
}

// A favicon that does not parse renders as nothing at all, in every tab, and
// the failure is silent: the browser simply shows its blank page glyph. This
// caught a real one — an XML comment may not contain a double hyphen, and the
// note above the drawing had been written with a command line in it.
func TestFaviconIsWellFormedXML(t *testing.T) {
	dec := xml.NewDecoder(strings.NewReader(asset(t, imgDir+"/favicon.svg")))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatalf("favicon.svg does not parse, so no browser will draw it: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 47. "In the brokers page the 'Radius (km)' make like that so that km is next
//     to radius, so remove km behind each number. Then on the right side of the
//     radius number feld make the same up-down arrows as in the search-menu.
//     And the up-down step is 10 when click the arrow, so if user has set 5 at
//     first and clicks up then it will be 15, but manually can user set any
//     number."
// ---------------------------------------------------------------------------

func TestBrokerRadiusIsATypedNumberInKilometres(t *testing.T) {
	body := mustGet(t, newServer(t), "/brokers")

	// The unit is in the label, once, instead of on every value.
	mustContain(t, body, `<label class="field__label" for="b-radius">Radius (km)</label>`,
		"the unit belongs beside the word Radius")
	mustNotContain(t, body, "50 km<", "no value may carry the unit after it")

	// A number field, not a list: "manually can user set any number".
	mustNotContain(t, body, `<select class="select" id="b-radius"`,
		"the radius must not be a fixed list of choices any more")
	radius := section(t, body, `id="b-radius"`, ">", "radius field")
	for _, want := range []string{`type="number"`, `step="10"`, `data-step-loose="on"`} {
		mustContain(t, radius, want, "the radius field must carry "+want)
	}
}

// The arrows themselves are built by previa.js for every number field on the
// site — the same control the search menu uses, which is what the client asked
// for. What is specific here is how they step: 5 + 10 has to be 15, and the
// platform's own stepUp snaps to the step grid and would give 10. Verified in
// Chrome: 5 → 15 → 25, and clamped to min on the way back down.
func TestSteppingIsLooseWhereTheClientAskedForIt(t *testing.T) {
	js := asset(t, jsDir+"/previa.js")
	mustContain(t, js, "if (input.dataset.stepLoose === 'on') {",
		"a field may ask for plain addition instead of the step grid")
	mustContain(t, js, "var next = cur + (dir === 'up' ? step : -step);",
		"loose stepping must add the step to whatever is in the field")
	mustContain(t, js, "if (input.min !== '' && next < Number(input.min)) next = Number(input.min);",
		"loose stepping must still respect min")
	// Everything else keeps the native behaviour, which is right for a year or
	// a floor number.
	mustContain(t, js, "dir === 'up' ? input.stepUp() : input.stepDown();",
		"fields without the flag must keep the platform's stepping")
}

// ---------------------------------------------------------------------------
// 48. "In the broker single page the page content move more up, next to the
//     header, there is too big empty area."
// ---------------------------------------------------------------------------

func TestBrokerPageOpensTightUnderItsHead(t *testing.T) {
	h := newServer(t)
	for _, path := range []string{"/broker/liis-kask", "/agency/kadaka-kinnisvara"} {
		body := mustGet(t, h, path)
		mustContain(t, body, `class="container section section--after-head"`,
			path+" must open on the tighter step under its page head")
	}
	// 24px, the same figure the account screens and the broker directory use.
	mustContain(t, asset(t, cssDir+"/layout.css"), ".section--after-head { padding-top: var(--sp-6); }",
		"the tighter step must still be the 24px one")
}

// ---------------------------------------------------------------------------
// 49. "Under the article single page the preview icons are too big, make them
//     the same size as in article page."
// ---------------------------------------------------------------------------

func TestRelatedArticlesAreTheSameSizeAsTheIndex(t *testing.T) {
	h := newServer(t)
	index := mustGet(t, h, "/articles")
	detail := mustGet(t, h, "/article/buying-in-spain-as-a-non-resident")

	mustContain(t, index, `class="grid grid--articles"`, "the index ladder must be unchanged")
	mustContain(t, detail, `class="grid grid--articles"`,
		"related reading must use the index's ladder, not the three-across one")

	// The class carries the card's tighter padding and type as well as the
	// column width, which is what makes the two cards identical rather than
	// merely equally wide.
	pages := asset(t, cssDir+"/pages.css")
	mustContain(t, pages, ".grid--articles .acard__body", "the tighter card must come with the grid")
}

// ---------------------------------------------------------------------------
// 50. "The main search menu is very good, just make it more compact … In
//     left-right scale no changes, the width of the menu stays like is, just
//     linespacing smaller to make it in up-down scale more compact."
// ---------------------------------------------------------------------------

func TestFilterPanelIsMoreCompactVerticallyOnly(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")

	header := section(t, layout, ".filter-panel__header {", "}", "filter panel header")
	mustContain(t, header, "padding: var(--sp-2) var(--sp-5);",
		"the Filters row must lose half its height and keep its side padding")

	body := section(t, layout, ".filter-panel__body {", "}", "filter panel body")
	mustContain(t, body, "padding: var(--sp-4) var(--sp-5);", "the body must keep --sp-5 at the sides")
	mustContain(t, body, "gap: var(--sp-3);", "the rows must sit closer together")

	// Titles up against what they open, and the rule under each group with them.
	mustContain(t, layout, "padding: 0 0 var(--sp-2);", "a group title must sit tight to its content")
	mustContain(t, layout, ".filter-group { border-bottom: 1px solid var(--border); padding-bottom: var(--sp-3); }",
		"the space under a group must be smaller")
	mustContain(t, layout, ".filter-panel .field { gap: 6px; }",
		"label to control, inside this panel only")

	// The width is untouched — that is the client's one constraint here.
	mustContain(t, layout, ".search-layout {", "the layout rule must still exist")
	panel := section(t, layout, ".filter-panel {", "}", "filter panel")
	mustNotContain(t, panel, "width:", "the panel's width must not have been touched")
}

// ---------------------------------------------------------------------------
// 51. "The label-filters in the main listing page are very good. Just the
//     'Clear all' button replace with red cross (#FF0000) inside rounded
//     rectangle."
// ---------------------------------------------------------------------------

func TestClearAllIsARedCrossInARoundedRectangle(t *testing.T) {
	body := mustGet(t, newServer(t), "/search?deal=sale&property_type=house")

	bar := section(t, body, `class="active-filters"`, "</div>", "active filter bar")
	mustNotContain(t, bar, ">Clear all<", "the word must be gone from the tag bar")
	mustContain(t, bar, `class="clear-filters"`, "a control must have taken its place")
	// A cross alone says nothing to a screen reader or to a first-time visitor.
	mustContain(t, bar, `aria-label="Clear all filters"`, "the control must keep its name")
	mustContain(t, bar, `title="Clear all filters"`, "and its tooltip")

	rule := section(t, asset(t, cssDir+"/components.css"), ".clear-filters {", "}", "clear-filters rule")
	mustContain(t, rule, "color: #FF0000;", "the cross must be the red the client named")
	mustContain(t, rule, "border-radius: var(--r-md);", "inside a rounded rectangle")
	mustContain(t, rule, "width: 34px; height: 34px;", "sized to the chips beside it")
}

// ---------------------------------------------------------------------------
// 52. "This featured heart borderline make #FF0000 (in day and dark theme the
//     same) and if clicked then the whole heart color the same #FF0000."
// ---------------------------------------------------------------------------

func TestHeartsAreThatExactRedInBothThemes(t *testing.T) {
	tokens := asset(t, cssDir+"/tokens.css")
	components := asset(t, cssDir+"/components.css")

	mustContain(t, tokens, "--fav: #FF0000;", "the heart's red must be the literal value")
	// Declared once and never per theme: that is what "in day and dark theme
	// the same" means, and the reason it is not --error, which is redefined.
	if strings.Count(tokens, "--fav:") != 1 {
		t.Error("--fav must be declared once, so no theme can repaint the heart")
	}

	mustContain(t, components, "stroke: var(--fav);", "the outline must use it")
	mustContain(t, components, ".pcard__fav[aria-pressed='true'] svg { fill: var(--fav); }",
		"a saved heart must fill with the same red")
	fav := section(t, components, ".pcard__fav {", "}", "favourite button")
	mustNotContain(t, fav, "var(--error)", "the heart must no longer take the themed error colour")
}

// ---------------------------------------------------------------------------
// 53. "When in main listing page and open the sidebar menu on the right, then
//     the menu on the left disappears — make that it will not disappear. In the
//     layout no changes needed if right sidebar menu opens."
// ---------------------------------------------------------------------------

func TestOpeningTheDrawerLeavesTheFilterPanelWhereItIs(t *testing.T) {
	body := mustGet(t, newServer(t), "/search")

	// The cause: x-trap.noscroll puts overflow:hidden on <html>, which stops
	// the document being a scroll container, and every position:sticky element
	// on the page falls back to its static position. Measured in Chrome at
	// scrollY 900: the panel's top went from 92 to -778 with the modifier, and
	// stays at 92 without it.
	drawer := section(t, body, `<div class="drawer" id="mobile-drawer"`, ">", "account drawer")
	mustContain(t, drawer, `x-trap="menuOpen"`, "the drawer must still trap focus")
	mustNotContain(t, drawer, "x-trap.noscroll", "but must not stop the document scrolling")

	// The panel is still sticky — the fix is that nothing breaks it, not that
	// it was pinned some other way.
	panel := section(t, asset(t, cssDir+"/layout.css"), ".filter-panel {", "}", "filter panel")
	mustContain(t, panel, "position: sticky;", "the panel must still be sticky")
}

// ---------------------------------------------------------------------------
// 54. "This area restrict to a rounded corners rectangle with the left-right
//     border limits the same as the 'work with us' below it is", and
// 55. "All the areas on the screen (except the header part) what stretch from
//     one screen side to another, make inside ronded corners rectangle with
//     limited size — so the same width as the footer."
// ---------------------------------------------------------------------------

func TestColouredBlocksArePanelsAtTheFooterWidth(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")

	// The width formula is the footer's, character for character. If one of
	// them is ever edited the other has to be too, and this is what says so.
	const width = "width: min(100% - var(--gutter) * 2, calc(var(--container) - var(--gutter) * 2));"
	footer := section(t, layout, ".site-footer {", "}", "footer")
	mustContain(t, footer, width, "the footer sets the width every panel matches")

	panel := section(t, layout, ".section--band::before,", "}", "band panel")
	mustContain(t, panel, width, "a coloured block must be exactly as wide as the footer")
	mustContain(t, panel, "border-radius: var(--r-xl);", "with rounded corners")

	// Painted behind the content, and kept inside the section: a negative
	// z-index without the isolation would slide under the page itself.
	mustContain(t, panel, "z-index: -1;", "the panel must sit behind the content")
	band := section(t, layout, ".section--band,\n.section--navy {", "}", "band")
	mustContain(t, band, "isolation: isolate;", "the negative layer must stay inside the section")

	// And room inside it. The panel's edge is a .container's content edge, so
	// without a second gutter the heading and the cards land exactly on it —
	// which is what the client saw next: "the container size corner etc is okay,
	// just adjust the inner boxes, add space between … make sure the text looks
	// good, move away from the left side." One more gutter puts the content
	// where the footer's own links sit inside their panel.
	mustContain(t, layout, "padding-inline: calc(var(--gutter) * 2);",
		"a panel's content must be inset from its edge")
	inset := section(t, layout, ".section--band > .container,", "}", "panel content")
	mustContain(t, inset, ".section--navy > .container", "both kinds of panel, not only the band")

	// Nothing may paint the full width any more.
	mustNotContain(t, layout, ".section--band { background: var(--surface-2); }",
		"a band must no longer paint edge to edge")
	mustNotContain(t, layout, ".section--navy { background: var(--navy); color: var(--text-inverse); }",
		"nor the navy one")
}

// The same rule reaches every page that had a full-width block, not only the
// two the client happened to screenshot.
func TestEveryPageWithABandGetsThePanel(t *testing.T) {
	h := newServer(t)
	for path, what := range map[string]string{
		"/":      "the homepage sections",
		"/about": "the About Previa statistics",
		"/article/buying-in-spain-as-a-non-resident": "related reading",
	} {
		body := mustGet(t, h, path)
		if !strings.Contains(body, "section--band") && !strings.Contains(body, "section--navy") {
			t.Errorf("%s: expected %s to still be a coloured block", path, what)
		}
	}
}

// ---------------------------------------------------------------------------
// 56. "Remove this block around the footer and footer stays into the same
//     background as in other pages. So only the header part stays at the moment
//     to a big area from left to right", and "reduce the spacing between these
//     menu blocks compat together, so move all a bit up."
// ---------------------------------------------------------------------------

func TestFooterStandsOnThePageColourEverywhere(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")

	mustContain(t, layout, ".page__foot { margin-top: auto; padding-top: var(--sp-6); background: var(--bg); }",
		"the footer's ground is the page colour, with a gap above its panel")
	// The rules that used to carry the last section's colour down behind it.
	mustNotContain(t, layout, ".page:has(.page__main > .section--band:last-child) .page__foot",
		"no page may wrap its footer in a band any more")
	mustNotContain(t, layout, ".page:has(.page__main > .section--navy:last-child) .page__foot",
		"nor in the navy one")
}

func TestBlocksSitCloserTogether(t *testing.T) {
	tokens := asset(t, cssDir+"/tokens.css")

	// A third off, which is also a third off the padding inside every panel,
	// since a panel now spans its section's box.
	//
	// Cut again on 19 August — "then move it all more a bit up, more compact
	// together … then look that the distance gaps between the blocks would be
	// the same" — so the figures below are the current ones and the two before
	// them are both named as gone. The property being kept is the direction:
	// the rhythm only ever tightens, and neither of the older ladders may come
	// back by accident.
	mustContain(t, tokens, "--section-y: clamp(24px, 2.6vw, 40px);", "the rhythm must be tighter")
	mustContain(t, tokens, "--section-y-lg: clamp(30px, 3.4vw, 52px);", "and its larger step with it")
	mustNotContain(t, tokens, "clamp(48px, 6vw, 96px)", "the original rhythm must be gone")
	mustNotContain(t, tokens, "clamp(36px, 4.2vw, 64px)", "and the 18 August one with it")
}

// ---------------------------------------------------------------------------
// 57. "In the single ad page this link works perfectly, just make the color of
//     it #0000FF."
// ---------------------------------------------------------------------------

func TestTheAddressLinkIsPureBlue(t *testing.T) {
	components := asset(t, cssDir+"/components.css")

	mustContain(t, components, ".address-link { color: #0000FF; text-decoration: none; }",
		"the address link must be the exact blue the client named")
	// It was --link, which is a different blue in each theme; the client asked
	// for one value, so the theme token must no longer be in the rule.
	rule := section(t, components, ".address-link {", "}", "address link")
	mustNotContain(t, rule, "var(--link)", "the link must not take the themed blue any more")

	// And it must still be the directions link it was: the client's "works
	// perfectly" is about the behaviour, which nothing here may change.
	body := mustGet(t, newServer(t), "/property/apartment-in-a-new-helsinki-development-pr-034")
	link := section(t, body, `class="address-link"`, "</a>", "address link markup")
	mustContain(t, link, "google.com/maps/dir/", "the address must still start navigation")
}

// ---------------------------------------------------------------------------
// 58. "In single ad page this 'view all 5 photos' make just '5 photos' (the
//     photo camera icon stays there as well) and the backgroun color of this
//     button is bad, make it better. Make it the same half-way transparent like
//     the choose your market button in the frontpage banner."
// ---------------------------------------------------------------------------

func TestThePhotosButtonIsShorterAndTakesTheMarketChipsTone(t *testing.T) {
	body := mustGet(t, newServer(t), "/property/apartment-in-a-new-helsinki-development-pr-034")

	button := section(t, body, `class="btn btn--sm gallery__open"`, "</button>", "gallery button")
	mustContain(t, button, "5 photos", "the label is the count and the word")
	mustContain(t, button, "<svg", "the camera icon stays")

	// The words left the button face but not the accessible name — a screen
	// reader still gets the whole phrase. Counted rather than searched for,
	// because the name is inside the element the visible label is in: exactly
	// one occurrence means it survives as the name and nowhere else.
	mustContain(t, body, `aria-label="View all 5 photos"`, "the full phrase must survive as the name")
	if n := strings.Count(body, "View all 5 photos"); n != 1 {
		t.Errorf("expected the long phrase once, as the accessible name; found it %d times", n)
	}

	pages := asset(t, cssDir+"/pages.css")
	chip := section(t, pages, ".market-picker--light .market-picker__btn {", "}", "market picker chip")
	open := section(t, pages, ".gallery__open {", "}", "gallery button rule")
	for _, want := range []string{
		"background: rgba(12, 45, 72, 0.55);",
		"backdrop-filter: blur(8px);",
	} {
		mustContain(t, chip, want, "the market chip must still be the reference")
		mustContain(t, open, want, "the photos button must take the same treatment")
	}
	mustNotContain(t, body, "btn--on-image btn--sm gallery__open",
		"the button must not keep the washed-out over-image treatment")
}

// ---------------------------------------------------------------------------
// 59. "If the seller has chosen he's googlemaps location, then in the single ad
//     page where the seller/broker profile is on the right, make this location
//     logo icon with the text what the seller entered into the field 'edit your
//     location like you want other users to see it' … So under the seller's
//     profile the googlemaps loction add the same googlemaps fields as in the
//     'add listing' page."
// ---------------------------------------------------------------------------

func TestTheListingShowsTheSellersOwnWordsForWhereTheyAre(t *testing.T) {
	body := mustGet(t, newServer(t), "/property/apartment-in-a-new-helsinki-development-pr-034")

	place := section(t, body, `class="seller-place"`, "</p>", "seller location line")
	mustContain(t, place, "<svg", "the pin marks the line")
	mustContain(t, place, "Helsinki, Finland", "and the seller's own text follows it")

	// Never the street the pin sits on: that is the whole point of the field.
	mustNotContain(t, body, "Aleksanterinkatu 17", "the office street must not be published")
}

func TestTheProfileCarriesTheAddListingLocationFields(t *testing.T) {
	body := mustGet(t, newServer(t), "/settings")
	wizard := mustGet(t, newServer(t), "/add-listing")

	// The one editable line, in the same green box the wizard marks it with.
	for _, want := range []string{
		`class="field field--editable-location"`,
		"Edit your location as you want other users to see it",
	} {
		mustContain(t, wizard, want, "the wizard must still be the reference")
		mustContain(t, body, want, "the profile must offer the same editable line")
	}
	mustContain(t, body, `name="office_public"`, "and submit it")
	mustContain(t, body, "Rotermanni quarter, Tallinn", "prefilled from what is saved")

	// And the geocoder's own reading of the pin, read-only, under it.
	mustContain(t, body, ">From the map<", "the map's data must be shown")
	for _, want := range []string{
		">Country<", ">Country code<", ">State / region<", ">City<",
		">District<", ">Street / address<", ">Latitude<", ">Longitude<",
	} {
		mustContain(t, body, want, "the profile must show "+want+" as the wizard does")
	}
	mustContain(t, body, `id="p-office-country" type="text" readonly`,
		"a field that follows the pin must not be typeable")

	// The component that fills them.
	js := asset(t, jsDir+"/previa.js")
	mustContain(t, js, "publicLabel: start.Public || '',",
		"the public line must be seeded from what is saved")
	mustContain(t, js, "if (this.lat || this.lng) this.lookup(this.lat, this.lng);",
		"a profile that already has a pin must open with its fields filled")
}

// ---------------------------------------------------------------------------
// 60. "In the filter image you can see that where the word or text Austria in
//     location search box, the remove like cross is not in proper place."
// ---------------------------------------------------------------------------

// The cross is a child of .input-icon, whose rule absolutely positions the icon
// it was written for — the leading pin — at left: var(--sp-4). That reached the
// cross too and took it out of its button's flex centring: measured in Chrome,
// the glyph sat 11px right of the 22px button's centre, hanging past its edge
// and over the field's border. It also inherited that rule's colour, so it
// never brightened on hover.
func TestTheLocationFieldsClearButtonIsCentredInIt(t *testing.T) {
	components := asset(t, cssDir+"/components.css")

	rule := section(t, components, ".location-field__clear svg {", "}", "the clear glyph")
	mustContain(t, rule, "position: static;", "the glyph must sit in the button's own flex centring")
	mustContain(t, rule, "color: inherit;", "and take the button's colour, including on hover")

	// The button itself is unchanged: 22px, centred on the field's right edge.
	// Both numbers now come from the component's own tokens, because the input's
	// right-hand padding is calculated from them — it has to leave the button
	// room plus one gap — and two places holding the same figure is how the
	// homepage panel came to overlap them. The tokens are checked below, so the
	// guarantee this test was written for is unchanged: 8px in, 22px across.
	btn := section(t, components, ".location-field__clear {", "}", "the clear button")
	mustContain(t, btn, "right: var(--loc-clear-inset);", "8px inside the field's right edge")
	mustContain(t, btn, "width: var(--loc-clear-size); height: var(--loc-clear-size);",
		"and 22px across")
	mustContain(t, btn, "top: 50%;", "and vertically centred")
	mustContain(t, btn, "justify-content: center;", "with its content centred in it")

	field := section(t, components, ".location-field {", "}", "the location field")
	mustContain(t, field, "--loc-clear-inset: var(--sp-2);", "the inset token is still 8px")
	mustContain(t, field, "--loc-clear-size: 22px;", "the size token is still 22px")

	// The rule it was inheriting from is still there for the pin it belongs to.
	mustContain(t, components, ".input-icon svg {", "the leading icon keeps its own positioning")
}
