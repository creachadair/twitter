// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package users supports queries for user lookup.
//
// To look up one or more users by ID, use users.Lookup. Additional IDs can be
// given in the options:
//
//   single := users.Lookup("12", nil)
//   multi := users.Lookup("12", &users.LookupOpts{
//      More: []string{"16431281", "2805856351"},
//   })
//
// By default only the default fields are returned (see types.User).  To
// request additional fields or expansions, include them in the options:
//
//   q := users.Lookup("12", &users.LookupOpts{
//      Optional: []types.Fields{
//         types.UserFields{Description: true, PublicMetrics: true},
//      },
//   })
//
// To look up users by username, use users.LookupByName. As above, additional
// usernames can be included in the option keys.
//
package users

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
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
	req := &jhttp.Request{
		Method: method,
		Params: make(jhttp.Params),
	}
	req.Params.Add(param, key)
	opts.addRequestParams(param, req)
	return Query{Request: req}
}

// Followers returns a query for the followers of the specified user ID.
func Followers(userID string, opts *ListOpts) Query {
	req := &jhttp.Request{
		Method: "2/users/" + userID + "/followers",
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// A Query performs a lookup query for one or more users.
type Query struct {
	*jhttp.Request
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
		return nil, &jhttp.Error{Data: rsp.Data, Message: "decoding users data", Err: err}
	}
	out := &Reply{Reply: rsp, Users: users}
	q.Request.Params.Set(twitter.NextTokenParam, "")
	if len(rsp.Meta) != 0 {
		if err := json.Unmarshal(rsp.Meta, &out.Meta); err != nil {
			return nil, &jhttp.Error{Data: rsp.Meta, Message: "decoding response metadata", Err: err}
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

func (o *LookupOpts) addRequestParams(param string, req *jhttp.Request) {
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

func (o *ListOpts) addRequestParams(req *jhttp.Request) {
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
