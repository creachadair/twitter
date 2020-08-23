package types

import "encoding/json"

type Place struct {
	ID       string `json:"id" twitter:"default"`
	FullName string `json:"full_name" twitter:"default"` // e.g., "Manhattan, New York"
	Name     string `json:"name"`                        // short name, e.g., "Manhattan"
	Type     string `json:"place_type"`                  // e.g., "city"

	ContainedIn []string        `json:"contained_within"`
	CountryName string          `json:"country"`      // e.g., "United States"
	CountryCode string          `json:"country_code"` // e.g., "US"; https://www.iso.org/obp/ui/#search
	Location    json.RawMessage `json:"geo"`          // in GeoJSON; https://geojson.org/

	Attachments `json:"attachments"`
}
