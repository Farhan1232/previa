package view

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"previa/pkg/assets"
	"previa/pkg/data"
	"previa/pkg/models"
)

// Funcs is the template function map available to every template.
func Funcs() template.FuncMap {
	return template.FuncMap{
		// formatting
		"money":      Money,
		"moneyShort": MoneyShort,
		"number":     Number,
		"area":       Area,
		"pricePerM2": PricePerM2,
		"date":       Date,
		"dateLong":   DateLong,
		"timeAgo":    TimeAgo,
		"pluralize":  Pluralize,
		"truncate":   Truncate,
		"initials":   Initials,
		"percent":    Percent,

		// labels
		"typeLabel":      data.TypeLabel,
		"conditionLabel": data.ConditionLabel,
		"dealLabel":      DealLabel,
		"dealTypes":      func() any { return models.DealTypes },
		"propertyTypes":  func() any { return models.PropertyTypes },
		// One catalogue behind two screens: the add-listing form ticks it and
		// the search sidebar filters on it, "with the same subtitles".
		"amenityGroups":  func() any { return models.AmenityGroups },
		"messengerKinds": func() any { return models.MessengerKinds },
		"messengerLabel": models.MessengerLabel,
		"messengerHref":  MessengerHref,
		"searchURL":      SearchURL,
		// Deal type is a multiple choice, so the filter panel has to ask
		// "is this one of the selected ones?" rather than compare a single value.
		"hasDeal": func(selected []models.DealType, v models.DealType) bool {
			for _, d := range selected {
				if d == v {
					return true
				}
			}
			return false
		},
		"placeIcon":   PlaceIcon,
		"placeLabel":  PlaceLabel,
		"statusLabel": StatusLabel,
		"statusHint":  StatusHint,
		"statusTone":  StatusTone,
		"paymentTone": PaymentTone,
		"methodLabel": MethodLabel,

		// Languages of communication. One catalogue serves the tag picker in
		// account settings, the filter at the foot of the search panel, the
		// badges on a broker's profile and the broker directory, so a language
		// chosen in one is findable in the others.
		"spokenLanguages": data.SpokenLanguages,
		"countryName":     data.CountryName,
		"langName":        data.LanguageName,
		"langNames":       data.LanguageNames,
		"langFlagFor":     data.LanguageFlag,
		// "is this code in that list?", for ticking the boxes a seller has
		// already chosen and for marking a filter as applied.
		"hasString": func(list []string, v string) bool {
			for _, s := range list {
				if strings.EqualFold(s, v) {
					return true
				}
			}
			return false
		},
		// "did the seller enable this app?", paired with messengerHandle below
		// so the settings form can tick the box and refill the address in one
		// pass over models.MessengerKinds.
		"hasMessenger": func(list []models.Messenger, kind models.MessengerKind) bool {
			for _, m := range list {
				if m.Kind == kind {
					return true
				}
			}
			return false
		},
		"messengerHandle": func(list []models.Messenger, kind models.MessengerKind) string {
			for _, m := range list {
				if m.Kind == kind {
					return m.Handle
				}
			}
			return ""
		},

		// urls
		"queryAdd":    QueryAdd,
		"queryDrop":   QueryDrop,
		"propertyURL": PropertyURL,
		"directions":  DirectionsURL,
		"langFlag":    LangFlag,
		"flagPath":    FlagPath,
		"srcset":      Srcset,

		// logic helpers
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
		"seq":      Seq,
		"pctOf":    PctOf,
		"pctOfF":   PctOfF,
		"dict":     Dict,
		"list":     func(v ...any) []any { return v },
		"default":  Default,
		"hasField": HasField,
		"divf": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"maxf":      math.Max,
		"json":      JSONAttr,
		"jsonValue": JSONValue,
		"safeHTML":  func(s string) template.HTML { return template.HTML(s) },
		"attr":      func(s string) template.HTMLAttr { return template.HTMLAttr(s) },
		"now":       time.Now,
		// A paid period's end, stamped in UTC — "as our website is global we
		// use UTC time". A broker in Lisbon and one in Helsinki have to read
		// the same expiry off the same ad, and neither of them should have to
		// work out whose afternoon it is.
		"utc": UTC,
	}
}

// ---------------------------------------------------------------------------
// Formatting
// ---------------------------------------------------------------------------

var currencySymbols = map[string]string{
	"EUR": "€", "GBP": "£", "USD": "$", "CZK": "Kč", "SEK": "kr", "PLN": "zł",
}

// Money renders a price with its currency symbol and thin-space grouping.
func Money(m models.Money) string {
	sym, ok := currencySymbols[m.Currency]
	if !ok {
		sym = m.Currency + " "
	}
	v := Number(m.Amount)
	// Suffixed currencies read better after the number.
	if m.Currency == "CZK" || m.Currency == "SEK" || m.Currency == "PLN" {
		return v + " " + sym
	}
	return sym + v
}

// MoneyShort abbreviates large sums for dense contexts (map pins, charts).
func MoneyShort(m models.Money) string {
	sym, ok := currencySymbols[m.Currency]
	if !ok {
		sym = ""
	}
	a := m.Amount
	switch {
	case a >= 1_000_000:
		return fmt.Sprintf("%s%.1fM", sym, a/1_000_000)
	case a >= 10_000:
		return fmt.Sprintf("%s%.0fk", sym, a/1_000)
	case a >= 1_000:
		return fmt.Sprintf("%s%.1fk", sym, a/1_000)
	}
	return fmt.Sprintf("%s%.0f", sym, a)
}

// Number groups an amount with thin spaces, dropping a zero fraction.
func Number(f float64) string {
	neg := f < 0
	if neg {
		f = -f
	}
	whole := int64(f)
	frac := f - float64(whole)

	s := strconv.FormatInt(whole, 10)
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteString(" ") // thin space
		}
		b.WriteString(s[i : i+3])
	}
	out := b.String()
	if frac > 0.005 {
		out += "." + strconv.FormatFloat(frac, 'f', 2, 64)[2:]
	}
	if neg {
		out = "-" + out
	}
	return out
}

// UTC renders a moment as the date and time an expiry is quoted in, with the
// zone named so it cannot be misread as local.
//
// Zero is empty rather than "1 Jan 0001": a placement that has not been bought
// has no end, and a date on it would be a lie rather than a blank.
func UTC(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2 Jan 2006, 15:04") + " UTC"
}

// Area renders a square-metre value.
func Area(f float64) string {
	if f == 0 {
		return "—"
	}
	if f == math.Trunc(f) {
		return Number(f) + " m²"
	}
	return strconv.FormatFloat(f, 'f', 1, 64) + " m²"
}

// PricePerM2 renders the derived unit price.
func PricePerM2(p models.Property) string {
	if p.PricePerM2 <= 0 {
		return ""
	}
	sym := currencySymbols[p.Price.Currency]
	return sym + Number(p.PricePerM2) + "/m²"
}

// Date renders a short date.
func Date(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2 Jan 2006")
}

// DateLong renders a date with the time.
func DateLong(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2 January 2006, 15:04")
}

// TimeAgo renders a coarse relative time.
func TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return Pluralize(int(d.Minutes()), "minute", "minutes") + " ago"
	case d < 24*time.Hour:
		return Pluralize(int(d.Hours()), "hour", "hours") + " ago"
	case d < 48*time.Hour:
		return "yesterday"
	case d < 30*24*time.Hour:
		return Pluralize(int(d.Hours()/24), "day", "days") + " ago"
	case d < 365*24*time.Hour:
		return Pluralize(int(d.Hours()/(24*30)), "month", "months") + " ago"
	}
	return Pluralize(int(d.Hours()/(24*365)), "year", "years") + " ago"
}

// Pluralize renders "1 day" / "3 days".
func Pluralize(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

// Truncate shortens text on a word boundary.
func Truncate(n int, s string) string {
	if len(s) <= n {
		return s
	}
	cut := s[:n]
	if i := strings.LastIndex(cut, " "); i > n/2 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,.;:") + "…"
}

// Initials builds a monogram from a name.
func Initials(name string) string {
	parts := strings.Fields(name)
	switch len(parts) {
	case 0:
		return "?"
	case 1:
		return strings.ToUpper(parts[0][:1])
	}
	return strings.ToUpper(parts[0][:1] + parts[len(parts)-1][:1])
}

// Percent renders a bounded percentage.
func Percent(n int) string { return strconv.Itoa(n) + "%" }

// ---------------------------------------------------------------------------
// Labels
// ---------------------------------------------------------------------------

// SearchURL is the search page replaying a stored query string.
//
// It exists because `href="/search?{{ .Query }}"` does not work: html/template
// sees the value land in a URL's query and escapes it as a single parameter,
// so "deal=sale&country=EE" was written as "deal%3dsale%26country%3dEE" and
// every saved search ran unfiltered. Building the whole URL here puts the
// value in the right context — and the query is re-parsed rather than trusted,
// so a stored string can only ever produce ordinary key=value pairs.
func SearchURL(query string) template.URL {
	v, err := url.ParseQuery(query)
	if err != nil || len(v) == 0 {
		return template.URL("/search")
	}
	return template.URL("/search?" + v.Encode())
}

// MessengerHref marks a messenger deep link as safe to put in an href.
//
// html/template allows only http, https and mailto in a URL context and
// rewrites everything else to "#ZgotmplZ" — which silently broke the Viber
// link, whose scheme is viber://. The link is trustworthy: Messenger.Link
// either builds the whole URL itself from digits, or passes through a value it
// has already checked begins with http:// or https://.
//
// The prefixes are re-checked here rather than assumed, so a later change to
// Link cannot quietly introduce a javascript: URL through this door. An
// unrecognised scheme returns "", and the caller renders no link at all.
func MessengerHref(raw string) template.URL {
	for _, scheme := range []string{"https://", "http://", "viber://chat?", "tg://", "sgnl://"} {
		if strings.HasPrefix(raw, scheme) {
			return template.URL(raw)
		}
	}
	return ""
}

// DealLabel renders the deal type for badges.
func DealLabel(d models.DealType) string {
	switch d {
	case models.DealRent:
		return "For rent"
	case models.DealShortRent:
		return "Short rent"
	default:
		return "For sale"
	}
}

// StatusLabel renders a listing status.
//
// Three labels, because there are three states. "Pending review" and "Rejected"
// were dropped at the client's request — nothing looks at a listing before it
// goes live, so neither could ever be reached.
//
// Lower case, at the client's request ("this active and draft make with small
// letters"). It matches the "add listing" nav item and the deal-type tags,
// which are already set that way — a status is a state the listing is in, not
// a proper noun.
func StatusLabel(s models.ListingStatus) string {
	switch s {
	case models.StatusActive:
		return "active"
	case models.StatusDraft:
		return "draft"
	case models.StatusExpired:
		return "expired"
	case models.StatusSold:
		return "sold"
	}
	return string(s)
}

// StatusHint is the one-line explanation shown beside a status, so a seller
// looking at their listings can tell what each state means without being told.
//
// Expired and draft leave the seller in the same position, which is how the
// client described it, so the two hints deliberately end the same way.
func StatusHint(s models.ListingStatus) string {
	switch s {
	case models.StatusActive:
		return "Online and visible to buyers until it expires."
	case models.StatusDraft:
		return "Saved but not published. Activate it to put it online."
	case models.StatusExpired:
		return "The paid period has ended. Re-activate it to put it back online."
	case models.StatusSold:
		return "Marked as sold and taken off the market."
	}
	return ""
}

// StatusTone maps a status onto a badge variant.
//
// Expired is a warning rather than the neutral it used to be: it is the one
// state that needs the seller to do something, and it now sits beside draft in
// the same table, where two grey badges would have read as the same thing.
func StatusTone(s models.ListingStatus) string {
	switch s {
	case models.StatusActive:
		return "success"
	case models.StatusExpired:
		return "warning"
	case models.StatusDraft:
		return "neutral"
	case models.StatusSold:
		return "info"
	}
	return "neutral"
}

// PaymentTone maps a payment status onto a badge variant.
func PaymentTone(s models.PaymentStatus) string {
	switch s {
	case models.PaymentPaid:
		return "success"
	case models.PaymentPending:
		return "warning"
	case models.PaymentFailed, models.PaymentCancelled:
		return "error"
	case models.PaymentRefunded:
		return "info"
	}
	return "neutral"
}

// MethodLabel renders a payment provider name.
func MethodLabel(m string) string {
	switch m {
	case "card":
		return "Credit card"
	case "stripe":
		return "Stripe"
	case "paypal":
		return "PayPal"
	case "paysera":
		return "Paysera bank links"
	case "crypto":
		return "Crypto (NOWPayments)"
	}
	return m
}

// ---------------------------------------------------------------------------
// URLs
// ---------------------------------------------------------------------------

// QueryAdd returns the query string with key set to value, resetting the page.
func QueryAdd(q url.Values, key, value string) string {
	n := cloneValues(q)
	if value == "" {
		n.Del(key)
	} else {
		n.Set(key, value)
	}
	n.Del("page")
	if len(n) == 0 {
		return ""
	}
	return "?" + n.Encode()
}

// QueryDrop removes a filter. When value is non-empty only that one entry of a
// repeatable parameter is removed; otherwise the whole key goes.
func QueryDrop(q url.Values, key, value string) string {
	n := cloneValues(q)
	switch {
	case key == "price":
		n.Del("price_min")
		n.Del("price_max")
	case key == "area":
		n.Del("area_min")
		n.Del("area_max")
	case value == "":
		n.Del(key)
	default:
		var kept []string
		for _, v := range n[key] {
			if v != value {
				kept = append(kept, v)
			}
		}
		if len(kept) == 0 {
			n.Del(key)
		} else {
			n[key] = kept
		}
	}
	n.Del("page")
	if len(n) == 0 {
		return ""
	}
	return "?" + n.Encode()
}

func cloneValues(q url.Values) url.Values {
	n := url.Values{}
	for k, vs := range q {
		n[k] = append([]string(nil), vs...)
	}
	return n
}

// Srcset builds a WebP srcset for an image from the variants that exist beside
// it. "/static/img/properties/p001.jpg" with widths 400,800 yields
//
//	/static/img/properties/p001-400.webp 400w, /static/img/properties/p001-800.webp 800w
//
// Widths with no generated file are skipped. The original JPEG stays as the
// <img src> fallback, so a browser without WebP support still gets a picture.
func Srcset(url string, widths ...int) string {
	if url == "" || !strings.HasSuffix(url, ".jpg") {
		return ""
	}
	stem := strings.TrimSuffix(url, ".jpg")
	parts := make([]string, 0, len(widths))
	for _, w := range widths {
		p := fmt.Sprintf("%s-%d.webp", stem, w)
		if assets.HasVariant(p) {
			parts = append(parts, fmt.Sprintf("%s %dw", p, w))
		}
	}
	return strings.Join(parts, ", ")
}

// PropertyURL is the canonical detail-page path for a listing.
func PropertyURL(p models.Property) string { return "/property/" + p.Slug }

// DirectionsURL is a Google Maps *navigation* link for one listing: it opens
// Maps with the property already set as the destination, so a visitor on a
// phone lands in the Google Maps app with the route ready to start rather than
// on a pin they then have to turn into a journey themselves.
//
// The destination is the address as the page prints it, because that is what
// the client asked for — "this address is there as final destination". Google's
// cross-platform Maps URL scheme accepts either an address or a coordinate
// pair, and a listing whose address fields are all empty falls back to the pin
// so the link can never resolve to nothing.
//
// Both the address line at the top of the listing and the button under the map
// use this one function: the client's note asks for exactly the same behaviour
// in both places.
func DirectionsURL(p models.Property) string {
	parts := make([]string, 0, 4)
	for _, s := range []string{p.Address, p.District, p.City, p.Country} {
		if s = strings.TrimSpace(s); s != "" {
			parts = append(parts, s)
		}
	}
	const base = "https://www.google.com/maps/dir/?api=1&destination="
	if len(parts) == 0 {
		return base + fmt.Sprintf("%f,%f", p.Coords.Lat, p.Coords.Lng)
	}
	// %20 rather than the + that QueryEscape produces: html/template escapes a
	// literal + inside an attribute to &#43;, which is correct HTML and decodes
	// fine, but leaves the address unreadable in the page source and in a
	// status bar. Percent-encoded spaces survive both.
	return base + strings.ReplaceAll(url.QueryEscape(strings.Join(parts, ", ")), "+", "%20")
}

// langFlags maps an interface language onto the country whose flag represents
// it. The flag component is keyed by country code, not language code, and the
// two only coincide by accident — "en" is drawn with the British flag, "cs"
// with the Czech one.
var langFlags = map[string]string{
	"en": "GB",
	"de": "DE",
	"es": "ES",
	"et": "EE",
	"fi": "FI",
	"pt": "PT",
	"nl": "NL",
	"cs": "CZ",
}

// LangFlag returns the country code whose flag stands for a language.
//
// Falls back to Great Britain rather than an empty string: the flag component
// renders nothing for an unknown code, which would leave a ragged gap in the
// language menu instead of a slightly wrong flag.
func LangFlag(lang string) string {
	if c, ok := langFlags[strings.ToLower(lang)]; ok {
		return c
	}
	return "GB"
}

// ---------------------------------------------------------------------------
// Template plumbing
// ---------------------------------------------------------------------------

// PctOf returns part as a whole-number percentage of total, clamped to 0–100.
func PctOf(part, total int) int {
	if total <= 0 {
		return 0
	}
	p := part * 100 / total
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

// PctOfF is PctOf for float values, used by the chart bars.
func PctOfF(part, total float64) int {
	if total <= 0 {
		return 0
	}
	p := int(part / total * 100)
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

// Seq returns [0,n) for repeat loops (star ratings, skeleton rows).
func Seq(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// Dict builds a map inline so components can be called with named arguments:
//
//	{{ template "property-card" dict "Property" $p "Variant" "compact" }}
func Dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("dict expects an even number of arguments, got %d", len(pairs))
	}
	out := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		k, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict key %d is not a string", i)
		}
		out[k] = pairs[i+1]
	}
	return out, nil
}

// Default returns fallback when value is empty.
func Default(fallback, value any) any {
	switch v := value.(type) {
	case nil:
		return fallback
	case string:
		if v == "" {
			return fallback
		}
	case int:
		if v == 0 {
			return fallback
		}
	}
	return value
}

// HasField reports whether a dict carries a key, so shared components can take
// optional arguments.
func HasField(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

// JSONAttr escapes a string for embedding inside a JS string literal in an
// Alpine attribute.
// JSONValue marshals any value for embedding in an Alpine expression, which is
// how the statistics dialog receives a listing's whole series at the moment its
// button is pressed.
//
// Distinct from JSONAttr above, which escapes a *string* for the same position.
// Both are needed: one hands Alpine a quoted title, this one hands it an object.
//
// Returned as template.JS so html/template leaves the braces and quotes alone
// as JavaScript; the attribute it lands in is then HTML-escaped as usual, and
// the browser unescapes it before Alpine ever sees it.
func JSONValue(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		// A value that will not marshal is a programming error, but it must not
		// take the page down: null is valid JavaScript and the dialog's x-if
		// treats it as "nothing to show".
		return template.JS("null")
	}
	return template.JS(b)
}

func JSONAttr(s string) template.JS {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`, "\n", `\n`, "\r", "", "<", `<`, ">", `>`, "&", `&`)
	return template.JS(r.Replace(s))
}

// PlaceIcon maps a location-suggestion kind onto an icon name, so a country,
// a city and a street are distinguishable in the list at a glance rather than
// only by reading the label.
func PlaceIcon(kind string) string {
	switch kind {
	case models.PlaceCountry:
		return "globe"
	case models.PlaceCity:
		return "building"
	case models.PlaceDistrict:
		return "layers"
	case models.PlaceStreet:
		return "map"
	default:
		return "map-pin"
	}
}

// PlaceLabel is the human name for a suggestion kind, shown as the right-hand
// caption on each row.
func PlaceLabel(kind string) string {
	switch kind {
	case models.PlaceCountry:
		return "Country"
	case models.PlaceCity:
		return "City"
	case models.PlaceDistrict:
		return "District"
	case models.PlaceStreet:
		return "Street"
	default:
		return "Address"
	}
}

// FlagPath returns the local SVG flag URL for a country code, or "" when no
// flag ships for it. The template renders a neutral placeholder for "" rather
// than an <img> that would 404.
func FlagPath(code string) string {
	path, ok := assets.FlagPath(code)
	if !ok {
		return ""
	}
	return path
}
