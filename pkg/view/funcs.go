package view

import (
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
		"money":        Money,
		"moneyShort":   MoneyShort,
		"number":       Number,
		"area":         Area,
		"pricePerM2":   PricePerM2,
		"date":         Date,
		"dateLong":     DateLong,
		"timeAgo":      TimeAgo,
		"pluralize":    Pluralize,
		"truncate":     Truncate,
		"initials":     Initials,
		"percent":      Percent,

		// labels
		"typeLabel":      data.TypeLabel,
		"conditionLabel": data.ConditionLabel,
		"dealLabel":      DealLabel,
		"statusLabel":    StatusLabel,
		"statusTone":     StatusTone,
		"paymentTone":    PaymentTone,
		"methodLabel":    MethodLabel,

		// urls
		"queryAdd":    QueryAdd,
		"queryDrop":   QueryDrop,
		"propertyURL": PropertyURL,
		"mapsURL":     MapsURL,
		"srcset":      Srcset,

		// logic helpers
		"add":      func(a, b int) int { return a + b },
		"sub":      func(a, b int) int { return a - b },
		"mul":      func(a, b int) int { return a * b },
		"mod":      func(a, b int) int { if b == 0 { return 0 }; return a % b },
		"seq":      Seq,
		"pctOf":    PctOf,
		"pctOfF":   PctOfF,
		"dict":     Dict,
		"list":     func(v ...any) []any { return v },
		"default":  Default,
		"hasField": HasField,
		"divf":     func(a, b float64) float64 { if b == 0 { return 0 }; return a / b },
		"maxf":     math.Max,
		"json":     JSONAttr,
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"attr":     func(s string) template.HTMLAttr { return template.HTMLAttr(s) },
		"now":      time.Now,
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

// DealLabel renders the deal type for badges.
func DealLabel(d models.DealType) string {
	if d == models.DealRent {
		return "For rent"
	}
	return "For sale"
}

// StatusLabel renders a listing status.
func StatusLabel(s models.ListingStatus) string {
	switch s {
	case models.StatusActive:
		return "Active"
	case models.StatusPending:
		return "Pending review"
	case models.StatusDraft:
		return "Draft"
	case models.StatusExpired:
		return "Expired"
	case models.StatusRejected:
		return "Rejected"
	case models.StatusSold:
		return "Sold"
	}
	return string(s)
}

// StatusTone maps a status onto a badge variant.
func StatusTone(s models.ListingStatus) string {
	switch s {
	case models.StatusActive:
		return "success"
	case models.StatusPending:
		return "warning"
	case models.StatusRejected:
		return "error"
	case models.StatusExpired, models.StatusDraft:
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
	case "stripe":
		return "Card (Stripe)"
	case "paypal":
		return "PayPal"
	case "paysera":
		return "Paysera bank link"
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

// MapsURL is an external directions link for the location section.
func MapsURL(c models.Coordinates) string {
	return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%f,%f", c.Lat, c.Lng)
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
func JSONAttr(s string) template.JS {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`, `"`, `\"`, "\n", `\n`, "\r", "", "<", `<`, ">", `>`, "&", `&`)
	return template.JS(r.Replace(s))
}
