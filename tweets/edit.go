// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tweets

import (
	"encoding/json"
	"time"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter/types"
)

// Create constructs a query to create a new tweet from the given settings.
//
// API: POST 2/tweets
func Create(opts CreateOpts) Query {
	req := &jhttp.Request{
		Method:     "2/tweets",
		HTTPMethod: "POST",
		Params:     make(jhttp.Params),
	}
	tweet := &postTweet{Text: opts.Text, QuotedID: opts.QuoteOf}
	if opts.InReplyTo != "" {
		tweet.Reply = &replyOpts{InReplyTo: opts.InReplyTo}
	}
	if len(opts.PollOptions) != 0 {
		tweet.Poll = &pollOpts{
			Options:  opts.PollOptions,
			Duration: types.Minutes(opts.PollDuration),
		}
	}

	data, err := json.Marshal(tweet)
	req.Data = data
	req.ContentType = "application/json"
	return Query{Request: req, encodeErr: err}
}

// CreateOpts are the settings needed to create a new tweet.
type CreateOpts struct {
	Text         string        // the text of the tweet (required)
	QuoteOf      string        // the ID of a tweet to quote
	InReplyTo    string        // the ID of a tweet to reply to
	PollOptions  []string      // options to create a poll (if non-empty)
	PollDuration time.Duration // poll duration (required with poll options)
}

type postTweet struct {
	Text       string     `json:"text" twitter:"required"`
	QuotedID   string     `json:"quote_tweet_id,omitempty"`
	LimitReply string     `json:"reply_settings,omitempty"` // mentionedUsers, following
	Poll       *pollOpts  `json:"poll,omitempty"`
	Reply      *replyOpts `json:"reply,omitempty"`

	// TODO: DM links, super followers, geo, media
}

type pollOpts struct {
	Duration types.Minutes `json:"duration_minutes,omitempty"`
	Options  []string      `json:"options"`
}

type replyOpts struct {
	InReplyTo string   `json:"in_reply_to_tweet_id,omitempty"`
	Exclude   []string `json:"exclude_reply_user_ids,omitempty"`
}
