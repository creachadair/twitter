// Copyright (C) 2021 Michael J. Fromberger. All Rights Reserved.

// Package lists supports queries for lists.
package lists

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
	"github.com/creachadair/twitter/users"
)

// Lookup constructs a query for the metadata of a list by ID.  A successful
// reply contains a single List value for the matching list.
//
// API: 2/lists
func Lookup(id string, opts *ListOpts) Query {
	req := &jhttp.Request{
		Method: "2/lists/" + id,
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// OwnedBy constructs a query for the metadata of lists owned by the specified
// user ID.
//
// API: 2/users/:id/owned_lists
func OwnedBy(userID string, opts *ListOpts) Query {
	req := &jhttp.Request{
		Method: "2/users/" + userID + "/owned_lists",
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// FollowedBy constructs a query for the metadata of lists followed by the
// specified user ID.
//
// API: 2/users/:id/followed_lists
func FollowedBy(userID string, opts *ListOpts) Query {
	req := &jhttp.Request{
		Method: "2/users/" + userID + "/followed_lists",
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// MemberOf constructs a query for the metadata of lists the specified user ID
// belongs to.
//
// API: 2/users/:id/list_memberships
func MemberOf(userID string, opts *ListOpts) Query {
	req := &jhttp.Request{
		Method: "2/users/" + userID + "/list_memberships",
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return Query{Request: req}
}

// Create constructs a query to create a new list. A successful reply contains
// a single List value for the created list.
//
// API: POST 2/lists
func Create(name, description string, private bool) Query {
	req := &jhttp.Request{
		Method:     "2/lists",
		HTTPMethod: "POST",
		Params:     make(jhttp.Params),
	}
	body, err := json.Marshal(struct {
		Name    string `json:"name"`
		Desc    string `json:"description,omitempty"`
		Private bool   `json:"private,omitempty"`
	}{Name: name, Desc: description, Private: private})
	req.Data = body
	req.ContentType = "application/json"
	return Query{Request: req, encodeErr: err}
}

// Members constructs a query to list the members of a list.  Note that the
// query reply contains user data, not lists.
//
// API: 2/lists/:id/members
func Members(listID string, opts *ListOpts) users.Query {
	req := &jhttp.Request{
		Method: "2/lists/" + listID + "/members",
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return users.Query{Request: req}
}

// Followers constructs a query to list the followers of a list. Note that the
// query reply contains user data, not lists.
//
// API: 2/lists/:id/followers
func Followers(listID string, opts *ListOpts) users.Query {
	req := &jhttp.Request{
		Method: "2/lists/" + listID + "/followers",
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return users.Query{Request: req}
}

// Tweets constructs a query for the tweets by members of a list. Note that the
// query reply contains tweets, not lists.
//
// API: 2/lists/:id/tweets
func Tweets(listID string, opts *ListOpts) tweets.Query {
	req := &jhttp.Request{
		Method: "2/lists/" + listID + "/tweets",
		Params: make(jhttp.Params),
	}
	opts.addRequestParams(req)
	return tweets.Query{Request: req}
}

// A Query performs a query for list metadata.
type Query struct {
	*jhttp.Request
	encodeErr error
}

// Invoke executes the query on the given context and client.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	if q.encodeErr != nil {
		return nil, q.encodeErr // deferred encoding error
	}
	rsp, err := cli.Call(ctx, q.Request)
	if err != nil {
		return nil, err
	}
	var lists types.Lists
	if len(rsp.Data) == 0 {
		// no results
	} else if rsp.Data[0] == '{' {
		// single-value return
		lists = append(lists, new(types.List))
		err = json.Unmarshal(rsp.Data, lists[0])
	} else {
		// multiple-value return
		err = json.Unmarshal(rsp.Data, &lists)
	}
	if err != nil {
		return nil, &jhttp.Error{Data: rsp.Data, Message: "decoding lists data", Err: err}
	}
	out := &Reply{Reply: rsp, Lists: lists}
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
	Lists types.Lists
	Meta  *twitter.Pagination
}

// ListOpts provide parameters for list queries.  A nil *ListOpts provides
// empty values for all fields.
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
