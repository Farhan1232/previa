// Package config reads runtime settings from the environment.
//
// No secret is ever committed. The Google Maps key is supplied through
// PREVIA_MAPS_API_KEY; when it is absent the application renders a fully
// functional mock map instead (see docs/backend-integration-points.md).
package config

import (
	"os"
	"strconv"
)

// Config holds runtime settings.
type Config struct {
	Addr           string
	BaseURL        string
	Dev            bool
	MapsKey        string
	DefaultCountry string
	TemplateDir    string
	StaticDir      string
}

// Load reads configuration from the environment, applying safe defaults so the
// project runs with no setup at all.
func Load() Config {
	return Config{
		Addr:           env("PREVIA_ADDR", ":8080"),
		BaseURL:        env("PREVIA_BASE_URL", "https://previa.estate"),
		Dev:            envBool("PREVIA_DEV", false),
		MapsKey:        os.Getenv("PREVIA_MAPS_API_KEY"), // never defaulted, never logged
		DefaultCountry: env("PREVIA_DEFAULT_COUNTRY", "EE"),
		TemplateDir:    env("PREVIA_TEMPLATE_DIR", "web/templates"),
		StaticDir:      env("PREVIA_STATIC_DIR", "public/static"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
