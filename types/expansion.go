// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

const (
	// Expansions is the parameter name for object expansions.
	Expansions = "expansions"

	// ExpandAuthorID returns a user object representing the Tweet’s author.
	ExpandAuthorID = "author_id"

	// ExpandReferencedTweetID returns a Tweet object that this Tweet is
	// referencing (either as a Retweet, Quoted Tweet, or reply).
	ExpandReferencedTweetID = "referenced_tweets.id"

	// ExpandInReplyTo returns a user object representing the Tweet author this
	// requested Tweet is a reply of.
	ExpandInReplyTo = "in_reply_to_user_id"

	// ExpandMediaKeys returns a media object representing the images, videos,
	// GIFs included in the Tweet.
	ExpandMediaKeys = "attachments.media_keys"

	// ExpandPollID returns a poll object containing metadata for the poll
	// included in the Tweet.
	ExpandPollID = "attachments.poll_ids"

	// ExpandPlaceID returns a place object containing metadata for the location
	// tagged in the Tweet.
	ExpandPlaceID = "geo.place_id"

	// ExpandMentionUsername returns a user object for the user mentioned in the Tweet.
	ExpandMentionUsername = "entities.mentions.username"

	// ExpandReferencedAuthorID returns a user object for the author of the referenced Tweet.
	ExpandReferencedAuthorID = "referenced_tweets.id.author_id"

	// ExpandPinnedTweetID returns a Tweet object representing the Tweet pinned
	// to the top of the user’s profile.
	ExpandPinnedTweetID = "pinned_tweet_id"
)
