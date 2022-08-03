// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

// ErrorDetail describes an error condition reported in an otherwise successful
// reply from the API, such as missing expansion data.
//
// See https://developer.twitter.com/en/support/twitter-api/error-troubleshooting
type ErrorDetail struct {
	// omitted: required_enrollment, registration_url

	ClientID     string `json:"client_id,omitempty"` // e.g., "1011011"
	Title        string `json:"title"`               // e.g., "Not Found Error"
	Detail       string `json:"detail"`              // for human consumption
	Parameter    string `json:"parameter"`           // e.g., "pinned_tweet_id"
	Value        string `json:"value"`               // e.g., "12345"
	Reason       string `json:"reason,omitempty"`    // e.g., "client-not-enrolled"
	ResourceType string `json:"resource_type"`       // e.g., "tweet"
	TypeURL      string `json:"type"`                // link to problem definition
}
