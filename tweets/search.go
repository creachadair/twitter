// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tweets

import (
	"strconv"
	"time"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// SearchRecent conducts a search query on recent tweets matching the specified
// query filter.
//
// For query syntax, see
// https://developer.twitter.com/en/docs/twitter-api/tweets/search/integrate/build-a-rule
//
// API: tweets/search/recent
func SearchRecent(query string, opts *SearchOpts) Query {
	req := &twitter.Request{
		Method: "tweets/search/recent",
		Params: make(twitter.Params),
	}
	req.Params.Set("query", query)
	opts.addRequestParams(req)
	return Query{request: req}
}

// Meta records server metadata reported in a search reply.
type Meta struct {
	ResultCount int    `json:"result_count"`
	NewestID    string `json:"newest_id"`
	OldestID    string `json:"oldest_id"`
	NextToken   string `json:"next_token"`
}

// SearchOpts provides parameters for tweet search. A nil *SearchOpts provides
// empty or zero values for all fields.
type SearchOpts struct {
	// A pagination token provided by the server.
	PageToken string

	// The oldest UTC time from which results will be provided.
	StartTime time.Time

	// The latest (most recent) UTC time to which results will be provided.
	EndTime time.Time

	// The maximum number of results to return; 0 means let the server choose.
	// Non-zero values < 10 or > 100 are invalid.
	MaxResults int

	// If set, return results with IDs greater than this (exclusive).
	SinceID string

	// If set, return results with IDs smaller than this (exclusive).
	UntilID string

	Expansions []string
	Optional   []types.Fields // optional response fields
}

func (o *SearchOpts) addRequestParams(req *twitter.Request) {
	if o == nil {
		return // nothing to do
	}
	if o.PageToken != "" {
		req.Params.Set("next_token", o.PageToken)
	}
	if !o.StartTime.IsZero() {
		req.Params.Set("start_time", o.StartTime.Format(types.DateFormat))
	}
	if !o.EndTime.IsZero() {
		req.Params.Set("end_time", o.EndTime.Format(types.DateFormat))
	}
	if o.MaxResults > 0 {
		req.Params.Set("max_results", strconv.Itoa(o.MaxResults))
	}
	if o.SinceID != "" {
		req.Params.Set("since_id", o.SinceID)
	}
	if o.UntilID != "" {
		req.Params.Set("until_id", o.UntilID)
	}
	req.Params.Add(types.Expansions, o.Expansions...)
	for _, fs := range o.Optional {
		req.Params.AddFields(fs)
	}
}
