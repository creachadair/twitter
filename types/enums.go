// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

// Fields defines a set of optional response fields to request.  This interface
// is satisfied by the generated enumeration types.
type Fields interface {
	// Return the parameter label for this field type.
	Label() string

	// Return the values selected for this field type.
	Values() []string
}

// MiscFields adapts an arbitrary label and values to the Fields interface.
func MiscFields(label string, values []string) Fields {
	return miscFields{label: label, values: values}
}

type miscFields struct {
	label  string
	values []string
}

func (m miscFields) Label() string    { return m.label }
func (m miscFields) Values() []string { return m.values }

// Expansions represents a set of object field expansions.
type Expansions []string

// Label satisfies part of the Fields interface.
func (Expansions) Label() string { return "expansions" }

// Values satisfies part of the Fields interface.
func (e Expansions) Values() []string { return []string(e) }

const (
	// Expand_AuthorID returns a user object representing the Tweet’s author.
	Expand_AuthorID = "author_id"

	// Expand_ReferencedTweetID returns a Tweet object that this Tweet is
	// referencing (either as a Retweet, Quoted Tweet, or reply).
	Expand_ReferencedTweetID = "referenced_tweets.id"

	// Expand_InReplyTo returns a user object representing the Tweet author this
	// requested Tweet is a reply of.
	Expand_InReplyTo = "in_reply_to_user_id"

	// Expand_MediaKeys returns a media object representing the images, videos,
	// GIFs included in the Tweet.
	Expand_MediaKeys = "attachments.media_keys"

	// Expand_PollID returns a poll object containing metadata for the poll
	// included in the Tweet.
	Expand_PollID = "attachments.poll_ids"

	// Expand_PlaceID returns a place object containing metadata for the
	// location tagged in the Tweet.
	Expand_PlaceID = "geo.place_id"

	// Expand_MentionUsername returns a user object for the user mentioned in
	// the Tweet.
	Expand_MentionUsername = "entities.mentions.username"

	// Expand_ReferencedAuthorID returns a user object for the author of the
	// referenced Tweet.
	Expand_ReferencedAuthorID = "referenced_tweets.id.author_id"

	// Expand_PinnedTweetID returns a Tweet object representing the Tweet pinned
	// to the top of the user’s profile.
	Expand_PinnedTweetID = "pinned_tweet_id"
)

// Constants for the names of various metrics reported in a Metrics map.  The
// comment beside each constant describes its visibility.
//
// See https://developer.twitter.com/en/docs/twitter-api/metrics
const (
	Metric_FollowersCount    = "followers_count"     // public
	Metric_FollowingCount    = "following_count"     // public
	Metric_ImpressionCount   = "impression_count"    // non-public, organic, promoted
	Metric_LikeCount         = "like_count"          // public, organic, promoted
	Metric_ListedCount       = "listed_count"        // public
	Metric_QuoteCount        = "quote_count"         // public
	Metric_ReplyCount        = "reply_count"         // public, organic, promoted
	Metric_RetweetCount      = "retweet_count"       // public, organic, promoted
	Metric_TweetCount        = "tweet_count"         // public
	Metric_URLLinkClicks     = "url_link_clicks"     // non-public, organic, promoted
	Metric_UserProfileClicks = "user_profile_clicks" // non-public, organic, promoted
	Metric_ViewCount         = "view_count"          // public, organic, promoted

	// Video view quartile metrics. Non-public, organic, promoted.
	Metric_Playback0Count   = "playback_0_count"
	Metric_Playback25Count  = "playback_25_count"
	Metric_Playback50Count  = "playback_50_count"
	Metric_Playback75Count  = "playback_75_count"
	Metric_Playback100Count = "playback_100_count"
)
