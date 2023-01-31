// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package ocall carries some shared utility code for calling the
// Twitter API v1.1.
package ocall

import (
	"context"
	"encoding/json"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/internal/otypes"
	"github.com/creachadair/twitter/jape"
	"github.com/creachadair/twitter/types"
)

const nextTokenParam = "cursor"

// A UsersReply is the response from a request that returns users.
type UsersReply struct {
	Data      []byte
	Users     []*types.User
	NextToken string
}

// HasMorePages reports whether the request has more pages to fetch.  This is
// true for a freshly-constructed request, and for an invoked request where the
// server not reported a next-page token.
func HasMorePages(req *jape.Request) bool {
	v, ok := req.Params[nextTokenParam]
	return !ok || (v[0] != "" && v[0] != "0")
}

// ResetPageToken resets (clears) the request's current page token.
// Subsequently invoking the query will then fetch the first page of results.
func ResetPageToken(req *jape.Request) { req.Params.Reset(nextTokenParam) }

// GetUsers invokes an API method that returns API v1.1 user objects and
// pagination metadata.
func GetUsers(ctx context.Context, req *jape.Request, opts types.UserFields, cli *twitter.Client) (*UsersReply, error) {
	data, err := cli.CallRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	var rsp struct {
		U []*otypes.User `json:"users"`
		C string         `json:"next_cursor_str"` // N.B. abbreviated
	}
	if err := json.Unmarshal(data, &rsp); err != nil {
		return nil, &jape.Error{Message: "decoding response body", Err: err}
	}
	nextPage := rsp.C
	if nextPage == "0" {
		nextPage = ""
	}
	req.Params.Set(nextTokenParam, nextPage)
	out := &UsersReply{Data: data, NextToken: nextPage}
	for _, u := range rsp.U {
		out.Users = append(out.Users, u.ToUserV2(opts))
	}
	return out, nil
}
