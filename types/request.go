// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package types

import (
	"bytes"
	"io"
	"net/url"
	"path"
	"strings"
)

// DefaultContentType is the default content-type reported for a request body.
const DefaultContentType = "application/json"

// A Request is the generic format for a request.
type Request struct {
	// The fully-expanded method path for the API to call, including parameters.
	// For example: "2/tweets/12345678".
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

// Encode encodes p as a query string.
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
