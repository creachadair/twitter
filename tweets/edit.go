// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tweets

import (
	"encoding/json"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter/types"
)

// Post constructs a query to create a new tweet from the given settings.
//
// API: POST 2/tweets
func Post(opts PostOpts) Query {
	req := &jhttp.Request{
		Method:     "2/tweets",
		HTTPMethod: "POST",
		Params:     make(jhttp.Params),
	}
	data, err := json.Marshal(opts)
	req.Data = data
	req.ContentType = "application/json"
	return Query{Request: req, encodeErr: err}
}

// PostOpts is the encoding of data to create a new tweet.
type PostOpts struct {
	Text string `json:"text" twitter:"required"`

	QuotedID   string     `json:"quote_tweet_id,omitempty"`
	LimitReply string     `json:"reply_settings,omitempty"` // mentionedUsers, following
	Poll       *PollOpts  `json:"poll,omitempty"`
	Reply      *ReplyOpts `json:"reply,omitempty"`

	// TODO: DM links, super followers, geo, media
}

// PollOpts is the encoding of data to create a poll tweet.
type PollOpts struct {
	Duration types.Minutes `json:"duration_minutes,omitempty"`
	Options  []string      `json:"options"`
}

// ReplyOpts is the encoding of data to filter reply mentions.
type ReplyOpts struct {
	InReplyTo string   `json:"in_reply_to_tweet_id,omitempty"`
	Exclude   []string `json:"exclude_reply_user_ids,omitempty"`
}
