// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

const (
	// Expansions is the parameter name for object expansions.
	Expansions = "expansions"

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

// Constants for the names of public metrics.
const (
	Metric_FollowersCount = "followers_count"
	Metric_FollowingCount = "following_count"
	Metric_TweetCount     = "tweet_count"
	Metric_ListedCount    = "listed_count"
)
