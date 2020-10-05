// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

import "time"

// A User contains Twitter user account metadata describing a Twitter user.
// The fields marked "default" will always be populated by the API; other
// fields are filled in based on the parameters in the request.
type User struct {
	ID       string `json:"id" twitter:"default"`
	Name     string `json:"name" twitter:"default"`     // e.g., "User McJones"
	Username string `json:"username" twitter:"default"` // e.g., "mcjonesey"

	CreatedAt       *time.Time    `json:"created_at,omitempty"`
	Description     string        `json:"description,omitempty"` // profile bio
	ProfileURL      string        `json:"url,omitempty"`
	Entities        *UserEntities `json:"entities,omitempty"`
	FuzzyLocation   string        `json:"location,omitempty"` // human-readable
	PinnedTweetID   string        `json:"pinned_tweet_id,omitempty"`
	ProfileImageURL string        `json:"profile_image_url,omitempty"`

	Protected bool `json:"protected,omitempty"`
	Verified  bool `json:"verified,omitempty"`

	PublicMetrics Metrics      `json:"public_metrics,omitempty"`
	Withheld      *Withholding `json:"withheld,omitempty"`
}

// UserEntities describe entities found in a user's profile.
type UserEntities struct {
	URL         Entities `json:"url,omitempty"`
	Description Entities `json:"description,omitempty"`
}
