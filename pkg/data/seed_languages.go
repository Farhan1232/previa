package data

import "strings"

// Languages of communication.
//
// The client's 17 August note: "under the user's profile add more field
// 'Languages of communication'. There the user can choose in dropdown menu in
// all the languages with flags (the same menu as choose your market in the
// frontpage). Then he can add these languages there as tag's and remove them if
// needed. And then into the search-filter menu at the end make the options for
// 'Language of communication' where the user can select the languages in what
// the property is sold."
//
// So this is one catalogue serving three places, exactly as the country list
// serves the market picker, the account settings and the search filter:
//
//   · the tag picker under Account → Settings → Profile
//   · the "Language of communication" filter at the foot of the search panel
//   · the language badges on a broker's profile and the broker directory filter
//
// It has to be one list, because a language chosen in the first has to be
// findable by the second. The brokers directory used to carry its own
// hand-written array of thirteen English names in the handler, and a broker
// whose language was spelled differently there simply could not be found.
//
// Stored as ISO 639-1 codes rather than names, for the same reason deal types
// and property types are stored as slugs: the code travels in the URL —
// `language=de` — and stays valid when the label is translated.
//
// The flag is a country, not a language: the flag component is keyed by ISO
// 3166-1 alpha-2 and the two only coincide by accident. English is drawn with
// the British flag, Czech with the Czech one; the choice is the country a
// European audience most associates with the language, which is why Arabic
// takes the Arab League's most-populous state rather than any one member.

// SpokenLanguage is one entry in the languages-of-communication catalogue.
type SpokenLanguage struct {
	Code   string // ISO 639-1, lowercase — what travels in a URL
	Name   string // English name, the label in an English interface
	Native string // endonym, shown beside the name so a speaker recognises it
	Flag   string // ISO 3166-1 alpha-2 for the flag image
}

// spokenLanguages is the catalogue, ordered so the eight markets Previa
// operates in come first — those are the languages a listing is most likely to
// be sold in — and the rest follow alphabetically by English name.
//
// The picker has a search box, exactly like the market selector, so length is
// not a usability problem: a user types "pol" rather than scrolling to Polish.
var spokenLanguages = []SpokenLanguage{
	// The eight Previa markets.
	{"en", "English", "English", "GB"},
	{"et", "Estonian", "Eesti", "EE"},
	{"de", "German", "Deutsch", "DE"},
	{"es", "Spanish", "Español", "ES"},
	{"fi", "Finnish", "Suomi", "FI"},
	{"pt", "Portuguese", "Português", "PT"},
	{"nl", "Dutch", "Nederlands", "NL"},
	{"cs", "Czech", "Čeština", "CZ"},

	// Everything else, alphabetically.
	{"sq", "Albanian", "Shqip", "AL"},
	{"ar", "Arabic", "العربية", "EG"},
	{"hy", "Armenian", "Հայերեն", "AM"},
	{"az", "Azerbaijani", "Azərbaycan", "AZ"},
	{"eu", "Basque", "Euskara", "ES"},
	{"be", "Belarusian", "Беларуская", "BY"},
	{"bn", "Bengali", "বাংলা", "BD"},
	{"bs", "Bosnian", "Bosanski", "BA"},
	{"bg", "Bulgarian", "Български", "BG"},
	{"ca", "Catalan", "Català", "AD"},
	{"zh", "Chinese", "中文", "CN"},
	{"hr", "Croatian", "Hrvatski", "HR"},
	{"da", "Danish", "Dansk", "DK"},
	{"fa", "Farsi", "فارسی", "IR"},
	{"fr", "French", "Français", "FR"},
	{"ka", "Georgian", "ქართული", "GE"},
	{"el", "Greek", "Ελληνικά", "GR"},
	{"he", "Hebrew", "עברית", "IL"},
	{"hi", "Hindi", "हिन्दी", "IN"},
	{"hu", "Hungarian", "Magyar", "HU"},
	{"is", "Icelandic", "Íslenska", "IS"},
	{"id", "Indonesian", "Bahasa Indonesia", "ID"},
	{"ga", "Irish", "Gaeilge", "IE"},
	{"it", "Italian", "Italiano", "IT"},
	{"ja", "Japanese", "日本語", "JP"},
	{"kk", "Kazakh", "Қазақша", "KZ"},
	{"ko", "Korean", "한국어", "KR"},
	{"lv", "Latvian", "Latviešu", "LV"},
	{"lt", "Lithuanian", "Lietuvių", "LT"},
	{"lb", "Luxembourgish", "Lëtzebuergesch", "LU"},
	{"mk", "Macedonian", "Македонски", "MK"},
	{"ms", "Malay", "Bahasa Melayu", "MY"},
	{"mt", "Maltese", "Malti", "MT"},
	{"no", "Norwegian", "Norsk", "NO"},
	{"pl", "Polish", "Polski", "PL"},
	{"pa", "Punjabi", "ਪੰਜਾਬੀ", "PK"},
	{"ro", "Romanian", "Română", "RO"},
	{"ru", "Russian", "Русский", "RU"},
	{"sr", "Serbian", "Српски", "RS"},
	{"sk", "Slovak", "Slovenčina", "SK"},
	{"sl", "Slovenian", "Slovenščina", "SI"},
	{"sv", "Swedish", "Svenska", "SE"},
	{"tl", "Tagalog", "Tagalog", "PH"},
	{"th", "Thai", "ไทย", "TH"},
	{"tr", "Turkish", "Türkçe", "TR"},
	{"uk", "Ukrainian", "Українська", "UA"},
	{"ur", "Urdu", "اردو", "PK"},
	{"vi", "Vietnamese", "Tiếng Việt", "VN"},
}

// SpokenLanguages returns the whole catalogue, in catalogue order.
func SpokenLanguages() []SpokenLanguage { return spokenLanguages }

// spokenByCode indexes the catalogue for the lookups below.
var spokenByCode = func() map[string]SpokenLanguage {
	m := make(map[string]SpokenLanguage, len(spokenLanguages))
	for _, l := range spokenLanguages {
		m[l.Code] = l
	}
	return m
}()

// SpokenLanguageByCode returns one entry and whether the code is known.
func SpokenLanguageByCode(code string) (SpokenLanguage, bool) {
	l, ok := spokenByCode[strings.ToLower(strings.TrimSpace(code))]
	return l, ok
}

// IsSpokenLanguage reports whether a raw value names a language in the
// catalogue. Used when parsing a URL, so a hand-edited `language=` cannot
// filter on something that exists nowhere and silently return an empty page.
func IsSpokenLanguage(code string) bool {
	_, ok := SpokenLanguageByCode(code)
	return ok
}

// LanguageName is the English label for a code, falling back to the code itself
// so an unrecognised value is visible rather than blank.
func LanguageName(code string) string {
	if l, ok := SpokenLanguageByCode(code); ok {
		return l.Name
	}
	return code
}

// LanguageFlag is the country code whose flag stands for a language, falling
// back to Great Britain rather than an empty string: the flag component renders
// nothing for an unknown code, which would leave a ragged gap in a list of tags
// instead of a slightly wrong flag.
func LanguageFlag(code string) string {
	if l, ok := SpokenLanguageByCode(code); ok {
		return l.Flag
	}
	return "GB"
}

// LanguageNames maps a list of codes onto their English labels, dropping any
// the catalogue does not know.
func LanguageNames(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		if l, ok := SpokenLanguageByCode(c); ok {
			out = append(out, l.Name)
		}
	}
	return out
}
