package app_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"previa/pkg/models"
)

// Property-type catalogue, multi-select filtering and the new menu control.
//
// The server side is covered here: the catalogue, the repeated-parameter URL
// contract, OR matching and the markup the browser then drives. The Any-type
// bookkeeping and the drawer are browser behaviour and are verified separately
// against a real browser.

// cardCount counts rendered property cards, not the inner pcard__* elements.
func cardCount(body string) int {
	return strings.Count(body, `<article class="pcard`)
}

// ---------------------------------------------------------------------------
// Catalogue
// ---------------------------------------------------------------------------

func TestPropertyTypeCatalogue(t *testing.T) {
	want := []struct{ value, label string }{
		{"apartment", "Apartment"},
		{"house", "House"},
		{"house-part", "House part"},
		{"cottage", "Cottage"},
		{"modular-house", "Modular house"},
		{"panelized-house", "Panelized house"},
		{"trailer-house", "Trailer house"},
		{"sauna", "Sauna"},
		{"commercial", "Commercial"},
		{"industrial", "Industrial property"},
		{"land", "Land"},
		{"garage", "Garage"},
		{"new-development", "New development"},
	}
	if len(models.PropertyTypes) != len(want) {
		t.Fatalf("catalogue has %d types, want %d", len(models.PropertyTypes), len(want))
	}
	for i, w := range want {
		got := models.PropertyTypes[i]
		if string(got.Value) != w.value || got.Label != w.label {
			t.Errorf("type %d = %s/%s, want %s/%s", i, got.Value, got.Label, w.value, w.label)
		}
	}
	if models.IsPropertyType("villa") {
		t.Error("villa is still a recognised property type")
	}
}

// Villa is gone from the controls, the data and every rendered page.
func TestVillaIsAbsent(t *testing.T) {
	h := newServer(t)
	for _, path := range []string{
		"/", "/search", "/search?view=map", "/add-listing", "/pricing", "/about",
		"/search?per_page=60",
	} {
		body := mustGet(t, h, path)
		if regexp.MustCompile(`(?i)villa`).MatchString(body) {
			t.Errorf("%s still mentions villa", path)
		}
	}
	// A stale villa link must not filter on nothing and return an empty page.
	all := cardCount(mustGet(t, h, "/search?per_page=60"))
	stale := cardCount(mustGet(t, h, "/search?property_type=villa&per_page=60"))
	if stale != all {
		t.Errorf("a stale villa value returned %d of %d listings; it should be ignored", stale, all)
	}
}

// Every category has dummy stock, so the new options are testable.
func TestEveryCategoryHasDummyListings(t *testing.T) {
	h := newServer(t)
	for _, pt := range models.PropertyTypes {
		body := mustGet(t, h, "/search?per_page=60&property_type="+string(pt.Value))
		if n := cardCount(body); n == 0 {
			t.Errorf("no dummy listings for %s", pt.Value)
		}
	}
}

// ---------------------------------------------------------------------------
// Multi-select filtering
// ---------------------------------------------------------------------------

// Selecting several types returns the union of them, not the intersection.
func TestPropertyTypeFilterIsUnion(t *testing.T) {
	h := newServer(t)
	count := func(q string) int {
		return cardCount(mustGet(t, h, "/search?per_page=60&"+q))
	}

	house := count("property_type=house")
	modular := count("property_type=modular-house")
	panelized := count("property_type=panelized-house")
	union := count("property_type=house&property_type=modular-house&property_type=panelized-house")

	if house == 0 || modular == 0 || panelized == 0 {
		t.Fatalf("a category has no stock: house=%d modular=%d panelized=%d", house, modular, panelized)
	}
	if union != house+modular+panelized {
		t.Errorf("union = %d, want %d (%d+%d+%d) — OR logic, not AND",
			union, house+modular+panelized, house, modular, panelized)
	}
	if union <= house {
		t.Error("adding types did not widen the result set")
	}
}

// All the chosen types come back checked, and Any type does not.
func TestSelectedTypesAreRestoredFromTheURL(t *testing.T) {
	body := squash(mustGet(t, newServer(t),
		"/search?property_type=house&property_type=modular-house&property_type=panelized-house"))

	for _, v := range []string{"house", "modular-house", "panelized-house"} {
		if !regexp.MustCompile(`value="` + v + `" checked`).MatchString(body) {
			t.Errorf("%s did not come back checked", v)
		}
	}
	for _, v := range []string{"cottage", "sauna", "garage"} {
		if regexp.MustCompile(`value="` + v + `" checked`).MatchString(body) {
			t.Errorf("%s was checked but was never selected", v)
		}
	}
	// Any type is off while specific types are on.
	anyTag := between(body, `id="f-type-any"`, ">")
	if strings.Contains(anyTag, "checked") {
		t.Error("Any type is checked even though specific types are selected")
	}
}

// With no property_type in the URL, Any type is the state.
func TestAnyTypeIsTheDefault(t *testing.T) {
	body := squash(mustGet(t, newServer(t), "/search"))

	anyTag := between(body, `id="f-type-any"`, ">")
	if !strings.Contains(anyTag, "checked") {
		t.Error("Any type is not checked on a fresh search")
	}
	if regexp.MustCompile(`name="property_type" value="[a-z-]+" checked`).MatchString(body) {
		t.Error("a specific property type is checked by default")
	}
}

// The controls are checkboxes, which is what makes multiple selection,
// keyboard operation and screen-reader announcement work.
func TestPropertyTypeControlsAreAccessibleCheckboxes(t *testing.T) {
	h := newServer(t)

	for _, path := range []string{"/search", "/search?view=map", "/"} {
		body := mustGet(t, h, path)
		grid := between(body, `class="type-check-grid"`, "</div>")
		if grid == "" {
			t.Errorf("%s has no property-type grid", path)
			continue
		}
		if strings.Contains(grid, `type="radio"`) {
			t.Errorf("%s still uses radios for property type", path)
		}
		if !strings.Contains(grid, `type="checkbox"`) {
			t.Errorf("%s does not use checkboxes for property type", path)
		}
		if !strings.Contains(body, `aria-label="Property type"`) {
			t.Errorf("%s property-type group has no accessible name", path)
		}
	}

	// Any type is never submitted as a category — it carries no name.
	anyTag := between(squash(mustGet(t, h, "/search")), `id="f-type-any"`, ">")
	if strings.Contains(anyTag, "name=") {
		t.Error("the Any type checkbox has a name and would be submitted as a category")
	}
}

// One consistent URL format: repeated property_type, everywhere.
func TestPropertyTypeURLFormatIsConsistent(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h,
		"/search?property_type=house&property_type=cottage")

	// Every view-switch link carries the whole selection forward. Scoped to
	// the switcher itself — "Clear all" also links to ?view=grid and is
	// supposed to arrive with nothing selected.
	links := regexp.MustCompile(`class="view-switch__btn" href="([^"]*)"`).FindAllStringSubmatch(body, -1)
	if len(links) == 0 {
		t.Fatal("no view-switch links found")
	}
	for _, l := range links {
		if n := strings.Count(l[1], "property_type="); n != 2 {
			t.Errorf("a view-switch link carries %d property_type values, want 2: %s", n, l[1])
		}
	}

	// The homepage's browse-by-type tiles use the same parameter.
	home := mustGet(t, h, "/")
	if strings.Contains(home, `href="/search?type=`) {
		t.Error("the homepage still links with the old single-value `type` parameter")
	}
	if !strings.Contains(home, `href="/search?property_type=`) {
		t.Error("the homepage type tiles do not use property_type")
	}
}

// Removing one chip leaves the other selections alone.
func TestRemovingOneTypeChipKeepsTheRest(t *testing.T) {
	body := mustGet(t, newServer(t),
		"/search?property_type=house&property_type=cottage&property_type=sauna")

	// Each selected type gets its own removable chip.
	for _, label := range []string{"House", "Cottage", "Sauna"} {
		if !strings.Contains(body, ">\n        "+label+"\n") && !strings.Contains(squash(body), "> "+label+" <") {
			// Chips render the label as the anchor's text; be tolerant of layout.
			if !strings.Contains(body, label) {
				t.Errorf("no chip for %s", label)
			}
		}
	}
	// A chip's remove link drops only its own value.
	rm := regexp.MustCompile(`href="/search\?([^"]*property_type[^"]*)"[^>]*aria-label="Remove filter Cottage"`).
		FindStringSubmatch(body)
	if rm == nil {
		t.Fatal("the Cottage chip has no remove link")
	}
	if strings.Contains(rm[1], "cottage") {
		t.Error("removing the Cottage chip does not drop cottage")
	}
	for _, keep := range []string{"house", "sauna"} {
		if !strings.Contains(rm[1], keep) {
			t.Errorf("removing the Cottage chip also dropped %s", keep)
		}
	}
}

// ---------------------------------------------------------------------------
// Single-selection controls stay single
// ---------------------------------------------------------------------------

func TestAddListingTakesExactlyOnePropertyType(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")
	section := between(body, `id="ls-category"`, "</section>")
	if section == "" {
		t.Fatal("add-listing has no property category section")
	}

	if strings.Contains(section, `type="checkbox"`) {
		t.Error("add-listing offers property type as checkboxes; a listing has one type")
	}
	if n := strings.Count(section, `type="radio"`); n != len(models.PropertyTypes) {
		t.Errorf("add-listing has %d property-type radios, want %d", n, len(models.PropertyTypes))
	}
	if strings.Contains(section, "Any type") {
		t.Error("add-listing offers Any type; a new listing must have a real category")
	}
	// Every catalogue entry is offered.
	for _, pt := range models.PropertyTypes {
		if !strings.Contains(section, `value="`+string(pt.Value)+`"`) {
			t.Errorf("add-listing does not offer %s", pt.Value)
		}
	}
}

func TestDealTypeStaysSingleSelect(t *testing.T) {
	h := newServer(t)
	for _, path := range []string{"/search", "/search?view=map"} {
		group := between(mustGet(t, h, path),
			`class="segmented segmented--full segmented--deal"`, "</div>")
		if strings.Contains(group, `type="checkbox"`) {
			t.Errorf("%s made deal type multi-select; it must stay single", path)
		}
		if n := strings.Count(group, `type="radio"`); n != 3 {
			t.Errorf("%s has %d deal-type radios, want 3", path, n)
		}
	}
}

// ---------------------------------------------------------------------------
// The menu control
// ---------------------------------------------------------------------------

func TestMenuControlsShareTheSidebarIcon(t *testing.T) {
	body := mustGet(t, newServer(t), "/")

	// Both controls that open the drawer use the same class and icon.
	if n := strings.Count(body, `class="menu-trigger`); n < 2 {
		t.Errorf("only %d menu triggers use the shared class, want the header and floating one", n)
	}
	for _, want := range []string{
		`class="menu-trigger header__burger"`,
		`class="menu-trigger floating-menu"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing menu control: %s", want)
		}
	}
	// The sidebar/panel glyph, not the old hamburger.
	if !strings.Contains(body, `<rect width="18" height="18" x="3" y="3" rx="2" /><path d="M9 3v18" />`) {
		t.Error("the menu controls do not use the sidebar/panel icon")
	}
	// Both carry the accessible name and open the same drawer.
	if n := strings.Count(body, `aria-label="Open navigation menu"`); n < 2 {
		t.Errorf("only %d menu controls carry the accessible name", n)
	}
	if n := strings.Count(body, `aria-controls="mobile-drawer"`); n < 2 {
		t.Errorf("only %d menu controls point at the drawer", n)
	}
}

// The header must still scroll away rather than stick.
func TestHeaderIsNotSticky(t *testing.T) {
	body := mustGet(t, newServer(t), "/static/css/layout.css")
	header := between(body, ".site-header {", "}")
	if strings.Contains(header, "position: sticky") || strings.Contains(header, "position: fixed") {
		t.Error("the header was made sticky; it should scroll out of view")
	}
	if !strings.Contains(header, "position: relative") {
		t.Error("the header lost its relative positioning")
	}
}

// The stylesheet must not reintroduce a coloured plate behind the icon.
func TestMenuTriggerIsTransparent(t *testing.T) {
	css := mustGet(t, newServer(t), "/static/css/layout.css")
	rule := between(css, ".menu-trigger {", "}")
	if rule == "" {
		t.Fatal("no .menu-trigger rule")
	}
	if !strings.Contains(rule, "background: none") {
		t.Error("the menu trigger does not have a transparent background")
	}
	if !strings.Contains(rule, "border: 0") {
		t.Error("the menu trigger still draws a border")
	}
	if strings.Contains(rule, "var(--success)") {
		t.Error("the menu trigger is still tinted green")
	}
	// 44px minimum hit area.
	if !strings.Contains(rule, "width: 44px") || !strings.Contains(rule, "height: 44px") {
		t.Error("the menu trigger is not a 44x44 target")
	}
}

// The filter fragment must push the page URL, not its own endpoint — otherwise
// reloading after a filter change returns bare markup.
func TestFilterFragmentPushesThePageURL(t *testing.T) {
	h := newServer(t)
	req, _ := http.NewRequest(http.MethodGet,
		"/search/results?property_type=house&property_type=cottage", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	push := rec.Header().Get("HX-Push-Url")
	if !strings.HasPrefix(push, "/search?") {
		t.Errorf("HX-Push-Url = %q, want a /search URL", push)
	}
	if n := strings.Count(push, "property_type="); n != 2 {
		t.Errorf("the pushed URL carries %d property_type values, want 2: %s", n, push)
	}
}
