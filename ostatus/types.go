// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package ostatus

import (
	"encoding/json"
	"time"

	"github.com/creachadair/twitter/types"
)

const dateFormat = "Mon Jan _2 15:04:05 -0700 2006"

type dateTime time.Time

func (d dateTime) String() string { return time.Time(d).Format(dateFormat) }

func (d *dateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ts, err := time.Parse(dateFormat, s)
	if err != nil {
		return err
	}
	*d = dateTime(ts)
	return nil
}

// oldTweet captures a subset of the fields of the v1.1 API Tweet object, as
// needed to populate some of the essential fields of a v2 Tweet.
//
// See https://developer.twitter.com/en/docs/twitter-api/v1/data-dictionary/overview/tweet-object
type oldTweet struct {
	CreatedAt       dateTime     `json:"created_at"`
	ID              string       `json:"id_str"` // N.B. the "id" field is a number
	Text            string       `json:"text"`
	Source          string       `json:"source"`
	Truncated       bool         `json:"truncated"`
	InReplyToStatus string       `json:"in_reply_to_status_id_str"`
	InReplyToUser   string       `json:"in_reply_to_user_id_str"`
	Sensitive       bool         `json:"possibly_sensitive"`
	Language        string       `json:"lang"`
	Entities        *oldEntities `json:"entities"`

	// Public metrics
	LikeCount    int `json:"favorite_count"` // note name difference
	QuoteCount   int `json:"quote_count"`
	ReplyCount   int `json:"reply_count"`
	RetweetCount int `json:"retweet_count"`

	// Author information (we only need the ID)
	User *struct {
		ID string `json:"id_str"` // N.B. the "id" field is a number
	} `json:"user"`
}

func (o oldTweet) toNewTweet(opt types.TweetFields) *types.Tweet {
	t := &types.Tweet{
		ID:        o.ID,
		Text:      o.Text,
		Sensitive: o.Sensitive,
	}
	if opt.AuthorID && o.User != nil {
		t.AuthorID = o.User.ID
	}
	if opt.CreatedAt && !time.Time(o.CreatedAt).IsZero() {
		ts := time.Time(o.CreatedAt)
		t.CreatedAt = &ts
	}
	if opt.InReplyTo {
		t.InReplyTo = o.InReplyToUser
	}
	if opt.Language && o.Language != "und" { // means "undetected"
		t.Language = o.Language
	}
	if opt.PublicMetrics {
		t.MetricSet.PublicMetrics = types.Metrics{
			"like_count":    o.LikeCount,
			"quote_count":   o.QuoteCount,
			"reply_count":   o.ReplyCount,
			"retweet_count": o.RetweetCount,
		}
	}
	if opt.Referenced {
		if o.InReplyToStatus != "" {
			t.Referenced = append(t.Referenced, &types.Ref{
				Type: "replied_to",
				ID:   o.InReplyToStatus,
			})
		}
	}
	if opt.Entities && o.Entities != nil {
		t.Entities = o.Entities.toNewEntities()
	}

	// TODO: Handle other fields.

	return t
}

func newSpan(zs []int) types.Span {
	var out types.Span
	if len(zs) > 0 {
		out.Start = zs[0]
	}
	if len(zs) > 1 {
		out.End = zs[1]
	}
	return out
}

type spanText struct {
	Span []int  `json:"indices"`
	Text string `json:"text"`
}

func (s spanText) toNewTag() *types.Tag {
	return &types.Tag{
		Span: newSpan(s.Span),
		Tag:  s.Text,
	}
}

type oldEntities struct {
	Hashtags []spanText `json:"hashtags"`
	Cashtags []spanText `json:"symbols"`

	URLs []struct {
		Span     []int  `json:"indices"`
		URL      string `json:"url"`
		Expanded string `json:"expanded_url"`
		Display  string `json:"display_url"`

		// Omitted: unwound submessage
	} `json:"urls"`

	Mentions []struct {
		Span     []int  `json:"indices"`
		Username string `json:"screen_name"`

		// Omitted other fields not used by the v2 mention
	} `json:"user_mentions"`

	// Omitted: media, polls
}

func (e *oldEntities) toNewEntities() *types.Entities {
	var out types.Entities
	for _, v := range e.Hashtags {
		out.HashTags = append(out.HashTags, v.toNewTag())
	}
	for _, v := range e.Cashtags {
		out.CashTags = append(out.CashTags, v.toNewTag())
	}
	for _, u := range e.URLs {
		out.URLs = append(out.URLs, &types.URL{
			Span:     newSpan(u.Span),
			URL:      u.URL,
			Expanded: u.Expanded,
			Display:  u.Display,
		})
	}
	for _, m := range e.Mentions {
		out.Mentions = append(out.Mentions, &types.Mention{
			Span:     newSpan(m.Span),
			Username: m.Username,
		})
	}
	return &out
}
