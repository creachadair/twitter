// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

// A Poll is the encoded description of a Twitter poll.
// The fields marked "default" will always be populated by the API; other
// fields are filled in based on the parameters in the request.
type Poll struct {
	ID      string        `json:"id" twitter:"default"`
	Options []*PollOption `json:"options" twitter:"default"`

	Duration     Minutes `json:"duration_minutes"`
	EndTime      *Date   `json:"end_datetime"`
	VotingStatus string  `json:"voting_status"` // e.g., "closed"

	Attachments `json:"attachments"`
}

// A PollOption is a single choice item in a poll.
type PollOption struct {
	Position int    `json:"position"`
	Label    string `json:"label"`
	Votes    int    `json:"votes"`
}
