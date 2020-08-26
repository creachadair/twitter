// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tweets

import (
	"context"
	"encoding/json"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// SampleStream constructs a streaming sample query that delivers results to f.
func SampleStream(f Callback, opts *StreamOpts) Stream {
	req := &twitter.Request{
		Method: "tweets/sample/stream",
		Params: make(twitter.Params),
	}
	opts.addRequestParams(req)
	return Stream{callback: f, request: req}
}

// SearchStream constructs a streaming search query that delivers results to f.
func SearchStream(f Callback, opts *StreamOpts) Stream {
	req := &twitter.Request{
		Method: "tweets/search/stream",
		Params: make(twitter.Params),
	}
	opts.addRequestParams(req)
	return Stream{callback: f, request: req}
}

// A Stream performs a streaming search or sampling query.
type Stream struct {
	callback Callback
	request  *twitter.Request
}

// StreamOpts provides parameters for tweet streaming. A nil *StreamOpts
// provides empty values for all fields.
type StreamOpts struct {
	Expansions  []string
	MediaFields []string
	PlaceFields []string
	PollFields  []string
	TweetFields []string
	UserFields  []string
}

func (o *StreamOpts) addRequestParams(req *twitter.Request) {
	if o == nil {
		return // nothing to do
	}
	req.Params.Add(types.Expansions, o.Expansions...)
	req.Params.Add(types.MediaFields, o.MediaFields...)
	req.Params.Add(types.PlaceFields, o.PlaceFields...)
	req.Params.Add(types.PollFields, o.PollFields...)
	req.Params.Add(types.TweetFields, o.TweetFields...)
	req.Params.Add(types.UserFields, o.UserFields...)
}

// A Callback receives streaming replies from a sample or streaming search
// query. If the callback returns an error, the stream is terminated. If the
// error is not twitter.ErrStopStreaming, that error is reported to the caller.
type Callback func(*Reply) error

// Invoke executes the streaming query on the given context and client.
func (s Stream) Invoke(ctx context.Context, cli *twitter.Client) error {
	return cli.Stream(ctx, s.request, func(rsp *twitter.Reply) error {
		var tweet types.Tweet
		if err := json.Unmarshal(rsp.Data, &tweet); err != nil {
			return twitter.Errorf(rsp.Data, "decoding tweet data", err)
		}
		return s.callback(&Reply{
			Reply:  rsp,
			Tweets: types.Tweets{&tweet},
		})
	})
}
