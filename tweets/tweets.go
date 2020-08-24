// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package tweets supports queries for tweet lookup and search.
package tweets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// Lookup constructs a lookup query for one or more tweet IDs.  To look up
// multiple IDs, add subsequent values the opts.IDs field.
func Lookup(id string, opts *LookupOpts) Query {
	req := &twitter.Request{
		Method: "tweets",
		Params: make(twitter.Params),
	}
	req.Params.Add("ids", id)
	opts.addRequestParams(req)
	return Query{request: req}
}

// A Query performs a lookup or search query.
type Query struct {
	request *twitter.Request
}

// Invoke executes the query on the given context and client.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	rsp, err := cli.Call(ctx, q.request)
	if err != nil {
		return nil, err
	}
	var tweets types.Tweets
	if len(rsp.Data) == 0 {
		// no results
	} else if err := json.Unmarshal(rsp.Data, &tweets); err != nil {
		return nil, fmt.Errorf("decoding tweet data: %v", err)
	}
	return &Reply{
		Reply:  rsp,
		Tweets: tweets,
	}, nil
}

// A Reply is the response from a Query.
type Reply struct {
	*twitter.Reply
	Tweets types.Tweets
}

// LookupOpts provides parameters for tweet lookup. A nil *LookupOpts provides
// empty values for all fields.
type LookupOpts struct {
	IDs []string // additional tweet IDs to query

	Expansions  []string
	MediaFields []string
	PlaceFields []string
	PollFields  []string
	TweetFields []string
	UserFields  []string
}

func (o *LookupOpts) addRequestParams(req *twitter.Request) {
	if o == nil {
		return // nothing to do
	}
	req.Params.Add("ids", o.IDs...)
	req.Params.Add(types.Expansions, o.Expansions...)
	req.Params.Add(types.MediaFields, o.MediaFields...)
	req.Params.Add(types.PlaceFields, o.PlaceFields...)
	req.Params.Add(types.PollFields, o.PollFields...)
	req.Params.Add(types.TweetFields, o.TweetFields...)
	req.Params.Add(types.UserFields, o.UserFields...)
}

// TODO: SearchStream, SampleStream
