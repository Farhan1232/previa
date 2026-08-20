package data

import (
	"strings"

	"previa/pkg/models"
)

// ---------------------------------------------------------------------------
// The world list
// ---------------------------------------------------------------------------

// worldCountries is the full ISO 3166-1 list the market selector offers.
//
// The eight entries in `countries` (seed_content.go) are the markets that carry
// seeded listings, brokers and banners. This table is the rest of the world: a
// visitor can switch the market to any of them, and the pages then render their
// ordinary empty state rather than pretending to have stock.
//
// Coordinates are the capital, used only to centre the map when a market with
// no seeded property is selected. Zoom is deliberately wide for the same
// reason. The second milestone replaces this table with the countries row in
// MySQL; nothing outside data.Mock reads it directly.
//
// Codes are ISO 3166-1 alpha-2 so the search-by-code requirement ("DE" finds
// Germany) works against the same value the backend will store.
var worldCountries = []models.Country{
	{Code: "AF", Name: "Afghanistan", Lat: 34.5553, Lng: 69.2075},
	{Code: "AL", Name: "Albania", Lat: 41.3275, Lng: 19.8187},
	{Code: "DZ", Name: "Algeria", Lat: 36.7538, Lng: 3.0588},
	{Code: "AD", Name: "Andorra", Lat: 42.5063, Lng: 1.5218},
	{Code: "AO", Name: "Angola", Lat: -8.8390, Lng: 13.2894},
	{Code: "AG", Name: "Antigua and Barbuda", Lat: 17.1274, Lng: -61.8468},
	{Code: "AR", Name: "Argentina", Lat: -34.6037, Lng: -58.3816},
	{Code: "AM", Name: "Armenia", Lat: 40.1792, Lng: 44.4991},
	{Code: "AU", Name: "Australia", Lat: -35.2809, Lng: 149.1300},
	{Code: "AZ", Name: "Azerbaijan", Lat: 40.4093, Lng: 49.8671},
	{Code: "BS", Name: "Bahamas", Lat: 25.0343, Lng: -77.3963},
	{Code: "BH", Name: "Bahrain", Lat: 26.2285, Lng: 50.5860},
	{Code: "BD", Name: "Bangladesh", Lat: 23.8103, Lng: 90.4125},
	{Code: "BB", Name: "Barbados", Lat: 13.1939, Lng: -59.5432},
	{Code: "BY", Name: "Belarus", Lat: 53.9006, Lng: 27.5590},
	{Code: "BE", Name: "Belgium", Lat: 50.8503, Lng: 4.3517},
	{Code: "BZ", Name: "Belize", Lat: 17.2510, Lng: -88.7590},
	{Code: "BJ", Name: "Benin", Lat: 6.4969, Lng: 2.6283},
	{Code: "BT", Name: "Bhutan", Lat: 27.4728, Lng: 89.6390},
	{Code: "BO", Name: "Bolivia", Lat: -16.4897, Lng: -68.1193},
	{Code: "BA", Name: "Bosnia and Herzegovina", Lat: 43.8563, Lng: 18.4131},
	{Code: "BW", Name: "Botswana", Lat: -24.6282, Lng: 25.9231},
	{Code: "BR", Name: "Brazil", Lat: -15.7942, Lng: -47.8822},
	{Code: "BN", Name: "Brunei", Lat: 4.9031, Lng: 114.9398},
	{Code: "BG", Name: "Bulgaria", Lat: 42.6977, Lng: 23.3219},
	{Code: "BF", Name: "Burkina Faso", Lat: 12.3714, Lng: -1.5197},
	{Code: "BI", Name: "Burundi", Lat: -3.3614, Lng: 29.3599},
	{Code: "CV", Name: "Cabo Verde", Lat: 14.9330, Lng: -23.5133},
	{Code: "KH", Name: "Cambodia", Lat: 11.5564, Lng: 104.9282},
	{Code: "CM", Name: "Cameroon", Lat: 3.8480, Lng: 11.5021},
	{Code: "CA", Name: "Canada", Lat: 45.4215, Lng: -75.6972},
	{Code: "CF", Name: "Central African Republic", Lat: 4.3947, Lng: 18.5582},
	{Code: "TD", Name: "Chad", Lat: 12.1348, Lng: 15.0557},
	{Code: "CL", Name: "Chile", Lat: -33.4489, Lng: -70.6693},
	{Code: "CN", Name: "China", Lat: 39.9042, Lng: 116.4074},
	{Code: "CO", Name: "Colombia", Lat: 4.7110, Lng: -74.0721},
	{Code: "KM", Name: "Comoros", Lat: -11.7172, Lng: 43.2473},
	{Code: "CG", Name: "Congo", Lat: -4.2634, Lng: 15.2429},
	{Code: "CD", Name: "Congo (DRC)", Lat: -4.4419, Lng: 15.2663},
	{Code: "CR", Name: "Costa Rica", Lat: 9.9281, Lng: -84.0907},
	{Code: "CI", Name: "Côte d'Ivoire", Lat: 6.8276, Lng: -5.2893},
	{Code: "HR", Name: "Croatia", Lat: 45.8150, Lng: 15.9819},
	{Code: "CU", Name: "Cuba", Lat: 23.1136, Lng: -82.3666},
	{Code: "CY", Name: "Cyprus", Lat: 35.1856, Lng: 33.3823},
	{Code: "DK", Name: "Denmark", Lat: 55.6761, Lng: 12.5683},
	{Code: "DJ", Name: "Djibouti", Lat: 11.5721, Lng: 43.1456},
	{Code: "DM", Name: "Dominica", Lat: 15.3092, Lng: -61.3794},
	{Code: "DO", Name: "Dominican Republic", Lat: 18.4861, Lng: -69.9312},
	{Code: "EC", Name: "Ecuador", Lat: -0.1807, Lng: -78.4678},
	{Code: "EG", Name: "Egypt", Lat: 30.0444, Lng: 31.2357},
	{Code: "SV", Name: "El Salvador", Lat: 13.6929, Lng: -89.2182},
	{Code: "GQ", Name: "Equatorial Guinea", Lat: 3.7523, Lng: 8.7742},
	{Code: "ER", Name: "Eritrea", Lat: 15.3229, Lng: 38.9251},
	{Code: "SZ", Name: "Eswatini", Lat: -26.3054, Lng: 31.1367},
	{Code: "ET", Name: "Ethiopia", Lat: 9.0300, Lng: 38.7400},
	{Code: "FJ", Name: "Fiji", Lat: -18.1416, Lng: 178.4419},
	{Code: "FR", Name: "France", Lat: 48.8566, Lng: 2.3522},
	{Code: "GA", Name: "Gabon", Lat: 0.4162, Lng: 9.4673},
	{Code: "GM", Name: "Gambia", Lat: 13.4549, Lng: -16.5790},
	{Code: "GE", Name: "Georgia", Lat: 41.7151, Lng: 44.8271},
	{Code: "GH", Name: "Ghana", Lat: 5.6037, Lng: -0.1870},
	{Code: "GR", Name: "Greece", Lat: 37.9838, Lng: 23.7275},
	{Code: "GD", Name: "Grenada", Lat: 12.0561, Lng: -61.7488},
	{Code: "GT", Name: "Guatemala", Lat: 14.6349, Lng: -90.5069},
	{Code: "GN", Name: "Guinea", Lat: 9.6412, Lng: -13.5784},
	{Code: "GW", Name: "Guinea-Bissau", Lat: 11.8817, Lng: -15.6178},
	{Code: "GY", Name: "Guyana", Lat: 6.8013, Lng: -58.1551},
	{Code: "HT", Name: "Haiti", Lat: 18.5944, Lng: -72.3074},
	{Code: "HN", Name: "Honduras", Lat: 14.0723, Lng: -87.1921},
	{Code: "HU", Name: "Hungary", Lat: 47.4979, Lng: 19.0402},
	{Code: "IS", Name: "Iceland", Lat: 64.1466, Lng: -21.9426},
	{Code: "IN", Name: "India", Lat: 28.6139, Lng: 77.2090},
	{Code: "ID", Name: "Indonesia", Lat: -6.2088, Lng: 106.8456},
	{Code: "IR", Name: "Iran", Lat: 35.6892, Lng: 51.3890},
	{Code: "IQ", Name: "Iraq", Lat: 33.3152, Lng: 44.3661},
	{Code: "IE", Name: "Ireland", Lat: 53.3498, Lng: -6.2603},
	{Code: "IL", Name: "Israel", Lat: 31.7683, Lng: 35.2137},
	{Code: "IT", Name: "Italy", Lat: 41.9028, Lng: 12.4964},
	{Code: "JM", Name: "Jamaica", Lat: 17.9714, Lng: -76.7931},
	{Code: "JP", Name: "Japan", Lat: 35.6762, Lng: 139.6503},
	{Code: "JO", Name: "Jordan", Lat: 31.9454, Lng: 35.9284},
	{Code: "KZ", Name: "Kazakhstan", Lat: 51.1694, Lng: 71.4491},
	{Code: "KE", Name: "Kenya", Lat: -1.2921, Lng: 36.8219},
	{Code: "KI", Name: "Kiribati", Lat: 1.3278, Lng: 172.9797},
	{Code: "KW", Name: "Kuwait", Lat: 29.3759, Lng: 47.9774},
	{Code: "KG", Name: "Kyrgyzstan", Lat: 42.8746, Lng: 74.5698},
	{Code: "LA", Name: "Laos", Lat: 17.9757, Lng: 102.6331},
	{Code: "LV", Name: "Latvia", Lat: 56.9496, Lng: 24.1052},
	{Code: "LB", Name: "Lebanon", Lat: 33.8938, Lng: 35.5018},
	{Code: "LS", Name: "Lesotho", Lat: -29.3151, Lng: 27.4869},
	{Code: "LR", Name: "Liberia", Lat: 6.2907, Lng: -10.7605},
	{Code: "LY", Name: "Libya", Lat: 32.8872, Lng: 13.1913},
	{Code: "LI", Name: "Liechtenstein", Lat: 47.1410, Lng: 9.5209},
	{Code: "LT", Name: "Lithuania", Lat: 54.6872, Lng: 25.2797},
	{Code: "LU", Name: "Luxembourg", Lat: 49.6116, Lng: 6.1319},
	{Code: "MG", Name: "Madagascar", Lat: -18.8792, Lng: 47.5079},
	{Code: "MW", Name: "Malawi", Lat: -13.9626, Lng: 33.7741},
	{Code: "MY", Name: "Malaysia", Lat: 3.1390, Lng: 101.6869},
	{Code: "MV", Name: "Maldives", Lat: 4.1755, Lng: 73.5093},
	{Code: "ML", Name: "Mali", Lat: 12.6392, Lng: -8.0029},
	{Code: "MT", Name: "Malta", Lat: 35.8989, Lng: 14.5146},
	{Code: "MH", Name: "Marshall Islands", Lat: 7.0897, Lng: 171.3803},
	{Code: "MR", Name: "Mauritania", Lat: 18.0735, Lng: -15.9582},
	{Code: "MU", Name: "Mauritius", Lat: -20.1609, Lng: 57.5012},
	{Code: "MX", Name: "Mexico", Lat: 19.4326, Lng: -99.1332},
	{Code: "FM", Name: "Micronesia", Lat: 6.9248, Lng: 158.1611},
	{Code: "MD", Name: "Moldova", Lat: 47.0105, Lng: 28.8638},
	{Code: "MC", Name: "Monaco", Lat: 43.7384, Lng: 7.4246},
	{Code: "MN", Name: "Mongolia", Lat: 47.8864, Lng: 106.9057},
	{Code: "ME", Name: "Montenegro", Lat: 42.4304, Lng: 19.2594},
	{Code: "MA", Name: "Morocco", Lat: 34.0209, Lng: -6.8416},
	{Code: "MZ", Name: "Mozambique", Lat: -25.9692, Lng: 32.5732},
	{Code: "MM", Name: "Myanmar", Lat: 19.7633, Lng: 96.0785},
	{Code: "NA", Name: "Namibia", Lat: -22.5609, Lng: 17.0658},
	{Code: "NR", Name: "Nauru", Lat: -0.5477, Lng: 166.9209},
	{Code: "NP", Name: "Nepal", Lat: 27.7172, Lng: 85.3240},
	{Code: "NZ", Name: "New Zealand", Lat: -41.2866, Lng: 174.7756},
	{Code: "NI", Name: "Nicaragua", Lat: 12.1150, Lng: -86.2362},
	{Code: "NE", Name: "Niger", Lat: 13.5117, Lng: 2.1254},
	{Code: "NG", Name: "Nigeria", Lat: 9.0765, Lng: 7.3986},
	{Code: "KP", Name: "North Korea", Lat: 39.0392, Lng: 125.7625},
	{Code: "MK", Name: "North Macedonia", Lat: 41.9973, Lng: 21.4280},
	{Code: "NO", Name: "Norway", Lat: 59.9139, Lng: 10.7522},
	{Code: "OM", Name: "Oman", Lat: 23.5880, Lng: 58.3829},
	{Code: "PK", Name: "Pakistan", Lat: 33.6844, Lng: 73.0479},
	{Code: "PW", Name: "Palau", Lat: 7.5000, Lng: 134.6242},
	{Code: "PS", Name: "Palestine", Lat: 31.9038, Lng: 35.2034},
	{Code: "PA", Name: "Panama", Lat: 8.9824, Lng: -79.5199},
	{Code: "PG", Name: "Papua New Guinea", Lat: -9.4438, Lng: 147.1803},
	{Code: "PY", Name: "Paraguay", Lat: -25.2637, Lng: -57.5759},
	{Code: "PE", Name: "Peru", Lat: -12.0464, Lng: -77.0428},
	{Code: "PH", Name: "Philippines", Lat: 14.5995, Lng: 120.9842},
	{Code: "PL", Name: "Poland", Lat: 52.2297, Lng: 21.0122},
	{Code: "QA", Name: "Qatar", Lat: 25.2854, Lng: 51.5310},
	{Code: "RO", Name: "Romania", Lat: 44.4268, Lng: 26.1025},
	{Code: "RU", Name: "Russia", Lat: 55.7558, Lng: 37.6173},
	{Code: "RW", Name: "Rwanda", Lat: -1.9441, Lng: 30.0619},
	{Code: "KN", Name: "Saint Kitts and Nevis", Lat: 17.3026, Lng: -62.7177},
	{Code: "LC", Name: "Saint Lucia", Lat: 14.0101, Lng: -60.9875},
	{Code: "VC", Name: "Saint Vincent and the Grenadines", Lat: 13.1600, Lng: -61.2248},
	{Code: "WS", Name: "Samoa", Lat: -13.8507, Lng: -171.7514},
	{Code: "SM", Name: "San Marino", Lat: 43.9424, Lng: 12.4578},
	{Code: "ST", Name: "São Tomé and Príncipe", Lat: 0.3302, Lng: 6.7333},
	{Code: "SA", Name: "Saudi Arabia", Lat: 24.7136, Lng: 46.6753},
	{Code: "SN", Name: "Senegal", Lat: 14.7167, Lng: -17.4677},
	{Code: "RS", Name: "Serbia", Lat: 44.7866, Lng: 20.4489},
	{Code: "SC", Name: "Seychelles", Lat: -4.6796, Lng: 55.4920},
	{Code: "SL", Name: "Sierra Leone", Lat: 8.4657, Lng: -13.2317},
	{Code: "SG", Name: "Singapore", Lat: 1.3521, Lng: 103.8198},
	{Code: "SK", Name: "Slovakia", Lat: 48.1486, Lng: 17.1077},
	{Code: "SI", Name: "Slovenia", Lat: 46.0569, Lng: 14.5058},
	{Code: "SB", Name: "Solomon Islands", Lat: -9.4456, Lng: 159.9729},
	{Code: "SO", Name: "Somalia", Lat: 2.0469, Lng: 45.3182},
	{Code: "ZA", Name: "South Africa", Lat: -25.7479, Lng: 28.2293},
	{Code: "KR", Name: "South Korea", Lat: 37.5665, Lng: 126.9780},
	{Code: "SS", Name: "South Sudan", Lat: 4.8594, Lng: 31.5713},
	{Code: "LK", Name: "Sri Lanka", Lat: 6.9271, Lng: 79.8612},
	{Code: "SD", Name: "Sudan", Lat: 15.5007, Lng: 32.5599},
	{Code: "SR", Name: "Suriname", Lat: 5.8520, Lng: -55.2038},
	{Code: "SE", Name: "Sweden", Lat: 59.3293, Lng: 18.0686},
	{Code: "CH", Name: "Switzerland", Lat: 46.9480, Lng: 7.4474},
	{Code: "SY", Name: "Syria", Lat: 33.5138, Lng: 36.2765},
	{Code: "TW", Name: "Taiwan", Lat: 25.0330, Lng: 121.5654},
	{Code: "TJ", Name: "Tajikistan", Lat: 38.5598, Lng: 68.7870},
	{Code: "TZ", Name: "Tanzania", Lat: -6.1630, Lng: 35.7516},
	{Code: "TH", Name: "Thailand", Lat: 13.7563, Lng: 100.5018},
	{Code: "TL", Name: "Timor-Leste", Lat: -8.5569, Lng: 125.5603},
	{Code: "TG", Name: "Togo", Lat: 6.1256, Lng: 1.2254},
	{Code: "TO", Name: "Tonga", Lat: -21.1393, Lng: -175.2049},
	{Code: "TT", Name: "Trinidad and Tobago", Lat: 10.6596, Lng: -61.4789},
	{Code: "TN", Name: "Tunisia", Lat: 36.8065, Lng: 10.1815},
	{Code: "TR", Name: "Türkiye", Lat: 39.9334, Lng: 32.8597},
	{Code: "TM", Name: "Turkmenistan", Lat: 37.9601, Lng: 58.3261},
	{Code: "TV", Name: "Tuvalu", Lat: -8.5211, Lng: 179.1962},
	{Code: "UG", Name: "Uganda", Lat: 0.3476, Lng: 32.5825},
	{Code: "UA", Name: "Ukraine", Lat: 50.4501, Lng: 30.5234},
	{Code: "AE", Name: "United Arab Emirates", Lat: 24.4539, Lng: 54.3773},
	{Code: "GB", Name: "United Kingdom", Lat: 51.5074, Lng: -0.1278},
	{Code: "US", Name: "United States", Lat: 38.9072, Lng: -77.0369},
	{Code: "UY", Name: "Uruguay", Lat: -34.9011, Lng: -56.1645},
	{Code: "UZ", Name: "Uzbekistan", Lat: 41.2995, Lng: 69.2401},
	{Code: "VU", Name: "Vanuatu", Lat: -17.7334, Lng: 168.3273},
	{Code: "VA", Name: "Vatican City", Lat: 41.9029, Lng: 12.4534},
	{Code: "VE", Name: "Venezuela", Lat: 10.4806, Lng: -66.9036},
	{Code: "VN", Name: "Vietnam", Lat: 21.0278, Lng: 105.8342},
	{Code: "YE", Name: "Yemen", Lat: 15.3694, Lng: 44.1910},
	{Code: "ZM", Name: "Zambia", Lat: -15.3875, Lng: 28.3228},
	{Code: "ZW", Name: "Zimbabwe", Lat: -17.8252, Lng: 31.0335},
}

// allCountries is every market the selector offers: the seeded eight first, in
// their curated order, then the rest of the world alphabetically.
//
// otherCountries is that same tail on its own — the selector renders the seeded
// markets under their own heading, so it needs the remainder without them or
// every seeded market would appear twice.
//
// Both are built once at init rather than per request: the selector renders on
// almost every page and the list never changes within a run.
var allCountries, otherCountries = buildCountryLists()

// CountryName is the display name for an ISO 3166-1 alpha-2 code, over the
// whole world list rather than the eight seeded markets.
//
// Templates need this wherever a country is stored as a code and shown as a
// name — the "Active in" block on a broker's profile is the first such place,
// and a broker active in a market Previa has no listings in is exactly the case
// the eight-market lookup could not answer.
//
// Falls back to the code, so an unrecognised value is visible rather than a
// blank space where a country should be.
func CountryName(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	for _, c := range countries {
		if c.Code == code {
			return c.Name
		}
	}
	for _, c := range worldCountries {
		if c.Code == code {
			return c.Name
		}
	}
	return code
}

func buildCountryLists() (all, other []models.Country) {
	seeded := make(map[string]bool, len(countries))
	for _, c := range countries {
		seeded[c.Code] = true
	}

	other = make([]models.Country, 0, len(worldCountries))
	for _, c := range worldCountries {
		if seeded[c.Code] {
			continue // the curated record already carries cities and a tuned zoom
		}
		// Defaults that make an unseeded market behave: euro pricing so the
		// EUR-only currency control stays truthful, English content, and a
		// country-level zoom because there is no city to open on.
		c.Currency = "EUR"
		c.Locale = "en"
		c.Zoom = 6
		other = append(other, c)
	}

	all = make([]models.Country, 0, len(countries)+len(other))
	all = append(all, countries...)
	all = append(all, other...)
	return all, other
}

// HasListings reports whether a market carries seeded stock. The selector uses
// it to group the eight live markets above the rest.
func hasListings(code string) bool {
	for _, c := range countries {
		if c.Code == code {
			return true
		}
	}
	return false
}
