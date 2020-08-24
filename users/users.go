// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package users supports queries for user lookup.
package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// Lookup constructs a lookup query for one or more users by ID.  To look up
// multiple IDs, add subsequent values to the opts.Keys field.
func Lookup(id string, opts *LookupOpts) LookupQuery {
	return newLookup("users", "ids", id, opts)
}

// LookupByName constructs a lookup query for one or more users by username.
// To look up multiple usernames, add subsequent values to the opts.Keys field.
func LookupByName(name string, opts *LookupOpts) LookupQuery {
	return newLookup("users/by", "usernames", name, opts)
}

func newLookup(method, param, key string, opts *LookupOpts) LookupQuery {
	req := &twitter.Request{
		Method: method,
		Params: make(twitter.Params),
	}
	req.Params.Add(param, key)
	opts.addRequestParams(param, req)
	return LookupQuery{request: req}
}

// A LookupQuery performs a lookup query for one or more users.
type LookupQuery struct {
	request *twitter.Request
}

// Invoke executes the query on the given context and client.
func (q LookupQuery) Invoke(ctx context.Context, cli *twitter.Client) (*LookupReply, error) {
	rsp, err := cli.Call(ctx, q.request)
	if err != nil {
		return nil, err
	}
	var users types.Users
	if err := json.Unmarshal(rsp.Data, &users); err != nil {
		return nil, fmt.Errorf("decoding user data: %v", err)
	}
	return &LookupReply{
		Reply: rsp,
		Users: users,
	}, nil
}

// A LookupReply is the response from a LookupQuery.
type LookupReply struct {
	*twitter.Reply
	Users types.Users
}

// LookupOpts provide parameters for user lookup. A nil *LookupOpts provides
// empty values for all fields.
type LookupOpts struct {
	Keys []string // additional user keys to query

	Expansions  []string
	TweetFields []string
	UserFields  []string
}

func (o *LookupOpts) addRequestParams(param string, req *twitter.Request) {
	if o == nil {
		return // nothing to do
	}
	req.Params.Add(param, o.Keys...)
	req.Params.Add(types.Expansions, o.Expansions...)
	req.Params.Add(types.TweetFields, o.TweetFields...)
	req.Params.Add(types.UserFields, o.UserFields...)
}
