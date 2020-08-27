// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package tweets supports queries for tweet lookup and search.
//
// Lookup
//
// To look up one or more tweets by ID, use tweets.Lookup. Additional IDs can
// be given in the options:
//
//   single := tweets.Lookup(id, nil)
//   multi := tweets.Lookup(id1, &tweets.LookupOpts{
//      IDs: []string{id2, id3},
//   })
//
// By default only the default fields are returned (see types.Tweet).  To
// request additional fields or expansions, include them in the options:
//
//   q := tweets.Lookup(id, &tweets.LookupOpts{
//      TweetFields: []string{types.Tweet_AuthorID},
//      Expansions:  []string{types.ExpandAuthorID},
//   })
//
// Invoke the query to fetch the tweets:
//
//   rsp, err := q.Invoke(ctx, cli)
//
// The Tweets field of the response contains the requested tweets.  In
// addition, any attachments resulting from expansions can be fetched using
// methods on the *Reply, e.g., rsp.IncludedTweets.
//
// Search
//
// To search recent tweets, use tweets.SearchRecent:
//
//   q := tweets.SearchRecent(`from:jack has:mentions -has:media`, nil)
//
// For search query syntax, see
// https://developer.twitter.com/en/docs/twitter-api/tweets/search/integrate/build-a-rule
//
// Streaming
//
// Streaming queries take a callback that receives each response sent by the
// server. Streaming continues as long as there are more results, or until the
// callback reports an error. The tweets.SearchStream and tweets.SampleStream
// functions use this interface.
//
// For example:
//
//    q := tweets.SearchStream(func(rsp *tweets.Reply) error {
//       handle(rsp)
//       if !wantMore() {
//          return twitter.ErrStopStreaming
//       }
//       return nil
//    }, nil)
//
// If the callback returns twitter.ErrStopStreaming, the stream is terminated
// without error; otherwise the error returned by the callback is reported to
// the caller of the query.
//
// Expansions and non-default fields can be requested using *StreamOpts:
//
//    opts := &tweets.StreamOpts{
//       Expansions:  []string{types.ExpandMediaKeys},
//       MediaFields: []string{types.Media_PublicMetrics},
//    }
//
package tweets

import (
	"context"
	"encoding/json"

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
	out := &Reply{Reply: rsp}
	if len(rsp.Data) == 0 {
		// no results
	} else if err := json.Unmarshal(rsp.Data, &out.Tweets); err != nil {
		return nil, &twitter.Error{Data: rsp.Data, Message: "decoding tweet data", Err: err}
	}
	if len(rsp.Meta) != 0 {
		if err := json.Unmarshal(rsp.Meta, &out.Meta); err != nil {
			return nil, &twitter.Error{Data: rsp.Meta, Message: "decoding response metadata", Err: err}
		}
	}
	return out, nil
}

// A Reply is the response from a Query.
type Reply struct {
	*twitter.Reply
	Tweets types.Tweets
	Meta   *Meta
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
