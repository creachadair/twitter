package types

import (
	"encoding/json"
	"fmt"
	"time"
)

//go:generate go run mkenum/mkenum.go -output field_enum.go

// DateFormat defines the encoding format for a Date string.
const DateFormat = "2006-01-02T15:04:05.999Z"

// A Date defines the JSON encoding of an ISO 8601 date.
type Date time.Time

// UnmarshalJSON decodes d from a JSON string value.
func (d *Date) UnmarshalJSON(bits []byte) error {
	var s string
	if err := json.Unmarshal(bits, &s); err != nil {
		return fmt.Errorf("cannot decode %q as a date", string(bits))
	}
	ts, err := time.Parse(DateFormat, s)
	if err != nil {
		return fmt.Errorf("invalid date: %v", err)
	}
	*d = Date(ts)
	return nil
}

// MarshalJSON encodes d as a JSON string value.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format(DateFormat))
}

type Tweet struct {
	ID   string `json:"id" twitter:"default"`
	Text string `json:"text" twitter:"default"`

	AuthorID       string    `json:"author_id"`
	ConversationID string    `json:"conversation_id"`
	CreatedAt      *Date     `json:"created_at"`
	InReplyTo      string    `json:"in_reply_to_user_id"`
	Language       string    `json:"lang"` // https://tools.ietf.org/html/bcp47
	Location       *Location `json:"geo"`
	Sensitive      bool      `json:"possibly_sensitive"`
	Referenced     []*Ref    `json:"referenced_tweets"`
	Source         string    `json:"source"` // e.g., "Twitter Web App"

	ContextAnnotations []*ContextAnnotation `json:"context_annotations"`
	Withheld           *Withholding         `json:"withheld"`
	Attachments        `json:"attachments"`
	MetricSet
}

// Attachments is a map of attachment type keys to string IDs for objects
// attached to a reply.
type Attachments map[string][]string

type ContextAnnotation struct {
	Domain *Domain `json:"domain"`
	Entity *Entity `json:"entity"`
}

type Domain struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Entity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Entities struct {
	Annotations []*Annotation `json:"annotations"`
	CashTags    []*Tag        `json:"cashtags"`
	HashTags    []*Tag        `json:"hashtags"`
	Mentions    []*Tag        `json:"mentions"`
	URLs        []*URL        `json:"urls"`
}

type Annotation struct {
	Span
	Probability    float64 `json:"probability"`
	Type           string  `json:"type"`
	NormalizedText string  `json:"normalized_text"`
}

type Tag struct {
	Span
	Tag string `json:"tag"`
}

type URL struct {
	Span
	URL         string `json:"url"`
	Expanded    string `json:"expanded_url"`
	Display     string `json:"display_url"`
	Unwound     string `json:"unwound_url"`
	HTTPStatus  string `json:"status"` // e.g., "200"
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Location struct {
	PlaceID     string          `json:"place_id"`
	Coordinates json.RawMessage `json:"coordinates"` // as GeoJSON
}

type Metrics map[string]int

type MetricSet struct {
	PublicMetrics    Metrics `json:"public_metrics"`
	NonPublicMetrics Metrics `json:"non_public_metrics"`
	OrganicMetrics   Metrics `json:"organic_metrics"`
	PromotedMetrics  Metrics `json:"promoted_metrics"`
}

type Ref struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Withholding struct {
	Copyright    bool     `json:"copyright"`
	CountryCodes []string `json:"country_codes"`
}
