// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package otypes

import (
	"encoding/json"
	"time"

	"github.com/creachadair/twitter/types"
)

// DateFormat is the timestamp format used by the Twitter v1.1 API.
const DateFormat = "Mon Jan _2 15:04:05 -0700 2006"

// DateTime represents a timestamp encoded in JSON.
type DateTime time.Time

func (d DateTime) String() string { return time.Time(d).Format(DateFormat) }

func (d *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ts, err := time.Parse(DateFormat, s)
	if err != nil {
		return err
	}
	*d = DateTime(ts)
	return nil
}

// Tweet captures a subset of the fields of the v1.1 API Tweet object, as
// needed to populate some of the essential fields of a v2 Tweet.
//
// See https://developer.twitter.com/en/docs/twitter-api/v1/data-dictionary/overview/tweet-object
type Tweet struct {
	CreatedAt       DateTime  `json:"created_at"`
	ID              string    `json:"id_str"` // N.B. the "id" field is a number
	Text            string    `json:"text"`
	Source          string    `json:"source"`
	Truncated       bool      `json:"truncated"`
	InReplyToStatus string    `json:"in_reply_to_status_id_str"`
	InReplyToUser   string    `json:"in_reply_to_user_id_str"`
	Sensitive       bool      `json:"possibly_sensitive"`
	Language        string    `json:"lang"`
	Entities        *Entities `json:"entities"`

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

// ToTweetV2 converts o into an approximately equivalent API v2 Tweet value.
func (o Tweet) ToTweetV2(opt types.TweetFields) *types.Tweet {
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
		t.Entities = o.Entities.ToEntitiesV2()
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

// SpanText represents a span of text.
type SpanText struct {
	Span []int  `json:"indices"`
	Text string `json:"text"`
}

// ToTagV2 converts s into an equivalent API v2 Tag.
func (s SpanText) ToTagV2() *types.Tag {
	return &types.Tag{
		Span: newSpan(s.Span),
		Tag:  s.Text,
	}
}

// Entities encodes a subset of API v1.1 Tweet entities.
type Entities struct {
	Hashtags []SpanText `json:"hashtags"`
	Cashtags []SpanText `json:"symbols"`

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

// ToEntitiesV2 converts o into an approximately equivalent API v2 value.
func (e *Entities) ToEntitiesV2() *types.Entities {
	var out types.Entities
	for _, v := range e.Hashtags {
		out.HashTags = append(out.HashTags, v.ToTagV2())
	}
	for _, v := range e.Cashtags {
		out.CashTags = append(out.CashTags, v.ToTagV2())
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
