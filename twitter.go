// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package twitter implements a client for the Twitter API v2.  This package is
// in development and is not yet ready for production use.
//
// Usage outline
//
// The general structure of an API call is to first construct a query, then
// invoke that query with a context on a client:
//
//    cli := twitter.NewClient(&twitter.ClientOpts{
//       Authorize: twitter.BearerTokenAuthorizer(token),
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
	"net/http"

	"github.com/creachadair/twitter/jhttp"
)

const (
	// BaseURL is the default base URL for the production Twitter API.
	// This is the default base URL if one is not given in the client.
	BaseURL = "https://api.twitter.com"
)

// NewClient returns a new client for the Twitter API.
// If opts == nil, the production API endpoint is used (BaseURL).
func NewClient(opts *ClientOpts) *Client {
	if opts == nil {
		opts = new(ClientOpts)
	}
	base := opts.BaseURL
	if base == "" {
		base = BaseURL
	}
	return (*Client)(&jhttp.Client{
		HTTPClient: opts.HTTPClient,
		Authorize:  opts.Authorize,
		BaseURL:    base,
		Log:        opts.Log,
	})
}

// A Client serves as a client for the Twitter API v2.
type Client jhttp.Client

// ClientOpts provide settings for a client. A nil *ClientOpts provides default
// values for the production API.
type ClientOpts struct {
	// The HTTP client used to issue requests to the API.
	// If nil, use http.DefaultClient.
	HTTPClient *http.Client

	// If set, this is called prior to issuing the request to the API.  If it
	// reports an error, the request is aborted and the error is returned to the
	// caller.
	Authorize func(*http.Request) error

	// If set, override the base URL for the API v2 endpoint.
	// This is mainly useful for testing.
	BaseURL string

	// If set, this function is called to log interesting events during the
	// transaction. See jhttp.Client for details.
	Log func(tag, message string)
}

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
// callback. Errors from Stream have concrete type *twitter.Error.
func (c *Client) Stream(ctx context.Context, req *jhttp.Request, f Callback) error {
	return (*jhttp.Client)(c).Stream(ctx, req, func(body []byte) error {
		var reply Reply
		if err := json.Unmarshal(body, &reply); err != nil {
			return &Error{Data: body, Message: "decoding stream response", Err: err}
		}
		return f(&reply)
	})
}
