// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package twitter

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/creachadair/twitter/jhttp"
	"github.com/creachadair/twitter/types"
)

// A Reply is a wrapper for the reply object returned by successful calls to
// the Twitter API v2.
type Reply struct {
	// The root reply object from the query.
	Data json.RawMessage `json:"data"`

	// For expansions that generate attachments, a map of attachment type to
	// JSON arrays of attachment objects.
	Includes map[string]json.RawMessage `json:"includes,omitempty"`

	// Server metadata reported with search replies.
	Meta json.RawMessage `json:"meta,omitempty"`

	// Error details reported with lookup or search replies.
	Errors []*types.ErrorDetail `json:"errors,omitempty"`

	// Rate limit metadata reported by the server. If the server did not return
	// these data, this field will be nil.
	RateLimit *RateLimit `json:"-"`
}

// IncludedMedia decodes any media objects in the includes of r.
// It returns nil without error if there are no media inclusions.
func (r *Reply) IncludedMedia() (types.Medias, error) {
	media, ok := r.Includes["media"]
	if !ok || len(media) == 0 {
		return nil, nil
	}
	var out types.Medias
	if err := json.Unmarshal(media, &out); err != nil {
		return nil, &jhttp.Error{Data: media, Message: "decoding media", Err: err}
	}
	return out, nil
}

// IncludedTweets decodes any tweet objects in the includes of r.
// It returns nil without error if there are no tweet inclusions.
func (r *Reply) IncludedTweets() (types.Tweets, error) {
	tweets, ok := r.Includes["tweets"]
	if !ok || len(tweets) == 0 {
		return nil, nil
	}
	var out types.Tweets
	if err := json.Unmarshal(tweets, &out); err != nil {
		return nil, &jhttp.Error{Data: tweets, Message: "decoding tweets", Err: err}
	}
	return out, nil
}

// IncludedUsers decodes any user objects in the includes of r.
// It returns nil without error if there are no user inclusions.
func (r *Reply) IncludedUsers() (types.Users, error) {
	users, ok := r.Includes["users"]
	if !ok || len(users) == 0 {
		return nil, nil
	}
	var out types.Users
	if err := json.Unmarshal(users, &out); err != nil {
		return nil, &jhttp.Error{Data: users, Message: "decoding users", Err: err}
	}
	return out, nil
}

// IncludedPolls decodes any poll objects in the includes of r.
// It returns nil without error if there are no poll inclusions.
func (r *Reply) IncludedPolls() (types.Polls, error) {
	polls, ok := r.Includes["polls"]
	if !ok || len(polls) == 0 {
		return nil, nil
	}
	var out types.Polls
	if err := json.Unmarshal(polls, &out); err != nil {
		return nil, &jhttp.Error{Data: polls, Message: "decoding polls", Err: err}
	}
	return out, nil
}

// IncludedPlaces decodes any place objects in the includes of r.
// It returns nil without error if there are no place inclusions.
func (r *Reply) IncludedPlaces() (types.Places, error) {
	places, ok := r.Includes["places"]
	if !ok || len(places) == 0 {
		return nil, nil
	}
	var out types.Places
	if err := json.Unmarshal(places, &out); err != nil {
		return nil, &jhttp.Error{Data: places, Message: "decoding places", Err: err}
	}
	return out, nil
}

// RateLimit records metadata about API rate limits reported by the server.
type RateLimit struct {
	Ceiling   int       // rate limit ceiling for this endpoint
	Remaining int       // requests remaining in the current window
	Reset     time.Time // time of next window reset
}

func decodeRateLimits(h http.Header) *RateLimit {
	ceiling := h.Get("x-rate-limit-limit")
	remaining := h.Get("x-rate-limit-remaining")
	reset := h.Get("x-rate-limit-reset")
	if ceiling == "" && remaining == "" && reset == "" {
		return nil
	}
	out := new(RateLimit)
	if v, err := strconv.Atoi(ceiling); err == nil {
		out.Ceiling = v
	}
	if v, err := strconv.Atoi(remaining); err == nil {
		out.Remaining = v
	}
	if v, err := strconv.ParseInt(reset, 10, 64); err == nil {
		out.Reset = time.Unix(v, 0)
	}
	return out
}
