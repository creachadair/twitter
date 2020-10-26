// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package olists implements queries that operate on lists using the
// Twitter API v1.1.
package olists

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/internal/otypes"
	"github.com/creachadair/twitter/types"
)

// Members constructs a query for the members of a list.
//
// API: 1.1/lists/members
func Members(listID string, opts *ListOpts) Query {
	q := Query{
		Request: &jhttp.Request{
			Method: "1.1/lists/members.json",
			Params: make(jhttp.Params),
		},
	}
	q.Request.Params.Set("list_id", listID)
	opts.addQueryParams(&q)
	return q
}

// Subscribers constructs a query for the subscribers to a list.
//
// API: 1.1/lists/subscribers
func Subscribers(listID string, opts *ListOpts) Query {
	q := Query{
		Request: &jhttp.Request{
			Method: "1.1/lists/subscribers.json",
			Params: make(jhttp.Params),
		},
	}
	q.Request.Params.Set("list_id", listID)
	opts.addQueryParams(&q)
	return q
}

// Query is a query for list memberships.
type Query struct {
	*jhttp.Request
	opts types.UserFields
}

const nextTokenParam = "cursor"

// HasMorePages reports whether the query has more pages to fetch.  This is
// true for a freshly-constructed query, and for an invoked query where the
// server not reported a next-page token.
func (q Query) HasMorePages() bool {
	v, ok := q.Request.Params[nextTokenParam]
	return !ok || (v[0] != "" && v[0] != "0")
}

// ResetPageToekn resets (clears) the query's current page token.  Subsequently
// invoking the query will then fetch the first page of results.
func (q Query) ResetPageToken() { q.Request.Params.Reset(nextTokenParam) }

// Invoke posts the update and reports the resulting tweet.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	data, err := cli.CallRaw(ctx, q.Request)
	if err != nil {
		return nil, err
	}
	var rsp struct {
		U []*otypes.User `json:"users"`
		C string         `json:"next_cursor_str"` // N.B. abbreviated
	}
	if err := json.Unmarshal(data, &rsp); err != nil {
		return nil, &jhttp.Error{Message: "decoding response body", Err: err}
	}
	nextPage := rsp.C
	if nextPage == "0" {
		nextPage = ""
	}
	q.Request.Params.Set(nextTokenParam, nextPage)
	out := &Reply{Data: data, NextToken: nextPage}
	for _, u := range rsp.U {
		out.Users = append(out.Users, u.ToUserV2(q.opts))
	}
	return out, nil
}

// ListOpts provides parameters for list queries.  A nil *ListOpts provides
// zero values for all fields.
type ListOpts struct {
	// A pagination token provided by the server.
	PageToken string

	// The number of results to return per page (maximum 200).
	// If zero, use the server default (20).
	PerPage int

	// Optional user fields to report with a successful update.
	Optional types.UserFields
}

func (o *ListOpts) addQueryParams(q *Query) {
	if o == nil {
		return
	}
	q.Request.Params.Set("skip_status", "true") // don't return tweets
	q.Request.Params.Set("include_entities", strconv.FormatBool(o.Optional.Entities))
	if o.PageToken != "" {
		q.Request.Params.Set("cursor", o.PageToken)
	}
	if o.PerPage > 0 {
		q.Request.Params.Set("count", strconv.Itoa(o.PerPage))
	}
	q.opts = o.Optional
}

// A Reply is the response from a Query.
type Reply struct {
	Data      []byte
	Users     []*types.User
	NextToken string
}
