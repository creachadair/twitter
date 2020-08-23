package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Minutes defines the JSON encoding of a duration in minutes.
type Minutes time.Duration

// UnmarshalJSON decodes d from a JSON integer number of minutes.
func (m *Minutes) UnmarshalJSON(bits []byte) error {
	var min int64
	if err := json.Unmarshal(bits, &min); err != nil {
		return fmt.Errorf("cannot decode %q as an integer", string(bits))
	}
	*m = Minutes(time.Duration(min) * time.Minute)
	return nil
}

// MarshalJSON encodes d as a JSON integer number of minutes.  Time intervals
// smaller than a minute are rounded toward zero.
func (m Minutes) MarshalJSON() ([]byte, error) {
	min := strconv.FormatInt(int64(time.Duration(m)/time.Minute), 10)
	return []byte(min), nil
}

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
