// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package ousers implements queries that operate on user data using the
// Twitter API v1.1.
package ousers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/internal/otypes"
	"github.com/creachadair/twitter/types"
)

// Followers constructs a query for the followers of a user.
//
// API: 1.1/followers/list
func Followers(user string, opts *FollowOpts) Query {
	q := Query{
		Request: &jhttp.Request{
			Method: "1.1/followers/list.json",
			Params: make(jhttp.Params),
		},
	}
	opts.addQueryParams(user, &q)
	return q
}

// Following constructs a query for the "friends" of a user, which are those
// accounts the user is following.
//
// API: 1.1/friends/list
func Following(user string, opts *FollowOpts) Query {
	q := Query{
		Request: &jhttp.Request{
			Method: "1.1/friends/list.json",
			Params: make(jhttp.Params),
		},
	}
	opts.addQueryParams(user, &q)
	return q
}

// Query is a query for user relationships.
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

// FollowOpts provides parameters for follower/following queries.  A nil
// *FollowOpts provides zero values for all fields.
type FollowOpts struct {
	// Look up following by user ID rather than username.
	ByID bool

	// A pagination token provided by the server.
	PageToken string

	// The number of results to return per page (maximum 200).
	// If zero, use the server default (20).
	PerPage int

	// Optional user fields to report with a successful update.
	Optional types.UserFields
}

func (o *FollowOpts) keyField() string {
	if o != nil && o.ByID {
		return "user_id"
	}
	return "screen_name"
}

func (o *FollowOpts) addQueryParams(key string, q *Query) {
	q.Request.Params.Set(o.keyField(), key)
	if o == nil {
		return
	}
	q.Request.Params.Set("skip_status", "true") // don't return tweets
	q.Request.Params.Set("include_user_entities", strconv.FormatBool(o.Optional.Entities))
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
