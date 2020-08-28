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

	CreatedAt       *time.Time    `json:"created_at"`
	Description     string        `json:"description"` // profile bio
	ProfileURL      string        `json:"url"`
	Entities        *UserEntities `json:"entities"`
	FuzzyLocation   string        `json:"location"` // human-readable
	PinnedTweetID   string        `json:"pinned_tweet_id"`
	ProfileImageURL string        `json:"profile_image_url"`

	Protected bool `json:"protected"`
	Verified  bool `json:"verified"`

	PublicMetrics Metrics      `json:"public_metrics"`
	Withheld      *Withholding `json:"withheld"`
}

// UserEntities describe entities found in a user's profile.
type UserEntities struct {
	URL         *Entities `json:"url"`
	Description *Entities `json:"description"`
}
