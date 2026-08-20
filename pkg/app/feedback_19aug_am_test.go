package app_test

import (
	"regexp"
	"strings"
	"testing"
)

// Tests for the client's 19 August notes (10:59–11:50).
//
// Nine messages, numbered 81–91 on from the eighty before them, each section
// named after what was asked rather than after the code that answers it.
//
// Three of them are one figure each and are checked as one; the two advertising
// dialogs are the largest thing in the round and carry a section apiece,
// because "a popup with a bill in it" is a specification with about six
// separable promises in it and a single assertion would only ever prove one.

// ---------------------------------------------------------------------------
// 81. "In the frontpage the articles section the article previews make the same
//     style and size as they are in article page. So in the frontpage as well 5
//     pcs in one row and two rows of them. Then in the frontpage in the articles
//     block move all the content more up compact together. At the moment there
//     is too much empty room."
// ---------------------------------------------------------------------------

func TestTheHomepageArticlesAreTheIndexsOwnCards(t *testing.T) {
	h := newServer(t)
	home := mustGet(t, h, "/")

	// The same grid as /articles, which is what makes the cards the same size:
	// .grid--articles carries the index's column width and the tighter card
	// padding and type scale that go with it.
	articles := section(t, home, "Articles &amp; real-estate advice", "</section>",
		"the homepage articles block")
	mustContain(t, articles, `class="grid grid--articles"`,
		"the homepage must use the article index's own grid")
	mustNotContain(t, articles, `class="grid grid--3"`,
		"three across was what made the previews bigger than the index's")

	// Five in a row on a wide screen, two rows of them: ten cards.
	if n := strings.Count(articles, `<article class="acard">`); n != 10 {
		t.Errorf("the homepage shows %d articles, want 10 — five across, two rows", n)
	}

	// And the index still shows them the same way, or "the same as they are in
	// article page" would be true of nothing.
	index := mustGet(t, h, "/articles")
	mustContain(t, index, `class="grid grid--articles"`, "the index keeps the grid it lends")

	layout := cssCode(t, "layout.css")
	rule := section(t, layout, ".grid--articles {", "}", "the article grid")
	mustContain(t, rule, "repeat(5, minmax(0, 1fr))", "five in one row on a wide screen")
}

func TestTheHomepageArticlesHeadingIsTucked(t *testing.T) {
	home := mustGet(t, newServer(t), "/")

	// "Move all the content more up compact together." The eyebrow, the title
	// and the subtitle read as one block rather than three separated lines.
	//
	// Matched from the heading block down to the title it introduces, because
	// the class sits above the words that identify which block this is.
	tucked := regexp.MustCompile(`(?s)section__head-text--tight.{0,200}Articles &amp; real-estate advice`)
	if !tucked.MatchString(home) {
		t.Error("the homepage articles heading block must be the compact one")
	}
}

// ---------------------------------------------------------------------------
// 82. "The content order in the frontpage make: the main banner, featured
//     properties, recently added (then the recently added previe icon size make
//     the same as featured properties), then the brokers, new developments,
//     articles, popular locations, Why people search on Previa, What people say.
//     Then move it all more a bit up, more compact together, the content blocks
//     closer the it's title. If there is small text under title then this move
//     ore up closer to title and the line spacing of it reduce. Then look that
//     the distance gaps between the blocks would be the same."
// ---------------------------------------------------------------------------

func TestTheHomepageRunsInTheOrderTheClientGave(t *testing.T) {
	home := mustGet(t, newServer(t), "/")

	want := []string{
		`class="hero hero--flush"`,
		"Featured properties",
		"Recently added",
		"Brokers to work with in",
		"New developments",
		"Articles &amp; real-estate advice",
		"Popular locations",
		"Why people search on Previa",
		"What people say",
	}

	at := -1
	for _, marker := range want {
		i := strings.Index(home, marker)
		if i < 0 {
			t.Fatalf("the homepage is missing %q entirely", marker)
		}
		if i <= at {
			t.Errorf("%q is out of order — it must come after everything above it in the client's list", marker)
		}
		at = i
	}
}

func TestRecentlyAddedIsTheSameCardAsFeatured(t *testing.T) {
	home := mustGet(t, newServer(t), "/")

	recent := section(t, home, "Recently added", "</section>", "the recently added block")
	featured := section(t, home, "Featured properties", "</section>", "the featured block")

	// "Then the recently added previe icon size make the same as featured
	// properties." Same grid, so the same column width, and no compact variant,
	// so the same card inside it.
	mustContain(t, featured, `class="grid grid--properties"`, "featured sets the density")
	mustContain(t, recent, `class="grid grid--properties"`, "…and recently added matches it")
	mustNotContain(t, recent, `"Variant" "compact"`,
		"the compact card is what made these previews smaller than the row above")
	mustNotContain(t, recent, `class="grid grid--4"`, "…and four across is what made them wider")
}

func TestTheGapsBetweenHomepageBlocksAreEqual(t *testing.T) {
	layout := cssCode(t, "layout.css")

	// One step, so the distance between any two blocks is 2 × --section-y
	// whichever pair you measure. The ladder of per-section overrides that used
	// to sit here is what made some gaps read wider than others.
	mustContain(t, layout, ".section { padding-block: var(--section-y); }",
		"every block takes the same padding above and below")
	mustNotContain(t, layout, ".section--after-hero .section__head { margin-bottom:",
		"no section may take a heading gap of its own")

	// And the small text under a title sits closer to it, with its own line
	// spacing reduced: that is what --tight means, and the homepage uses it on
	// every block.
	tight := section(t, layout, ".section__head-text--tight {", "}", "the compact heading block")
	mustContain(t, tight, "gap: 2px;", "the eyebrow, title and subtitle read as one block")

	home := mustGet(t, newServer(t), "/")
	if n := strings.Count(home, "section__head-text--tight"); n < 8 {
		t.Errorf("only %d homepage headings are compact; every block was asked for", n)
	}
}

// ---------------------------------------------------------------------------
// 83. "In the single article page the article title is in two rows, reduce the
//     line spacing. And the small text under it move more up as well and reduce
//     line spacing. Now in everywhere need to just pull content more compact
//     together."
// ---------------------------------------------------------------------------

func TestTheArticleTitleAndItsStandfirstAreTight(t *testing.T) {
	layout := cssCode(t, "layout.css")

	title := section(t, layout, ".page-head__title {", "}", "the page head title")
	mustContain(t, title, "line-height: var(--lh-tight);",
		"a two-line display title must not be spaced like body copy")

	inner := section(t, layout, ".page-head__inner {", "}", "the page head block")
	mustContain(t, inner, "gap: var(--sp-2);",
		"the line under the title moves up to 8px from it")

	lede := section(t, layout, ".page-head__inner .lede {", "}", "the standfirst")
	mustContain(t, lede, "line-height: var(--lh-snug);",
		"and its own line spacing comes down: it is a standfirst, not a paragraph")

	// It is the article page the client was looking at, so that page has to be
	// the one carrying it.
	h := newServer(t)
	body := mustGet(t, h, "/article/what-energy-class-actually-costs-you")
	mustContain(t, body, `class="page-head__title"`, "the article title takes the page head")
	mustContain(t, body, `class="lede"`, "…and the excerpt under it is the standfirst")
}

// ---------------------------------------------------------------------------
// 84. "In related reading move the conent more up inside it's menu block, and
//     reduce the distance between title and contend."
// ---------------------------------------------------------------------------

func TestRelatedReadingSitsHighInItsPanel(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/article/what-energy-class-actually-costs-you")

	tucked := regexp.MustCompile(`(?s)section__head-text--tight.{0,120}Related reading`)
	if !tucked.MatchString(body) {
		t.Error("the distance between the Related reading title and the cards must come down")
	}

	// "More up inside it's menu block": the panel's own padding, which is what
	// was holding the heading a full section step below the rounded corner.
	mustContain(t, body, "section--band-tight", "the panel takes the reduced padding")

	layout := cssCode(t, "layout.css")
	mustContain(t, layout, ".section--band-tight { padding-block: calc(var(--section-y) * 0.7); }",
		"and the rule that does it must actually be there")
}

// ---------------------------------------------------------------------------
// 85. "In the user account settings, where is the 'Choose every market you are
//     active in…' there after the title 'Contry' make a violet #8B008B button
//     'advertise' if click there then a popup will open with the options to
//     insert your broker ad into the broker section. … So make there at first
//     multiselect menu for the countries … these come there as tags. In the
//     backend there is option to set the price per day for each country. For
//     example in Germany 3 € per day, in Poland 1 € per day. So the broker
//     choosess the Germany and Poland for 10 days. So the system will calculate
//     a bill: Germany 3 € per day x 10 = 30 €; for Estonia 1 € per day x 10 =
//     10 €; total: 40 €. … In the frontpage there are only 2 rows broker ads. In
//     the broker page it stays till to the end of payd periode. If the user has
//     activated paid ad then under he's account he sees the end time/date of the
//     periode. As our website is global we use UTC time."
// ---------------------------------------------------------------------------

func TestTheAdvertiseButtonSitsAfterTheCountryTitle(t *testing.T) {
	body := mustGet(t, newServer(t), "/settings")

	// In the Country field's own heading row — "after the title 'Contry'" — and
	// not under the field, where it would read as part of what is being edited.
	row := section(t, body, `<div class="field__label-row">`, "</div>", "the Country heading row")
	mustContain(t, row, "Country", "the row is the Country label's")
	mustContain(t, row, "btn--advertise", "…and the advertise button sits in it")
	if !regexp.MustCompile(`>\s*advertise\s*</button>`).MatchString(row) {
		t.Error("the button must be labelled the way the client wrote it: advertise")
	}

	// Violet, and the colour the client named.
	components := cssCode(t, "components.css")
	rule := section(t, components, ".btn--advertise {", "}", "the advertise button")
	mustContain(t, rule, "background: #8B008B;", "the client named this colour")
}

func TestTheAdvertiseButtonOpensADialog(t *testing.T) {
	ad := asset(t, componentDir+"/broker-ad.html")

	// It is on the page it was asked for, and the structure below is what that
	// page renders.
	mustContain(t, mustGet(t, newServer(t), "/settings"), `x-data="previaBrokerAd(`,
		"the control is on the settings screen")

	// "A popup will open with the options to insert your broker ad into the
	// broker section." A real dialog, opened by the button and closable.
	mustContain(t, ad, `class="modal-backdrop"`, "the button opens a popup")
	mustContain(t, ad, `role="dialog"`, "…which is a dialog")
	mustContain(t, ad, `aria-modal="true"`, "…a modal one")
	mustContain(t, ad, "openDialog()", "the button is what opens it")
	mustContain(t, ad, "closeDialog()", "…and there is a way out of it")
}

func TestTheAdDialogPicksSeveralMarketsAsTags(t *testing.T) {
	body := mustGet(t, newServer(t), "/settings")
	ad := asset(t, componentDir+"/broker-ad.html")

	// "So make there at first multiselect menu for the countries, that can
	// multiselect all the countries he wants to advertise, these come there as
	// tags." The tag picker is exactly that control, and the same one the
	// profile's own Country field uses.
	mustContain(t, ad, `"Kind" "country"`, "markets are chosen with the tag picker")
	mustContain(t, ad, `"Name" "broker_ad_countries"`, "…submitted as the ad's markets")
	mustContain(t, ad, `"Countries" .Countries`, "…over the full market list")

	// Which renders as the picker, with a removable tag per choice, over every
	// market rather than only the eight that carry stock: a broker advertises
	// where the buyers are.
	picker := section(t, body, `id="ad-countries-label"`, "Run for", "the market picker")
	mustContain(t, picker, "tagpick__tag", "a choice becomes a removable tag")
	for _, code := range []string{"DE", "PL", "EE", "BR"} {
		mustContain(t, picker, `name="broker_ad_countries" value="`+code+`"`,
			"the picker must offer "+code)
	}
}

func TestTheAdIsPricedPerCountryPerDay(t *testing.T) {
	// The rate table is the client's own worked example, so the two figures
	// they named are pinned literally: Germany at €3 a day and Poland at €1.
	seed := asset(t, "../../pkg/data/seed_content.go")
	plan := section(t, seed, "var brokerAdPlan = models.BrokerAdPlan{", "\n}", "the ad plan")
	mustContain(t, plan, `{Country: "DE", PerDay: models.Money{Amount: 3, Currency: "EUR"}}`,
		"Germany is 3 € per day, which is the client's example")
	mustContain(t, plan, `{Country: "PL", PerDay: models.Money{Amount: 1, Currency: "EUR"}}`,
		"and Poland 1 €")
	mustContain(t, plan, "DefaultPerDay:", "a market nobody has priced still has a price")

	// The ladder of fixed run lengths this replaced is gone: a tier prices a
	// run, and here the run is the same length in every market while the price
	// is not.
	mustNotContain(t, plan, "Tiers:", "the homepage placement is not sold in tiers any more")

	// The dialog bills from it: a number of days, and a bill built line by line.
	ad := mustGet(t, newServer(t), "/settings")
	mustContain(t, ad, `name="broker_ad_days"`, "the broker chooses how many days")
	mustContain(t, ad, `class="ad-bill"`, "…and is shown the bill")
	mustContain(t, ad, "per day ×", "…a line per market, with its own rate")
	mustContain(t, ad, "Total", "…and a total")
}

func TestTheAdminSetsThePricePerDayForEachCountry(t *testing.T) {
	body := mustGet(t, newServer(t), "/admin/packages")

	// "In the backend there is option to set the price per day for each
	// country." One row per market, each with its own editable rate.
	mustContain(t, body, "Homepage broker advertising", "the price list is in the admin panel")
	mustContain(t, body, `name="broker_ad_default_per_day"`, "a default for the unpriced markets")
	for _, code := range []string{"DE", "PL", "EE"} {
		mustContain(t, body, `name="broker_ad_rate_`+code+`"`, "a rate for "+code)
	}

	// Every market, not a shortlist — and searchable, because there are two
	// hundred of them.
	if n := strings.Count(body, `name="broker_ad_rate_`); n < 150 {
		t.Errorf("the price list covers %d markets; it must cover every one the picker offers", n)
	}
	mustContain(t, body, "previaAdRates()", "…with a search over them")
}

func TestTheHomepageStripIsAQueueTwoRowsDeep(t *testing.T) {
	h := newServer(t)

	// "In the frontpage there are only 2 rows broker ads." The strip is four
	// across, so that is eight.
	strip := brokerNames(t, stripSection(t, getWithCountry(t, h, "/", "EE")))
	if len(strip) != 8 {
		t.Errorf("the homepage shows %d broker ads, want 8 — two rows of four: %q", len(strip), strip)
	}

	// "If next ad will come then the last one will be pushed futher till it
	// disappears from the frontpage. … In the broker page it stays till to the
	// end of payd periode." So the directory holds more than the homepage, and
	// what it holds extra is what the homepage pushed off.
	page := brokerNames(t, getWithCountry(t, h, "/brokers", "EE"))
	if len(page) <= len(strip) {
		t.Fatalf("the directory must outlast the homepage: page %q, strip %q", page, strip)
	}
	for _, name := range strip {
		if !contains(page, name) {
			t.Errorf("%q is on the homepage but not in the directory", name)
		}
	}

	// Newest purchase first, which is what makes it a queue rather than a
	// ranking: the seed's most recent Estonian buyer leads and its oldest is
	// the one missing from the homepage.
	seed := asset(t, "../../pkg/data/seed_content.go")
	mustContain(t, seed, `"br-03": {{"EE", 30, 0}}`, "the newest Estonian purchase")
	mustContain(t, seed, `"br-01": {{"EE", 30, 20}}`, "and the oldest")
	if strip[0] != "Liis Kask" {
		t.Errorf("the newest advertiser must lead the strip, got %q", strip[0])
	}
	if contains(strip, "Kadri Tamm") {
		t.Error("the oldest purchase must have been pushed off the homepage")
	}
	if !contains(page, "Kadri Tamm") {
		t.Error("…and must still be running in the directory until its period ends")
	}
}

func TestAPaidPeriodShowsItsEndInUTC(t *testing.T) {
	body := mustGet(t, newServer(t), "/settings")

	// "If the user has activated paid ad then under he's account he sees the
	// end time/date of the periode. As our website is global we use UTC time."
	runs := section(t, body, `<ul class="broker-ad__runs">`, "</ul>", "the running ads")
	stamp := regexp.MustCompile(`ends \d{1,2} [A-Z][a-z]{2} \d{4}, \d{2}:\d{2} UTC`)
	if got := stamp.FindAllString(runs, -1); len(got) < 2 {
		t.Errorf("each running market must show its end date and time in UTC, found %q in %q", got, runs)
	}

	// The map placement says the same thing about its own period.
	mustContain(t, body, "On the search map —", "the map placement reports itself too")

	// And it is UTC because the formatter says so, not because the seed happens
	// to be in it.
	funcs := asset(t, "../../pkg/view/funcs.go")
	utc := section(t, funcs, "func UTC(t time.Time) string {", "\n}", "the UTC stamp")
	mustContain(t, utc, "t.UTC().Format", "the moment is converted, not merely labelled")
	mustContain(t, utc, `" UTC"`, "…and labelled as well, so it cannot be read as local")
}

// ---------------------------------------------------------------------------
// 86. "Then after title 'Your location on the map' make the same advertise
//     button. And if open it then is popup menu with option 'Make your broker
//     profile visible on the Google map'. There is no country menu. There is
//     just option make your broker profile visible on the Google map for set
//     amount of days. Can choose 5 days - 1 €, 10 days - 2 €, 20 days - 3 € etc."
// ---------------------------------------------------------------------------

func TestTheMapPlacementHasItsOwnAdvertiseButton(t *testing.T) {
	body := mustGet(t, newServer(t), "/settings")

	// Same button, same place in the row, after the section's own title.
	row := section(t, body, `<span class="field__label" id="p-office-label">`,
		"</div>", "the map heading row")
	mustContain(t, row, "Your location on the map", "the row is the map section's")
	mustContain(t, row, "btn--advertise", "…and carries the same violet button")

	mustContain(t, body, `x-data="previaMapAd()"`, "the control is on the settings screen")
	mapAd := asset(t, componentDir+"/broker-map-ad.html")
	mustContain(t, mapAd, "Make your broker profile visible on the Google map",
		"the dialog says what the client wrote")
	mustContain(t, mapAd, `role="dialog"`, "…and is a dialog like the other one")

	// "There is no country menu."
	mustNotContain(t, mapAd, "previaTagPicker()", "this placement has no market picker")
	mustNotContain(t, mapAd, "broker_ad_countries", "…and nothing that would submit one")
}

func TestTheMapPlacementIsSoldByTheDayLadderTheClientNamed(t *testing.T) {
	body := mustGet(t, newServer(t), "/settings")
	mapAd := section(t, body, `<legend class="field__label">Show it for</legend>`,
		"</fieldset>", "the map placement's price ladder")

	// "Can choose 5 days - 1 €, 10 days - 2 €, 20 days - 3 € etc." The three the
	// client named, in their words, plus the one the "etc" continues to.
	for _, tier := range []struct{ days, price string }{
		{"5", "€1"}, {"10", "€2"}, {"20", "€3"}, {"30", "€4"},
	} {
		block := section(t, mapAd, `value="`+tier.days+`" data-price=`, "</label>",
			tier.days+"-day tier")
		mustContain(t, block, tier.days+" days", "the run length")
		mustContain(t, block, tier.price, "…and what the client priced it at")
	}
}

// ---------------------------------------------------------------------------
// 87. "Then with the logo lets try no this that the same rounded rectangle
//     background, what is behind favicon, make it in the header as well. The
//     current logo in the header stays the same place and same size, just put
//     this background there. And then the same logo insert into footer left side
//     as well. So the footer text 'Listing packages' and '© 2026 Previa, all
//     rights reserved!' move to the right and place the logo there, in the same
//     place as it is in the header - so they are aligned to each other."
// ---------------------------------------------------------------------------

func TestTheLogoCarriesTheFaviconsTile(t *testing.T) {
	body := mustGet(t, newServer(t), "/")

	// The favicon's rectangle, in the header.
	mustContain(t, body, `class="logo__tile"`, "the mark draws the tile behind it")
	mustContain(t, body, `rx="11.6" fill="#8B008B"`, "…in the favicon's colour, rounded the same way")

	// And it is the favicon's colour because the favicon says so, not because
	// two files happen to agree.
	favicon := asset(t, imgDir+"/favicon.svg")
	mustContain(t, favicon, `fill="#8B008B"`, "the favicon's tile is the one being copied")

	// Same size and same place: the lockup's own dimensions are untouched.
	pages := cssCode(t, "pages.css")
	mustContain(t, pages, ".logo__mark { flex: none; width: 72px; height: 72px; }",
		"the header lockup keeps the size it had")
}

func TestTheFooterCarriesTheSameLogoAlignedWithTheHeaders(t *testing.T) {
	body := mustGet(t, newServer(t), "/")
	footer := section(t, body, `<footer class="site-footer">`, "</footer>", "the footer")

	mustContain(t, footer, `class="footer__logo"`, "the footer has the mark on its left")
	mustContain(t, footer, "logo__tile", "…the same one, tile and all")

	// "So the footer text … move to the right and place the logo there." Both
	// rows sit in the second column, to the right of the mark.
	mustContain(t, footer, `class="footer__grid"`, "the footer is a two-column block now")
	if strings.Index(footer, "footer__logo") > strings.Index(footer, "Listing packages") {
		t.Error("the logo must come before the links it pushed to the right")
	}

	// Aligned with the header's, which is a property of the box rather than of
	// a hand-tuned offset: both marks start at the same container's inline edge.
	layout := cssCode(t, "layout.css")
	grid := section(t, layout, ".footer__grid {", "}", "the footer grid")
	mustContain(t, grid, "grid-template-columns: auto minmax(0, 1fr);",
		"the mark takes the first column and everything else the second")
	logo := section(t, layout, ".footer__logo {", "}", "the footer logo cell")
	mustContain(t, logo, "grid-row: 1 / span 2;",
		"…spanning both rows, so the links and the legal strip both clear it")
}

// ---------------------------------------------------------------------------
// 88. "The 'Features and amenities' we just added into 'add listing' page add to
//     the search menu as well, with the same subtitles. And there make search
//     field as well to type to look the desired one, as there are quite many of
//     them."
// ---------------------------------------------------------------------------

func TestTheSearchMenuOffersTheWholeAmenityCatalogue(t *testing.T) {
	h := newServer(t)
	search := mustGet(t, h, "/search")

	block := section(t, search, `x-data="previaAmenityFilter()"`, "</div>\n    </div>",
		"the features filter")

	// "With the same subtitles" — the same five groups the add-listing form
	// draws, because both render the same catalogue.
	for _, group := range []string{
		"The property", "Parking and access", "Living and comfort", "Kitchen", "Safety",
	} {
		mustContain(t, block, `<legend class="amenity-group__title">`+group+`</legend>`,
			"the search menu must carry the subtitle: "+group)
	}

	// And the same ticks. Every one of them, not a selection.
	for _, name := range append(append([]string{}, newAmenities...), oldAmenities...) {
		mustContain(t, block, `<span class="check__text">`+name+`</span>`,
			"the search menu must offer: "+name)
	}
}

func TestOneCatalogueBacksBothScreens(t *testing.T) {
	// The list is Go now, not a template literal, which is what makes "the same
	// subtitles" a guarantee rather than a coincidence: a seller's tick and a
	// buyer's filter are the same key or the filter finds nothing.
	models := asset(t, "../../pkg/models/amenities.go")
	mustContain(t, models, "var AmenityGroups = []AmenityGroup{", "one catalogue, in one place")

	for _, page := range []string{"add-listing.html", "../components/filters.html"} {
		body := asset(t, "../../web/templates/pages/"+page)
		mustContain(t, body, "{{ range amenityGroups }}",
			page+" must render the catalogue rather than a copy of it")
	}
}

func TestTheAmenityFilterHasASearchField(t *testing.T) {
	search := mustGet(t, newServer(t), "/search")

	// "And there make search field as well to type to look the desired one, as
	// there are quite many of them."
	mustContain(t, search, `class="amenity-search"`, "the list has a search field above it")
	mustContain(t, search, `aria-label="Search features and amenities"`, "…named for a screen reader")
	mustContain(t, search, `data-amenity="Coffee maker"`,
		"every tick carries the text the field matches on")
	mustContain(t, search, "data-amenity-group",
		"…and a group hides itself when nothing in it matches")

	js := asset(t, jsDir+"/previa.js")
	mustContain(t, js, "window.previaAmenityFilter = function ()", "the component behind it exists")
	mustContain(t, js, "item.hidden = !hit;",
		"filtering hides ticks rather than removing them, so a ticked one still submits")
}

func TestTheNewAmenitiesActuallyFilter(t *testing.T) {
	h := newServer(t)

	count := func(query string) int {
		body := mustGet(t, h, "/search"+query)
		m := regexp.MustCompile(`<strong class="numeric">([0-9 ]+)</strong>`).FindStringSubmatch(body)
		if m == nil {
			t.Fatalf("no result count on /search%s", query)
		}
		n := 0
		for _, r := range m[1] {
			if r >= '0' && r <= '9' {
				n = n*10 + int(r-'0')
			}
		}
		return n
	}

	all := count("")
	one := count("?amenity=coffee-maker")
	two := count("?amenity=coffee-maker&amenity=sauna")

	if one >= all {
		t.Errorf("ticking an amenity must narrow the results: %d of %d", one, all)
	}
	// Two ticks means both, which is what ticking two boxes has always meant.
	if two >= one {
		t.Errorf("a second amenity must narrow it further: %d then %d", one, two)
	}
	// A hand-edited URL filters on nothing that exists rather than returning an
	// unexplained empty page.
	if count("?amenity=nonsense") != all {
		t.Error("an unknown amenity must be ignored, not obeyed")
	}

	// And each one is a removable chip, so a search narrowed by six can drop
	// back one at a time.
	chips := mustGet(t, h, "/search?amenity=coffee-maker&amenity=sauna")
	mustContain(t, chips, `href="/search?amenity=sauna"`, "each amenity chip removes only itself")
}

// ---------------------------------------------------------------------------
// 89. "In the add listing page this right side progress bar, at first make this
//     title 'your progress' more up and cut it's section together, there is too
//     much empty room at the moment. Then the green there make #008000."
// ---------------------------------------------------------------------------

func TestTheProgressRailIsCutTogether(t *testing.T) {
	pages := cssCode(t, "pages.css")

	// "Make this title 'your progress' more up": the panel's own padding, which
	// is what was holding the header away from the top edge.
	panel := section(t, pages, ".stepper {\n", "}", "the progress panel")
	mustContain(t, panel, "padding: var(--sp-4);", "the panel's padding comes down a step")

	head := section(t, pages, ".stepper__header {", "}", "the panel header")
	mustContain(t, head, "padding: var(--sp-2) var(--sp-4);", "…and the header's with it")
	mustContain(t, head, "margin: calc(var(--sp-4) * -1) calc(var(--sp-4) * -1) var(--sp-3);",
		"…so the header still spans the panel's full width")

	// "Cut it's section together": the twelve waypoints lose the empty room
	// between them. Nothing is smaller — the markers and the type are untouched.
	mustContain(t, pages,
		".stepper__item { position: relative; display: flex; align-items: center; gap: var(--sp-3); padding: 3px 0; }",
		"the waypoints sit closer together")
	mustContain(t, pages, ".stepper__marker {", "the markers themselves are unchanged")
	mustContain(t, pages, "width: 25px; height: 25px;", "…still 25px")
}

func TestTheProgressRailsGreenIsTheOneTheClientNamed(t *testing.T) {
	pages := cssCode(t, "pages.css")

	mustContain(t, pages, ".stepper { --stepper-done: #008000; }",
		"the client named this green for this rail")
	mustContain(t, pages, ".stepper__item.is-done::after { background: var(--stepper-done); }",
		"the connector takes it")
	mustContain(t, pages,
		".stepper__item.is-done .stepper__marker { background: var(--stepper-done); border-color: var(--stepper-done); color: #fff; }",
		"…and so does the marker")

	// Scoped to the rail. --success paints the paid badge, the saved-listing
	// notice and the published state, none of which the client asked to move.
	tokens := cssCode(t, "tokens.css")
	mustContain(t, tokens, "--success: #2E7D5B;", "the site's status green is untouched")
}

// ---------------------------------------------------------------------------
// 90. "The sidebar menu on the right, if start to scroll page down, it moves too
//     far away, make that it stays always inside the page, so it is always under
//     the header right side. So that this menu's right side is aligned to header
//     right side."
// ---------------------------------------------------------------------------

func TestTheFloatingMenuStaysInsideThePage(t *testing.T) {
	body := mustGet(t, newServer(t), "/")

	mustContain(t, body, `<div class="floating-menu-shell">`,
		"the floating control sits in a shell that is the header's box")

	layout := cssCode(t, "layout.css")
	shell := section(t, layout, ".floating-menu-shell {", "}", "the shell")

	// The header's own box, so the button's right edge is the same pixel as the
	// "add listing" button's above it — at every width, not only past 1440px.
	mustContain(t, shell, "width: var(--page-width);", "same width as the header")
	mustContain(t, shell, "margin-inline: auto;", "…centred the same way")
	mustContain(t, shell, "padding-inline: max(var(--gutter)", "…with the header's inside gutter")
	mustContain(t, shell, "justify-content: flex-end;", "…and the button pushed to its right edge")
	mustContain(t, shell, "pointer-events: none;",
		"the strip is inert, or it would swallow clicks across the top of every page")

	// The button itself takes clicks again, and no longer positions itself.
	button := section(t, layout, "\n.floating-menu {", "}", "the button")
	mustContain(t, button, "pointer-events: auto;", "the button is still clickable")
	mustNotContain(t, button, "position: fixed;", "the shell is what is pinned now, not the button")

	// The header it aligns with reads the same token, which is what makes this
	// hold on the wide-container and full-bleed map layouts too.
	header := section(t, layout, ".site-header {", "}", "the header")
	mustContain(t, header, "width: var(--page-width);", "the header is the thing being matched")
}
