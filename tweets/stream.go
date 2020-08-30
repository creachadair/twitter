// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tweets

import (
	"context"
	"encoding/json"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// SampleStream constructs a streaming sample query that delivers results to f.
//
// API: tweets/sample/stream
func SampleStream(f Callback, opts *StreamOpts) Stream {
	req := &twitter.Request{
		Method: "tweets/sample/stream",
		Params: make(twitter.Params),
	}
	opts.addRequestParams(req)
	return Stream{Request: req, callback: f, maxResults: opts.maxResults()}
}

// SearchStream constructs a streaming search query that delivers results to f.
//
// API: tweets/search/stream
func SearchStream(f Callback, opts *StreamOpts) Stream {
	req := &twitter.Request{
		Method: "tweets/search/stream",
		Params: make(twitter.Params),
	}
	opts.addRequestParams(req)
	return Stream{Request: req, callback: f, maxResults: opts.maxResults()}
}

// A Stream performs a streaming search or sampling query.
type Stream struct {
	*twitter.Request
	callback   Callback
	maxResults int
}

// StreamOpts provides parameters for tweet streaming. A nil *StreamOpts
// provides empty values for all fields.
type StreamOpts struct {
	// If positive, stop streaming after this many results have been reported.
	MaxResults int

	// Optional response fields and expansions.
	Optional []types.Fields
}

func (o *StreamOpts) addRequestParams(req *twitter.Request) {
	if o == nil {
		return // nothing to do
	}
	for _, fs := range o.Optional {
		if vs := fs.Values(); len(vs) != 0 {
			req.Params.Add(fs.Label(), vs...)
		}
	}
}

func (o *StreamOpts) maxResults() int {
	if o == nil {
		return 0
	}
	return o.MaxResults
}

// A Callback receives streaming replies from a sample or streaming search
// query. If the callback returns an error, the stream is terminated. If the
// error is not twitter.ErrStopStreaming, that error is reported to the caller.
type Callback func(*Reply) error

// Invoke executes the streaming query on the given context and client.
func (s Stream) Invoke(ctx context.Context, cli *twitter.Client) error {
	var nr int
	return cli.Stream(ctx, s.Request, func(rsp *twitter.Reply) error {
		nr++
		var tweet types.Tweet
		if err := json.Unmarshal(rsp.Data, &tweet); err != nil {
			return &twitter.Error{Data: rsp.Data, Message: "decoding tweet data", Err: err}
		}
		if err := s.callback(&Reply{
			Reply:  rsp,
			Tweets: types.Tweets{&tweet},
		}); err != nil {
			return err
		} else if s.maxResults > 0 && nr == s.maxResults {
			return twitter.ErrStopStreaming
		}
		return nil
	})
}
