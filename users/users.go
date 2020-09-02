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
	req := &types.Request{
		Method: method,
		Params: make(types.Params),
	}
	req.Params.Add(param, key)
	opts.addRequestParams(param, req)
	return Query{Request: req}
}

// A Query performs a lookup query for one or more users.
type Query struct {
	*types.Request
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
		return nil, &twitter.Error{Data: rsp.Data, Message: "decoding users data", Err: err}
	}
	return &Reply{
		Reply: rsp,
		Users: users,
	}, nil
}

// A Reply is the response from a Query.
type Reply struct {
	*twitter.Reply
	Users types.Users
}

// LookupOpts provide parameters for user lookup. A nil *LookupOpts provides
// empty values for all fields.
type LookupOpts struct {
	// Additional usernames or IDs to query
	More []string

	// Optional response fields and expansions.
	Optional []types.Fields
}

func (o *LookupOpts) addRequestParams(param string, req *types.Request) {
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
