// Copyright (C) 2021 Michael J. Fromberger. All Rights Reserved.

// Package lists supports queries for lists.
package lists

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
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

// Delete constructs a query to delete an existing list.
//
// API: DELETE 2/lists/:id
func Delete(id string) Edit {
	req := &jhttp.Request{
		Method:     "2/lists/" + id,
		HTTPMethod: "DELETE",
	}
	return Edit{Request: req, tag: "deleted"}
}

// Update constructs a query to update an existing list.
//
// API: PUT 2/lists/:id
func Update(id string, opts UpdateOpts) Edit {
	req := &jhttp.Request{
		Method:     "2/lists/" + id,
		HTTPMethod: "PUT",
	}
	body, err := json.Marshal(opts)
	req.Data = body
	req.ContentType = "application/json"
	return Edit{Request: req, tag: "updated", encodeErr: err}
}

// AddMember constructs a query to add a member to an existing list.
//
// API: POST 2/lists/:id/members
func AddMember(listID, userID string) Edit {
	req := &jhttp.Request{
		Method:     "2/lists/" + listID + "/members",
		HTTPMethod: "POST",
	}
	body, err := json.Marshal(struct {
		U string `json:"user_id"`
	}{U: userID})
	req.Data = body
	req.ContentType = "application/json"
	return Edit{Request: req, tag: "is_member", encodeErr: err}
}

// DeleteMember constructs a query to remove a member from a list.
//
// API: DELETE 2/lists/:id/members/:userid
func DeleteMember(listID, userID string) Edit {
	req := &jhttp.Request{
		Method:     "2/lists/" + listID + "/members/" + userID,
		HTTPMethod: "DELETE",
	}
	return Edit{Request: req, tag: "is_member"}
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
	q.Request.Params.Set(nextTokenParam, "")
	if len(rsp.Meta) != 0 {
		if err := json.Unmarshal(rsp.Meta, &out.Meta); err != nil {
			return nil, &jhttp.Error{Data: rsp.Meta, Message: "decoding response metadata", Err: err}
		}
		// Update the query page token. Do this even if next_token is empty; the
		// HasMorePages method uses the presence of the parameter to distinguish
		// a fresh query from end-of-pages.
		q.Request.Params.Set(nextTokenParam, out.Meta.NextToken)
	}
	return out, nil
}

// nextTokenParam is the name of the pagination tokenq uery parameter.
const nextTokenParam = "pagination_token"

// HasMorePages reports whether the query has more pages to fetch. This is true
// for a freshly-constructed query, and for an invoked query where the server
// has not reported a next-page token.
func (q Query) HasMorePages() bool {
	v, ok := q.Request.Params[nextTokenParam]
	return !ok || v[0] != ""
}

// ResetPageToken clears (resets) the query's current page token. Subsequently
// invoking the query will then fetch the first page of results.
func (q Query) ResetPageToken() { q.Request.Params.Reset(nextTokenParam) }

// An Edit is a query to edit or delete a list.
type Edit struct {
	*jhttp.Request
	tag       string
	encodeErr error
}

// Invoke executes the query on the given context and client. A successful
// response reports whether the edit took effect.
func (e Edit) Invoke(ctx context.Context, cli *twitter.Client) (bool, error) {
	if e.encodeErr != nil {
		return false, e.encodeErr // deferred encoding error
	}
	rsp, err := cli.Call(ctx, e.Request)
	if err != nil {
		return false, err
	}
	m := make(map[string]*bool)
	if err := json.Unmarshal(rsp.Data, &m); err != nil {
		return false, &jhttp.Error{Data: rsp.Data, Message: "decoding response", Err: err}
	}
	if v := m[e.tag]; v != nil {
		return *v, nil
	}
	return false, fmt.Errorf("tag %q not found", e.tag)
}

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
		req.Params.Set(nextTokenParam, o.PageToken)
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

// UpdateOpts provide parameters for list update queries.  The fields that are
// non-nil are modified to the given values. Fields that are nil are not
// changed from their existing settings.
type UpdateOpts struct {
	Name    *string `json:"name,omitempty"`
	Desc    *string `json:"description,omitempty"`
	Private *bool   `json:"private,omitempty"`
}

// SetName sets the option to update the name of the list.
func (u *UpdateOpts) SetName(name string) *UpdateOpts { u.Name = &name; return u }

// SetDescription sets the option to update the description of the list.
func (u *UpdateOpts) SetDescription(desc string) *UpdateOpts { u.Desc = &desc; return u }

// SetPrivate sets the option to mark a list private or non-private.
func (u *UpdateOpts) SetPrivate(private bool) *UpdateOpts { u.Private = &private; return u }
