package app_test

import (
	"regexp"
	"strings"
	"testing"
)

// Tests for the client's 17 August corrections.
//
// Thirteen notes, thirteen sections, each named after what was asked rather
// than after the code that answers it — the split the earlier rounds use.
// Anything the server renders is asserted against a real response; the parts
// that only exist once a browser has laid the page out or run JavaScript are
// asserted at the level of the rule or the handler that produces them, and were
// driven in a real headless Chrome as well. Where a number was measured there,
// it is quoted in the comment so a later change that moves it is obvious.

// ---------------------------------------------------------------------------
// 1. "when the heart in the ad is not checked, then make it's background
//    transparent, the borderline red and background transparent ... So that
//    only the red heart bordereline on the image and that's it."
//
// Covered by TestHeartIsRedOutlinedUntilSaved in feedback_16aug_test.go, which
// was the test asserting the grey body this note replaced. Amending it there
// rather than adding a second, contradictory test is what keeps the suite
// describing one interface.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 2. "the green earth is wrong shape. Take it from the carwelt logo, there are
//    the continents exactly the right shape. The 3 houses are on the right
//    place, these stay like they are. But they have 2 borderlines, remove the
//    borderlines. There is only the background color on the circle. And this
//    background color make blue red and green."
// ---------------------------------------------------------------------------

func TestGlobeCarriesRealContinents(t *testing.T) {
	logo := asset(t, "../../web/templates/components/logo.html")

	// The land is drawn as many separate coastlines rather than four blobs.
	// Each landmass is its own path inside the group filled with the land
	// gradient, so a count of paths there is a count of landmasses.
	land := section(t, logo, `<g fill="url(#pvN{{ $uid }})">`, "</g>", "the landmasses")
	if n := strings.Count(land, "<path"); n < 6 {
		t.Errorf("the globe carries %d landmasses; the client asked for real "+
			"continents, which needs at least Europe, Africa, Asia and the "+
			"larger islands drawn separately", n)
	}

	// Real coastlines have far more vertices than a four-blob approximation.
	// The old drawing was four Bézier paths of about a dozen points each; this
	// is the floor a projected outline clears comfortably and a hand-drawn blob
	// cannot.
	if n := strings.Count(land, " L"); n < 250 {
		t.Errorf("the landmasses have %d line segments; a coastline traced from "+
			"real geography has hundreds, a decorative blob has tens", n)
	}
}

func TestPropertyDiscsAreOneFlatColourEach(t *testing.T) {
	logo := asset(t, "../../web/templates/components/logo.html")
	favicon := asset(t, imgDir+"/favicon.svg")

	badges := section(t, logo, `<g class="logo__badges">`, "</g>", "the property discs")

	// Three discs, three flat fills, and the client's three colours.
	circles := regexp.MustCompile(`<circle [^>]*fill="(#[0-9A-Fa-f]{6})"`).FindAllStringSubmatch(badges, -1)
	if len(circles) != 3 {
		t.Fatalf("found %d flat-filled discs, want 3", len(circles))
	}
	got := map[string]bool{}
	for _, m := range circles {
		got[strings.ToUpper(m[1])] = true
	}
	// The client named the exact values on 18 August, replacing the muted
	// blue, red and green this note first asked for: "the house on the green
	// background, this background make #7CFC00, the red background make
	// #FF0000 and the blue make #FF00FF."
	for _, want := range []string{"#FF00FF", "#FF0000", "#7CFC00"} {
		if !got[want] {
			t.Errorf("disc colour %s is missing; the client named these three exactly", want)
		}
	}

	// No gradient survives on a disc: "there is only the background color on
	// the circle".
	mustNotContain(t, badges, "url(#pvG", "a disc must be a flat colour, not a gradient")
	mustNotContain(t, logo, "pvG{{ $uid }}", "the disc gradient is unused and must not be defined")

	// And nothing rings one. The collar, the navy hairline and the warm sheen
	// arc were the "2 borderlines" the client counted; all three were drawn
	// from disc centres, so no stroked circle may share a disc's centre.
	for _, m := range regexp.MustCompile(`<circle cx="([\d.]+)" cy="([\d.]+)"`).FindAllStringSubmatch(badges, -1) {
		ring := regexp.MustCompile(
			`<circle cx="` + regexp.QuoteMeta(m[1]) + `" cy="` + regexp.QuoteMeta(m[2]) +
				`"[^>]*stroke=`)
		if ring.MatchString(logo) {
			t.Errorf("a disc at %s,%s still has a stroked ring around it", m[1], m[2])
		}
	}
	mustNotContain(t, logo, `stroke="#FFF3D6"`, "the sheen arc around each disc is one of the borderlines")

	// One drawing, two files.
	for _, m := range circles {
		mustContain(t, favicon, m[1], "the favicon must carry the same disc colours as the mark")
	}
	mustNotContain(t, favicon, "pvGf", "the favicon must not keep the dropped disc gradient")
}

func TestHousesAreUnmoved(t *testing.T) {
	// "The 3 houses are on the right place, these stay like they are."
	//
	// The disc centres are what position the houses, and they are computed from
	// the same three (angle, distance, radius) placements as before. Pinning
	// the centres pins the houses.
	logo := asset(t, "../../web/templates/components/logo.html")
	badges := section(t, logo, `<g class="logo__badges">`, "</g>", "the property discs")

	for _, want := range []string{
		`cx="19.3" cy="19.7" r="7"`,   // upper left, overhanging
		`cx="48.4" cy="24" r="6.4"`,   // right, overhanging
		`cx="30.3" cy="44.3" r="6.6"`, // lower, inside
	} {
		mustContain(t, badges, want, "a disc, and so its house, must not have moved")
	}
}

// ---------------------------------------------------------------------------
// 3. "in the narrower screen the frontpage search menu brokes, make it nice
//    there - otherwise this search menu is good"
//
// The placement rules are asserted in TestMiniSearchIsTwoRowsOnPhones
// (update_test.go), which is where the phone layout is described. This is the
// bug itself: two controls that shared a class were assigned one cell.
// ---------------------------------------------------------------------------

func TestHeroPickersAreDistinguishable(t *testing.T) {
	h := newServer(t)
	home := mustGet(t, h, "/")

	panel := section(t, home, `class="searchbox__body searchbox__body--three"`,
		`class="searchbox__footer"`, "the hero search panel")

	// The two pickers must be tellable apart by class alone, or no grid rule
	// can place them in different cells.
	mustContain(t, panel, `class="field type-picker deal-picker"`,
		"the deal picker needs a class of its own")

	pickers := regexp.MustCompile(`class="field type-picker[^"]*"`).FindAllString(panel, -1)
	if len(pickers) != 2 {
		t.Fatalf("found %d pickers in the hero panel, want 2 (deal and property type)", len(pickers))
	}
	if pickers[0] == pickers[1] {
		t.Error("the two pickers carry identical class lists, so no rule can " +
			"place them in different grid cells — this is exactly what drew " +
			"them on top of each other on a narrow screen")
	}
}

// ---------------------------------------------------------------------------
// 4–5. "in the map page these ad previews what open ... make the same system as
//      in the frontpage previews, that if move the mouse to the right side of
//      the image then these left-right arrows will come and can change the
//      images. At the moment there is only finger and no matter where click
//      this ad will open in single page"
//
//      "and the image scrolling by holding down the green dot as well, like in
//      the frontpage"
//
// Driven in Chrome: with a popup open, a click on the right-hand zone advances
// the track to translate3d(-100%,0,0) and moves the active dot; a press on the
// dots and an 80px drag right advances three pictures and leaves the strip in
// its held state.
// ---------------------------------------------------------------------------

func TestMapPopupHasTheCardsPagingZones(t *testing.T) {
	js := asset(t, jsDir+"/previa.js")
	pages := asset(t, cssDir+"/pages.css")

	// The zones exist and page in both directions.
	mustContain(t, js, "map-popup__zone map-popup__zone--prev", "the popup needs a previous-photo zone")
	mustContain(t, js, "map-popup__zone map-popup__zone--next", "the popup needs a next-photo zone")

	// A click on one is handled, and handled as paging rather than as a follow
	// of the link underneath it.
	handler := section(t, js, "function popupPaging(container)", "\n    }", "the popup click handler")
	mustContain(t, handler, ".map-popup__zone", "a click on a zone must be picked up")
	mustContain(t, handler, "e.preventDefault()", "a click on a zone must not open the listing")

	// The zones carry the property card's directional cursors, which is what
	// makes the behaviour visible before anything is clicked.
	prev := section(t, pages, ".map-popup__zone--prev", "}", "the previous zone rule")
	mustContain(t, prev, "cursor: w-resize", "the left zone must read as paging backwards")
	next := section(t, pages, ".map-popup__zone--next", "}", "the next zone rule")
	mustContain(t, next, "cursor: e-resize", "the right zone must read as paging forwards")

	// The middle of the photograph still opens the listing.
	mustContain(t, js, "map-popup__media-link", "the middle of the image must still open the ad")

	// And the pager fades in on hover, as it does on a card, rather than
	// sitting on the photograph permanently.
	pager := section(t, pages, ".map-popup__pager {", "}", "the popup pager")
	mustContain(t, pager, "opacity: 0", "the pager must be revealed on hover")
	mustContain(t, pages, ".map-popup__media:hover .map-popup__pager",
		"hovering the photograph is what brings the arrows in")
}

func TestMapPopupDotsScrub(t *testing.T) {
	js := asset(t, jsDir+"/previa.js")
	pages := asset(t, cssDir+"/pages.css")

	scrub := section(t, js, "function popupScrubbing(container)", "\n    }\n", "the popup scrub handler")
	for _, needle := range []string{
		"pointerdown", "pointermove", "pointerup", "pointercancel",
		"setPointerCapture", "DOT_DRAG_SLOP", "DOT_DRAG_STEP",
	} {
		mustContain(t, scrub, needle, "the popup drag must be the card's gesture, not a new one")
	}

	// The same two constants drive the card and the popup, so the two gestures
	// cannot come to feel different.
	mustContain(t, js, "var DOT_DRAG_STEP = 22", "one distance-per-picture for both carousels")
	mustContain(t, js, "var DOT_DRAG_SLOP = 6", "one drag threshold for both carousels")

	// A vertical swipe must still scroll the page, which is what pan-y buys.
	dots := section(t, pages, ".map-popup__dots {", "}", "the popup dot strip")
	mustContain(t, dots, "touch-action: pan-y",
		"a vertical swipe over the dots must still scroll the page")
	mustContain(t, dots, "cursor: grab", "the strip must read as draggable")
	mustContain(t, pages, ".map-popup__dots.is-holding .map-popup__dot.is-on",
		"the held dot must grow, as it does on a card")

	// The handlers are delegated from the map container, not bound per popup.
	// Leaflet rebuilds a popup's inner DOM after popupopen — Marker.openPopup
	// passes a latlng, Popup.setLatLng calls update(), and update() reassigns
	// the content node's innerHTML — so anything bound at popupopen is thrown
	// away a moment later, which is what made the first attempt inert.
	mustContain(t, js, "popupScrubbing(map.getContainer())",
		"the scrub handlers must survive Leaflet rebuilding the popup's contents")
	mustNotContain(t, js, "map.on('popupopen'",
		"binding at popupopen loses the listeners on the next update()")
}

// ---------------------------------------------------------------------------
// 6. "in map page overall all good, but sometimes in list view can not scroll
//    down any more, and grid view is broken, there should be 3 or 4 ads in one
//    row. Make that it would always work and with every screen size the layout
//    would be fine"
// ---------------------------------------------------------------------------

// Both halves had one cause. The fragment this endpoint returns is itself a
// <div id="results">, and it was being swapped *into* #results, so every filter
// change buried another copy of the wrapper inside the last one. The nested
// wrapper is a .map-results__list in its own right: in grid mode it took a 50%
// flex basis from the list it landed in and squeezed the whole result set into
// half the panel, and in list mode it became a second scroll container with
// nothing to scroll, whose overscroll-behavior: contain stopped the wheel
// reaching the real one underneath.
//
// Verified in Chrome: after a filter change on /search?view=map there is
// exactly one #results and one .map-results__list on the page.
func TestResultsFragmentReplacesRatherThanNests(t *testing.T) {
	h := newServer(t)
	filters := asset(t, "../../web/templates/components/filters.html")
	results := asset(t, "../../web/templates/components/results.html")

	form := section(t, filters, `<form id="filter-form"`, ">", "the filter form")
	mustContain(t, form, `hx-swap="outerHTML"`,
		"the response is a #results, so it must replace the target rather than fill it")
	mustNotContain(t, form, `hx-swap="innerHTML"`, "innerHTML is what nested the wrapper")

	// The sort control is one shared definition since the 18 August evening
	// round — note 59 put the same menu on the map page — so its id is a
	// parameter now and the block is found by its define rather than by the
	// literal id="sort" this used to match on.
	sort := section(t, results, `{{ define "sort-select" }}`, "</select>", "the sort control")
	mustContain(t, sort, `hx-swap="outerHTML"`, "the sort control swaps the same fragment")

	// The endpoint really does return the wrapper, in both modes — which is why
	// the swap has to be outerHTML.
	for _, path := range []string{
		"/search/results?view=map&deal=sale",
		"/search/results?view=grid&deal=sale",
	} {
		body := mustGet(t, h, path)
		if !strings.Contains(body, `id="results"`) {
			t.Errorf("%s: the fragment must carry the #results wrapper", path)
		}
		if n := strings.Count(body, `id="results"`); n != 1 {
			t.Errorf("%s: the fragment carries %d #results wrappers, want 1", path, n)
		}
	}
}

// The grid is a real grid, sized by how wide the results column actually is.
//
// Measured in Chrome, with the results column widened by is-grid-mode: 4 cards
// per row at 1920 and 1440, 3 at 1280 and 1024, 2 at 480. Cards keep their full
// height in every one — the row track was collapsing to 73px before, because
// `.pcard` sets overflow: hidden and Chromium sizes an auto grid row against a
// scroll-container item as though its contents were not there.
func TestMapGridGivesThreeOrFourPerRow(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	// The column count is a function of the container, not the viewport: the
	// two differ by the width of the map on a desktop and by nothing on a phone.
	mustContain(t, pages, "container-type: inline-size",
		"the results column must be a query container")
	mustContain(t, pages, "container-name: map-results", "…with a name the grid rules can query")

	grid := section(t, pages, ".map-shell.is-grid-mode .map-results__list {", "}", "the grid")
	mustContain(t, grid, "display: grid", "a wrapped flex row cannot describe three column counts")
	mustContain(t, grid, "repeat(2, minmax(0, 1fr))", "two per row is the narrow default")

	for _, want := range []string{
		"@container map-results (min-width: 560px)",
		"repeat(3, minmax(0, 1fr))",
		"@container map-results (min-width: 800px)",
		"repeat(4, minmax(0, 1fr))",
	} {
		mustContain(t, pages, want, "the client asked for three or four cards in a row")
	}

	// The card must not be a scroll container here, or the row that holds it
	// cannot be sized from its contents.
	card := section(t, pages, ".map-shell.is-grid-mode .map-results__list .pcard--map {", "}", "the grid card")
	mustContain(t, card, "overflow: visible",
		"an overflow-hidden grid item makes Chromium size the row without its contents")

	// And the column widens in grid mode, or three cards could not fit at a
	// readable width.
	mustContain(t, pages, ".map-shell.is-grid-mode { grid-template-columns: minmax(420px, 56%)",
		"the results column must widen to fit three or four cards")
}

// The layout class lives on an element HTMX never replaces.
//
// It was bound on #results itself. That is the one element the fragment swap
// replaces, and the incoming copy came back without the class: the toolbar said
// Grid and the results were laid out as a list.
func TestGridModeIsBoundAboveTheSwappedRegion(t *testing.T) {
	split := asset(t, "../../web/templates/pages/search-map-split.html")
	results := asset(t, "../../web/templates/components/results.html")

	mustContain(t, split, "'is-grid-mode': listMode === 'grid'",
		"the layout class belongs on the shell, which is never swapped")

	region := section(t, results, `<div id="results" class="map-results__list`, "</div>",
		"the swapped results element")
	mustNotContain(t, region, ":class",
		"binding the layout on the swapped element loses it on the next filter change")
}

// Nothing on the map screen scrolls the page, so the shell has to be exactly
// the height of what is left below the header — measured with dvh, since 100vh
// on a phone is the height with the browser chrome retracted and left the last
// cards unreachable.
func TestMapShellUsesDynamicViewportHeight(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")
	shell := section(t, pages, ".map-shell {", "}", "the map shell")
	mustContain(t, shell, "height: calc(100vh - var(--header-h))", "the fallback for older browsers")
	mustContain(t, shell, "height: calc(100dvh - var(--header-h))",
		"the shell must measure what is actually on screen")
}

// ---------------------------------------------------------------------------
// 7. "this avatar, what at the moment is round, make rectangle with rounded
//    corners. then this text there 'JPEG or PNG, at least 400×400' make
//    'profile picture rec. dimensions 400 x 400' there is no point to mention
//    the jpeg or png"
// ---------------------------------------------------------------------------

func TestProfilePictureIsARoundedRectangle(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")
	components := asset(t, cssDir+"/components.css")

	mustContain(t, settings, `class="avatar-tile`, "the profile picture must be the rectangular tile")
	mustNotContain(t, settings, `class="avatar avatar--xl"`, "the round avatar is what the client asked to replace")

	tile := section(t, components, ".avatar-tile {", "}", "the profile picture tile")
	mustContain(t, tile, "border-radius: var(--r-md)", "rectangle with rounded corners")
	mustNotContain(t, tile, "border-radius: 50%", "a circle is what was asked to go")

	// The wording, and nothing of the wording it replaced.
	mustContain(t, settings, "Profile picture rec. dimensions 400 × 400", "the client's exact replacement")
	mustNotContain(t, settings, "JPEG or PNG", "the file types are not worth mentioning — everything is converted to webp")
}

// ---------------------------------------------------------------------------
// 8. "make there under user profile settings, where the profile picture upload
//    is, there make 'your company logo' where the brokers can upload their
//    compay logo, it is rectangle with rounded corners and a bit bigger. In the
//    single ad page it will be displayed under 'view listings of this broker'"
// ---------------------------------------------------------------------------

func TestCompanyLogoUploadSitsBesideTheProfilePicture(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")
	components := asset(t, cssDir+"/components.css")

	uploads := section(t, settings, `class="upload-row"`, `class="grid grid--2"`, "the upload row")
	mustContain(t, uploads, "Profile picture", "the two uploads sit together")
	mustContain(t, uploads, "Your company logo", "the client's exact wording")
	mustContain(t, uploads, "Upload logo", "the logo needs its own control")

	// Rounded rectangle, and larger than the avatar beside it.
	logo := section(t, components, ".logo-tile {", "}", "the company logo tile")
	mustContain(t, logo, "border-radius: var(--r-md)", "rectangle with rounded corners")
	mustContain(t, logo, "width: 160px", "a logo is wider than tall — a square frame crops a wordmark")
	mustContain(t, logo, "object-fit: contain", "somebody else's logo is never cropped")
}

func TestCompanyLogoAppearsUnderTheBrokerLink(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/property/altbau-apartment-with-stucco-ceilings-in-prenzlauer-berg-pr-007")

	// The client placed it precisely: under "view listings of this broker".
	after := section(t, body, "View all listings from this broker", "</aside>",
		"the foot of the seller box")
	mustContain(t, after, `class="seller-logo"`, "the company logo goes below the broker link")

	// And it is genuinely optional, so a seller without one gets no empty frame.
	tmpl := asset(t, "../../web/templates/pages/property-detail.html")
	mustContain(t, tmpl, "{{ if $d.Broker.CompanyLogo }}",
		"a broker who has not uploaded a logo must not get an empty frame")
}

// ---------------------------------------------------------------------------
// 9. "in the frontpage where are the ad banners of the brokers (a payd
//    service), make the broker profile picture the same size (only with rounded
//    corners) as in kinnisvara24"
// ---------------------------------------------------------------------------

func TestPromotedBrokerPhotoFillsTheCard(t *testing.T) {
	h := newServer(t)
	home := mustGet(t, h, "/")
	components := asset(t, cssDir+"/components.css")

	strip := section(t, home, `class="broker-strip"`, "</section>", "the promoted broker strip")
	mustContain(t, strip, `class="bcard__photo"`, "the photograph is the card's own block now")
	mustNotContain(t, strip, `class="bcard__avatar"`, "the 88px medallion is what the client asked to replace")

	photo := section(t, components, ".bcard__photo {", "}", "the broker photograph")
	mustContain(t, photo, "width: 100%", "the reference fills the top of the card")
	mustContain(t, photo, "aspect-ratio: 4 / 5", "a portrait crop, as the reference has")
	mustContain(t, photo, "object-position: center top", "keep the face in frame whatever the source ratio")

	// "Only with rounded corners": the top two follow the card, the bottom two
	// are square because the card carries on underneath.
	media := section(t, components, ".bcard__media {", "}", "the photograph's frame")
	mustContain(t, media, "border-radius: calc(var(--r-lg) - 1px) calc(var(--r-lg) - 1px) 0 0",
		"the rounded corners the client asked for, on the two that meet the card's edge")
	mustNotContain(t, media, "border-radius: 50%", "the medallion is gone")
}

// ---------------------------------------------------------------------------
// 10. "under user's profile the 'full name' make 'Name' as full name is not
//     obligatory. Then next to name field make 'your company' ... if he does
//     it, then this is displayed in the frontend next to he's name."
// ---------------------------------------------------------------------------

func TestNameAndCompanyFields(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")

	mustContain(t, settings, `<label class="field__label" for="p-name">Name</label>`,
		"a full name is not obligatory, so the field is not called one")
	mustNotContain(t, settings, "Full name", "the old label must be gone")

	// The company field, beside the name and optional.
	company := section(t, settings, `for="p-company"`, "</div>", "the company field")
	mustContain(t, company, "Your company", "the client's exact wording")
	mustContain(t, company, "Best House Ltd", "the client's own example, as the placeholder")

	// The two sit next to each other in the same two-column row.
	row := section(t, settings, `for="p-name"`, `for="p-email"`, "the name row")
	mustContain(t, row, "p-company", "the company field goes next to the name field")
}

// ---------------------------------------------------------------------------
// 11. "add more field 'Languages of communication'. There the user can choose
//     in dropdown menu in all the languages with flags (the same menu as choose
//     your market in the frontpage). Then he can add these languages there as
//     tag's and remove them if needed. And then into the search-filter menu at
//     the end make the options for 'Language of communication' ... This option
//     is not obligatory."
// ---------------------------------------------------------------------------

func TestLanguagesOfCommunicationArePickedAsTags(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")

	picker := section(t, settings, "Languages of communication", `id="p-countries-label"`,
		"the languages picker")

	// The market picker's menu: a search box over a flagged list.
	mustContain(t, picker, "market-menu__search", "the same menu as choose your market")
	mustContain(t, picker, "market-menu__list", "…including its scrolling list")
	mustContain(t, picker, `class="flag"`, "all the languages with flags")

	// Tags that can be removed.
	mustContain(t, picker, "tagpick__tag", "chosen languages appear as tags")
	mustContain(t, picker, "tagpick__remove", "…and can be removed again")

	// The checkboxes are the form state: what the menu holds is what the form
	// sends, with no hidden inputs to keep in step with a parallel array.
	mustContain(t, picker, `type="checkbox" class="tagpick__input" name="languages"`,
		"the options must be real form controls")

	// The seeded account's languages arrive ticked.
	if !regexp.MustCompile(`name="languages" value="et"[^>]*checked`).MatchString(picker) {
		t.Error("a language already chosen must arrive ticked")
	}
}

func TestLanguageFilterIsLastAndOptional(t *testing.T) {
	h := newServer(t)
	page := mustGet(t, h, "/search")

	// Last in the panel, as the client placed it: nothing but the footer
	// between it and the end of the form.
	tail := section(t, page, "Language of communication", "</form>", "the end of the filter panel")
	mustNotContain(t, tail, "filter-group__toggle", "the language filter must be the last group")

	mustContain(t, tail, `name="language"`, "the filter submits a language parameter")
	mustContain(t, tail, `class="flag"`, "the options carry their flags")

	// Optional: with nothing selected, nothing is removed.
	all := mustGet(t, h, "/search?deal=sale")
	one := mustGet(t, h, "/search?deal=sale&language=de")
	two := mustGet(t, h, "/search?deal=sale&language=de&language=et")

	count := func(body string) string {
		m := regexp.MustCompile(`<strong class="numeric">(\d+)</strong>`).FindStringSubmatch(body)
		if m == nil {
			t.Fatal("no result count in the response")
		}
		return m[1]
	}
	if count(all) == count(one) {
		t.Error("selecting a language must narrow the results")
	}
	if count(two) == count(one) {
		t.Error("languages must be OR-ed: adding a second must widen the results, not narrow them")
	}

	// A hand-edited URL cannot filter on a language that exists nowhere.
	bogus := mustGet(t, h, "/search?deal=sale&language=zz")
	if count(bogus) != count(all) {
		t.Error("an unknown language code must be ignored, not obeyed")
	}
}

// One catalogue serves the profile picker, the search filter, the broker
// directory and the badges on a profile — so a language chosen in one is
// findable in the others. The directory used to carry its own hand-written list
// of English names in the handler, compared against names on the broker record.
func TestOneLanguageCatalogue(t *testing.T) {
	h := newServer(t)

	// The directory offers codes now, and filtering by one returns brokers.
	// They are checkboxes in the market picker's menu rather than a select
	// since the 17 August evening round — see TestBrokerLanguageIsAMultiSelect
	// in feedback_17aug_night_test.go — but they are the same catalogue codes.
	brokers := mustGet(t, h, "/brokers")
	mustContain(t, brokers, `value="cs"`, "the directory offers language codes")
	mustContain(t, brokers, "Language of communication", "…under the client's wording")

	filtered := mustGet(t, h, "/brokers?language=cs")
	if !strings.Contains(filtered, "bcard__name") {
		t.Error("filtering the directory by a language a broker speaks must return them")
	}

	// A broker's own profile names the languages from the same catalogue.
	profile := mustGet(t, h, "/broker/petra-novak")
	langs := section(t, profile, "Languages of communication", "Specialisations", "the language badges")
	mustContain(t, langs, "Czech", "a code must be rendered as its name")
	mustNotContain(t, langs, ">cs<", "a raw code must never reach the page")
}

// ---------------------------------------------------------------------------
// 12. "in the account settings country menu make all countries there with flags
//     (the same menu as choose your market) and the user can multi-select them,
//     as nowadays some brokers (living near the border) can be active in more
//     countries at once. These 'active in' countries will be displayed under
//     the user profile in the frontend."
// ---------------------------------------------------------------------------

func TestCountryIsAMultiSelectOfEveryCountry(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")

	picker := section(t, settings, `id="p-countries-label"`, `type="submit"`, "the country picker")

	// Not a <select> of eight seeded markets any more.
	mustNotContain(t, picker, `<select class="select" id="p-country">`,
		"the single-choice select is what the client asked to replace")
	mustContain(t, picker, `name="active_countries"`, "several countries can be chosen")
	mustContain(t, picker, "market-menu__search", "the same menu as choose your market")

	// Every country, not only the markets carrying listings: a broker can be
	// active somewhere Previa has no stock yet.
	for _, code := range []string{"EE", "FI", "DE", "FR", "PL", "SE"} {
		if !strings.Contains(picker, `value="`+code+`"`) {
			t.Errorf("the country menu must offer %s — the client asked for all countries", code)
		}
	}

	// And the seeded account's two arrive ticked.
	for _, code := range []string{"EE", "FI"} {
		if !regexp.MustCompile(`value="` + code + `"[^>]*checked`).MatchString(picker) {
			t.Errorf("%s is one of the account's active markets and must arrive ticked", code)
		}
	}
}

func TestActiveInIsShownOnTheProfile(t *testing.T) {
	h := newServer(t)

	// br-02 is seeded active either side of the gulf, which is the client's own
	// example of why one country was not enough.
	profile := mustGet(t, h, "/broker/marten-sepp")
	block := section(t, profile, ">Active in<", "Languages of communication", "the active-in block")
	for _, want := range []string{"Estonia", "Finland"} {
		mustContain(t, block, want, "every market the broker works in must be listed")
	}
	mustContain(t, block, `class="flag"`, "markets are drawn with their flags, as everywhere else")
}

// ---------------------------------------------------------------------------
// 13. "under the phone make the social icons with checkboxes and Signal /
//     Telegram adress bars like in sexydate. So each user can check with social
//     apps he is using and add Signal / Telegram link or username ... The phone
//     number is conneced to Viber and WhatsApp anyway."
// ---------------------------------------------------------------------------

func TestMessengerTogglesSitUnderThePhoneField(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")

	// Directly under the phone field, which is the arrangement in the
	// reference — and the same control the add-listing form already uses, so
	// the two cannot drift apart.
	block := section(t, settings, `for="p-phone"`, "</div>", "the phone field")
	mustContain(t, block, `class="msg-picker"`, "the messenger row belongs under the phone number")
	for _, kind := range []string{"whatsapp", "telegram", "viber", "signal", "teams"} {
		mustContain(t, block, `name="messengers" value="`+kind+`"`,
			"every app the client listed must be offered")
	}
	mustContain(t, block, "msg-toggle__box", "each app is a tickbox beside its brand tile")

	// The apps the account already uses arrive ticked.
	for _, kind := range []string{"whatsapp", "viber", "telegram"} {
		if !regexp.MustCompile(`value="` + kind + `"[^>]*checked`).MatchString(block) {
			t.Errorf("%s is enabled on the account and must arrive ticked", kind)
		}
	}

	// Telegram and Signal have their own address fields, refilled from the
	// account — a Telegram account is not always findable by number and
	// Signal's is a share link rather than a number.
	mustContain(t, settings, `id="p-telegram"`, "Telegram needs its own address field")
	mustContain(t, settings, `id="p-signal"`, "Signal needs its own address field")
	mustContain(t, settings, `value="t.me/annalehtinen"`, "a saved Telegram link must come back")

	// WhatsApp and Viber need no handle: they are reached on the phone number.
	mustContain(t, settings, "reached on the number above",
		"the form should say why WhatsApp and Viber have no field")
}
