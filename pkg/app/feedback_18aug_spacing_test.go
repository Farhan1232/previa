package app_test

import (
	"testing"
)

// Tests for the 18 August spacing pass — the round the client raised with four
// screenshots of the homepage's panels and one of a Location field, asking for
// "spacing, padding, margins, alignment and overall visual balance" rather than
// for any change of design.
//
// The panel inset the screenshots were taken against was already answered by
// the round before this one, and TestPanelContentIsInsetFromThePanelEdge below
// pins it so it cannot quietly come back. What this round changes is the rest
// of what those screenshots showed: grids that ran on gaps of their own, a
// feature row whose columns each found their own line, a caption sitting closer
// to the bottom of its tile than to its sides, and one Location field whose
// text ran under its clear button.
//
// Every figure below was measured in a real headless Chrome at 1440, 1280,
// 1024, 960, 900, 820, 768, 640, 520, 480, 390 and 360, and the measured
// numbers are quoted in the comments so a later change that moves them is
// obvious.

// ---------------------------------------------------------------------------
// The spacing system: one gap token for every card grid
// ---------------------------------------------------------------------------

// --card-gap is 24 / 20 / 16 across the breakpoints, and a grid that sets its
// own figure instead is a grid that will drift away from the ones above and
// below it. Popular locations was the one doing it — 16px at every width,
// against 24px for the property, development and article rows in the same
// column — and inside a panel the difference showed against the panel's own
// 32px inset in a single glance.
func TestEveryHomepageCardGridRunsOnTheSameGapToken(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")
	layout := asset(t, cssDir+"/layout.css")

	mustContain(t, layout, ".grid { display: grid; gap: var(--card-gap); }",
		"the shared grid keeps the token")

	loc := section(t, pages, ".loc-grid {", "}", "the popular-locations grid")
	mustContain(t, loc, "gap: var(--card-gap);",
		"the location tiles must sit on the same gap as every other card row")

	trust := section(t, pages, ".trust-grid {", "}", "the why-Previa grid")
	mustContain(t, trust, "gap: var(--card-gap);",
		"the feature columns must sit on the same gap as every other row")
}

// The tokens the gap resolves to, so the desktop / tablet / mobile ladder the
// client asked for is pinned in one place rather than inferred from each grid.
func TestGapAndGutterTokensKeepTheirLadder(t *testing.T) {
	tokens := asset(t, cssDir+"/tokens.css")

	mustContain(t, tokens, "--gutter: 32px;", "desktop section padding is 32px")
	mustContain(t, tokens, "--card-gap: 24px;", "desktop card gap is 24px")
	mustContain(t, tokens, "--gutter: 24px; --card-gap: 20px;",
		"tablet steps down to 24px padding")
	mustContain(t, tokens, "--gutter: 16px; --card-gap: 16px;",
		"mobile steps down to 16px padding")
}

// ---------------------------------------------------------------------------
// Panels: the content boundary every heading, link and card shares
// ---------------------------------------------------------------------------

// Measured at 1440: the panel spans 112–1328 and its content 144–1296, so
// heading, subtitle, action link and cards all start 32px inside the panel and
// end 32px short of it. At 1024 that inset is 24px and at 390 it is 16px,
// because it is one --gutter throughout.
func TestPanelContentIsInsetFromThePanelEdge(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")

	inset := section(t, layout, ".section--band > .container,", "}", "the panel inset")
	mustContain(t, inset, "padding-inline: calc(var(--gutter) * 2);",
		"content inside a panel must clear the panel's own edge by one gutter")
	mustContain(t, inset, ".section--navy > .container {",
		"the navy panel is inset the same way as the band")
}

// Heading to subtitle, and the header block to what it introduces: 8px and
// 20px, one pair of figures for every section on the site.
//
// 20px since 19 August, and the per-section exception with it: "the content
// blocks closer the it's title … then look that the distance gaps between the
// blocks would be the same." A first block tucked closer to its cards than the
// eight below it is the mismatch that note is about, so there is now exactly
// one figure and no override may reintroduce a second.
func TestSectionHeaderRhythm(t *testing.T) {
	layout := asset(t, cssDir+"/layout.css")

	head := section(t, layout, ".section__head {", "}", "the section header")
	mustContain(t, head, "margin-bottom: var(--sp-5);", "header to content is 20px")

	text := section(t, layout, ".section__head-text {", "}", "the header's text column")
	mustContain(t, text, "gap: var(--sp-2);", "heading to subtitle is 8px")

	mustNotContain(t, layout, ".section--after-hero .section__head { margin-bottom:",
		"no section may take a heading gap of its own")
}

// ---------------------------------------------------------------------------
// Why people search on Previa: four columns, one row
// ---------------------------------------------------------------------------

// The four features are a row, and a row that only lines up when every title
// happens to be one line long is not lined up at all: "Search the map, not a
// list" wrapped to two, and its description dropped 26px below the other
// three. The columns are subgrids of the same three rows now, so the icons,
// the titles and the descriptions each share a line whatever the wrap.
//
// Measured at 1440: title tops 5788 × 4, description tops 5829 × 4. At 960 and
// 768, where the grid is two columns, the same holds per band.
func TestFeatureColumnsShareTheirRows(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	grid := section(t, pages, ".trust-grid {", "}", "the feature grid")
	mustContain(t, grid, "grid-template-rows: auto auto auto;",
		"the grid must declare the three rows the columns line up on")

	item := section(t, pages, ".trust-item {", "}", "a feature column")
	mustContain(t, item, "grid-template-rows: subgrid;", "a column takes its rows from the grid")
	mustContain(t, item, "grid-row: span 3;", "and spans all three of them")
	mustContain(t, item, "row-gap: var(--sp-3);",
		"the icon-title-text rhythm inside a column is unchanged at 12px")
}

// Once the columns wrap, the gap between one feature and the next is doing a
// different job from the gap beside it: at --card-gap's mobile value of 16 it
// was barely clear of the 12px inside a feature. It holds at 24 from the point
// the grid stops being four across.
func TestStackedFeaturesKeepAClearRowGap(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	wrapped := section(t, pages, "@media (max-width: 960px) {\n  .trust-grid", "}", "the wrapped feature grid")
	mustContain(t, wrapped, "row-gap: var(--sp-6);",
		"stacked features stay 24px apart, clear of their own 12px rhythm")
}

// ---------------------------------------------------------------------------
// Popular locations: the caption inside a tile
// ---------------------------------------------------------------------------

// The city and its count sat 20px from the sides of the tile and 16px from the
// bottom, which read as text dropped near a corner rather than as a padded
// caption. One inset on all three edges it touches.
func TestLocationTileCaptionIsEvenlyInset(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	caption := section(t, pages, ".loc-tile__text {", "}", "the tile caption")
	mustContain(t, caption, "left: var(--sp-5); right: var(--sp-5); bottom: var(--sp-5);",
		"the caption keeps the same 20px from every edge it touches")
}

// ---------------------------------------------------------------------------
// The Location field: one formula, every variant
// ---------------------------------------------------------------------------

// The control carries a pin on the left and a clear button on the right, and
// its two paddings have to clear them both. Those figures live on the field as
// tokens now, so a variant that wants a tighter control overrides a token
// rather than restating a padding — which is exactly how the homepage panel
// came to overlap them.
func TestLocationFieldSpacingComesFromTokens(t *testing.T) {
	components := asset(t, cssDir+"/components.css")

	field := section(t, components, ".location-field {", "}", "the location field")
	for _, want := range []string{
		"--loc-icon-inset: var(--sp-3);",
		"--loc-icon-size: 18px;",
		"--loc-gap: var(--sp-2);",
		"--loc-clear-size: 22px;",
		"--loc-clear-inset: var(--sp-2);",
	} {
		mustContain(t, field, want, "the field must carry "+want+" as a token")
	}

	control := section(t, components, ".location-field__control .input {", "}", "the field's gutters")
	mustContain(t, control, "padding-left: calc(var(--loc-icon-inset) + var(--loc-icon-size) + var(--loc-gap));",
		"the text must start one gap past the pin")
	mustContain(t, control, "padding-right: calc(var(--loc-clear-inset) + var(--loc-clear-size) + var(--loc-gap));",
		"and stop one gap short of the clear button")

	// The pin reads the same inset the padding is calculated from, so the two
	// cannot drift apart.
	icon := section(t, components, ".location-field__control > svg {", "}", "the pin")
	mustContain(t, icon, "left: var(--loc-icon-inset);", "the pin sits on the token")
	mustContain(t, icon, "transform: translateY(-50%);", "and is centred on the field")
}

// The homepage panel sets a flat 12px padding-inline for its compact controls,
// and it lands after the component's own rule in the cascade. The location
// field's right-hand gutter collapsed to 12px while its clear button still sat
// 8px in at 22px wide: measured in Chrome, a typed value — "Kalamaja, Tallinn,
// Estonia" — ran 18px under the cross. Restating both gutters from the same
// tokens puts it back to 8px of clearance, the sidebar's figure.
func TestHeroPanelKeepsTheLocationFieldsGutters(t *testing.T) {
	pages := asset(t, cssDir+"/pages.css")

	hero := section(t, pages, ".searchbox .location-field__control .input {", "}",
		"the panel's location field")
	mustContain(t, hero, "padding-left: calc(var(--loc-icon-inset) + var(--loc-icon-size) + var(--loc-gap));",
		"the panel keeps the component's left gutter")
	mustContain(t, hero, "padding-right: calc(var(--loc-clear-inset) + var(--loc-clear-size) + var(--loc-gap));",
		"and its right one, so a value cannot run under the clear button")

	// A label 4px off its control reads as sitting on it. The panel stays
	// compact — this is the sidebar's step, not the 8px the page forms use.
	mustContain(t, pages, ".searchbox .field { gap: 6px; }",
		"the panel's label-to-control step is 6px")
}
