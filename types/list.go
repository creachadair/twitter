// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

import (
	"time"
)

// A List is the decoded form of list metadata.  The fields marked "default"
// will always be populated by the API; other fields are filled in based on the
// parameters in the request.
type List struct {
	ID   string `json:"id" twitter:"default"`
	Name string `json:"name" twitter:"default"`

	CreatedAt   *time.Time `json:"created_at,omitempty"`
	Description string     `json:"description,omitempty"`
	Followers   int        `json:"follower_count,omitempty"`
	Members     int        `json:"member_count,omitempty"`
	OwnerID     string     `json:"owner_id,omitempty"`
	Private     bool       `json:"private,omitempty"`
}
