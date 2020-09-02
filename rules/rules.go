// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package rules implements queries for reading and modifying the
// rules used by streaming search queries.
//
// Reading Rules
//
// Use rules.Get to query for existing rules by ID. If no IDs are given, Get
// will return all available rules.
//
//   myRules := rules.Get(id1, id2, id3)
//   allRules := rules.Get()
//
// Invoke the query to fetch the rules:
//
//   rsp, err := allRules.Invoke(ctx, cli)
//
// The Rules field of the response contains the requested rules.
//
// Updating Rules
//
// Each rule update must either add or delete rules, but not both.  Use Adds to
// describe a Set of rules to add, or Deletes to identify a Set of rules to
// delete. For example:
//
//    adds := rules.Adds{
//       {Query: `cat has:images lang:en`, Tag: "cat pictures in English"},
//       {Query: `dog or puppy has:images`},
//    }
//    dels := rules.Deletes{id1, id2}
//
// Once you have a set, you can build a query to Update or Validate.  Update
// applies the rule change; Validate just reports whether the update would have
// succeeded (this corresponds to the "dry_run" parameter in the API):
//
//    apply := rules.Update(adds)
//    check := rules.Validate(dels)
//
// Invoke the query to execute the change or check:
//
//    rsp, err := apply.Invoke(ctx, cli)
//
// The response will include the updated rules, along with server metadata
// indicating the effective time of application and summary statistics.
//
package rules

import (
	"context"
	"encoding/json"
	"time"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/jhttp"
)

// Get constructs a query to fetch the specified streaming search rule IDs.  If
// no rule IDs are given, all known rules are fetched.
//
// API: GET 2/tweets/search/stream/rules
func Get(ids ...string) Query {
	req := &jhttp.Request{Method: "2/tweets/search/stream/rules"}
	if len(ids) != 0 {
		req.Params = jhttp.Params{"ids": ids}
	}
	return Query{request: req}
}

// Update constructs a query to add or delete streaming search rules.
//
// API: POST 2/tweets/search/stream/rules
func Update(r Set) Query {
	enc, err := r.encode()
	req := &jhttp.Request{
		Method:     "2/tweets/search/stream/rules",
		HTTPMethod: "POST",
		Data:       enc,
	}
	return Query{request: req, encodeErr: err}
}

// Validate constructs a query to validate addition or deletion of streaming
// search rules, without actually modifying the rules.
//
// API: POST 2/tweets/search/stream/rules, dry_run=true
func Validate(r Set) Query {
	enc, err := r.encode()
	req := &jhttp.Request{
		Method:     "2/tweets/search/stream/rules",
		HTTPMethod: "POST",
		Params:     jhttp.Params{"dry_run": []string{"true"}},
		Data:       enc,
	}
	return Query{request: req, encodeErr: err}
}

// A Query performs a rule fetch or update query.
type Query struct {
	request   *jhttp.Request
	encodeErr error // an error from encoding the rules
}

// Invoke executes the query on the given context and client.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	// Report a deferred error from encoding.
	if q.encodeErr != nil {
		return nil, &jhttp.Error{Message: "encoding rule set", Err: q.encodeErr}
	}
	rsp, err := cli.Call(ctx, q.request)
	if err != nil {
		return nil, err
	}
	out := new(Reply)
	if len(rsp.Data) == 0 {
		// no rules returned
	} else if err := json.Unmarshal(rsp.Data, &out.Rules); err != nil {
		return nil, &jhttp.Error{Data: rsp.Data, Message: "decoding rules data", Err: err}
	}
	if err := json.Unmarshal(rsp.Meta, &out.Meta); err != nil {
		return nil, &jhttp.Error{Data: rsp.Meta, Message: "decoding rules metadata", Err: err}
	}
	return out, nil
}

// A Rule encodes a single streaming search rule.
type Rule struct {
	ID    string `json:"id,omitempty"`
	Value string `json:"value"`
	Tag   string `json:"tag,omitempty"`
}

// A Reply is the response from a Query.
type Reply struct {
	*twitter.Reply
	Rules []Rule
	Meta  *Meta
}

// Meta records rule set metadata reported by the service.
type Meta struct {
	Sent    time.Time `json:"sent"`
	Summary struct {
		Created    int `json:"created"`
		NotCreated int `json:"not_created"`
		Deleted    int `json:"deleted"`
		NotDeleted int `json:"not_deleted"`
		Valid      int `json:"valid"`
		Invalid    int `json:"invalid"`
	} `json:"summary,omitempty"`
}

// A Set encodes a set of rule additions or deletions.
type Set interface {
	encode() ([]byte, error)
}

// Add gives a query and optional tag to define a rule.
type Add struct {
	Query string
	Tag   string
}

// Adds is a Set of search rules to be added.
type Adds []Add

func (as Adds) encode() ([]byte, error) {
	rules := make([]Rule, len(as))
	for i, a := range as {
		rules[i] = Rule{Value: a.Query, Tag: a.Tag}
	}
	return json.Marshal(struct {
		A []Rule `json:"add"`
	}{A: rules})
}

// Deletes is a Set of search rule IDs to be deleted.
type Deletes []string

func (ds Deletes) encode() ([]byte, error) {
	type del struct {
		I []string `json:"ids"`
	}
	return json.Marshal(struct {
		D del `json:"delete"`
	}{
		D: del{I: []string(ds)},
	})
}
