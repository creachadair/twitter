// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package users supports queries for user lookup.
//
// To look up one or more users by ID, use users.Lookup. Additional IDs can be
// given in the options:
//
//	single := users.Lookup("12", nil)
//	multi := users.Lookup("12", &users.LookupOpts{
//	   More: []string{"16431281", "2805856351"},
//	})
//
// By default only the default fields are returned (see types.User).  To
// request additional fields or expansions, include them in the options:
//
//	q := users.Lookup("12", &users.LookupOpts{
//	   Optional: []types.Fields{
//	      types.UserFields{Description: true, PublicMetrics: true},
//	   },
//	})
//
// To look up users by username, use users.LookupByName. As above, additional
// usernames can be included in the option keys.
package users

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/jape"
	"github.com/creachadair/twitter/types"
)

// Lookup constructs a lookup query for one or more users by ID.  To look up
// multiple IDs, add subsequent values to the opts.More field.
//
// API: 2/users
func Lookup(id string, opts *LookupOpts) Query {
	return newLookup("2/users", "ids", id, opts)
}

// LookupByName constructs a lookup query for one or more users by username.
// To look up multiple usernames, add subsequent values to the opts.More field.
//
// API: 2/users/by
func LookupByName(name string, opts *LookupOpts) Query {
	return newLookup("2/users/by", "usernames", name, opts)
}

func newLookup(method, param, key string, opts *LookupOpts) Query {
	req := &jape.Request{
		Method: method,
		Params: make(jape.Params),
	}
	req.Params.Add(param, key)
	opts.addRequestParams(param, req)
	return Query{Request: req}
}

// FollowersOf returns a query for the followers of the specified user ID.
//
// API: 2/users/:id/followers
func FollowersOf(userID string, opts *ListOpts) Query {
	req := &jape.Request{
		Method: "2/users/" + userID + "/followers",
		Params: make(jape.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// FollowedBy returns a query for those the specified user ID is following.
//
// API: 2/users/:id/following
func FollowedBy(userID string, opts *ListOpts) Query {
	req := &jape.Request{
		Method: "2/users/" + userID + "/following",
		Params: make(jape.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// MutedBy returns a query for those the specified user ID is muting.
//
// API: 2/users/:id/muting
func MutedBy(userID string, opts *ListOpts) Query {
	req := &jape.Request{
		Method: "2/users/" + userID + "/muting",
		Params: make(jape.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// BlockedBy returns a query for those the specified user ID is blocking.
//
// API: 2/users/:id/blocking
func BlockedBy(userID string, opts *ListOpts) Query {
	req := &jape.Request{
		Method: "2/users/" + userID + "/blocking",
		Params: make(jape.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// RetweetersOf returns a query for users who retweeted the specified tweet ID.
//
// API: 2/tweets/:id/retweeted_by
func RetweetersOf(tweetID string, opts *ListOpts) Query {
	req := &jape.Request{
		Method: "2/tweets/" + tweetID + "/retweeted_by",
		Params: make(jape.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// LikersOf constructs a query for the users who like a given tweet ID.
//
// API: 2/tweets/:id/liking_users
//
// BUG: The service does not understand pagination for this endpoint.
// It appears to return a fixed number of responses regardless how many there
// actually are. If you set MaxResults or PageToken in the options, the request
// will report an error.
func LikersOf(id string, opts *ListOpts) Query {
	req := &jape.Request{
		Method: "2/tweets/" + id + "/liking_users",
		Params: make(jape.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// A Query performs a lookup query for one or more users.
type Query struct {
	*jape.Request
}

// Invoke executes the query on the given context and client.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	rsp, err := cli.Call(ctx, q.Request)
	if err != nil {
		return nil, err
	}
	var users types.Users
	if len(rsp.Data) == 0 {
		// no results
	} else if err := json.Unmarshal(rsp.Data, &users); err != nil {
		return nil, &jape.Error{Data: rsp.Data, Message: "decoding users data", Err: err}
	}
	out := &Reply{Reply: rsp, Users: users}
	q.Request.Params.Set(twitter.NextTokenParam, "")
	if len(rsp.Meta) != 0 {
		if err := json.Unmarshal(rsp.Meta, &out.Meta); err != nil {
			return nil, &jape.Error{Data: rsp.Meta, Message: "decoding response metadata", Err: err}
		}
		// Update the query page token. Do this even if next_token is empty; the
		// HasMorePages method uses the presence of the parameter to distinguish
		// a fresh query from end-of-pages.
		q.Request.Params.Set(twitter.NextTokenParam, out.Meta.NextToken)
	}
	return out, nil
}

// HasMorePages reports whether the query has more pages to fetch. This is true
// for a freshly-constructed query, and for an invoked query where the server
// has not reported a next-page token.
func (q Query) HasMorePages() bool {
	v, ok := q.Request.Params[twitter.NextTokenParam]
	return !ok || v[0] != ""
}

// ResetPageToken clears (resets) the query's current page token. Subsequently
// invoking the query will then fetch the first page of results.
func (q Query) ResetPageToken() { q.Request.Params.Reset(twitter.NextTokenParam) }

// A Reply is the response from a Query.
type Reply struct {
	*twitter.Reply
	Users types.Users
	Meta  *twitter.Pagination
}

// LookupOpts provide parameters for user lookup. A nil *LookupOpts provides
// empty values for all fields.
type LookupOpts struct {
	// Additional usernames or IDs to query
	More []string

	// Optional response fields and expansions.
	Optional []types.Fields
}

func (o *LookupOpts) addRequestParams(param string, req *jape.Request) {
	if o == nil {
		return // nothing to do
	}
	req.Params.Add(param, o.More...)
	for _, fs := range o.Optional {
		if vs := fs.Values(); len(vs) != 0 {
			req.Params.Add(fs.Label(), vs...)
		}
	}
}

// ListOpts provide parameters for listing user memberships. A nil *ListOpts
// provides empty values for all fields.
type ListOpts struct {
	// A pagination token provided by the server.
	PageToken string

	// The maximum number of results to return; 0 means let the server choose.
	// The service will accept values up to 100.
	MaxResults int

	// Optional response fields and expansions.
	Optional []types.Fields
}

func (o *ListOpts) addRequestParams(req *jape.Request) {
	if o == nil {
		return // nothing to do
	}
	if o.PageToken != "" {
		req.Params.Set(twitter.NextTokenParam, o.PageToken)
	}
	if o.MaxResults > 0 {
		req.Params.Set("max_results", strconv.Itoa(o.MaxResults))
	}
	for _, fs := range o.Optional {
		if vs := fs.Values(); len(vs) != 0 {
			req.Params.Add(fs.Label(), vs...)
		}
	}
}
