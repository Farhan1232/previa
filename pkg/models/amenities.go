package models

import "strings"

// ---------------------------------------------------------------------------
// Features and amenities
// ---------------------------------------------------------------------------

// AmenityGroup is one titled block of the "Features and amenities" list.
//
// The list lived inside web/templates/pages/add-listing.html as a template
// literal until the client asked for it twice over: once when publishing a
// listing, and again when searching for one — "the 'Features and amenities' we
// just added into 'add listing' page add to the search menu as well, with the
// same subtitles."
//
// "The same subtitles" is the reason this is one catalogue rather than two
// lists that happen to agree today. A seller ticking Coffee maker and a buyer
// filtering on Coffee maker have to be talking about the same thing, and the
// only way to guarantee that is for both screens to render the same slice.
type AmenityGroup struct {
	Name  string
	Items []Amenity
}

// Amenity is one tick.
//
// Key is what travels in a URL and what a listing is matched on. Filter is the
// query parameter for the handful that the search already ran on before this
// list existed — furnished, parking, balcony, garden, elevator — so their
// bookmarks, saved searches and active-filter chips keep working unchanged.
// Everything else is submitted as amenity=<key>.
type Amenity struct {
	Key    string
	Label  string
	Filter string
}

// Param is the query parameter this tick is submitted under.
func (a Amenity) Param() string {
	if a.Filter != "" {
		return a.Filter
	}
	return "amenity"
}

// Value is what that parameter carries. The five legacy filters are booleans in
// the URL, which is what they have always been.
func (a Amenity) Value() string {
	if a.Filter != "" {
		return "1"
	}
	return a.Key
}

// AmenityGroups is the catalogue, in the client's own order and grouping.
//
// The groups are the client's list of 19 August read as what it is — parking,
// then the connected things, then the kitchen, then the safety alarms — and the
// twelve that predate it keep the first group, so nothing anybody has already
// read has moved.
var AmenityGroups = []AmenityGroup{
	{Name: "The property", Items: []Amenity{
		{Key: "parking", Label: "Parking", Filter: "parking"},
		{Key: "balcony", Label: "Balcony", Filter: "balcony"},
		{Key: "terrace", Label: "Terrace"},
		{Key: "garden", Label: "Garden", Filter: "garden"},
		{Key: "elevator", Label: "Elevator", Filter: "elevator"},
		{Key: "sauna", Label: "Sauna"},
		{Key: "seaview", Label: "Sea or water view"},
		{Key: "furnished", Label: "Furnished", Filter: "furnished"},
		{Key: "storage", Label: "Storage room"},
		{Key: "air-conditioning", Label: "Air conditioning"},
		{Key: "fireplace", Label: "Fireplace"},
		{Key: "alarm-system", Label: "Alarm system"},
	}},
	{Name: "Parking and access", Items: []Amenity{
		{Key: "free-parking", Label: "Free parking"},
		{Key: "paid-parking", Label: "Paid parking"},
		{Key: "personal-parking", Label: "Personal parking place"},
		{Key: "garage", Label: "Garage"},
	}},
	{Name: "Living and comfort", Items: []Amenity{
		{Key: "wifi", Label: "Wi-Fi"},
		{Key: "tv", Label: "TV"},
		{Key: "bed-linen", Label: "Bed linen"},
		{Key: "blackout-shades", Label: "Room-darkening shades"},
		{Key: "baby-bath", Label: "Baby bath"},
		{Key: "regulated-heating", Label: "Regulated heating"},
		{Key: "heater", Label: "Heater"},
		{Key: "washing-machine", Label: "Washing machine"},
		{Key: "dryer", Label: "Dryer"},
	}},
	{Name: "Kitchen", Items: []Amenity{
		{Key: "kitchen", Label: "Kitchen"},
		{Key: "cooking-basics", Label: "Cooking basics"},
		{Key: "refrigerator", Label: "Refrigerator"},
		{Key: "freezer", Label: "Freezer"},
		{Key: "dishwasher", Label: "Dishwasher"},
		{Key: "microwave", Label: "Microwave"},
		{Key: "coffee-maker", Label: "Coffee maker"},
	}},
	{Name: "Safety", Items: []Amenity{
		{Key: "smoke-alarm", Label: "Smoke alarm"},
		{Key: "co-alarm", Label: "Carbon monoxide alarm"},
	}},
}

// IsAmenity reports whether a raw value names a tick in the catalogue. Used
// when parsing a URL, so a hand-edited one filters on nothing that exists
// rather than returning an unexplained empty page.
func IsAmenity(key string) bool {
	for _, g := range AmenityGroups {
		for _, a := range g.Items {
			if a.Key == key {
				return true
			}
		}
	}
	return false
}

// AmenityLabel is the display name for a key, used by the active-filter chips.
// An unknown key returns itself rather than an empty string, so a chip is never
// blank even if the catalogue and a stored search have drifted apart.
func AmenityLabel(key string) string {
	for _, g := range AmenityGroups {
		for _, a := range g.Items {
			if a.Key == key {
				return a.Label
			}
		}
	}
	return key
}

// HasAmenity reports whether a listing carries a tick.
func HasAmenity(list []string, key string) bool {
	for _, k := range list {
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}
