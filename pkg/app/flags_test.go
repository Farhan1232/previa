package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"previa/pkg/assets"
	"previa/pkg/data"
)

// Country-to-flag coverage.
//
// Every market country must resolve to a local SVG that exists, parses, is not
// empty, and is not shared with another country. The dataset is read from the
// same store the application serves, so this cannot drift from what a visitor
// sees.

const flagDir = "../../public/static/img/flags"

// marketCountries returns every selectable market, from the live store.
func marketCountries(t *testing.T) []string {
	t.Helper()
	store := data.NewStore(time.Now())
	var codes []string
	for _, c := range store.Catalog.AllCountries(context.Background()) {
		codes = append(codes, c.Code)
	}
	if len(codes) == 0 {
		t.Fatal("the market dataset is empty")
	}
	return codes
}

func TestEveryMarketCountryHasAValidFlag(t *testing.T) {
	codes := marketCountries(t)

	var missing, empty, invalid []string
	seen := map[string]string{} // file digest -> first country using it
	dupes := map[string][]string{}

	for _, code := range codes {
		// 1. Resolve the flag path.
		url, ok := assets.FlagPath(code)
		if !ok {
			missing = append(missing, code)
			continue
		}

		// 2. Confirm the file exists on disk.
		path := filepath.Join(flagDir, strings.ToLower(code)+".svg")
		body, err := os.ReadFile(path)
		if err != nil {
			missing = append(missing, code+" ("+url+")")
			continue
		}

		// 3. Confirm it is non-empty.
		if len(strings.TrimSpace(string(body))) == 0 {
			empty = append(empty, code)
			continue
		}

		// 4. Confirm it is valid, parseable SVG.
		var root struct {
			XMLName xml.Name
			ViewBox string `xml:"viewBox,attr"`
		}
		if err := xml.Unmarshal(body, &root); err != nil {
			invalid = append(invalid, code+": "+err.Error())
			continue
		}
		if root.XMLName.Local != "svg" {
			invalid = append(invalid, code+": root element is <"+root.XMLName.Local+">")
			continue
		}
		if root.ViewBox == "" {
			invalid = append(invalid, code+": no viewBox, so it cannot scale")
			continue
		}

		// 5. No emoji anywhere in the asset.
		if regexp.MustCompile(`[\x{1F1E6}-\x{1F1FF}]`).Match(body) {
			invalid = append(invalid, code+": contains emoji flag characters")
			continue
		}

		// 6. Detect two countries pointing at identical artwork.
		sum := sha256.Sum256(body)
		digest := hex.EncodeToString(sum[:])
		if first, clash := seen[digest]; clash {
			dupes[digest] = append(dupes[digest], code)
			_ = first
		} else {
			seen[digest] = code
		}
	}

	if len(missing) > 0 {
		t.Errorf("countries with no flag (%d): %v", len(missing), missing)
	}
	if len(empty) > 0 {
		t.Errorf("empty flag files (%d): %v", len(empty), empty)
	}
	if len(invalid) > 0 {
		t.Errorf("invalid flag SVGs (%d): %v", len(invalid), invalid)
	}
	for digest, codes := range dupes {
		t.Errorf("countries sharing identical artwork: %s and %v", seen[digest], codes)
	}

	resolved := len(codes) - len(missing) - len(empty) - len(invalid)
	if resolved != len(codes) {
		t.Errorf("resolved %d flags for %d countries; want a one-to-one mapping", resolved, len(codes))
	}
	t.Logf("market countries: %d, valid flags: %d, missing: %d", len(codes), resolved, len(missing))
}

// Every flag the app serves must actually be reachable over HTTP.
func TestEveryFlagIsServed(t *testing.T) {
	h := newServer(t)
	var bad []string

	for _, code := range marketCountries(t) {
		url, ok := assets.FlagPath(code)
		if !ok {
			bad = append(bad, code+": unresolved")
			continue
		}
		status, body := get(t, h, url)
		if status != http.StatusOK {
			bad = append(bad, code+": HTTP "+http.StatusText(status))
			continue
		}
		if !strings.Contains(body, "<svg") {
			bad = append(bad, code+": response is not SVG")
		}
	}
	if len(bad) > 0 {
		t.Errorf("flags not served correctly (%d): %v", len(bad), bad)
	}
}

// No emoji flag survives anywhere in the rendered interface.
func TestNoEmojiFlagsAreRendered(t *testing.T) {
	h := newServer(t)
	emoji := regexp.MustCompile(`[\x{1F1E6}-\x{1F1FF}]{2}`)

	for _, path := range []string{
		"/", "/search", "/search?view=map", "/dashboard",
		"/add-listing", "/checkout", "/admin/packages", "/admin/languages",
	} {
		body := mustGet(t, h, path)
		if m := emoji.FindString(body); m != "" {
			t.Errorf("%s renders an emoji flag (%q)", path, m)
		}
	}
}

// The selector must not have silently shrunk while flags were being added.
func TestMarketDatasetIsNotReduced(t *testing.T) {
	codes := marketCountries(t)
	if len(codes) < 190 {
		t.Errorf("market dataset has %d countries, want ~192 or more", len(codes))
	}

	uniq := map[string]bool{}
	for _, c := range codes {
		if uniq[c] {
			t.Errorf("country %s appears twice in the dataset", c)
		}
		uniq[c] = true
		if len(c) != 2 || c != strings.ToUpper(c) {
			t.Errorf("country code %q is not an uppercase alpha-2 code", c)
		}
	}
}

// Flags render as local <img> assets, sized and shaped as the client asked.
func TestFlagsRenderAsLocalImages(t *testing.T) {
	body := mustGet(t, newServer(t), "/")

	if !strings.Contains(body, `<img class="flag" src="/static/img/flags/`) {
		t.Error("flags are not rendered as local <img> assets")
	}
	if strings.Contains(body, "cdn.") || strings.Contains(body, "//flagcdn") {
		t.Error("a flag is being loaded from a CDN")
	}
	// Dimensions are declared so the row does not reflow as flags arrive.
	if !strings.Contains(body, `width="22" height="16"`) {
		t.Error("flags do not declare their 22x16 dimensions")
	}
	// No decorative frame is drawn in the markup.
	if regexp.MustCompile(`<img class="flag"[^>]*style=`).MatchString(body) {
		t.Error("a flag carries an inline style; sizing belongs in the stylesheet")
	}
}
