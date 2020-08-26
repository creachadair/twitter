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
	"github.com/creachadair/twitter/rules"
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
		Method: "users/by/username/jack",
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
		Method: "tweets/sample/stream",
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

func TestRules(t *testing.T) {
	checkManual(t)
	cli := newClient(t)
	ctx := context.Background()
	logResponse := func(t *testing.T, rsp *rules.Reply) {
		t.Helper()
		for i, r := range rsp.Rules {
			t.Logf("Rule %d: id=%q, value=%q, tag=%q", i+1, r.ID, r.Value, r.Tag)
		}
		t.Logf("Sent: %s", rsp.Meta.Sent)
		t.Logf("Summary: c=%d, nc=%d, d=%d, nd=%d",
			rsp.Meta.Summary.Created, rsp.Meta.Summary.NotCreated,
			rsp.Meta.Summary.Deleted, rsp.Meta.Summary.NotDeleted)
	}

	const testRuleTag = "test english kittens whargarbl"
	var testRuleID string

	t.Run("Update", func(t *testing.T) {
		r, err := rules.Add(rules.Rule{
			Value: `cat has:images lang:en`,
			Tag:   testRuleTag,
		})
		if err != nil {
			t.Fatalf("Creating rules: %v", err)
		}

		rsp, err := rules.Update(r).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		logResponse(t, rsp)
	})

	t.Run("Get", func(t *testing.T) {
		rsp, err := rules.Get(nil).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		for _, r := range rsp.Rules {
			if r.Tag == testRuleTag {
				testRuleID = r.ID
			}
		}
		logResponse(t, rsp)
	})

	if testRuleID == "" {
		t.Fatalf("Test rule with tag %q was not found", testRuleTag)
	} else {
		t.Logf("Found test rule with id=%q", testRuleID)
	}

	del, err := rules.Delete(testRuleID)
	if err != nil {
		t.Fatalf("Creating rules: %v", err)
	}

	t.Run("Validate", func(t *testing.T) {
		rsp, err := rules.Validate(del).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		logResponse(t, rsp)
	})

	t.Run("Update", func(t *testing.T) {
		rsp, err := rules.Update(del).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		logResponse(t, rsp)
	})
}
