package handlers

import (
	"html/template"
	"net/http"

	"previa/internal/models"
)

// WizardStep is one entry in the vertical publishing progression.
type WizardStep struct {
	Number int
	Key    string
	Label  string
	State  string // "done" | "current" | "todo" | "error"
}

// AddListingData drives the add-listing wizard.
type AddListingData struct {
	Steps      []WizardStep
	Current    WizardStep
	CurrentNum int
	TotalSteps int
	Packages   []models.Package
	Countries  []models.Country
	SampleImgs []string
	PrevStep   int
	NextStep   int
	Progress   int
	WizardMap  template.JS
}

var wizardSteps = []struct{ key, label string }{
	{"deal", "Sale or rent"},
	{"category", "Property category"},
	{"location", "Address and map pin"},
	{"privacy", "Public location display"},
	{"details", "Property information"},
	{"rooms", "Rooms and dimensions"},
	{"features", "Features and amenities"},
	{"description", "Description"},
	{"media", "Photos and media"},
	{"price", "Price"},
	{"contact", "Contact details"},
	{"package", "Package and promotion"},
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
	cur := clampInt(atoi(r.URL.Query().Get("step"), 1), 1, total)

	steps := make([]WizardStep, 0, total)
	for i, s := range wizardSteps {
		state := "todo"
		switch {
		case i+1 < cur:
			state = "done"
		case i+1 == cur:
			state = "current"
		}
		// One step is shown in the error state so the design covers it.
		if i+1 == 6 && cur > 6 {
			state = "error"
		}
		steps = append(steps, WizardStep{Number: i + 1, Key: s.key, Label: s.label, State: state})
	}

	var imgs []string
	for i := 1; i <= 6; i++ {
		imgs = append(imgs, sampleImage(i))
	}

	pd := h.base(r, "", "Add a listing — Previa",
		"Publish your property on Previa in a few guided steps.")
	pd.Meta.NoIndex = true
	pd.NeedsMap = cur == 3 // only the address step shows a map
	pd.Data = AddListingData{
		Steps:      steps,
		Current:    steps[cur-1],
		CurrentNum: cur,
		TotalSteps: total,
		Packages:   h.Store.Catalog.Packages(ctx),
		Countries:  h.Store.Catalog.Countries(ctx),
		SampleImgs: imgs,
		PrevStep:   clampInt(cur-1, 1, total),
		NextStep:   clampInt(cur+1, 1, total),
		Progress:   cur * 100 / total,
		// A single draggable pin centred on the market being listed in.
		WizardMap: buildMapConfig([]models.Property{{
			ID: "draft", Title: "Your property", City: pd.Country.Name,
			Coords: models.Coordinates{Lat: pd.Country.Lat, Lng: pd.Country.Lng},
			Price:  models.Money{Amount: 429000, Currency: pd.Country.Currency},
		}}, pd.Country, h.Cfg.MapsKey),
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
