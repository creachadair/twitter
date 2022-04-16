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
	body, err := json.Marshal(struct {
		H bool `json:"hidden"`
	}{H: hidden})
	return Edit{
		Request: &jhttp.Request{
			Method:      "2/tweets/" + tweetID + "/hidden",
			HTTPMethod:  "PUT",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "hidden",
		encodeErr: err,
	}
}

// Like constructs a query for the given user ID to like the given tweet ID.
//
// API: POST 2/users/:id/likes
func Like(userID, tweetID string) Edit {
	body, err := json.Marshal(struct {
		ID string `json:"tweet_id"`
	}{ID: tweetID})
	return Edit{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/likes",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "liked",
		encodeErr: err,
	}
}

// Unlike constructs a query for the given user ID to un-like the given tweet ID.
//
// API: DELETE 2/users/:id/likes/:tid
func Unlike(userID, tweetID string) Edit {
	return Edit{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/likes/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "liked",
	}
}

// Bookmark constructs a query for the given user ID to bookmark the given
// tweet ID.
//
// API: 2/users/:id/bookmarks
func Bookmark(userID, tweetID string) Edit {
	body, err := json.Marshal(struct {
		ID string `json:"tweet_id"`
	}{ID: tweetID})
	return Edit{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/bookmarks",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "bookmarked",
		encodeErr: err,
	}
}

// Unbookmark constructs a query for the given user ID to un-like the given
// tweet ID.
//
// API: DELETE 2/users/:id/bookmarks/:tid
func Unbookmark(userID, tweetID string) Edit {
	return Edit{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/bookmarks/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "bookmarked",
	}
}

// Retweet constructs a query for the given user ID to retweet the given tweet ID.
//
// API: POST 2/users/:id/retweets
func Retweet(userID, tweetID string) Edit {
	body, err := json.Marshal(struct {
		ID string `json:"tweet_id"`
	}{ID: tweetID})
	return Edit{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/retweets",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "retweeted",
		encodeErr: err,
	}
}

// Unretweet constructs a query for the given user ID to un-retweet the given
// tweet ID.
//
// API: DELETE 2/users/:id/retweets/:tid
func Unretweet(userID, tweetID string) Edit {
	return Edit{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/retweets/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "retweeted",
	}
}
