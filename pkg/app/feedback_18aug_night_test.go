package app_test

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"previa/pkg/config"
	"previa/pkg/data"
	"previa/pkg/handlers"
	"previa/pkg/models"
	"previa/pkg/view"
)

// Tests for the client's last batch of 18 August notes (22:03–22:26).
//
// Nine notes, numbered 66–74 on from the sixty-five before them, each section
// named after what was asked rather than after the code that answers it. Six of
// the nine are small and exact. The other three — the double-header pages, the
// listings page and the homepage banner, sent as three consecutive messages —
// are one instruction described three times ("cut the header the same width as
// the page content, and make the below corners round") and are tested as the
// one system they describe.
//
// Where a note is about a number the browser computes, the number in the
// comment was measured in a real headless Chrome at 1440 with the dev server
// running, so a later change that moves it is obvious rather than invisible.

// cssCode reads a stylesheet with its comments stripped.
//
// Several notes in this round are about something being absent, and the
// comments in these files quote the client's words — which include the very
// strings the assertions forbid. Prose explaining why a property is not set
// must not read as the property being set.
var cssComment = regexp.MustCompile(`(?s)/\*.*?\*/`)

// The same, for the one asset that is XML rather than CSS.
var xmlComment = regexp.MustCompile(`(?s)<!--.*?-->`)

func cssCode(t *testing.T, name string) string {
	t.Helper()
	return cssComment.ReplaceAllString(asset(t, cssDir+"/"+name), "")
}

// ---------------------------------------------------------------------------
// 66. "Here in the main search menu remove the separator - between these
//     fields."
// ---------------------------------------------------------------------------

func TestThePairedFiltersHaveNoSeparatorBetweenThem(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, "/search")

	mustNotContain(t, body, "range-pair__dash", "no dash may be drawn between a Min and a Max")
	mustNotContain(t, body, `aria-hidden="true">–<`, "and none may survive as a bare character either")

	// Every pair, not only the price one the screenshot was taken of: the same
	// component draws the living area and the construction year, and leaving
	// the dash on two of the three would have answered the note by a third.
	filters := asset(t, "../../web/templates/components/filters.html")
	mustNotContain(t, filters, "range-pair__dash", "the dash is gone from the component, so from all three pairs")
	for _, name := range []string{"price_min", "price_max", "area_min", "area_max", "year_min", "year_max"} {
		mustContain(t, body, `name="`+name+`"`, "the field itself stays: "+name)
	}

	// Two columns now, not three. Without this the pair would keep a third
	// track sized to an element that no longer exists and the two boxes would
	// stop short of the sidebar's width.
	components := cssCode(t, "components.css")
	rule := section(t, components, ".range-pair {", "}", "the min/max pair")
	mustContain(t, rule, "grid-template-columns: 1fr 1fr;", "the two boxes must share the row evenly")
	mustNotContain(t, components, ".range-pair__dash", "the dash's own rule goes with it")
}

// ---------------------------------------------------------------------------
// 67. "And around these fields, when start to type there, then the bold
//     borderline will come. Make this borderline green #008000 and thinner —
//     it stays green in day and dark mode."
// ---------------------------------------------------------------------------

func TestTheFieldYouTypeInIsFramedInGreen(t *testing.T) {
	tokens := cssCode(t, "tokens.css")

	// The client's colour, literally, and a one-pixel ring rather than the
	// three-pixel halo that was there — "and thinner". Border plus ring was
	// four pixels of colour; it is two.
	mustContain(t, tokens, "--field-focus: #008000;", "the frame is the green the client named")
	mustContain(t, tokens, "--field-focus-ring: 0 0 0 1px #008000;", "…and one pixel of it, not three")

	// "It stays green in day and dark mode": declared once, and never again
	// inside the dark block. --focus-color flips navy → gold down there, which
	// is exactly what this must not do.
	dark := section(t, tokens, "[data-theme='dark'] {", "\n}", "the dark theme block")
	mustNotContain(t, dark, "--field-focus", "the night theme may not repaint the field frame")

	components := cssCode(t, "components.css")
	rule := section(t, components, ".input:focus,\n.select:focus,\n.textarea:focus {", "}", "a focused field")
	mustContain(t, rule, "border-color: var(--field-focus);", "the border takes the green")
	mustContain(t, rule, "box-shadow: var(--field-focus-ring);", "…and so does the ring around it")
	mustNotContain(t, rule, "var(--focus-ring);", "the themed ring must not be drawn as well")

	// Only the fields. Everything else that shows focus — buttons, checkboxes,
	// cards, links — keeps the theme's own ring, because the note is about what
	// happens "when start to type there".
	mustContain(t, components, ".btn:focus-visible { outline: none; box-shadow: var(--focus-ring); }",
		"a button keeps the themed ring")
	base := cssCode(t, "base.css")
	mustContain(t, base, "outline: 2px solid var(--focus-color);", "and so does everything a keyboard reaches")
}

// ---------------------------------------------------------------------------
// 68. "The favicon backgroun lets try: #8B008B"
// ---------------------------------------------------------------------------

func TestTheFaviconTileIsTheDarkerMagenta(t *testing.T) {
	favicon := asset(t, imgDir+"/favicon.svg")
	mustContain(t, favicon, `<rect width="64" height="64" rx="14" fill="#8B008B"/>`,
		"the tile takes the colour the client asked to try")
	mustNotContain(t, xmlComment.ReplaceAllString(favicon, ""), "#cc00cc",
		"the first magenta must not still be painted anywhere")

	// The mark is generated, not hand-placed, so the generator has to carry the
	// same colour: the next run of it would otherwise put the old one back.
	gen := asset(t, "../../docs/logo_gen.py")
	mustContain(t, gen, `TILE_FILL = "#8B008B"`, "the generator must produce the tile that ships")
}

// ---------------------------------------------------------------------------
// 69. "So this admin panel is the website backend? All user's do not have
//     access there"
// ---------------------------------------------------------------------------

// adminRoutes is every path behind the gate. The list is deliberately written
// out rather than derived: a route added to the panel and forgotten here is the
// one that would be left open, and a test that walks the same table as the code
// cannot notice that.
var adminRoutes = []string{
	"/admin", "/admin/listings", "/admin/users", "/admin/brokers", "/admin/agencies",
	"/admin/developments", "/admin/articles", "/admin/banners", "/admin/packages",
	"/admin/payments", "/admin/languages", "/admin/strings", "/admin/seo", "/admin/maps",
	"/admin/restricted", "/admin/settings", "/admin/backups", "/admin/files",
	"/admin/database", "/admin/system",
}

// plainAccount is the seeded account with its administrator role taken away —
// an ordinary signed-in visitor, which is who the client's note is about.
type plainAccount struct{ data.AccountRepository }

func (p plainAccount) CurrentUser(ctx context.Context) models.User {
	u := p.AccountRepository.CurrentUser(ctx)
	u.Role = "user"
	return u
}

// newServerAsPlainUser wires the same application with that account signed in.
func newServerAsPlainUser(t *testing.T) http.Handler {
	t.Helper()
	cfg := config.Load()
	cfg.TemplateDir = "../../web/templates"
	cfg.StaticDir = "../../public/static"

	engine, err := view.New(os.DirFS(cfg.TemplateDir), true)
	if err != nil {
		t.Fatalf("view.New: %v", err)
	}
	store := data.NewStore(time.Now())
	store.Account = plainAccount{store.Account}
	return handlers.New(store, engine, cfg).Routes()
}

func TestAnOrdinaryAccountCannotReachTheAdminPanel(t *testing.T) {
	h := newServerAsPlainUser(t)

	// Not a redirect and not a 403: both of those confirm there is something
	// there. The back office answers a stranger the way it would answer a typo.
	for _, path := range adminRoutes {
		code, body := get(t, h, path)
		if code != http.StatusNotFound {
			t.Errorf("GET %s as an ordinary user = %d, want 404", path, code)
		}
		mustNotContain(t, body, "admin-nav", path+": no part of the panel may render")
	}

	// The link is gone too. Hiding it without closing the routes would not be
	// access control, and closing the routes while still offering the link
	// would send the account's own owner to a 404.
	body := mustGet(t, h, "/")
	mustNotContain(t, body, `href="/admin"`, "an ordinary account is not offered the panel")
	mustNotContain(t, body, "Admin panel", "…not by that name either")
}

func TestTheAdministratorStillHasTheBackend(t *testing.T) {
	h := newServer(t) // the seeded account, which is an administrator

	for _, path := range adminRoutes {
		if code, _ := get(t, h, path); code != http.StatusOK {
			t.Errorf("GET %s as an administrator = %d, want 200", path, code)
		}
	}

	body := mustGet(t, h, "/")
	mustContain(t, body, `href="/admin"`, "an administrator is offered the panel")
	mustContain(t, body, "Admin panel", "…under the name it has in the menu")

	// One rule, asked in both places — the menu before it draws the link and
	// the router before it renders a page.
	header := asset(t, "../../web/templates/partials/header.html")
	mustContain(t, header, "{{ if .User.IsAdmin }}", "the menu asks the account, not the URL")
	routes := asset(t, "../../pkg/handlers/routes.go")
	for _, path := range adminRoutes {
		mustContain(t, routes, `"GET `+path+`", h.requireAdmin(`, path+" must go through the gate")
	}
	mustContain(t, routes, `"POST /admin/mock-action", h.requireAdmin(`,
		"the panel's one writing route is gated as well")
}

// ---------------------------------------------------------------------------
// 70. "The footer distance to the upper menu block make in every page like it
//     is in frontpage: so smaller"
// ---------------------------------------------------------------------------

func TestEveryPageClosesOnTheFooterLikeTheHomepage(t *testing.T) {
	layout := cssCode(t, "layout.css")

	// The homepage's gap is the 24px above the footer panel and nothing else:
	// it closes on a coloured panel whose own padding is inside it. Every other
	// page was adding a full --section-y on top of that — measured at 1440, an
	// 82px gap against the homepage's 24 — so the last block stops padding
	// itself at the bottom and the two agree.
	mustContain(t, layout, ".page__foot { margin-top: auto; padding-top: var(--sp-6); background: var(--bg); }",
		"the ground under the footer is what sets the gap")
	mustContain(t, layout, ".page__main > :last-child:not(.section--band):not(.section--navy) { padding-bottom: 0; }",
		"the last block on the page must not add a second gap under itself")

	// Panels are the exception, and have to be: their padding is the room
	// inside the shape. Taking it away would put the homepage's article cards
	// against the bottom edge of their own panel.
	mustContain(t, layout, ".section { padding-block: var(--section-y); }",
		"a section in the middle of a page keeps its rhythm")

	// An inline style is the one thing that rule cannot correct, and the
	// listings page carried one.
	search := asset(t, "../../web/templates/pages/search.html")
	mustNotContain(t, search, "padding-bottom:var(--sp-11)", "the listings page must not set its own bottom gap")
}

// ---------------------------------------------------------------------------
// 71–73. "So now the header, in these pages where we have the 'double-header',
//        cut the header the same width as the page content, and make the below
//        corners round."
//        "In the main listing page cut the header as well, make it the same
//        width as the page content, and make the below corners round."
//        "In the frontpage make the width of the headet and banner the same as
//        the content in the listings page, and banner down corners cut round."
// ---------------------------------------------------------------------------

func TestTheHeaderIsCutToThePageContentWidth(t *testing.T) {
	tokens := cssCode(t, "tokens.css")
	layout := cssCode(t, "layout.css")

	// One number, defined once. Its default is the footer's formula character
	// for character — the content box of a .container — so the header and the
	// footer are the same shape at opposite ends of every page.
	const width = "min(100% - var(--gutter) * 2, calc(var(--container) - var(--gutter) * 2))"
	mustContain(t, tokens, "--page-width: "+width+";", "the page width is the footer's width")
	mustContain(t, section(t, layout, ".site-footer {", "}", "the footer"), "width: "+width,
		"…and the footer still states it itself, so the two cannot drift apart")

	header := section(t, layout, ".site-header {", "}", "the header")
	mustContain(t, header, "width: var(--page-width);", "the header stops where the page content stops")
	mustContain(t, header, "margin-inline: auto;", "…centred on the page")
	mustContain(t, header, "border-radius: 0 0 var(--r-xl) var(--r-xl);", "…with the bottom corners round")
	mustNotContain(t, header, "position: sticky", "and it still scrolls away rather than pinning")

	// Measured at 1440: the header runs 105 → 1321, which is where the footer
	// runs and where the cards between them run.
	mustContain(t, layout, ".site-header > .container,\n.page-head > .container { max-width: none; }",
		"the bar's own content sits one gutter inside the panel, as the footer's does")
}

func TestTheDoubleHeaderIsOneShape(t *testing.T) {
	layout := cssCode(t, "layout.css")

	// The double header is the bar and the page head: two elements, one colour,
	// read as one block. Both are cut to the same width and only the lower one
	// is rounded — a rounded bar with a square block under it would read as two
	// stacked cards with a notch between them.
	head := section(t, layout, ".page-head {", "}", "the page head")
	mustContain(t, head, "width: var(--page-width);", "the page head is cut to the same width as the bar")
	mustContain(t, head, "margin-inline: auto;", "…and centred with it")
	mustContain(t, head, "border-radius: 0 0 var(--r-xl) var(--r-xl);", "…and it carries the round corners")
	mustContain(t, layout, "body:has(.page-head) .site-header { border-radius: 0; }",
		"the bar above it must square off, or the block reads as two")

	// The hairline between them stays: it is the one place on the site where
	// two blocks share a colour, which is the client's own test for a
	// separator having earned its place.
	mustContain(t, layout, "body:not(:has(.page-head)) .site-header { border-bottom: 0; }",
		"a page with no page head still has no line under the bar")

	// Every double-header page really is one — the head is the block straight
	// after the bar, on both container widths.
	h := newServer(t)
	for _, path := range []string{"/articles", "/faq", "/brokers", "/developments", "/settings", "/admin"} {
		body := mustGet(t, h, path)
		mustContain(t, body, `class="page-head`, path+" is a double-header page")
	}
}

func TestTheWidePagesCutTheHeaderToTheirOwnContent(t *testing.T) {
	layout := cssCode(t, "layout.css")

	// "The same width as the page content" is a different number on the two
	// families laid out on the wide container: measured at 1440, the listings
	// grid and the admin shell run 32 → 1393, not 105 → 1321. A header cut to
	// the narrower figure would have answered the note with the wrong one.
	mustContain(t, layout, "body:has(.container--wide) { --page-width: calc(100% - var(--gutter) * 2); }",
		"a wide page cuts the header to its own content")
	mustContain(t, layout, "body:has(.container--wide) .site-footer { width: var(--page-width); }",
		"…and closes on a footer of the same width, or the page disagrees with itself")

	h := newServer(t)
	mustContain(t, mustGet(t, h, "/search"), "container--wide", "the listings page is a wide page")
	mustContain(t, mustGet(t, h, "/admin"), "container--wide", "so is the administration panel")

	// The map screens are the exception: a full-viewport shell with no content
	// column to match, where a panel-width bar would be the same mismatch drawn
	// the other way round.
	mustContain(t, layout, "body:has(.map-shell) { --page-width: 100%; }",
		"the map screens keep an edge-to-edge header")
	mustContain(t, layout, "body:has(.map-shell) .site-header { border-radius: 0; }",
		"…and square corners with it")
	mustContain(t, mustGet(t, h, "/search?view=map"), "map-shell", "the map screen is what that rule names")
}

func TestTheHomepageBannerIsCutAndRounded(t *testing.T) {
	pages := cssCode(t, "pages.css")
	layout := cssCode(t, "layout.css")

	hero := section(t, pages, ".hero--flush {", "}", "the homepage banner")
	mustContain(t, hero, "width: var(--page-width);", "the banner is cut to the same width as the header")
	mustContain(t, hero, "margin-inline: auto;", "…and centred under it")
	mustContain(t, hero, "margin-top: calc(var(--header-h) * -1);", "…and still runs up under the bar")
	mustNotContain(t, hero, "overflow: hidden", "clipping the hero would clip the panels the search box opens")

	// The rounding is drawn on the layer that actually shows a colour, which is
	// also the layer that already clips its own image.
	media := section(t, pages, ".hero__media {", "}", "the banner's photograph")
	mustContain(t, media, "border-radius: 0 0 var(--r-xl) var(--r-xl);", "the banner's bottom corners are round")
	mustContain(t, media, "overflow: hidden;", "…and the photograph is clipped to them")

	// The bar above it squares off for the same reason a page head makes it
	// square off: something else finishes the block.
	flush := section(t, layout, "body:has(.hero--flush) .site-header {", "}", "the header above a flush hero")
	mustContain(t, flush, "border-radius: 0;", "the bar must not round off in the middle of the banner")
	mustContain(t, flush, "border-bottom: 0;", "and it still draws no line across the photograph")

	mustContain(t, mustGet(t, newServer(t), "/"), "hero--flush", "the homepage is what those rules name")
}

// ---------------------------------------------------------------------------
// 74. "In the single listing page remove this: Optional — brokers usually reply
//     faster by phone."
// ---------------------------------------------------------------------------

func TestTheEnquiryFormNoLongerAdvertisesThePhone(t *testing.T) {
	h := newServer(t)
	body := mustGet(t, h, firstPropertyPath(t, h))

	mustNotContain(t, body, "brokers usually reply faster by phone",
		"the sentence the client pointed at must be gone")
	mustNotContain(t, body, "Optional — brokers", "…including any half of it")

	// The field itself stays, and stays optional — which the browser reads from
	// the absence of `required`, not from a sentence under the box.
	form := section(t, body, `id="c-phone"`, "</div>", "the phone field")
	mustContain(t, form, `name="phone"`, "the number can still be given")
	mustNotContain(t, form, "required", "…and is still not demanded")
}

// firstPropertyPath returns a real listing URL from the search results, so this
// test does not depend on a slug that seeding may renumber.
func firstPropertyPath(t *testing.T, h http.Handler) string {
	t.Helper()
	body := mustGet(t, h, "/search")
	const marker = `href="/property/`
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatal("no listing link on the search page")
	}
	rest := body[i+len(`href="`):]
	return rest[:strings.Index(rest, `"`)]
}
