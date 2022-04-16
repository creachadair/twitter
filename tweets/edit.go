// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tweets

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// Create constructs a query to create a new tweet from the given settings.
//
// API: POST 2/tweets
func Create(opts CreateOpts) Query {
	req := &jhttp.Request{
		Method:     "2/tweets",
		HTTPMethod: "POST",
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

// An Edit is a query to modify the contents or properties of tweets.
type Edit struct {
	*jhttp.Request
	tag       string
	encodeErr error
}

// Invoke executes the query on the given context and client. A successful
// response reports whether the edit took effect.
func (e Edit) Invoke(ctx context.Context, cli *twitter.Client) (bool, error) {
	if e.encodeErr != nil {
		return false, e.encodeErr // deferred encoding error
	}
	rsp, err := cli.Call(ctx, e.Request)
	if err != nil {
		return false, err
	}
	m := make(map[string]*bool)
	if err := json.Unmarshal(rsp.Data, &m); err != nil {
		return false, &jhttp.Error{Data: rsp.Data, Message: "decoding response", Err: err}
	}
	if v := m[e.tag]; v != nil {
		return *v, nil
	}
	return false, fmt.Errorf("tag %q not found", e.tag)
}

// SetHidden constructs a query to set whether replies to the given tweet ID
// should (hidden == true) or should not (hidden == false) be hidden.
func SetHidden(tweetID string, hidden bool) Edit {
	req := &jhttp.Request{
		Method:     "2/tweets/" + tweetID + "/hidden",
		HTTPMethod: "PUT",
	}
	body, err := json.Marshal(struct {
		H bool `json:"hidden"`
	}{H: hidden})
	req.Data = body
	req.ContentType = "application/json"
	return Edit{Request: req, tag: "hidden", encodeErr: err}
}

// Delete constructs a query to delete the given tweet ID.
func Delete(tweetID string) Edit {
	return Edit{
		Request: &jhttp.Request{
			Method:     "2/tweets/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "deleted",
	}
}
