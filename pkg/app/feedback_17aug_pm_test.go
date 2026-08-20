package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// post performs a POST with no body and returns the status and body. The
// shared get() helper only does GETs, and the clone endpoint is a POST because
// it creates something.
func post(t *testing.T, h http.Handler, path string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

// Tests for the client's 17 August afternoon corrections.
//
// Four notes, four sections, each named after what was asked rather than after
// the code that answers it — the split every earlier round uses. Anything the
// server renders is asserted against a real response; the parts that only exist
// once a browser has laid the page out or run JavaScript are asserted at the
// level of the rule or handler that produces them, and were driven in a real
// headless Chrome as well.

// ---------------------------------------------------------------------------
// 14. "next to the user name add more one thing 'direct from the owner' and for
//     this in a frontend there is a special label. As sometimes it is important
//     to note that the property owner itself is selling it."
// ---------------------------------------------------------------------------

func TestDirectFromOwnerIsSetBesideTheName(t *testing.T) {
	h := newServer(t)
	settings := mustGet(t, h, "/settings")

	// Beside the name it qualifies, inside the same field.
	field := section(t, settings, `for="p-name"`, `for="p-company"`, "the name field")
	mustContain(t, field, `name="direct_from_owner"`, "the control belongs next to the name")
	mustContain(t, field, "Direct from the owner", "the client's exact wording")

	// And it says what ticking it does — a label whose effect is invisible from
	// the form is one a seller ticks without meaning it.
	mustContain(t, field, "Adds a label to your listings",
		"the control should say what it produces")
}

func TestDirectFromOwnerLabelAppearsInTheFrontend(t *testing.T) {
	h := newServer(t)

	// On the cards, and on the listing itself.
	search := mustGet(t, h, "/search")
	mustContain(t, search, `class="badge badge--owner"`, "the label must reach the result cards")
	mustContain(t, search, "Direct from the owner", "…with the client's wording")

	// It is a real distinction, not a synonym for a private listing: some
	// listings carry it and some do not.
	total := strings.Count(search, `class="pcard `)
	owned := strings.Count(search, "badge--owner")
	if owned == 0 || owned >= total {
		t.Errorf("%d of %d cards carry the owner label; it must be a genuine "+
			"distinction rather than a label on everything or on nothing", owned, total)
	}

	// Its own colour. Gold is a paid placement and green is a Previa check;
	// this is neither, and on a card that carries all three they have to be
	// tellable apart.
	components := asset(t, cssDir+"/components.css")
	badge := section(t, components, ".badge--owner {", "}", "the owner badge")
	mustContain(t, badge, "var(--teal)", "the label needs a colour of its own")
	mustNotContain(t, badge, "var(--gold)", "gold means a paid placement")
	mustNotContain(t, badge, "var(--success)", "green means Previa verified it")
}

// Private seller and owner are two different questions, and the listing page no
// longer answers the second from the first.
func TestPrivateSellerNoLongerAssertsOwnership(t *testing.T) {
	detail := asset(t, "../../web/templates/pages/property-detail.html")

	private := section(t, detail, `badge--neutral">Private seller`, "Contact seller",
		"the private-seller box")
	mustContain(t, private, "{{ if $p.DirectFromOwner }}",
		"the ownership claim must be conditional, not assumed")
	mustContain(t, private, "Listed privately rather than through an agency",
		"a private listing that makes no ownership claim must not imply one")
	mustNotContain(t, private, "This property is listed directly by its owner. Messages",
		"the old copy asserted ownership of every private listing")
}

// ---------------------------------------------------------------------------
// 15. "remove this string: 'Everything you've published, and everything still
//     waiting to go live' ... There is no pending review as noone will not look
//     the ads before publishing, and so nothing is rejected. So these labele
//     remove. There are labels: draft, active, expired."
// ---------------------------------------------------------------------------

func TestMyListingsHasNoLede(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/my-listings")

	mustNotContain(t, body, "Everything you've published",
		"the client asked for the standing line to go")
	mustNotContain(t, body, "still waiting to go live",
		"nothing waits for anyone to look at it any more, so nothing can say it does")

	// The page keeps its heading — only the sentence under it went.
	mustContain(t, body, `class="page-head__title">My listings<`, "the title stays")
}

func TestListingStatusesAreDraftActiveExpired(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/my-listings")

	tabs := section(t, body, `aria-label="Listing status"`, "</div>", "the status tabs")
	for _, want := range []string{"/my-listings?status=draft", "/my-listings?status=active",
		"/my-listings?status=expired"} {
		mustContain(t, tabs, want, "the three states the client listed must each have a tab")
	}
	mustNotContain(t, tabs, "status=pending", "nothing is pending review")
	mustNotContain(t, tabs, "status=rejected", "nothing is rejected")

	// The labels are gone from the whole page, not only the tab bar.
	mustNotContain(t, body, "Pending review", "the label must not survive on a badge either")
	mustNotContain(t, body, ">Rejected<", "the label must not survive on a badge either")

	// Every state has listings behind it, so each tab returns something.
	for _, st := range []string{"draft", "active", "expired"} {
		page := mustGet(t, h, "/my-listings?status="+st)
		if strings.Contains(page, "Nothing here yet") {
			t.Errorf("the %s tab has no seeded listing behind it, so the state "+
				"cannot be checked in the interface", st)
		}
	}

	// A filter on a state that no longer exists returns nothing rather than
	// everything — a removed status must not quietly widen the view.
	gone := mustGet(t, h, "/my-listings?status=pending")
	mustContain(t, gone, "Nothing here yet", "a status that no longer exists matches nothing")
}

// Expired leaves the seller where a draft does — "it is the same as draft, just
// info for the user that listing active period is expired. So the user can edit
// it, clone it, re-activate it, or delete it."
func TestExpiredOffersTheSameActionsAsDraft(t *testing.T) {
	h := newServer(t)

	actions := func(status string) map[string]bool {
		body := mustGet(t, h, "/my-listings?status="+status)
		row := section(t, body, `data-label="Actions"`, "</td>", status+" row actions")
		found := map[string]bool{}
		for _, a := range []string{"Edit", "Clone", "Re-activate", "Activate", "Delete",
			"Statistics", "View", "Promote"} {
			if strings.Contains(row, `aria-label="`+a+" ") ||
				strings.Contains(row, `aria-label="`+a+` for`) {
				found[a] = true
			}
		}
		return found
	}

	draft, expired := actions("draft"), actions("expired")
	for _, a := range []string{"Edit", "Clone", "Delete"} {
		if !draft[a] || !expired[a] {
			t.Errorf("%s must be offered on both draft and expired listings "+
				"(draft=%v expired=%v)", a, draft[a], expired[a])
		}
	}
	if !draft["Activate"] {
		t.Error("a draft has to have a way to go online")
	}
	if !expired["Re-activate"] {
		t.Error("the client listed re-activation among what an expired listing must offer")
	}

	// And promotion is not offered on either: paying to feature something
	// nobody can see does nothing.
	if draft["Promote"] || expired["Promote"] {
		t.Error("a listing that is not online cannot usefully be promoted")
	}
}

// A draft has never been paid for, so it has nothing to expire.
func TestDraftsShowNoExpiryDate(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/my-listings?status=draft")

	cell := section(t, body, `data-label="Expires"`, "</td>", "the expiry cell")
	if regexp.MustCompile(`\d{4}`).MatchString(cell) {
		t.Errorf("a draft shows an expiry date (%q); it has never been online, "+
			"so there is nothing for it to expire from", strings.TrimSpace(cell))
	}
}

// ---------------------------------------------------------------------------
// 16. "on the right side there are the buttons featured and edit, there add
//     'clone' so can duplicate the ad. And then there somewhere has to be the
//     statistics, what shows how many visitors have seen this ad per day, so
//     statistic window."
// ---------------------------------------------------------------------------

func TestCloneDuplicatesAListingAsADraft(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/my-listings")

	mustContain(t, body, `hx-post="/listing/clone/`, "every row needs a clone action")

	code, out := post(t, h, "/listing/clone/pr-003")
	if code != http.StatusOK {
		t.Fatalf("POST /listing/clone/pr-003 = %d, want 200", code)
	}
	mustContain(t, out, "Copied to a new draft", "the copy lands as a draft, never as active")
	mustContain(t, out, "(copy)", "the copy must be tellable apart from the original")

	// A listing the user does not own cannot be copied — a hand-edited id must
	// not duplicate somebody else's advertisement.
	code, _ = post(t, h, "/listing/clone/pr-999")
	if code != http.StatusNotFound {
		t.Errorf("cloning a listing the user does not own = %d, want 404", code)
	}
}

func TestStatisticsWindowChartsVisitorsPerDay(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/my-listings")

	mustContain(t, body, `aria-label="Statistics for`, "every row needs a statistics action")
	mustContain(t, body, "Visitors per day, last 14 days", "the client asked for views per day")
	mustContain(t, body, `x-for="d in stats.days"`, "the chart iterates the daily series")

	// The series really is handed over, and it really is per day.
	m := regexp.MustCompile(`stats = (\{&#34;id&#34;.*?\})"`).FindStringSubmatch(body)
	if m == nil {
		t.Fatal("no statistics payload found on the page")
	}
	var stats struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Views int    `json:"views"`
		Days  []struct {
			Label string `json:"label"`
			Views int    `json:"views"`
		} `json:"days"`
	}
	raw := strings.NewReplacer("&#34;", `"`, "&amp;", "&", "&#39;", "'").Replace(m[1])
	if err := json.Unmarshal([]byte(raw), &stats); err != nil {
		t.Fatalf("statistics payload is not valid JSON: %v\n%s", err, raw)
	}
	if len(stats.Days) != 14 {
		t.Errorf("the series carries %d days, want 14", len(stats.Days))
	}
	if stats.Title == "" || stats.ID == "" {
		t.Error("the panel needs to know which listing it is describing")
	}
	total := 0
	for _, d := range stats.Days {
		if d.Label == "" {
			t.Error("every day needs an axis label")
		}
		if d.Views < 0 {
			t.Errorf("a day reports %d views", d.Views)
		}
		total += d.Views
	}
	if total == 0 {
		t.Error("a listing with hundreds of lifetime views must chart something")
	}
}

// The same listing must draw the same chart on every request. A series that
// moved between two page loads would look like a bug in the numbers.
func TestStatisticsAreStableBetweenRequests(t *testing.T) {
	h := newServer(t)
	first := mustGet(t, h, "/my-listings")
	second := mustGet(t, h, "/my-listings")

	pat := regexp.MustCompile(`stats = (\{&#34;id&#34;.*?\})"`)
	a := pat.FindAllString(first, -1)
	b := pat.FindAllString(second, -1)
	if len(a) == 0 {
		t.Fatal("no statistics payloads on the page")
	}
	if strings.Join(a, "|") != strings.Join(b, "|") {
		t.Error("the statistics changed between two identical requests")
	}
}

// ---------------------------------------------------------------------------
// 17. "the up/down arrows for the numbers, use the same style as in sexydate,
//     so these arrows are on the right corner of the field. At the moment they
//     are centered, this is bad UX. And the m2 remove from the field, this m2
//     add in (m2) behind the text, like 'Living area (m2)'"
// ---------------------------------------------------------------------------

func TestUnitsMovedFromTheFieldIntoTheLabel(t *testing.T) {
	h := newServer(t)
	search := mustGet(t, h, "/search")
	wizard := mustGet(t, h, "/add-listing")

	mustContain(t, search, "Living area (m²)", "the client's exact example")
	mustContain(t, search, "Minimum land area (m²)", "the same treatment for the other area field")
	mustContain(t, wizard, "Living area (m²)", "the wizard asks the same question and reads the same")

	// The suffix drawn inside the box is gone everywhere, not only where the
	// client happened to photograph it.
	for _, body := range []string{search, wizard} {
		mustNotContain(t, body, "input-unit__suffix", "no unit may be drawn inside a field")
	}
	components := asset(t, cssDir+"/components.css")
	mustNotContain(t, components, ".input-unit__suffix", "the rule behind it must not linger as dead CSS")
	mustNotContain(t, components, ".input-unit .input", "…nor the padding it reserved")
}

// The arrows sit in the corner because nothing reserves space to their right
// any more, and they are drawn by us because the browser's own cannot be made
// to look like the reference.
//
// The client came back to this a second time on 17 August — "the up/down arrows
// need to make better, make the same style as in sexydate, that they sit on the
// right side of the field" — with a screenshot of a ruled column of carets
// pinned to the field's edge. Chrome's native stepper is an unstylable grey
// slab and Firefox's takes no styling at all, so it is hidden and replaced.
func TestNumberSteppersSitInTheCorner(t *testing.T) {
	components := asset(t, cssDir+"/components.css")

	native := section(t, components, "input[type='number']::-webkit-outer-spin-button,",
		"}", "the browser's own stepper")
	mustContain(t, native, "appearance: none", "the native stepper must be out of the way")

	ours := section(t, components, ".num-stepper__arrows {", "}", "our stepper's column")
	mustContain(t, ours, "position: absolute", "pinned to the field rather than in its text flow")
	mustContain(t, ours, "right: 1px", "…against the right edge, inside the border")
	mustContain(t, ours, "border-left: 1px solid var(--border)", "…ruled off from the digits")
	mustContain(t, ours, "flex-direction: column", "up above down, as in the reference")

	// Room for the column, or a five-digit price runs under it.
	mustContain(t, section(t, components, ".num-stepper .input {", "}", "the field itself"),
		"padding-right", "the field must leave room for the arrows")

	// Every number field gets one, wherever it is added.
	js := asset(t, "../../public/static/js/previa.js")
	mustContain(t, js, `querySelectorAll('input[type="number"]')`,
		"the arrows are attached to every number field rather than field by field")

	// Nothing may reintroduce padding on the right of a number field — that is
	// what pushed the arrows inward in the first place.
	filters := asset(t, "../../web/templates/components/filters.html")
	mustNotContain(t, filters, "input-unit",
		"a wrapper reserving room for a suffix is what centred the arrows")
}
