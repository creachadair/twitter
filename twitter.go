// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package twitter implements a client for the Twitter API v2.  This package is
// in development and is not yet ready for production use.
//
// Usage outline
//
// The general structure of an API call is to first construct a query, then
// invoke that query with a context on a client:
//
//    cli := &twitter.Client{
//       Authorizer: twitter.BearerTokenAuthorizer(token),
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
		return nil, &Error{Message: "invalid base URL", Err: err}
	}
	u.Path = path.Join(u.Path, req.Method)
	req.addQueryTerms(u)
	requestURL := u.String()
	c.log("RequestURL", requestURL)

	data, dlen := req.dataLen()
	hreq, err := http.NewRequestWithContext(ctx, req.HTTPMethod, requestURL, data)
	if err != nil {
		return nil, &Error{Message: "invalid request", Err: err}
	}
	hreq.ContentLength = dlen
	if req.ContentType != "" {
		hreq.Header.Set("Content-Type", req.ContentType)
	}

	if auth := c.Authorize; auth != nil {
		if err := auth(hreq); err != nil {
			return nil, &Error{Message: "attaching authorization", Err: err}
		}
	}

	rsp, err := c.httpClient().Do(hreq)
	if err != nil {
		return nil, &Error{Message: "issuing request", Err: err}
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
	body, err := c.receive(rsp)
	if err != nil {
		return nil, err
	}
	var reply Reply
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, &Error{Data: body, Message: "decoding response body", Err: err}
	}
	reply.RateLimit = decodeRateLimits(rsp.Header)
	return &reply, nil
}

// receive checks the status of a successful (non-nil) HTTP response returned
// by a call to start.  It returns the response body data on success.
func (c *Client) receive(rsp *http.Response) ([]byte, error) {
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
	switch rsp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		// ok
	default:
		return nil, &Error{
			Status:  rsp.StatusCode,
			Data:    body.Bytes(),
			Message: "request failed: " + rsp.Status,
		}
	}
	return body.Bytes(), nil
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

// CallRaw issues the specified API request and returns the raw response body
// without decoding. Errors from CallRaw have concrete type *twitter.Error
func (c *Client) CallRaw(ctx context.Context, req *Request) ([]byte, error) {
	hrsp, err := c.start(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.receive(hrsp)
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
		return &Error{
			Status:  rsp.StatusCode,
			Data:    data,
			Message: "request failed: " + rsp.Status,
		}
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
			return &Error{Message: "decoding message from stream", Err: err}
		}
		if c.hasLog() {
			c.log("StreamBody", string(next))
		}
		var reply Reply
		if err := json.Unmarshal(next, &reply); err != nil {
			return &Error{Data: next, Message: "decoding stream response", Err: err}
		} else if err := f(&reply); err != nil {
			return &Error{Message: "callback", Err: err}
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
		if _, ok := err.(*Error); ok {
			return err
		}
		return &Error{Message: "callback", Err: err}
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

	// If non-empty, send these data as the body of the request.
	Data []byte

	// If set, use this as the content-type for the request body.
	ContentType string
}

func (r *Request) dataLen() (io.Reader, int64) {
	if len(r.Data) == 0 {
		return nil, 0
	}
	return bytes.NewReader(r.Data), int64(len(r.Data))
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
