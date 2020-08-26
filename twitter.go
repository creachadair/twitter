// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package twitter implements a client for the Twitter API v2.  This package is
// in development and is not yet ready for production use.
//
// Usage outline:
//
//    cli := &twitter.Client{
//       Authorizer: twitter.NewBearerTokenAuthorizer(token),
//    }
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
// Limitations
//
// Currently the lookup APIs for tweets and users are supported, as well as the
// search API for recent tweets.
//
// Sampling and streaming are not yet supported.
//
package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// BaseURL is the default base URL for production Twitter API v2.
// This is the default base URL if one is not given in the client.
const BaseURL = "https://api.twitter.com/2"

// A Client serves as a client for the Twitter API v2.
type Client struct {
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
	// transaction.
	//
	// Tags include:
	//
	//    RequestURL   -- the request URL sent to the server
	//    HTTPStatus   -- the HTTP status string (e.g., "200 OK")
	//    ResponseBody -- the body of the response sent by the server
	//    StreamBody   -- the body of a stream response from the server
	//
	Log func(tag, message string)
}

func (c *Client) baseURL() (*url.URL, error) {
	if c.BaseURL != "" {
		return url.Parse(c.BaseURL)
	}
	return url.Parse(BaseURL)
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) log(tag, message string) {
	if c.Log != nil {
		c.Log(tag, message)
	}
}

func (c *Client) hasLog() bool { return c.Log != nil }

// start issues the specified API request and returns its HTTP response.  The
// caller is responsible for interpreting any errors or unexpected status codes
// from the request.
func (c *Client) start(ctx context.Context, req *Request) (*http.Response, error) {
	u, err := c.baseURL()
	if err != nil {
		return nil, Errorf(nil, "invalid base URL", err)
	}
	u.Path = path.Join(u.Path, req.Method)
	req.addQueryTerms(u)
	requestURL := u.String()
	c.log("RequestURL", requestURL)

	hreq, err := http.NewRequestWithContext(ctx, req.HTTPMethod, requestURL, req.Data)
	if err != nil {
		return nil, Errorf(nil, "invalid request", err)
	}
	if req.ContentType != "" {
		hreq.Header.Set("Content-Type", req.ContentType)
	}

	if auth := c.Authorize; auth != nil {
		if err := auth(hreq); err != nil {
			return nil, Errorf(nil, "attaching authorization", err)
		}
	}

	rsp, err := c.httpClient().Do(hreq)
	if err != nil {
		return nil, Errorf(nil, "issuing request", err)
	}
	return rsp, nil
}

// ErrStopStreaming is a sentinel error that a stream callback can use to
// signal it does not want any further results.
var ErrStopStreaming = errors.New("stop streaming")

// A Callback function is invoked for each reply received in a stream.  If the
// callback reports a non-nil error, the stream is terminated. If the error is
// anything other than ErrStopStreaming, it is reported to the caller.
type Callback func(*Reply) error

// finish cleans up and decodes a successful (non-nil) HTTP response returned
// by a call to start.
func (c *Client) finish(rsp *http.Response) (*Reply, error) {
	if rsp == nil { // safety check
		panic("cannot finish a nil *http.Response")
	}

	// The body must be fully read and closed to avoid orphaning resources.
	// See: https://godoc.org/net/http#Do
	var body bytes.Buffer
	io.Copy(&body, rsp.Body)
	rsp.Body.Close()
	c.log("HTTPStatus", rsp.Status)
	if c.hasLog() {
		c.log("ResponseBody", body.String())
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, newErrorf(nil, rsp.StatusCode, body.Bytes(), "request failed: %s", rsp.Status)
	}

	var reply Reply
	if err := json.Unmarshal(body.Bytes(), &reply); err != nil {
		return nil, Errorf(body.Bytes(), "decoding response body", err)
	}
	reply.RateLimit = decodeRateLimits(rsp.Header)
	return &reply, nil
}

// Call issues the specified API request and returns the decoded reply.
// Errors from Call have concrete type *twitter.Error.
func (c *Client) Call(ctx context.Context, req *Request) (*Reply, error) {
	hrsp, err := c.start(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.finish(hrsp)
}

// stream streams results from a successful (non-nil) HTTP response returned by
// a call to start. Results are delivered to the given callback until the
// stream ends, ctx ends, or the callback reports a non-nil error.  The error
// from the callback is propagated to the caller of stream.
func (c *Client) stream(ctx context.Context, rsp *http.Response, f Callback) error {
	if rsp == nil { // safety check
		panic("cannot stream a nil *http.Response")
	}
	body := rsp.Body
	defer body.Close()

	c.log("HTTPStatus", rsp.Status)
	if rsp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(body)
		if c.hasLog() {
			c.log("ResponseBody", string(data))
		}
		return newErrorf(nil, rsp.StatusCode, data, "request failed: %s", rsp.Status)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// When ctx ends, close the response body to unblock the reader.
	go func() {
		<-ctx.Done()
		body.Close()
	}()

	dec := json.NewDecoder(body)
	for {
		var next json.RawMessage
		if err := dec.Decode(&next); err == io.EOF {
			break
		} else if err != nil {
			return Errorf(nil, "decoding message from stream", err)
		}
		if c.hasLog() {
			c.log("StreamBody", string(next))
		}
		var reply Reply
		if err := json.Unmarshal(next, &reply); err != nil {
			return Errorf(next, "decoding stream response", err)
		} else if err := f(&reply); err != nil {
			return Errorf(nil, "callback", err)
		}
	}
	return nil
}

// Stream issues the specified API request and streams results to the given
// callback. Errors from Stream have concrete type *twitter.Error.
func (c *Client) Stream(ctx context.Context, req *Request, f Callback) error {
	hrsp, err := c.start(ctx, req)
	if err != nil {
		return err
	}
	if err := c.stream(ctx, hrsp, f); errors.Is(err, ErrStopStreaming) {
		return nil // the callback requested a stop
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

// An Authorizer attaches authorization metadata to an outbound request after
// it has been populated with the caller's query but before it is sent to the
// API.  The function modifies the request in-place as needed.
type Authorizer func(*http.Request) error

// BearerTokenAuthorizer returns an authorizer that injects the specified
// bearer token into the Authorization header of each request.
func BearerTokenAuthorizer(token string) Authorizer {
	authValue := "Bearer " + token
	return func(req *http.Request) error {
		req.Header.Add("Authorization", authValue)
		return nil
	}
}

// A Request is the generic format for a Twitter API v2 request.
type Request struct {
	// The fully-expanded method path for the API to call, including parameters.
	// For example: "tweets/12345678".
	Method string

	// Additional request parameters, including optional fields and expansions.
	Params Params

	// The HTTP method to use for the request; if unset the default is "GET".
	HTTPMethod string

	// If set, send these data as the body of the request.
	Data io.Reader

	// If set, use this as the content-type for the request body.
	ContentType string
}

// Params carries additional request parameters sent in the query URL.
type Params map[string][]string

// Add the given values for the specified parameter, in addition to any
// previously-defined values for that name.
func (p Params) Add(name string, values ...string) {
	if len(values) == 0 {
		return
	}
	p[name] = append(p[name], values...)
}

// Set sets the value of the specified parameter name, removing any
// previously-defined values for that name.
func (p Params) Set(name, value string) { p[name] = []string{value} }

// Reset removes any existing values for the specified parameter.
func (p Params) Reset(name string) { delete(p, name) }

func (p Params) addQueryTerms(query url.Values) {
	for name, values := range p {
		query.Set(name, strings.Join(values, ","))
	}
}

func (req *Request) addQueryTerms(u *url.URL) {
	if len(req.Params) == 0 {
		return // nothing to do
	}
	query := make(url.Values)
	req.Params.addQueryTerms(query)
	u.RawQuery = query.Encode()
}
