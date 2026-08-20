package app_test

import (
	"strings"
	"testing"
)

// Tests for the client's 19 August notes (01:55–02:25).
//
// Six notes across five messages, numbered 75–80 on from the seventy-four
// before them, each section named after what was asked rather than after the
// code that answers it. Four are exact and small; the amenity list is long and
// is checked item by item, because "add these twenty-two" is a note that is
// either completely answered or not answered at all.

// ---------------------------------------------------------------------------
// 75. "The payment methods make there in 2 columns."
// ---------------------------------------------------------------------------

func TestPaymentMethodsAreInTwoColumns(t *testing.T) {
	components := cssCode(t, "components.css")

	// Two, stated as two. An auto-fit track list reads "2 columns" as a minimum
	// and gives four on a wide screen, which is how this was first written and
	// is not what was asked for.
	grid := section(t, components, ".pay-grid {", "}", "the payment grid")
	mustContain(t, grid, "grid-template-columns: repeat(2, minmax(0, 1fr));",
		"the methods must be laid out in exactly two columns")
	mustContain(t, grid, "display: grid;", "…by a grid, not by the stack they used to sit in")
	mustContain(t, components, ".pay-grid { grid-template-columns: minmax(0, 1fr); }",
		"one column on a phone, where two would be about 155px each")

	body := mustGet(t, newServer(t), "/checkout")
	grid = section(t, body, `class="pay-grid"`, "</div>", "the rendered grid")
	for _, label := range []string{"Credit card", "PayPal", "Stripe", "Paysera bank links", "Crypto"} {
		mustContain(t, grid, ">\n                      "+label, "every method is inside the grid: "+label)
	}

	// The column the card sits in had to grow for two of anything to fit: at
	// the old reading width it was 404px, so a tile was 175px — narrower than
	// the words "Credit card" set on one line.
	mustContain(t, body, `class="container section section--after-head"`,
		"the checkout page is on the ordinary container now, not the reading one")
	mustNotContain(t, body, "container--narrow", "…and not on both at once")
}

// ---------------------------------------------------------------------------
// 76. "The text under credit card write 'processed via Stripe'. Under Paypal
//     and Stripe remove the text as everyone knows anyway what they are."
// ---------------------------------------------------------------------------

func TestOnlyTheMethodsThatNeedExplainingCarryALine(t *testing.T) {
	body := mustGet(t, newServer(t), "/checkout")

	// The card is the one row whose name does not say who processes it.
	mustContain(t, body, "Processed via Stripe.", "the card keeps the line the client dictated")
	mustNotContain(t, body, "Visa, Mastercard or American Express",
		"the longer card sentence was replaced, not added to")

	// The two everyone knows carry nothing at all.
	mustNotContain(t, body, "You'll be redirected to PayPal", "PayPal explains itself")
	mustNotContain(t, body, "Pay on Stripe's own checkout", "so does Stripe")

	// The two nobody can be expected to recognise keep theirs. The client named
	// PayPal and Stripe, and only those two.
	mustContain(t, body, "Pay directly with many bank links around Europe.",
		"Paysera still says what it is")
	mustContain(t, body, "Bitcoin, Ethereum, stablecoins and many others through NOWPayments.",
		"and so does Crypto")

	// An empty line is not the same as no line: the span it used to leave
	// behind still took the gap under the name.
	checkout := asset(t, "../../web/templates/pages/account/checkout.html")
	mustContain(t, checkout, `{{ if .Hint }}<span class="text-sm muted">{{ .Hint }}</span>{{ end }}`,
		"a method with no line must render no element for it")
}

// ---------------------------------------------------------------------------
// 77. "Here in the add listing page where can enter your location the green
//     bow, make the borderline there equal, so on the left remove this bolder
//     part."
// ---------------------------------------------------------------------------

func TestTheGreenLocationBoxHasOneBorderWeight(t *testing.T) {
	pages := cssCode(t, "pages.css")
	rule := section(t, pages, ".field--editable-location {", "}", "the editable location box")

	mustContain(t, rule, "border: 1px solid var(--success-border);",
		"one weight, all the way round")
	mustNotContain(t, rule, "border-left", "the 3px accent down the left-hand side is gone")

	// The box is still green, and still the exception among the read-only
	// fields under it — the accent was never what said so.
	mustContain(t, rule, "background: var(--success-bg);", "the green ground stays")
	mustContain(t, pages, ".field--editable-location .field__label { color: var(--success); }",
		"and so does the green label")

	mustContain(t, mustGet(t, newServer(t), "/add-listing"), `class="field field--editable-location"`,
		"the box is on the page the client pointed at")
}

// ---------------------------------------------------------------------------
// 78. "Under rooms and dimensions, additionally to living area, make there
//     'Total area (m2)' but these fields are not obligatory, so remove the red
//     error and restriction from there."
// ---------------------------------------------------------------------------

func TestTheAreaFieldsAreOptional(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")

	// The new field, beside the one it was asked to sit beside.
	mustContain(t, body, `<label class="field__label" for="w-total-area">Total area (m²)</label>`,
		"Total area must be offered")
	mustContain(t, body, `id="w-total-area"`, "…as a real field, not a label alone")

	// Neither area is demanded any more: no asterisk, no message, and nothing
	// for the section's required-field count to find.
	mustContain(t, body, `<label class="field__label" for="w-area">Living area (m²)</label>`,
		"the living area label must carry no asterisk")
	mustNotContain(t, body, "Living area is required before you can publish",
		"the red message must be gone")
	area := section(t, body, `id="w-area"`, ">", "the living area input")
	mustNotContain(t, area, "data-required", "…and the field must not be counted as required")
	total := section(t, body, `id="w-total-area"`, ">", "the total area input")
	mustNotContain(t, total, "data-required", "the new field is optional too")

	// The waypoint stops accusing the section before anything has been typed.
	mustContain(t, body, "stateOf('rooms', 'todo')",
		"the rooms waypoint must not be seeded red")

	// Total rooms was not named in the note and keeps its asterisk — which is
	// also what keeps the error state real: clear it and the waypoint turns.
	mustContain(t, body, `<label class="field__label" for="w-rooms">Total rooms <span class="req">*</span></label>`,
		"the one field the results cannot do without stays required")
	rooms := section(t, body, `id="w-rooms"`, ">", "the total rooms input")
	mustContain(t, rooms, "data-required", "…and is still what the live check counts")
}

// ---------------------------------------------------------------------------
// 79. "To features and amenities add: Free Parking, Paid Parking, Personal
//     Parking Place, Garage, WIFI, TV, Kitchen, Washing Maschine, Dryer, Bed
//     Linens, Room-darkening shades, Baby bath, Regulated Heating, Smoke Alarm,
//     Refrigerator, Cooking Basics, Freezer, Dishwasher, Heater, Microwave,
//     Coffee Maker, Carbon Monoxide Alarm"
// ---------------------------------------------------------------------------

// newAmenities is the client's list, in the client's order, in the spelling the
// page uses: sentence case like the twelve that were already there, and the two
// typed spellings corrected.
var newAmenities = []string{
	"Free parking", "Paid parking", "Personal parking place", "Garage",
	"Wi-Fi", "TV", "Kitchen", "Washing machine", "Dryer", "Bed linen",
	"Room-darkening shades", "Baby bath", "Regulated heating", "Smoke alarm",
	"Refrigerator", "Cooking basics", "Freezer", "Dishwasher", "Heater",
	"Microwave", "Coffee maker", "Carbon monoxide alarm",
}

// oldAmenities is what the section already offered. None of it was to be
// replaced — the note says "add".
var oldAmenities = []string{
	"Parking", "Balcony", "Terrace", "Garden", "Elevator", "Sauna",
	"Sea or water view", "Furnished", "Storage room", "Air conditioning",
	"Fireplace", "Alarm system",
}

func TestEveryAmenityTheClientNamedIsOffered(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")
	features := section(t, body, `id="ls-features"`, "</section>", "the amenities section")

	for _, name := range newAmenities {
		mustContain(t, features, `<span class="check__text">`+name+`</span>`,
			"the client asked for this amenity: "+name)
	}
	for _, name := range oldAmenities {
		mustContain(t, features, `<span class="check__text">`+name+`</span>`,
			"and this one was not to be removed: "+name)
	}

	// Every one of them is a real tick, not a line of text.
	//
	// Counted by the opening tag rather than by the whole element: each tick
	// carries the query parameter and value the search filters on since the
	// catalogue moved to models.AmenityGroups on 19 August, so they are no
	// longer bare `<input type="checkbox">`.
	if n := strings.Count(features, `<input type="checkbox" name=`); n != len(newAmenities)+len(oldAmenities) {
		t.Errorf("the section offers %d checkboxes, want %d", n, len(newAmenities)+len(oldAmenities))
	}

	// "Parking" survives alongside the three kinds of parking under it because
	// it is one of the five ticks the search filter actually runs on.
	mustContain(t, mustGet(t, newServer(t), "/search"), `name="parking"`,
		"the generic parking filter must not have been dropped by the new list")
}

func TestTheAmenitiesAreGroupedRatherThanRunTogether(t *testing.T) {
	body := mustGet(t, newServer(t), "/add-listing")
	features := section(t, body, `id="ls-features"`, "</section>", "the amenities section")

	// Thirty-four ticks in one undivided pair of columns is a wall. A fieldset
	// and a legend, because that is what a named group of checkboxes is — the
	// group is announced to a screen reader as well as drawn for a reader.
	for _, group := range []string{
		"The property", "Parking and access", "Living and comfort", "Kitchen", "Safety",
	} {
		mustContain(t, features, `<legend class="amenity-group__title">`+group+`</legend>`,
			"the amenities are grouped: "+group)
	}
	if n := strings.Count(features, `<fieldset class="amenity-group">`); n != 5 {
		t.Errorf("%d amenity groups, want 5", n)
	}

	// Each group keeps the two-column grid the twelve already used.
	if n := strings.Count(features, `class="grid grid--2"`); n != 5 {
		t.Errorf("%d two-column grids inside the section, want one per group", n)
	}

	pages := cssCode(t, "pages.css")
	mustContain(t, pages, ".amenity-group { border: 0; padding: 0; margin: 0; min-width: 0; }",
		"a fieldset draws a box by default, and this one must not")
}

// ---------------------------------------------------------------------------
// 80. "In the main search menu this text 'Show brokers advertising on the map
//     alongside the listings.' move more up closer to it's button and reduce
//     the line spacing of it."
// ---------------------------------------------------------------------------

func TestTheBrokersHintSitsCloseUnderItsButton(t *testing.T) {
	components := cssCode(t, "components.css")
	rule := section(t, components, ".segmented--brokers + .field__hint {", "}", "the brokers hint")

	// Both halves of the note. The fieldset is a .field, which spaces its
	// children with an 8px gap; half of that is given back, so the line sits
	// 4px under the button instead of 8.
	mustContain(t, rule, "margin-top: calc(var(--sp-1) - var(--sp-2));",
		"the line must move up towards its button")
	mustContain(t, rule, "line-height: var(--lh-snug);",
		"…and set its two lines tighter than body text")

	// Only this hint. Every other field on the site keeps the gap it had.
	mustContain(t, components, ".field { display: flex; flex-direction: column; gap: var(--sp-2); min-width: 0; }",
		"the field's own spacing must not have been changed for everyone")

	body := mustGet(t, newServer(t), "/search")
	hint := section(t, body, `class="segmented segmented--full segmented--brokers"`, "</fieldset>",
		"the brokers control")
	mustContain(t, hint, "Show brokers advertising on the map alongside the listings.",
		"the line is still there — it was moved, not removed")
	mustContain(t, hint, `<p class="field__hint">`,
		"…and is still the element the rule names, immediately after the control")
}
