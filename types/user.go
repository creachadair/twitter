package types

type User struct {
	ID       string `json:"id" twitter:"default"`
	Name     string `json:"name" twitter:"default"`     // e.g., "User McJones"
	Username string `json:"username" twitter:"default"` // e.g., "mcjonesey"

	CreatedAt       *Date         `json:"created_at"`
	Description     string        `json:"description"` // profile bio
	ProfileURL      string        `json:"url"`
	Entities        *UserEntities `json:"entities"`
	FuzzyLocation   string        `json:"location"` // human-readable
	PinnedTweetID   string        `json:"pinned_tweet_id"`
	ProfileImageURL string        `json:"profile_image_url"`

	Protected bool `json:"protected"`
	Verified  bool `json:"verified"`

	PublicMetrics Metrics      `json:"public_metrics"`
	Withheld      *Withholding `json:"withheld"`
}

type UserEntities struct {
	URL         *Entities `json:"url"`
	Description *Entities `json:"description"`
}
