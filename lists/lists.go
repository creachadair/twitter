// Copyright (C) 2021 Michael J. Fromberger. All Rights Reserved.

// Package lists supports queries for lists.
package lists

import (
	"context"
	"encoding/json"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// Lookup constructs a query for the metadata of a list by ID.  A successful
// response will contain exactly one list.
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

// A Query performs a query for list metadata.
type Query struct {
	*jhttp.Request
}

// Invoke executes the query on the given context and client.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
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
