// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package twitter implements a client for the Twitter API v2.  This package is
// in development and is not yet ready for production use.
//
// Usage outline
//
// The general structure of an API call is to first construct a query, then
// invoke that query with a context on a client:
//
//    cli := twitter.NewClient(&jhttp.Client{
//       Authorize: jhttp.BearerTokenAuthorizer(token),
//    })
//
//    ctx := context.Background()
//    rsp, err := users.LookupByName("jack", nil).Invoke(ctx, cli)
//    if err != nil {
//       log.Fatalf("Request failed: %v", err)
//    } else if len(rsp.Users) == 0 {
//       log.Fatal("No matches")
//    }
//    process(rsp.Users)
//
// Packages
//
// Package "types" contains the type and constant definitions for the API.
//
// Queries to look up tweets by ID or username, to search recent tweets, and to
// search or sample streams of tweets are defined in package "tweets".
//
// Queries to look up users by ID or user name are defined in package "users".
//
// Queries to read or update search rules are defined in package "rules".
//
package twitter

import (
	"context"
	"encoding/json"

	"github.com/creachadair/jhttp"
)

const (
	// BaseURL is the default base URL for the production Twitter API.
	// This is the default base URL if one is not given in the client.
	BaseURL = "https://api.twitter.com"
)

// NewClient returns a new client for the Twitter API.
// If cli == nil, default client options are used targeting the production API
// at BaseURL.
func NewClient(cli *jhttp.Client) *Client {
	if cli == nil {
		cli = new(jhttp.Client)
	}
	if cli.BaseURL == "" {
		cli.BaseURL = BaseURL
	}
	return (*Client)(cli)
}

// A Client serves as a client for the Twitter API v2.
type Client jhttp.Client

// A Callback function is invoked for each reply received in a stream.  If the
// callback reports a non-nil error, the stream is terminated. If the error is
// anything other than ErrStopStreaming, it is reported to the caller.
type Callback func(*Reply) error

// Call issues the specified API request and returns the decoded reply.
// Errors from Call have concrete type *jhttp.Error.
func (c *Client) Call(ctx context.Context, req *jhttp.Request) (*Reply, error) {
	header, body, err := (*jhttp.Client)(c).Call(ctx, req)
	if err != nil {
		return nil, err
	}
	var reply Reply
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, &jhttp.Error{Data: body, Message: "decoding response body", Err: err}
	}
	reply.RateLimit = decodeRateLimits(header)
	return &reply, nil
}

// CallRaw issues the specified API request and returns the raw response body
// without decoding. Errors from CallRaw have concrete type *jhttp.Error
func (c *Client) CallRaw(ctx context.Context, req *jhttp.Request) ([]byte, error) {
	_, body, err := (*jhttp.Client)(c).Call(ctx, req)
	return body, err
}

// Stream issues the specified API request and streams results to the given
// callback. Errors from Stream have concrete type *jhttp.Error.
func (c *Client) Stream(ctx context.Context, req *jhttp.Request, f Callback) error {
	return (*jhttp.Client)(c).Stream(ctx, req, func(body []byte) error {
		var reply Reply
		if err := json.Unmarshal(body, &reply); err != nil {
			return &jhttp.Error{Data: body, Message: "decoding stream response", Err: err}
		}
		return f(&reply)
	})
}
