package tweets

import (
	"context"
	"encoding/json"
	"fmt"
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
func SearchRecent(query string, opts *SearchOpts) SearchQuery {
	req := &twitter.Request{
		Method: "tweets/search/recent",
		Params: make(twitter.Params),
	}
	req.Params.Set("query", query)
	opts.addRequestParams(req)
	return SearchQuery{request: req}
}

// A SearchQuery performs a search query on recent tweets matching a query.
type SearchQuery struct {
	request *twitter.Request
}

// Invoke executes the query on the given context and client.
func (q SearchQuery) Invoke(ctx context.Context, cli *twitter.Client) (*SearchReply, error) {
	rsp, err := cli.Call(ctx, q.request)
	if err != nil {
		return nil, err
	}
	var tweets types.Tweets
	if err := json.Unmarshal(rsp.Data, &tweets); err != nil {
		return nil, fmt.Errorf("decoding tweet data: %v", err)
	}
	return &SearchReply{
		Reply:  rsp,
		Tweets: tweets,
	}, nil
}

// A SearchReply is the response from a SearchQuery.
type SearchReply struct {
	*twitter.Reply
	Tweets types.Tweets
}

// SearchOpts provides parameters for tweet search. A nil *SearchOpts provides
// empty or zero values for all fields.
type SearchOpts struct {
	// A pagination token provided by the server.
	NextPage string

	// The oldest UTC time from which results will be provided.
	StartTime time.Time

	// The latest (most recent) UTC time to which results will be provided.
	EndTime time.Time

	// The maximum number of results to return; 0 means let the server choose.
	// Values < 10 or > 100 are invalid.
	MaxResults int

	// If set, return results with IDs greater than this (exclusive).
	SinceID string

	// If set, return results with IDs smaller than this (exclusive).
	UntilID string

	Expansions  []string
	MediaFields []string
	PlaceFields []string
	PollFields  []string
	TweetFields []string
	UserFields  []string
}

func (o *SearchOpts) addRequestParams(req *twitter.Request) {
	if o == nil {
		return // nothing to do
	}
	if o.NextPage != "" {
		req.Params.Set("next_token", o.NextPage)
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
	req.Params.Add(types.MediaFields, o.MediaFields...)
	req.Params.Add(types.PlaceFields, o.PlaceFields...)
	req.Params.Add(types.PollFields, o.PollFields...)
	req.Params.Add(types.TweetFields, o.TweetFields...)
	req.Params.Add(types.UserFields, o.UserFields...)
}
