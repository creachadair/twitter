// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package rules implements queries for reading and modifying the
// rules used by streaming search queries.
package rules

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// Get constructs a query to fetch the specified streaming search rule IDs.  If
// no rule IDs are given, all known rules are fetched.
func Get(ids []string) Query {
	req := &twitter.Request{Method: "tweets/search/stream/rules"}
	if len(ids) != 0 {
		req.Params = twitter.Params{"ids": ids}
	}
	return Query{request: req}
}

// Update constructs a query to add and/or delete streaming search rules.
func Update(r Set) Query {
	req := &twitter.Request{
		Method:      "tweets/search/stream/rules",
		HTTPMethod:  "POST",
		Data:        bytes.NewReader(r.encoded),
		ContentType: "application/json",
	}
	return Query{request: req}
}

// Validate constructs a query to validate addition and/or deletion of
// streaming search rules, without actually modifying the rules.
func Validate(r Set) Query {
	req := &twitter.Request{
		Method:      "tweets/search/stream/rules",
		HTTPMethod:  "POST",
		Params:      twitter.Params{"dry_run": []string{"true"}},
		Data:        bytes.NewReader(r.encoded),
		ContentType: "application/json",
	}
	return Query{request: req}
}

// A Query performs a rule fetch or update query.
type Query struct {
	request *twitter.Request
}

// Invoke executes the query on the given context and client.
func (q Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	rsp, err := cli.Call(ctx, q.request)
	if err != nil {
		return nil, err
	}
	out := new(Reply)
	if len(rsp.Data) == 0 {
		// no rules returned
	} else if err := json.Unmarshal(rsp.Data, &out.Rules); err != nil {
		return nil, twitter.Errorf(rsp.Data, "decoding rules data", err)
	}
	if err := json.Unmarshal(rsp.Meta, &out.Meta); err != nil {
		return nil, twitter.Errorf(rsp.Meta, "decoding rules metadata", err)
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
	Sent    *types.Date `json:"sent"`
	Summary struct {
		Created    int `json:"created"`
		NotCreated int `json:"not_created"`
		Deleted    int `json:"deleted"`
		NotDeleted int `json:"not_deleted"`
	} `json:"summary,omitempty"`
}

// A Set encodes a set of rule additions and/or deletions.
type Set struct {
	encoded []byte
}

// Add constructs a set of add rules.
func Add(rules ...Rule) (Set, error) {
	enc, err := json.Marshal(struct {
		A []Rule `json:"add"`
	}{A: rules})
	if err != nil {
		return Set{}, err
	}
	return Set{encoded: enc}, nil
}

// Delete constructs a set of delete rules.
func Delete(ids ...string) (Set, error) {
	type del struct {
		I []string `json:"ids"`
	}
	enc, err := json.Marshal(struct {
		D del `json:"delete"`
	}{
		D: del{I: ids},
	})
	if err != nil {
		return Set{}, err
	}
	return Set{encoded: enc}, nil
}
