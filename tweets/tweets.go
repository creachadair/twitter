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
//      More: []string{id2, id3},
//   })
//
// By default only the default fields are returned (see types.Tweet).  To
// request additional fields or expansions, include them in the options:
//
//   q := tweets.Lookup(id, &tweets.LookupOpts{
//      Optional: []types.Fields{
//         types.TweetFields{AuthorID: true, PublicMetrics: true},
//         types.MediaFields{Duration: true},
//         types.Expansions{types.Expand_AuthorID},
//      },
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
// Search results can be paginated. Specifically, if there are more results
// available than the requested cap (max_results), the server response will
// contain a pagination token that can be used to fetch more. Invoking a search
// query automatically updates the query with this pagination token, so
// invoking the query again will fetch the remaining results:
//
//   for q.HasMorePages() {
//      rsp, err := q.Invoke(ctx, cli)
//      // ...
//   }
//
// Use q.ResetPageToken to reset the query.
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
// the caller of the query. For the common and simple case of limiting the
// number of results, you can use the MaxResults stream option.
//
// Expansions and non-default fields can be requested using *StreamOpts:
//
//    opts := &tweets.StreamOpts{
//       Optional:    []types.Fields{
//          types.Expansions{types.Expand_MediaKeys},
//          types.MediaFields{PublicMetrics: true},
//       },
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
// multiple IDs, add subsequent values the opts.More field.
//
// API: tweets
func Lookup(id string, opts *LookupOpts) Query {
	req := &twitter.Request{
		Method: "tweets",
		Params: make(twitter.Params),
	}
	req.Params.Add("ids", id)
	opts.addRequestParams(req)
	return Query{Request: req}
}

// A Query performs a lookup or search query.
type Query struct {
	*twitter.Request
}

// Invoke executes the query on the given context and client. If the reply
// contains a pagination token, q is updated in-place so that invoking the
// query again will fetch the next page.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	rsp, err := cli.Call(ctx, q.Request)
	if err != nil {
		return nil, err
	}
	out := &Reply{Reply: rsp}
	if len(rsp.Data) == 0 {
		// no results
	} else if err := json.Unmarshal(rsp.Data, &out.Tweets); err != nil {
		return nil, &twitter.Error{Data: rsp.Data, Message: "decoding tweet data", Err: err}
	}
	// Maintain the flag validity for lookup queries.
	q.Request.Params.Set(nextTokenParam, "")
	if len(rsp.Meta) != 0 {
		if err := json.Unmarshal(rsp.Meta, &out.Meta); err != nil {
			return nil, &twitter.Error{Data: rsp.Meta, Message: "decoding response metadata", Err: err}
		}
		// Update the query page token. Do this even if next_token is empty; the
		// HasMorePages method uses the presence of the parameter to distinguish
		// a fresh query from end-of-pages.
		q.Request.Params.Set(nextTokenParam, out.Meta.NextToken)
	}
	return out, nil
}

const nextTokenParam = "next_token"

// HasMorePages reports whether the query has more pages to fetch. This is true
// for a freshly-constructed query, and for an invoked query where the server
// has not reported a next-page token.
func (q Query) HasMorePages() bool {
	// To distinguish a fresh query from a query that has exhausted all pages,
	// we use the presence of nextTokenParam in the parameter map.
	//
	// If it's there but empty, there are no more pages.
	// If it's there but nonempty, there are more pages.
	// If it's not there, this is a fresh query.
	v, ok := q.Request.Params[nextTokenParam]
	return !ok || v[0] != ""
}

// PageToken reports the query's current page token, or "".
func (q Query) PageToken() string {
	v, ok := q.Request.Params[nextTokenParam]
	if ok && len(v) != 0 {
		return v[0]
	}
	return ""
}

// ResetPageToken clears (resets) the query's current page token. Subsequently
// invoking the query will then fetch the first page of results.
func (q Query) ResetPageToken() { q.Request.Params.Reset(nextTokenParam) }

// A Reply is the response from a Query.
type Reply struct {
	*twitter.Reply
	Tweets types.Tweets
	Meta   *Meta
}

// LookupOpts provides parameters for tweet lookup. A nil *LookupOpts provides
// empty values for all fields.
type LookupOpts struct {
	More      []string       // additional tweet IDs to query
	PageToken string         // a pagination token
	Optional  []types.Fields // optional response fields, expansions
}

func (o *LookupOpts) addRequestParams(req *twitter.Request) {
	if o == nil {
		return // nothing to do
	}
	if o.PageToken != "" {
		req.Params.Set("next_token", o.PageToken)
	}
	req.Params.Add("ids", o.More...)
	for _, fs := range o.Optional {
		if vs := fs.Values(); len(vs) != 0 {
			req.Params.Add(fs.Label(), vs...)
		}
	}
}
