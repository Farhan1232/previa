package data

import (
	"sort"
	"strings"

	"previa/pkg/models"
)

// ---------------------------------------------------------------------------
// Location suggestions and mock reverse geocoding
// ---------------------------------------------------------------------------
//
// The client asked for one Google-Maps-style Location box wherever a place is
// entered, resolving countries, cities, districts, streets and exact addresses.
// No Places key is configured in this milestone, so this file stands in for it.
//
// Everything here is derived from the seeded listings, which means a suggestion
// always corresponds to stock that actually exists. Swapping in Places is
// replacing LocationSuggestions and ReverseGeocode; PropertyFilter, the
// templates and the query-parameter contract stay as they are. See
// docs/backend-integration-points.md.

// buildLocationSuggestions derives the autocomplete list from seeded stock.
//
// Ordered broad to narrow — country, city, district, street — so the list reads
// the way a person narrows a search, and within each kind alphabetically so it
// is stable between runs.
func buildLocationSuggestions(properties []models.Property) []models.LocationSuggestion {
	seen := map[string]bool{}
	var out []models.LocationSuggestion

	add := func(s models.LocationSuggestion) {
		key := s.Kind + "|" + strings.ToLower(s.Label)
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, s)
	}

	for _, c := range countries {
		add(models.LocationSuggestion{
			Kind: models.PlaceCountry, Label: c.Name,
			CountryCode: c.Code, Lat: c.Lat, Lng: c.Lng,
		})
	}

	for _, p := range properties {
		if p.City != "" {
			add(models.LocationSuggestion{
				Kind: models.PlaceCity, Label: p.City + ", " + p.Country,
				CountryCode: p.CountryCode, City: p.City,
				Lat: p.Coords.Lat, Lng: p.Coords.Lng,
			})
		}
		if p.District != "" {
			add(models.LocationSuggestion{
				Kind: models.PlaceDistrict, Label: p.District + ", " + p.City + ", " + p.Country,
				CountryCode: p.CountryCode, City: p.City, District: p.District,
				Lat: p.Coords.Lat, Lng: p.Coords.Lng,
			})
		}
		if p.Address != "" {
			// The street on its own, and then the full address. A visitor
			// looking for "Vana-Kalamaja" should not have to know the number.
			if street := streetOf(p.Address); street != "" {
				add(models.LocationSuggestion{
					Kind: models.PlaceStreet, Label: street + ", " + p.City + ", " + p.Country,
					CountryCode: p.CountryCode, City: p.City, District: p.District,
					Address: street, Lat: p.Coords.Lat, Lng: p.Coords.Lng,
				})
			}
			add(models.LocationSuggestion{
				Kind: models.PlaceAddress, Label: p.Address + ", " + p.City + ", " + p.Country,
				CountryCode: p.CountryCode, City: p.City, District: p.District,
				Address: p.Address, Lat: p.Coords.Lat, Lng: p.Coords.Lng,
			})
		}
	}

	rank := map[string]int{
		models.PlaceCountry: 0, models.PlaceCity: 1, models.PlaceDistrict: 2,
		models.PlaceStreet: 3, models.PlaceAddress: 4,
	}
	sort.SliceStable(out, func(i, j int) bool {
		if rank[out[i].Kind] != rank[out[j].Kind] {
			return rank[out[i].Kind] < rank[out[j].Kind]
		}
		return out[i].Label < out[j].Label
	})
	return out
}

// streetOf strips a house number off an address, so "Vana-Kalamaja 21" yields
// "Vana-Kalamaja". Returns "" when the whole address is the street already, so
// the caller does not add a duplicate entry.
func streetOf(address string) string {
	fields := strings.Fields(address)
	if len(fields) < 2 {
		return ""
	}
	last := fields[len(fields)-1]
	if !strings.ContainsAny(last, "0123456789") {
		return ""
	}
	return strings.Join(fields[:len(fields)-1], " ")
}

// ResolveLocation matches a typed or selected label against the suggestion
// list and returns the structured place behind it.
//
// An exact label match wins. Otherwise the first suggestion whose label starts
// with the query is used, which is what makes a half-typed "Tallinn" still
// work when the form is submitted before a suggestion is picked. A query that
// matches nothing returns ok=false and the caller keeps it as free text.
func (m *Mock) ResolveLocation(label string) (models.LocationSuggestion, bool) {
	q := strings.TrimSpace(strings.ToLower(label))
	if q == "" {
		return models.LocationSuggestion{}, false
	}
	for _, s := range m.locations {
		if strings.ToLower(s.Label) == q {
			return s, true
		}
	}
	for _, s := range m.locations {
		if strings.HasPrefix(strings.ToLower(s.Label), q) {
			return s, true
		}
	}
	return models.LocationSuggestion{}, false
}

// ReverseGeocode turns a map click into a structured address.
//
// It answers with the nearest seeded listing's location, which is the closest
// this milestone can get to a real reverse-geocode and is enough to show the
// add-listing form filling its read-only fields from a click on the map. A real
// Geocoding API call replaces the body and nothing else.
func (m *Mock) ReverseGeocode(lat, lng float64) models.LocationSuggestion {
	best := models.LocationSuggestion{}
	bestDist := -1.0
	for _, p := range m.properties {
		// Squared planar distance: only the ordering matters here, so there is
		// no reason to pay for a great-circle calculation.
		dLat := p.Coords.Lat - lat
		dLng := p.Coords.Lng - lng
		d := dLat*dLat + dLng*dLng
		if bestDist < 0 || d < bestDist {
			bestDist = d
			best = models.LocationSuggestion{
				Kind: models.PlaceAddress, Label: p.Address + ", " + p.City + ", " + p.Country,
				CountryCode: p.CountryCode, Country: p.Country, City: p.City,
				District: p.District, Address: p.Address, PostalCode: p.PostalCode,
				Lat: lat, Lng: lng,
			}
		}
	}
	if best.CountryCode == "" {
		best = models.LocationSuggestion{Kind: models.PlaceAddress, Lat: lat, Lng: lng}
	}
	// The pin is where the user put it, not where the nearest listing is.
	best.Lat, best.Lng = lat, lng
	return best
}

// ApplyLocation resolves a filter's Location label into the structured fields
// the matcher uses.
//
// Called by the handler right after ParseFilter, because resolving needs the
// seeded suggestion list and ParseFilter deliberately has no store access.
//
// A label that resolves narrows to exactly what was chosen: a country
// suggestion sets only the country, a city suggestion sets country and city,
// an address sets all three. A label that resolves to nothing is kept as a
// keyword instead of silently returning everything, so a typo reads as "no
// matches" rather than as "no filter".
func ApplyLocation(f *PropertyFilter, resolve func(string) (models.LocationSuggestion, bool)) {
	if f.LocationLabel == "" {
		return
	}
	s, ok := resolve(f.LocationLabel)
	if !ok {
		f.Keyword = strings.TrimSpace(f.Keyword + " " + f.LocationLabel)
		return
	}
	f.CountryCode = s.CountryCode
	f.City = s.City
	f.District = s.District
	f.Address = s.Address
}
