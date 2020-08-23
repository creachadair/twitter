// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

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
