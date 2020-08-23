package types

// Code generated by mkenum. DO NOT EDIT.

// Legal values of the TweetFields enumeration.
const (
	// TweetFields is the label for optional Tweet field parameters.
	TweetFields = "tweet.fields"

	Tweet_Attachments        = "attachments"
	Tweet_AuthorID           = "author_id"
	Tweet_ContextAnnotations = "context_annotations"
	Tweet_ConversationID     = "conversation_id"
	Tweet_CreatedAt          = "created_at"
	Tweet_Location           = "geo"
	Tweet_InReplyTo          = "in_reply_to_user_id"
	Tweet_Language           = "lang"
	Tweet_Sensitive          = "possibly_sensitive"
	Tweet_Referenced         = "referenced_tweets"
	Tweet_Source             = "source"
	Tweet_Withheld           = "withheld"
)

// Legal values of the UserFields enumeration.
const (
	// UserFields is the label for optional User field parameters.
	UserFields = "user.fields"

	User_CreatedAt       = "created_at"
	User_Description     = "description"
	User_Entities        = "entities"
	User_FuzzyLocation   = "location"
	User_PinnedTweetID   = "pinned_tweet_id"
	User_ProfileImageURL = "profile_image_url"
	User_Protected       = "protected"
	User_PublicMetrics   = "public_metrics"
	User_ProfileURL      = "url"
	User_Verified        = "verified"
	User_Withheld        = "withheld"
)

// Legal values of the MediaFields enumeration.
const (
	// MediaFields is the label for optional Media field parameters.
	MediaFields = "media.fields"

	Media_Attachments     = "attachments"
	Media_Duration        = "duration_ms"
	Media_Height          = "height"
	Media_PreviewImageURL = "preview_image_url"
	Media_Width           = "width"
)

// Legal values of the PollFields enumeration.
const (
	// PollFields is the label for optional Poll field parameters.
	PollFields = "poll.fields"

	Poll_Attachments  = "attachments"
	Poll_Duration     = "duration_minutes"
	Poll_EndTime      = "end_datetime"
	Poll_VotingStatus = "voting_status"
)

// Legal values of the PlaceFields enumeration.
const (
	// PlaceFields is the label for optional Place field parameters.
	PlaceFields = "place.fields"

	Place_Attachments = "attachments"
	Place_ContainedIn = "contained_within"
	Place_CountryName = "country"
	Place_CountryCode = "country_code"
	Place_Location    = "geo"
	Place_Name        = "name"
	Place_Type        = "place_type"
)
