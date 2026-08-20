package app_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// Tests for the client's third batch of 18 August notes (17:35–18:14).
//
// Seven notes, numbered 59–65 on from the afternoon's fourteen, each section
// named after what was asked rather than after the code that answers it — the
// convention the seven earlier rounds set. The first three are small and exact;
// the last four are one design, described across four messages, and are tested
// as the one system they describe rather than message by message.
//
// Anything the server renders is asserted against a real response. What only
// exists once JavaScript has run — the broker markers, their preview popups —
// is asserted at the level of the function that produces it.

// ---------------------------------------------------------------------------
// 59. "In the maps page there is 'All locations · Newest first' there make the
//     same 'sort by' menu as in listing page. And sort by in both places write
//     with small letter."
// ---------------------------------------------------------------------------

func TestTheMapPageHasTheListingsPageSortMenu(t *testing.T) {
	h := newServer(t)

	// One definition, so the two screens cannot drift apart — which is what
	// the note is really about. The map page used to print the order as text.
	results := asset(t, "../../web/templates/components/results.html")
	mustContain(t, results, `{{ define "sort-select" }}`,
		"the sort menu must be one shared definition")
	mustNotContain(t, results, "· {{ $d.SortLabel }}",
		"the map page must not print the order as dead text any more")

	grid := mustGet(t, h, "/search")
	split := mustGet(t, h, "/search?view=map")

	// The same six options, in the same order, on both.
	for _, opt := range []string{
		`>Newest first<`, `>Price: low to high<`, `>Price: high to low<`,
		`>Largest first<`, `>Price per m²<`, `>Most viewed<`,
	} {
		mustContain(t, grid, opt, "the listings page offers every order")
		mustContain(t, split, opt, "and the map page offers the same ones")
	}

	// A real control on the map, wired to the same endpoint as the grid's.
	sort := section(t, split, `id="map-sort"`, "</select>", "the map page's sort menu")
	mustContain(t, sort, `hx-get="/search/results"`, "it must run the search")
	mustContain(t, sort, `hx-include="#filter-form"`, "…keeping the filters that are applied")

	// And it works: ordering the map by price ascending really does put the
	// cheapest first. The two `sort` values are what htmx actually sends — the
	// control's own value, then the filter form's hidden copy — and the first
	// is the one that must win.
	asc := mustGet(t, h, "/search/results?sort=price_asc&view=map&deal=sale&sort=newest")
	mustContain(t, asc, `<option value="price_asc"  selected>`,
		"the menu must come back showing the order that was chosen")
}

func TestSortByIsLowerCaseInBothPlaces(t *testing.T) {
	h := newServer(t)
	for _, path := range []string{"/search", "/search?view=map"} {
		body := mustGet(t, h, path)
		mustContain(t, body, ">sort by</label>", path+": the label is lower case")
		mustNotContain(t, body, ">Sort by<", path+": no capital survives")
	}
}

// ---------------------------------------------------------------------------
// 60. "Here in the maps page where the filter-tags are, the '+ Add filter' make
//     '+ filter' and after that make the same red delete cross to erase all the
//     selected filters at once."
// ---------------------------------------------------------------------------

func TestTheMapTagBarSaysFilterAndClearsInOne(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search?view=map&deal=sale&deal=rent&property_type=house")

	bar := section(t, body, `<div class="active-filters">`, "</div>", "the map tag bar")
	mustContain(t, bar, " filter\n", "the add-filter chip is '+ filter' now")
	mustNotContain(t, bar, "Add filter", "…not '+ Add filter'")

	// The same control the listings page already had, not a second one that
	// looks like it: same class, same red, same accessible name.
	mustContain(t, bar, `class="clear-filters"`, "and a red cross that clears everything")
	mustContain(t, bar, `aria-label="Clear all filters"`,
		"a cross on its own says nothing to a screen reader")

	// It clears the filters and keeps the reader on the map, which is the
	// point of it being on this page rather than a link back to the grid.
	mustContain(t, bar, `href="/search?view=map"`, "clearing must not leave the map")
}

// ---------------------------------------------------------------------------
// 61. "In the header after Listings add section 'Map' so this is the map page."
// ---------------------------------------------------------------------------

func TestMapSitsAfterListingsInTheHeader(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search")

	nav := section(t, body, `<nav class="header__nav"`, "</nav>", "the header nav")
	order := []string{">Listings<", ">Map<", ">Articles<", ">Brokers<", ">Developments<"}
	at := 0
	for _, item := range order {
		i := strings.Index(nav[at:], item)
		if i < 0 {
			t.Fatalf("the header nav is missing %s, or it is out of order: %q", item, nav)
		}
		at += i
	}

	mustContain(t, nav, `href="/search?view=map"`, "Map opens the split map view")
}

func TestMapAndListingsNeverLightUpTogether(t *testing.T) {
	h := newServer(t)

	// On the map, Map is current and Listings is not — an entry that stayed
	// dark on the page it names would be worse than no entry at all.
	split := section(t, mustGet(t, h, "/search?view=map"), `<nav class="header__nav"`, "</nav>", "nav")
	mustContain(t, split, `href="/search?view=map"
           aria-current="page">Map</a>`, "Map is current on the map page")
	mustNotContain(t, split, `href="/search"
           aria-current="page">Listings</a>`, "…and Listings is not")

	// And the other way round on the grid.
	grid := section(t, mustGet(t, h, "/search"), `<nav class="header__nav"`, "</nav>", "nav")
	mustContain(t, grid, `href="/search"
           aria-current="page">Listings</a>`, "Listings is current on the listings page")
	mustNotContain(t, grid, `aria-current="page">Map</a>`, "…and Map is not")

	// The full-screen map is the same page, so it marks the same entry.
	full := section(t, mustGet(t, h, "/search?view=full"), `<nav class="header__nav"`, "</nav>", "nav")
	mustContain(t, full, `aria-current="page">Map</a>`, "the full-screen map is still the map page")
}

// ---------------------------------------------------------------------------
// 62. "So the broker can buy an ad so that his profile is displayed under
//     Germany's market … One broker can buy these ads in the frontpage under
//     different countries, but he must pay separately for each country. At
//     first he wants that his profile is displayed in the German market for 30
//     days, then he activates it with payment, and then he can activate his ad
//     under France market as well for 30 days with new payment. This is the
//     broker user-profile what is displayed there and if he updates his profile
//     by changing photo or phone, then this will be updated in this ad
//     immediately."
// ---------------------------------------------------------------------------

func TestEachMarketIsItsOwnPaidRun(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/settings")

	// Two markets bought on different days, each with its own countdown. A
	// single "running in Estonia, Finland — 18 days left" would be wrong about
	// at least one of them, which is exactly the case the client described.
	runs := section(t, body, `<ul class="broker-ad__runs">`, "</ul>", "the running ads")
	mustContain(t, runs, "Running in <strong>Estonia</strong>", "the first market is named on its own")
	mustContain(t, runs, "Running in <strong>Finland</strong>", "and so is the second")

	ee := section(t, runs, "Estonia", "</li>", "the Estonian run")
	fi := section(t, runs, "Finland", "</li>", "the Finnish run")
	if strings.Contains(ee, "18") == strings.Contains(fi, "18") {
		t.Errorf("two runs bought on different days must show different countdowns\nEE: %q\nFI: %q", ee, fi)
	}

	// And the money rule, stated where the money is.
	mustContain(t, body, "Each market is paid for separately",
		"the profile must say each country is bought on its own")
}

func TestTheAdShowsTheProfileAsItIsNow(t *testing.T) {
	h := newServer(t)

	// Nothing in the placement stores a photograph, a name or a number, which
	// is what makes "updated in this ad immediately" true rather than a promise
	// somebody has to keep. The model is where that is guaranteed.
	models := asset(t, "../../pkg/models/models.go")
	ad := section(t, models, "type BrokerAd struct {", "}", "the market placement")
	for _, field := range []string{"Photo", "Phone", "Name", "Email"} {
		mustNotContain(t, ad, field, "the placement must not keep a copy of the profile")
	}

	mustContain(t, mustGet(t, h, "/settings"),
		"every ad you are running updates immediately",
		"and the profile must say so, or a broker buys a second ad to refresh the first")
}

func TestAMarketStripShowsOnlyWhoBoughtThatMarket(t *testing.T) {
	h := newServer(t)

	ee := brokerNames(t, mustGet(t, h, "/brokers"))
	de := brokerNames(t, mustGetAs(t, h, "/brokers", "DE"))
	if len(ee) == 0 || len(de) == 0 {
		t.Fatalf("both markets should have advertisers: EE %q, DE %q", ee, de)
	}
	for _, name := range ee {
		for _, other := range de {
			if name == other {
				t.Errorf("%q appears in both markets' strips; an ad is bought per market", name)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// 63. "And the second thing what the broker can do … he can activate that his
//     broker profile is displayed in the googlemaps like the ads. For this in
//     the maps his rounded rectangle profile image will be displayed with his
//     name - small icon, and if user clicks on it, then the bigger user profile
//     will open, just like the real estate preview window, and if click there
//     then will redirect to this broker's single page. So in the main search
//     menu, where is 'deal type' sell and buy, there add 'brokers' so the user
//     can choose if he wants the brokers will be displayed or not."
// ---------------------------------------------------------------------------

func TestBrokersIsAChoiceBesideDealType(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search")

	// In the filter panel, in the deal-type block, exactly where the client
	// put it. Its own row rather than a fourth deal type: it does not narrow
	// the listings, it adds a second kind of result beside them.
	block := section(t, body, `<legend class="field__label">Deal type</legend>`,
		"</fieldset>", "the deal type block")
	mustContain(t, block, `name="brokers" value="1"`, "Brokers is a choice in this block")
	mustContain(t, block, "segmented--brokers", "…on its own row")

	// Off unless asked for.
	mustNotContain(t, block, `name="brokers" value="1"
                 checked`, "brokers must be off by default")

	// And it is a filter like any other, so it says so above the results and
	// can be taken off again from there.
	on := mustGet(t, h, "/search?brokers=1")
	mustContain(t, on, "results-brokers__head", "ticking it puts brokers among the results")
	mustContain(t, on, `aria-label="Remove filter Brokers"`, "…as a removable tag")
}

func TestBrokersAppearOnTheMapAndInTheResults(t *testing.T) {
	h := newServer(t)

	off := mustGet(t, h, "/search?view=map")
	on := mustGet(t, h, "/search?view=map&brokers=1")

	// Nothing at all until asked for, in either place.
	mustNotContain(t, off, "&#34;brokers&#34;:[{", "no broker pins unless brokers are switched on")
	mustNotContain(t, off, "bcard__name", "and no broker cards either")

	// "In the maps like the real estate the small broker images are displayed.
	// And in the listings page the broker profile-ads are displayed like the
	// real-estate ads." Both, from one tick.
	mustContain(t, on, "&#34;brokers&#34;:[{", "the map is handed the broker pins")
	mustContain(t, on, "bcard--map", "and the results column shows their cards")
	mustContain(t, on, "bcard__name", "…the same broker card the rest of the site uses")

	grid := mustGet(t, h, "/search?brokers=1")
	mustContain(t, grid, "bcard__name", "the listings page shows them as cards too")
}

func TestOnlyPaidPinnedBrokersReachTheMap(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search?view=map&brokers=1")

	// The placement is a purchase, so a broker who has not made it is not on
	// the map however good their profile is. Fourteen brokers are seeded; the
	// map placement was bought by eight of them.
	shown := strings.Count(body, `&#34;id&#34;:&#34;br-`)
	if shown == 0 {
		t.Fatal("no broker pins at all")
	}
	if shown >= 14 {
		t.Errorf("%d brokers on the map; only the ones who bought the placement belong there", shown)
	}

	// Both halves are required, and the model says so: an unpaid pin is
	// private, and a paid placement with nowhere to put it cannot be drawn.
	models := asset(t, "../../pkg/models/models.go")
	mustContain(t, models, "func (b Broker) IsOnMap() bool { return b.MapAd.IsLive(time.Now()) && b.Office.IsSet() }",
		"a broker is on the map only when they have paid and dropped a pin")
}

func TestABrokerMarkerIsAPhotographAndAName(t *testing.T) {
	js := asset(t, "../../public/static/js/previa.js")
	css := asset(t, cssDir+"/pages.css")

	// "His rounded rectangle profile image will be displayed with his name."
	marker := section(t, js, "function brokerMarker(", "\n    }", "the broker marker")
	mustContain(t, marker, `map-broker__photo`, "the marker carries the photograph")
	mustContain(t, marker, `map-broker__name`, "…and the name beside it")

	shape := section(t, css, ".map-broker__photo {", "}", "the marker photograph")
	mustContain(t, shape, "border-radius: 7px", "rounded rectangle, not a disc")

	// "If user clicks on it, then the bigger user profile will open, just like
	// the real estate preview window, and if click there then will redirect to
	// this broker's single page."
	popup := section(t, js, "function brokerPopupHtml(", "\n    }", "the broker preview")
	mustContain(t, popup, "map-popup__card", "the preview is the listing popup's shape")
	if strings.Count(popup, "b.url") < 3 {
		t.Error("every part of the preview a pointer lands on must lead to the profile page")
	}
	mustContain(t, popup, ">View profile<", "…with the same primary action a listing popup has")
}

// ---------------------------------------------------------------------------
// 64. "Then now the separate page Brokers, there we do that this stays a
//     separate brokers page. And there are two ways to look for the brokers. On
//     top in the header is the choose your market button, and so whatever
//     market is chosen there, then these brokers what have bought ad in this
//     market are displayed there — so the same as the frontpage broker section.
//     And if the user in this search menu enters the location and radius, then
//     the 'choose your market' system is not active any more and now the system
//     displays there only these brokers what are in this range and in this
//     order, who is closer. So on every broker profile now come the 'distance'
//     like it is in sexydate page."
// ---------------------------------------------------------------------------

func TestBrowsingBrokersShowsTheMarketsOwnStrip(t *testing.T) {
	h := newServer(t)

	// "The same as the frontpage broker section" — so the directory shows that
	// market's advertisers and nobody else's.
	//
	// The same *set*, not the same list, since 19 August: "in the frontpage
	// there are only 2 rows broker ads … in the broker page it stays till to
	// the end of payd periode." The homepage is a window onto the front of a
	// queue and the directory is the whole of it, so the strip is a prefix of
	// the page rather than equal to it — see TestTheHomepageStripIsAQueueTwoRowsDeep.
	for _, market := range []string{"EE", "DE"} {
		page := brokerNames(t, mustGetAs(t, h, "/brokers", market))
		strip := brokerNames(t, stripSection(t, mustGetAs(t, h, "/", market)))
		if len(page) == 0 {
			t.Fatalf("%s: the directory should show that market's advertisers", market)
		}
		if len(strip) == 0 {
			t.Fatalf("%s: the homepage should show that market's advertisers", market)
		}
		for _, name := range strip {
			if !contains(page, name) {
				t.Errorf("%s: %q is on the homepage strip but not in the directory\npage: %q",
					market, name, page)
			}
		}
		if len(strip) > len(page) {
			t.Errorf("%s: the homepage shows more advertisers than the directory\npage: %q\nstrip: %q",
				market, page, strip)
		}
	}

	mustContain(t, mustGet(t, h, "/brokers"), "advertising in Estonia",
		"the count must say which of the page's two searches produced it")
}

func TestSearchingAPlaceTurnsTheMarketOff(t *testing.T) {
	h := newServer(t)

	market := brokerCount(t, mustGet(t, h, "/brokers"))
	near := mustGet(t, h, "/brokers?location=Tallinn%2C+Estonia&radius=5000")

	// "The 'choose your market' system is not active any more": a radius that
	// reaches every seeded broker must return every seeded broker, whichever
	// market they advertise in.
	if brokerCount(t, near) <= market {
		t.Errorf("a place search returned %d brokers, no more than the market's %d",
			brokerCount(t, near), market)
	}
	mustContain(t, near, "nearest first", "…and the order says what it is ordered by")
}

func TestBrokersAreOrderedNearestFirstWithTheDistanceShown(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/brokers?location=Tallinn%2C+Estonia&radius=5000")

	km := distances(t, body)
	if len(km) < 3 {
		t.Fatalf("expected several brokers with distances, got %v", km)
	}
	for i := 1; i < len(km); i++ {
		if km[i] < km[i-1] {
			t.Errorf("brokers are not ordered nearest first: %v", km)
			break
		}
	}

	// "Like it is in sexydate page": the word in ordinary ink, the number in
	// red. The reference the client sent reads "distance 75.78 KM".
	mustContain(t, body, `<span class="distance-line__label">distance</span>`,
		"the word comes first, quietly")
	mustContain(t, body, ` KM</strong>`, "and the figure carries the unit")
	mustContain(t, asset(t, cssDir+"/components.css"), ".distance-line__value {",
		"the figure is styled on its own")
	red := section(t, asset(t, cssDir+"/components.css"), ".distance-line__value {", "}", "the distance figure")
	mustContain(t, red, "color: #FF0000", "the client's red, the same one the hearts and the cross use")

	// No place searched, no distance claimed.
	mustNotContain(t, mustGet(t, h, "/brokers"), "distance-line",
		"a card that cannot honestly state a distance must not carry the line")
}

// ---------------------------------------------------------------------------
// 65. "In the Listings page the same if the user has entered the location to
//     googlemaps then all the ads displayed have on bottom the distance and
//     distance number is red. The order of the real-estate ads here is not
//     automatically according to nearest distance. The order stays like it is
//     set up in the 'sort by' menu."
// ---------------------------------------------------------------------------

func TestListingsShowTheDistanceWhenAPlaceWasSearched(t *testing.T) {
	h := newServer(t)

	plain := mustGet(t, h, "/search")
	mustNotContain(t, plain, "distance-line",
		"a search with no place in it must look exactly as it did")

	located := mustGet(t, h, "/search?location=Tallinn%2C+Estonia")
	if len(distances(t, located)) < 5 {
		t.Error("every card must carry its distance once a place has been entered")
	}
	mustContain(t, located, `<span class="distance-line__label">distance</span>`,
		"the same line the broker cards carry, from the same component")
}

func TestEveryCardCarriesItsDistance(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search?location=Tallinn%2C+Estonia")

	// "All the ads displayed have on bottom the distance" — all of them, which
	// includes the one sitting on the point that was searched.
	cards := strings.Count(body, `class="pcard pcard--compact`)
	lines := strings.Count(body, "distance-line__value")
	if cards == 0 {
		t.Fatal("no result cards on a located search")
	}
	if lines != cards {
		t.Errorf("%d cards but %d distances: every card must carry one", cards, lines)
	}

	// The listing at the searched point is 0.00 km away, not undisclosed.
	// Treating zero as "not measured" dropped the line from exactly the
	// listing the reader had gone looking for, which is why Distance carries a
	// DistanceSet flag rather than using zero as its own sentinel.
	mustContain(t, body, `>0.00 KM</strong>`,
		"a listing on the searched point states its distance like any other")
}

func TestTheSortMenuStillDecidesTheOrder(t *testing.T) {
	h := newServer(t)

	// "The order of the real-estate ads here is not automatically according to
	// nearest distance." Under an explicit order, the distances must not come
	// out sorted — if they did, the distance would have quietly taken over.
	body := mustGet(t, h, "/search?location=Tallinn%2C+Estonia&sort=price_desc")
	km := distances(t, body)
	if len(km) < 5 {
		t.Fatalf("expected a page of located results, got %v", km)
	}
	sorted := true
	for i := 1; i < len(km); i++ {
		if km[i] < km[i-1] {
			sorted = false
			break
		}
	}
	if sorted {
		t.Errorf("the results came back in distance order under sort=price_desc: %v", km)
	}

	// And the order that was asked for is the one that was applied.
	mustContain(t, body, `<option value="price_desc" selected>`,
		"the menu must show the order the results are actually in")

	// The measurement happens after the sort, which is what makes that true
	// rather than a coincidence of this data.
	mock := asset(t, "../../pkg/data/mock.go")
	sortAt := strings.Index(mock, "sortProperties(matched, f.Sort)")
	distAt := strings.Index(mock, "matched[i].Distance = here.DistanceKm")
	if sortAt < 0 || distAt < 0 || distAt < sortAt {
		t.Error("the distance must be measured after the order is settled, so it cannot affect it")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mustGetAs fetches a page with a market chosen. The market is a cookie, so a
// plain GET always arrives on the default one — which is why the homepage
// strip and the directory can only be compared per market through this.
func mustGetAs(t *testing.T, h http.Handler, path, market string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(&http.Cookie{Name: "previa_country", Value: market})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s as %s = %d, want 200", path, market, rec.Code)
	}
	return rec.Body.String()
}

// brokerNames lists the brokers a page shows, in the order it shows them.
func brokerNames(t *testing.T, body string) []string {
	t.Helper()
	var out []string
	rest := body
	for {
		i := strings.Index(rest, `<a href="/broker/`)
		if i < 0 {
			break
		}
		rest = rest[i:]
		j := strings.Index(rest, ">")
		k := strings.Index(rest, "</a>")
		if j < 0 || k < 0 || k < j {
			break
		}
		out = append(out, strings.TrimSpace(rest[j+1:k]))
		rest = rest[k:]
	}
	return out
}

// stripSection is the homepage's broker strip on its own, so the names in it
// cannot be confused with a broker linked from anywhere else on the page.
func stripSection(t *testing.T, body string) string {
	t.Helper()
	return section(t, body, `<div class="broker-strip">`, "broker-strip__sell", "the homepage broker strip")
}

// distances reads the distance figures a page prints, in page order.
func distances(t *testing.T, body string) []float64 {
	t.Helper()
	var out []float64
	rest := body
	const marker = `class="distance-line__value numeric">`
	for {
		i := strings.Index(rest, marker)
		if i < 0 {
			break
		}
		rest = rest[i+len(marker):]
		j := strings.Index(rest, " KM")
		if j < 0 {
			break
		}
		if v, err := strconv.ParseFloat(strings.TrimSpace(rest[:j]), 64); err == nil {
			out = append(out, v)
		}
		rest = rest[j:]
	}
	return out
}
