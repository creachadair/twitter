// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package jhttp implements a client for a JSON-based HTTP API.
//
// Usage outline
//
//    cli := &jhttp.Client{
//       BaseURL:   "https://api.whatevs.com/v2",
//       Authorize: jhttp.BearerTokenAuthorizer(token),
//    }
//
//    ctx := context.Background()
//    headers, body, err := cli.Call(ctx, &client.Request{
//       Method: "service/method",
//       Params: client.Params{
//          "ids": []string{"a", "b", "c"},
//       },
//       Data:        []byte("fly you fools"),
//       ContentType: "text/plain",
//    })
//    if err != nil {
//       log.Fatalf("Request failed: %v", err)
//    }
//    process(headers, body)
//
package jhttp

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

const (
	// DefaultContentType is the default content-type reported for a request body.
	DefaultContentType = "application/json"
)

// A Client serves as a client for an JSON-based HTTP API.
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
	requestURL, err := req.URL(c.BaseURL)
	if err != nil {
		return nil, &Error{Message: "invalid request URL", Err: err}
	}
	c.log("RequestURL", requestURL)

	data, dlen, dtype := req.Body()
	hreq, err := http.NewRequestWithContext(ctx, req.HTTPMethod, requestURL, data)
	if err != nil {
		return nil, &Error{Message: "invalid request", Err: err}
	}
	hreq.ContentLength = dlen
	if dlen > 0 {
		hreq.Header.Set("Content-Type", dtype)
	}

	if auth := c.Authorize; auth != nil {
		if err := auth(hreq); err != nil {
			return nil, &Error{Message: "attaching authorization", Err: err}
		}
		if c.hasLog() {
			c.log("Authorization", hreq.Header.Get("authorization"))
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
type Callback func([]byte) error

// receive checks the status of a successful (non-nil) HTTP response returned
// by a call to start.  It returns the response headers and response body data
// on success.
func (c *Client) receive(rsp *http.Response) (http.Header, []byte, error) {
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
		return rsp.Header, nil, &Error{
			Status:  rsp.StatusCode,
			Data:    body.Bytes(),
			Message: "request failed: " + rsp.Status,
		}
	}
	return rsp.Header, body.Bytes(), nil
}

// Call issues the specified API request and returns the HTTP response headers
// and response body without decoding. Errors from Call have type *jhttp.Error.
func (c *Client) Call(ctx context.Context, req *Request) (http.Header, []byte, error) {
	hrsp, err := c.start(ctx, req)
	if err != nil {
		return nil, nil, err
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
		if err := f(next); err != nil {
			return &Error{Message: "callback", Err: err}
		}
	}
	return nil
}

// Stream issues the specified API request and streams results to the given
// callback. Errors from Stream have concrete type *jhttp.Error.
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

// A Request is the generic format for a request.
type Request struct {
	// The fully-expanded method path for the API to call, including parameters.
	// For example: "service/method/12345".
	Method string

	// Additional request parameters, including optional fields and expansions.
	Params Params

	// The HTTP method to use for the request; if unset the default is "GET".
	HTTPMethod string

	// If non-empty, send these data as the body of the request.
	Data []byte

	// If set, use this as the content-type for the request body.
	// If unset, the value defaults to DefaultContentType (JSON).
	// A content-type is only set if Data is non-empty.
	ContentType string
}

// SetBodyToParams encodes r.Params in the request body.  This replaces the
// Data and ContentType fields, and leaves r.Params set to nil.
func (r *Request) SetBodyToParams() {
	r.Data = []byte(r.Params.Encode())
	r.ContentType = "application/x-www-form-urlencoded"
	r.Params = nil
}

// URL returns the complete request URL for r, using base as the base URL.
func (r *Request) URL(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, r.Method)
	r.addQueryTerms(u)
	return u.String(), nil
}

// Body returns the size and putative content-type of the request body, along
// with a reader that will deliver its contents.
//
// If no data are set on the request, Body returns nil, 0, "".
func (r *Request) Body() (data io.Reader, size int64, ctype string) {
	if len(r.Data) == 0 {
		return nil, 0, ""
	}
	ctype = r.ContentType
	if ctype == "" {
		ctype = DefaultContentType
	}

	// N.B. Do not change the type of the reader without first reading the
	// documentation for http.Request.GetBody.
	return bytes.NewReader(r.Data), int64(len(r.Data)), ctype
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

// Encode encodes p as a query string. If len(p) == 0, Encode returns "".
func (p Params) Encode() string {
	query := make(url.Values)
	p.addQueryTerms(query)
	return query.Encode()
}

func (p Params) addQueryTerms(query url.Values) {
	for name, values := range p {
		query.Set(name, strings.Join(values, ","))
	}
}

func (req *Request) addQueryTerms(u *url.URL) {
	if len(req.Params) == 0 {
		return // nothing to do
	}
	u.RawQuery = req.Params.Encode()
}
