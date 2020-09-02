// Package ostatus implements queries that operate on statuses (tweets)
// using the Twitter API v1.1.
package ostatus

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

// Update constructs a status update ("tweet") with the given text.
// This query requires user-context authorization.
//
// API: 1.1/statuses/update.json
func Update(text string, opts *UpdateOpts) Query {
	q := Query{
		Request: &types.Request{
			Method:     "1.1/statuses/update.json",
			HTTPMethod: "POST",
			Params: types.Params{
				"status":    []string{text},
				"trim_user": []string{"true"},
			},
		},
	}
	opts.addQueryParams(&q)
	return q
}

func modQuery(path, id string, opts *Options) Query {
	q := Query{
		Request: &types.Request{
			Method:     path + "/" + id + ".json", // N.B. parameter in path
			HTTPMethod: "POST",
			Params:     types.Params{"trim_user": []string{"true"}},
		},
	}
	opts.addQueryParams(&q)
	return q
}

// Delete constructs a query to delete ("destroy") a tweet with the given ID.
// This query requires user-context authorization.
//
// API: 1.1/statuses/destroy/:id.json
func Delete(id string, opts *Options) Query {
	return modQuery("1.1/statuses/destroy", id, opts)
}

// Retweet constructs a query to retweet a tweet with the given ID.
// This query requires user-context authorization.
//
// API: 1.1/statuses/retweet/:id.json
func Retweet(id string, opts *Options) Query {
	return modQuery("1.1/statuses/retweet", id, opts)
}

// UnRetweet constructs a query to un-retweet a tweet with the given ID.
// This query requires user-context authorization.
//
// API: 1.1/statuses/unretweet/:id.json
func Unretweet(id string, opts *Options) Query {
	return modQuery("1.1/statuses/unretweet", id, opts)
}

func likeQuery(path, id string, opts *Options) Query {
	q := Query{
		Request: &types.Request{
			Method:     path + ".json",
			HTTPMethod: "POST",
			Params:     types.Params{"id": []string{id}},
		},
	}
	opts.addQueryParams(&q)
	q.Request.Params.Set("include_entities", strconv.FormatBool(q.opts.Entities))
	return q
}

// Like constructs a query to like ("favorite") a tweet with the given ID.
// This query requires user-context authorization.
//
// API: 1.1/favorites/create.json
func Like(id string, opts *Options) Query {
	return likeQuery("1.1/favorites/create", id, opts)
}

// UnLike constructs a query to un-like ("unfavorite") a tweet with the given ID.
// This query requires user-context authorization.
//
// API: 1.1/favorites/destroy.json
func UnLike(id string, opts *Options) Query {
	return likeQuery("1.1/favorites/destroy", id, opts)
}

// Query is the query to post a status update.
type Query struct {
	*types.Request
	opts types.TweetFields
}

// Invoke posts the update and reports the resulting tweet.
func (o Query) Invoke(ctx context.Context, cli *twitter.Client) (*Reply, error) {
	data, err := cli.CallRaw(ctx, o.Request)
	if err != nil {
		return nil, err
	}
	var rsp oldTweet
	if err := json.Unmarshal(data, &rsp); err != nil {
		return nil, &twitter.Error{Message: "decoding response body", Err: err}
	}
	return &Reply{
		Data:  data,
		Tweet: rsp.toNewTweet(o.opts),
	}, nil
}

// UpdateOpts provides parameters for tweet creation. A nil *UpdateOpts
// provides zero values for all fields.
type UpdateOpts struct {
	// Record the update as a reply to this tweet ID.  This will be ignored
	// unless the update text includes an @mention of the author of that tweet.
	InReplyTo string

	// Ask the server to automatically populate the reply target and mentions.
	AutoPopulateReply bool

	// User IDs to exclude when auto-populating mentions.
	AutoExcludeMentions []string

	// Optional tweet fields to report with a successful update.
	Optional types.TweetFields
}

func (o *UpdateOpts) addQueryParams(q *Query) {
	if o != nil {
		if o.InReplyTo != "" {
			q.Request.Params.Set("in_reply_to_status_id", o.InReplyTo)
		}
		if o.AutoPopulateReply {
			q.Request.Params.Set("auto_populate_reply_metadata", "true")
			if len(o.AutoExcludeMentions) != 0 {
				q.Request.Params.Add("exclude_reply_user_ids", o.AutoExcludeMentions...)
			}
		}
		q.opts = o.Optional
	}
	// Move parameters to the request body.
	q.Request.Data = []byte(q.Request.Params.Encode())
	q.Request.ContentType = "application/x-www-form-urlencoded"
	q.Request.Params = nil
}

// Options provides parameters for tweet modification. A nil *Options provides
// zero values for all fields.
type Options struct {
	Optional types.TweetFields
}

func (o *Options) addQueryParams(q *Query) {
	if o != nil {
		q.opts = o.Optional
	}
}

// A Reply is the response from an Query.
type Reply struct {
	Data  []byte
	Tweet *types.Tweet
}
