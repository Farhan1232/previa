// Package assets records which image files ship with the site.
//
// Two things need this list at runtime: the mock data layer, which spreads
// property galleries across the available photographs, and the srcset helper,
// which must only emit responsive variants that actually exist.
//
// Both used to scan the filesystem at startup. That works for a long-lived
// server sitting next to its files, but not for a serverless deployment where
// the function bundle contains the compiled Go binary and nothing else — the
// images are served separately from a CDN. So the list is generated into Go
// source instead, and stays correct in both environments.
//
// Regenerate after adding or removing images:
//
//	go run ./pkg/assets/cmd/genmanifest
package assets

import "strings"

// PropertyPhotos returns the property photograph URLs, in filename order.
func PropertyPhotos() []string { return propertyPhotos }

// HasVariant reports whether a responsive WebP variant exists, e.g.
// "/static/img/properties/p001-400.webp".
func HasVariant(path string) bool {
	_, ok := variantSet[path]
	return ok
}

var variantSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(imageVariants))
	for _, v := range imageVariants {
		m[v] = struct{}{}
	}
	return m
}()

// Count reports how many files the manifest knows about, for startup logging.
func Count() (photos, variants int) { return len(propertyPhotos), len(imageVariants) }

// Describe is a short summary used in logs.
func Describe() string {
	p, v := Count()
	return strings.Join([]string{
		itoa(p) + " property photos",
		itoa(v) + " responsive variants",
	}, ", ")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// ---------------------------------------------------------------------------
// Country flags
// ---------------------------------------------------------------------------

// FlagPath returns the URL of the local SVG flag for an ISO 3166-1 alpha-2
// country code, and whether one exists.
//
// Callers must check ok before rendering an <img>: emitting a src for a code
// with no file behind it is what produces a broken-image placeholder, which is
// worse than the neutral fallback the template draws instead.
func FlagPath(code string) (string, bool) {
	c := strings.ToUpper(strings.TrimSpace(code))
	if _, ok := flagSet[c]; !ok {
		return "", false
	}
	return "/static/img/flags/" + strings.ToLower(c) + ".svg", true
}

// FlagCodes returns every country code with a local flag, for verification.
func FlagCodes() []string { return flagCodes }

var flagSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(flagCodes))
	for _, c := range flagCodes {
		m[c] = struct{}{}
	}
	return m
}()
