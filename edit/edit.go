// Copyright (C) 2022 Michael J. Fromberger. All Rights Reserved.

// Package edit implements editing operations on tweets and tweet metadata.
package edit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
)

// DeleteTweet constructs a query to delete the given tweet ID.
//
// API: DELETE 2/tweets/:tid
func DeleteTweet(tweetID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/tweets/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "deleted",
	}
}

// A Query is a query to modify the contents or properties of tweets.
type Query struct {
	*jhttp.Request
	tag       string
	encodeErr error
}

// Invoke executes the query on the given context and client. A successful
// response reports whether the edit took effect.
func (e Query) Invoke(ctx context.Context, cli *twitter.Client) (bool, error) {
	if e.encodeErr != nil {
		return false, e.encodeErr // deferred encoding error
	}
	rsp, err := cli.Call(ctx, e.Request)
	if err != nil {
		return false, err
	}
	m := make(map[string]*bool)
	if err := json.Unmarshal(rsp.Data, &m); err != nil {
		return false, &jhttp.Error{Data: rsp.Data, Message: "decoding response", Err: err}
	}
	if v := m[e.tag]; v != nil {
		return *v, nil
	}
	return false, fmt.Errorf("tag %q not found", e.tag)
}

// SetHidden constructs a query to set whether replies to the given tweet ID
// should (hidden == true) or should not (hidden == false) be hidden.
func SetHidden(tweetID string, hidden bool) Query {
	body, err := json.Marshal(struct {
		H bool `json:"hidden"`
	}{H: hidden})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/tweets/" + tweetID + "/hidden",
			HTTPMethod:  "PUT",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "hidden",
		encodeErr: err,
	}
}

// Like constructs a query for the given user ID to like the given tweet ID.
//
// API: POST 2/users/:id/likes
func Like(userID, tweetID string) Query {
	body, err := json.Marshal(struct {
		ID string `json:"tweet_id"`
	}{ID: tweetID})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/likes",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "liked",
		encodeErr: err,
	}
}

// Unlike constructs a query for the given user ID to un-like the given tweet ID.
//
// API: DELETE 2/users/:id/likes/:tid
func Unlike(userID, tweetID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/likes/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "liked",
	}
}

// Bookmark constructs a query for the given user ID to bookmark the given
// tweet ID.
//
// API: 2/users/:id/bookmarks
func Bookmark(userID, tweetID string) Query {
	body, err := json.Marshal(struct {
		ID string `json:"tweet_id"`
	}{ID: tweetID})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/bookmarks",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "bookmarked",
		encodeErr: err,
	}
}

// Unbookmark constructs a query for the given user ID to un-like the given
// tweet ID.
//
// API: DELETE 2/users/:id/bookmarks/:tid
func Unbookmark(userID, tweetID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/bookmarks/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "bookmarked",
	}
}

// Retweet constructs a query for the given user ID to retweet the given tweet ID.
//
// API: POST 2/users/:id/retweets
func Retweet(userID, tweetID string) Query {
	body, err := json.Marshal(struct {
		ID string `json:"tweet_id"`
	}{ID: tweetID})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/retweets",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "retweeted",
		encodeErr: err,
	}
}

// Unretweet constructs a query for the given user ID to un-retweet the given
// tweet ID.
//
// API: DELETE 2/users/:id/retweets/:tid
func Unretweet(userID, tweetID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/retweets/" + tweetID,
			HTTPMethod: "DELETE",
		},
		tag: "retweeted",
	}
}

// Block constructs a query for one user ID to block another user ID.
//
// API: POST 2/users/:id/blocking
func Block(userID, blockeeID string) Query {
	body, err := json.Marshal(struct {
		ID string `json:"target_user_id"`
	}{ID: blockeeID})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/blocking",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "blocking",
		encodeErr: err,
	}
}

// Unblock constructs a query for one user ID to un-block another user ID.
//
// API: DELETE 2/users/:id/blocking/:other
func Unblock(userID, blockeeID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/blocking/" + blockeeID,
			HTTPMethod: "DELETE",
		},
		tag: "blocking",
	}
}

// Follow constructs a query for one user ID to follow another user ID.
//
// API: POST 2/users/:id/following
func Follow(userID, followeeID string) Query {
	body, err := json.Marshal(struct {
		ID string `json:"target_user_id"`
	}{ID: followeeID})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/following",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "following",
		encodeErr: err,

		// TODO(creachadair): Do something about the pending status for target
		// users who are protected.
	}
}

// Unfollow constructs a query for one user ID to un-follow another user ID.
//
// API: DELETE 2/users/:id/following/:other
func Unfollow(userID, followeeID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/following/" + followeeID,
			HTTPMethod: "DELETE",
		},
		tag: "following",
	}
}

// Mute constructs a query for one user ID to mute another user ID.
//
// API: POST 2/users/:id/muting
func Mute(userID, muteeID string) Query {
	body, err := json.Marshal(struct {
		ID string `json:"target_user_id"`
	}{ID: muteeID})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/muting",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "muting",
		encodeErr: err,
	}
}

// Unmute constructs a query for one user ID to un-mute another user ID.
//
// API: DELETE 2/users/:id/muting/:other
func Unmute(userID, muteeID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/muting/" + muteeID,
			HTTPMethod: "DELETE",
		},
		tag: "muting",
	}
}

// PinList constructs a query for one user ID to pin a list ID.
//
// API: POST 2/users/:id/pinned_lists
func PinList(userID, listID string) Query {
	body, err := json.Marshal(struct {
		ID string `json:"list_id"`
	}{ID: listID})
	return Query{
		Request: &jhttp.Request{
			Method:      "2/users/" + userID + "/pinned_lists",
			HTTPMethod:  "POST",
			ContentType: "application/json",
			Data:        body,
		},
		tag:       "pinned",
		encodeErr: err,
	}
}

// UnpinList constructs a query for one user ID to un-pin a list ID.
//
// API: DELETE 2/users/:id/pinned_lists/:lid
func UnpinLists(userID, listID string) Query {
	return Query{
		Request: &jhttp.Request{
			Method:     "2/users/" + userID + "/pinned_lists/" + listID,
			HTTPMethod: "DELETE",
		},
		tag: "pinned",
	}
}
