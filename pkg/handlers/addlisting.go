package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"previa/pkg/models"
)

// WizardStep is one waypoint in the vertical publishing navigation.
//
// The form is a single continuous page, so a waypoint is an anchor rather than
// a separate screen: Number and Label name it, Key is its section id.
type WizardStep struct {
	Number int
	Key    string
	Label  string
	State  string // "todo" | "error" — the active one is tracked in the browser
}

// AddListingData drives the add-listing form.
type AddListingData struct {
	Steps      []WizardStep
	TotalSteps int
	Packages   []models.Package
	Promotions []models.Promotion
	Countries  []models.Country
	SampleImgs []string
	// Anchor is the section a "?step=" or "#" deep link should open on. Empty
	// when the visitor simply opened /add-listing.
	Anchor              string
	LocationSuggestions []models.LocationSuggestion
	WizardMap           template.JS
}

// wizardSteps is the waypoint list, in page order.
//
// Two of the client's changes are baked in here:
//   - waypoint 1 is "Deal type", matching the search filters and the homepage
//   - the old "Public location display" waypoint is gone, folded into Location,
//     and "Package and promotion" is gone, folded into Publish
//
// Numbering is derived from position, so removing a waypoint renumbers the rest
// on its own and no anchor has to be renumbered by hand.
var wizardSteps = []struct{ key, label string }{
	{"deal", "Deal type"},
	{"category", "Property category"},
	{"location", "Location"},
	{"details", "Property information"},
	{"rooms", "Rooms and dimensions"},
	{"features", "Features and amenities"},
	{"description", "Description"},
	{"media", "Photos & videos"},
	{"price", "Price"},
	{"contact", "Contact details"},
	{"preview", "Preview"},
	{"publish", "Publish"},
}

// AddListing renders the wizard at the requested step.
//
// A logged-out visitor is sent to the login page first, as the client
// specified.
func (h *Handler) AddListing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.Store.Account.CurrentUser(ctx).ID == "" {
		http.Redirect(w, r, "/login?return=/add-listing", http.StatusSeeOther)
		return
	}

	total := len(wizardSteps)

	// Every waypoint starts neutral.
	//
	// The rooms section used to be seeded as an error so the design covered
	// that state on first paint. It was seeded because Living area was required
	// and empty; the client made both areas optional on 19 August — "these
	// fields are not obligatory, so remove the red error and restriction from
	// there" — and a red marker on a section with nothing missing is exactly
	// the restriction they asked to be rid of.
	//
	// The state itself is not gone: recheck() in previa.js counts the empty
	// [data-required] fields in each section on load and on every keystroke, so
	// clearing Total rooms still turns this waypoint red. It is earned now
	// rather than staged.
	steps := make([]WizardStep, 0, total)
	for i, s := range wizardSteps {
		steps = append(steps, WizardStep{Number: i + 1, Key: s.key, Label: s.label, State: "todo"})
	}

	// Deep links. "?step=6" used to load a separate screen; now it names the
	// section to open on, so older links keep working.
	anchor := ""
	if q := r.URL.Query().Get("step"); q != "" {
		if n := atoi(q, 0); n >= 1 && n <= total {
			anchor = wizardSteps[n-1].key
		}
	}
	if q := r.URL.Query().Get("section"); q != "" {
		for _, s := range wizardSteps {
			if s.key == q {
				anchor = q
			}
		}
	}

	var imgs []string
	for i := 1; i <= 6; i++ {
		imgs = append(imgs, sampleImage(i))
	}

	pd := h.base(r, "", "Add a listing — Previa",
		"Publish your property on Previa in a few guided steps.")
	pd.Meta.NoIndex = true
	// Every section is on the page at once now, so the Location map is always
	// present and Leaflet is always needed.
	pd.NeedsMap = true
	pd.Data = AddListingData{
		Steps:               steps,
		TotalSteps:          total,
		Packages:            h.Store.Catalog.Packages(ctx),
		Promotions:          h.Store.Catalog.Promotions(ctx),
		Countries:           h.Store.Catalog.Countries(ctx),
		SampleImgs:          imgs,
		Anchor:              anchor,
		LocationSuggestions: h.Store.Catalog.LocationSuggestions(ctx),
		// A single draggable pin centred on the market being listed in.
		WizardMap: buildMapConfig([]models.Property{{
			ID: "draft", Title: "Your property", City: pd.Country.Name,
			Coords: models.Coordinates{Lat: pd.Country.Lat, Lng: pd.Country.Lng},
			Price:  models.Money{Amount: 429000, Currency: pd.Country.Currency},
		}}, nil, pd.Country, h.Cfg.MapsKey),
	}
	h.View.Render(w, http.StatusOK, "add-listing", pd)
}

// AddListingSave acknowledges a wizard autosave. Drafts are held client-side in
// this milestone; the future backend persists them per user.
func (h *Handler) AddListingSave(w http.ResponseWriter, r *http.Request) {
	h.View.RenderPartial(w, http.StatusOK, "add-listing", "autosave-indicator",
		map[string]any{"State": "saved"})
}

func sampleImage(n int) string {
	return "/static/img/properties/p" + pad3(n*7) + ".jpg"
}

func pad3(n int) string {
	switch {
	case n < 10:
		return "00" + itoa(n)
	case n < 100:
		return "0" + itoa(n)
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ReverseGeocode answers a map click with a structured address.
//
// A frontend mock, in the same family as /favourite and /set-country: it reads
// from the seeded catalogue and writes nothing. The second milestone swaps the
// body for a Google Geocoding call; the response shape and this URL stay put,
// so the add-listing form does not change with it.
func (h *Handler) ReverseGeocode(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lat := parseFloatOr(q.Get("lat"), 0)
	lng := parseFloatOr(q.Get("lng"), 0)

	place := h.Store.Catalog.ReverseGeocode(lat, lng)

	// Country name and ISO code come from the catalogue so the read-only
	// fields agree with the market selector.
	if c, ok := h.Store.Catalog.Country(r.Context(), place.CountryCode); ok {
		place.Country = c.Name
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(place); err != nil {
		http.Error(w, "encode", http.StatusInternalServerError)
	}
}
