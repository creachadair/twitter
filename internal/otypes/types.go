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
	FullText        string    `json:"full_text"` // requires tweet_mode=extended
	Source          string    `json:"source"`
	Truncated       bool      `json:"truncated"`
	InReplyToStatus string    `json:"in_reply_to_status_id_str"`
	QuotedStatusID  string    `json:"quoted_status_id_str"`
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
	if o.FullText != "" {
		t.Text = o.FullText
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
		if o.QuotedStatusID != "" {
			t.Referenced = append(t.Referenced, &types.Ref{
				Type: "quoted",
				ID:   o.QuotedStatusID,
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

func (e *Entities) isEmpty() bool {
	return e == nil || (len(e.Hashtags) == 0 && len(e.Cashtags) == 0 && len(e.URLs) == 0 && len(e.Mentions) == 0)
}

// ToEntitiesV2 converts o into an approximately equivalent API v2 value.
func (e *Entities) ToEntitiesV2() *types.Entities {
	if e.isEmpty() {
		return nil
	}
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

// User captures a subset of the fields of the v1.1 API User object, as needed
// to populate some of the essential fields of a v2 User.
//
// See https://developer.twitter.com/en/docs/twitter-api/v1/data-dictionary/overview/user-object
type User struct {
	CreatedAt       DateTime `json:"created_at"`
	Description     string   `json:"description"`
	FuzzyLocation   string   `json:"location"`
	ID              string   `json:"id_str"` // N.B. the "id" field is a number
	Name            string   `json:"name"`
	ProfileImageURL string   `json:"profile_image_url_https"`
	ProfileURL      string   `json:"url"`
	Protected       bool     `json:"protected"`
	Username        string   `json:"screen_name"`
	Verified        bool     `json:"verified"`

	Entities struct {
		URL  *Entities `json:"urls"`
		Desc *Entities `json:"description"`
	} `json:"entities"`

	// Public metrics
	FollowersCount int `json:"followers_count"`
	FollowingCount int `json:"friends_count"`
	ListedCount    int `json:"listed_count"`
	TweetCount     int `json:"statuses_count"`

	// TODO: Handle other fields.
	// Pinned tweet?
}

// ToUserV2 converts o to an approximately-equivalent API v2 User value.
func (o User) ToUserV2(opt types.UserFields) *types.User {
	u := &types.User{
		ID:        o.ID,
		Name:      o.Name,
		Username:  o.Username,
		Protected: o.Protected,
		Verified:  o.Verified,
	}
	if opt.CreatedAt && !time.Time(o.CreatedAt).IsZero() {
		ts := time.Time(o.CreatedAt)
		u.CreatedAt = &ts
	}
	if opt.Description {
		u.Description = o.Description
	}
	if opt.Entities {
		var ue types.UserEntities
		if o.Entities.URL != nil {
			ue.URL = *o.Entities.URL.ToEntitiesV2()
			u.Entities = &ue
		}
		if o.Entities.Desc != nil {
			ue.Description = *o.Entities.Desc.ToEntitiesV2()
			u.Entities = &ue
		}
	}
	if opt.FuzzyLocation {
		u.FuzzyLocation = o.FuzzyLocation
	}
	if opt.ProfileImageURL {
		u.ProfileImageURL = o.ProfileImageURL
	}
	if opt.PublicMetrics {
		u.PublicMetrics = types.Metrics{
			"followers_count": o.FollowersCount,
			"following_count": o.FollowingCount,
			"listed_count":    o.ListedCount,
			"tweet_count":     o.TweetCount,
		}
	}
	if opt.ProfileURL {
		u.ProfileURL = o.ProfileURL
	}
	return u
}
