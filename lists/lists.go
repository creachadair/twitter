// Copyright (C) 2021 Michael J. Fromberger. All Rights Reserved.

// Package lists supports queries for lists.
package lists

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
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
// API: PUT 2/lists
func Delete(id string) Edit {
	req := &jhttp.Request{
		Method:     "2/lists/" + id,
		HTTPMethod: "DELETE",
	}
	return Edit{Request: req, tag: "deleted"}
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
	return &Reply{
		Reply: rsp,
		Lists: lists,
	}, nil
}

// An Edit is a query to edit or delete a list.
type Edit struct {
	*jhttp.Request
	tag string
}

// Invoke executes the query on the given context and client. A successful
// response reports whether the edit took effect.
func (e Edit) Invoke(ctx context.Context, cli *twitter.Client) (bool, error) {
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
}

// ListOpts provide parameters for list queries.  A nil *ListOpts provides
// empty values for all fields.
type ListOpts struct {
	// Optional response fields and expansions.
	Optional []types.Fields
}

func (o *ListOpts) addRequestParams(req *jhttp.Request) {
	if o == nil {
		return // nothing to do
	}
	for _, fs := range o.Optional {
		if vs := fs.Values(); len(vs) != 0 {
			req.Params.Add(fs.Label(), vs...)
		}
	}
}
