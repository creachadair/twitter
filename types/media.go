package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Milliseconds defines the JSON encoding of a duration in milliseconds.
type Milliseconds time.Duration

// UnmarshalJSON decodes d from a JSON integer number of milliseconds.
func (m *Milliseconds) UnmarshalJSON(bits []byte) error {
	var ms int64
	if err := json.Unmarshal(bits, &ms); err != nil {
		return fmt.Errorf("cannot decode %q as an integer", string(bits))
	}
	*m = Milliseconds(time.Duration(ms) * time.Millisecond)
	return nil
}

// MarshalJSON encodes m as a JSON integer number of milliseconds.
func (m Milliseconds) MarshalJSON() ([]byte, error) {
	ms := strconv.FormatInt(int64(time.Duration(m)/time.Millisecond), 10)
	return []byte(ms), nil

}

// Media refers to any image, GIF, or video attached to a tweet.
// The fields marked "default" will always be populated by the API; other
// fields are filled in based on the parameters in the request.
type Media struct {
	Key  string `json:"media_key" twitter:"default"`
	Type string `json:"type" twitter:"default"` // e.g., "video"

	Duration        Milliseconds `json:"duration_ms"`
	Height          int          `json:"height"` // pixels
	Width           int          `json:"width"`  // pixels
	PreviewImageURL string       `json:"preview_image_url"`

	Attachments `json:"attachments"`
	MetricSet
}
