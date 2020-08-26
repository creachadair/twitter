// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package twitter_test

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
	"github.com/creachadair/twitter/users"
)

var (
	doManual     = flag.Bool("manual", false, "Run manual tests that query the network")
	doVerboseLog = flag.Bool("verbose-log", false, "Enable verbose client logging")
)

func newClient(t *testing.T) *twitter.Client {
	return &twitter.Client{
		Authorize: checkAuth(t),
		Log: func(tag, msg string) {
			if tag == "RequestURL" || *doVerboseLog {
				t.Logf("API %s :: %s", tag, msg)
			}
		},
	}
}

func checkAuth(t *testing.T) twitter.Authorizer {
	t.Helper()
	bearerToken := os.Getenv("TWITTER_TOKEN")
	if bearerToken == "" {
		t.Skip("No TWITTER_TOKEN found in the environment; test cannot run")
	}
	return twitter.BearerTokenAuthorizer(bearerToken)
}

func checkManual(t *testing.T) {
	t.Helper()
	if !*doManual {
		t.Skip("Skipping manual test because -manual=false")
	}
}

// Verify that the direct call plumbing works.
func TestClientCall(t *testing.T) {
	checkManual(t)
	cli := newClient(t)

	rsp, err := cli.Call(context.Background(), &twitter.Request{
		Method: "/users/by/username/jack",
		Params: twitter.Params{
			types.UserFields: []string{
				types.User_CreatedAt,
				types.User_Description,
				types.User_PublicMetrics,
				types.User_Verified,
			},
			types.Expansions: []string{
				types.ExpandPinnedTweetID,
			},
		},
	})
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	t.Logf("Rate limits: %+v", rsp.RateLimit)
	t.Logf("Reply: %s", string(rsp.Data))
	tweets, err := rsp.IncludedTweets()
	if err != nil {
		t.Fatalf("Decoding included tweets: %v", err)
	}
	for i, tweet := range tweets {
		t.Logf("Tweet [%d]: id=%s", i+1, tweet.ID)
	}
}

func TestTweetLookup(t *testing.T) {
	checkManual(t)
	cli := newClient(t)

	ctx := context.Background()
	rsp, err := tweets.Lookup("1297524288245895168", &tweets.LookupOpts{
		TweetFields: []string{
			types.Tweet_CreatedAt,
			types.Tweet_Entities,
			types.Tweet_AuthorID,
		},
		Expansions: []string{types.ExpandMentionUsername},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	t.Logf("Lookup request returned %d bytes", len(rsp.Reply.Data))

	for i, v := range rsp.Tweets {
		t.Logf("Tweet %d: id=%s, author=%s", i+1, v.ID, v.AuthorID)
	}
	ius, err := rsp.IncludedUsers()
	if err != nil {
		t.Fatalf("Decoding included users: %v", err)
	}
	for i, v := range ius {
		t.Logf("Included User %d: id=%s, username=%q, name=%q", i+1, v.ID, v.Username, v.Name)
	}
}

func TestUserIDLookup(t *testing.T) {
	checkManual(t)
	cli := newClient(t)

	ctx := context.Background()
	rsp, err := users.Lookup("12", nil).Invoke(ctx, cli) // @jack
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	t.Logf("Lookup request returned %d bytes", len(rsp.Reply.Data))

	for i, v := range rsp.Users {
		t.Logf("User %d: id=%s, username=%q, name=%q", i+1, v.ID, v.Username, v.Name)
	}
}

func TestUsernameLookup(t *testing.T) {
	checkManual(t)
	cli := newClient(t)

	ctx := context.Background()
	rsp, err := users.LookupByName("creachadair", &users.LookupOpts{
		Keys: []string{"jack", "inlieuoffunshow"},
		UserFields: []string{
			types.User_PinnedTweetID,
			types.User_ProfileImageURL,
			types.User_PublicMetrics,
		},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	t.Logf("Lookup request returned %d bytes", len(rsp.Reply.Data))

	for i, v := range rsp.Users {
		t.Logf("User %d: id=%s, username=%q, name=%q", i+1, v.ID, v.Username, v.Name)
		t.Logf("User %d public metrics: %+v", i+1, v.PublicMetrics)
	}
}

func TestSearchRecent(t *testing.T) {
	checkManual(t)
	cli := newClient(t)

	ctx := context.Background()
	const query = `from:benjaminwittes "Today on @inlieuoffunshow"`
	rsp, err := tweets.SearchRecent(query, &tweets.SearchOpts{
		MaxResults:  10,
		StartTime:   time.Now().Add(-24 * time.Hour),
		TweetFields: []string{types.Tweet_AuthorID, types.Tweet_Entities},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("SearchRecent failed: %v", err)
	}
	var meta tweets.SearchMeta
	if err := json.Unmarshal(rsp.Meta, &meta); err == nil {
		t.Logf("Response metadata: count=%d, oldest=%s, newest=%s",
			meta.ResultCount, meta.OldestID, meta.NewestID)
	}

	if len(rsp.Tweets) == 0 {
		t.Fatal("No matching results")
	}
	for i, tw := range rsp.Tweets {
		t.Logf("Match %d: id=%s, author=%s, text=%q", i+1, tw.ID, tw.AuthorID, tw.Text)
		for j, u := range tw.Entities.URLs {
			t.Logf("-- URL %d: (%d..%d) %s title=%q", j+1, u.Start, u.End, u.Expanded, u.Title)
		}
	}
}

func TestStream(t *testing.T) {
	checkManual(t)
	cli := newClient(t)
	req := &twitter.Request{
		Method: "/tweets/sample/stream",
		Params: twitter.Params{
			types.TweetFields: []string{
				types.Tweet_AuthorID,
				types.Tweet_Entities,
			},
		},
	}
	ctx := context.Background()
	const maxResults = 5

	nr := 0
	err := cli.Stream(ctx, req, func(rsp *twitter.Reply) error {
		nr++
		t.Logf("Msg %d: %s", nr, string(rsp.Data))
		if nr == maxResults {
			return twitter.ErrStopStreaming
		}
		return nil
	})
	if err != nil {
		t.Errorf("Error from Stream: %v", err)
	}
}
