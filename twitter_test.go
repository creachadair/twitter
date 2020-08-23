package twitter_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
	"github.com/creachadair/twitter/users"
)

func checkAuth(t *testing.T) twitter.Authorizer {
	t.Helper()
	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		t.Skip("No BEARER_TOKEN found in the environment; test cannot run")
	}
	return twitter.BearerTokenAuthorizer(bearerToken)
}

func logFunc(t *testing.T) func(tag, message string) {
	return func(tag, message string) {
		t.Logf("LOG tag=%s :: %s", tag, message)
	}
}

func TestClientCall(t *testing.T) {
	cli := &twitter.Client{
		Authorize: checkAuth(t),
		Log:       logFunc(t),
	}

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
		t.Logf("Tweet [%d]: %+v", i+1, tweet)
	}
}

func TestTweetLookup(t *testing.T) {
	cli := &twitter.Client{
		Authorize: checkAuth(t),
		Log:       logFunc(t),
	}

	ctx := context.Background()
	// jack 1247616214769086465
	rsp, err := tweets.Lookup("1297524288245895168", &tweets.LookupOpts{
		TweetFields: []string{
			types.Tweet_CreatedAt,
			types.Tweet_Entities,
		},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	t.Logf("Lookup reply: %s", string(rsp.Reply.Data))
	for i, v := range rsp.Tweets {
		t.Logf("Tweet %d: %+v", i+1, v)
	}
	its, err := rsp.IncludedTweets()
	if err != nil {
		t.Fatalf("Decoding included tweets: %v", err)
	}
	for i, v := range its {
		t.Logf("Included Tweet %d: %+v", i+1, v)
	}
}

func TestUserIDLookup(t *testing.T) {
	cli := &twitter.Client{
		Authorize: checkAuth(t),
		Log:       logFunc(t),
	}

	ctx := context.Background()
	rsp, err := users.Lookup("12", nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	t.Logf("Lookup reply: %s", string(rsp.Reply.Data))

	for i, v := range rsp.Users {
		t.Logf("User %d: %+v", i+1, v)
	}
}

func TestUsernameLookup(t *testing.T) {
	cli := &twitter.Client{
		Authorize: checkAuth(t),
		Log:       logFunc(t),
	}

	ctx := context.Background()
	rsp, err := users.LookupByName("creachadair", &users.LookupOpts{
		Keys: []string{"jack", "benjaminwittes"},
		UserFields: []string{
			types.User_PinnedTweetID,
			types.User_ProfileImageURL,
			types.User_PublicMetrics,
		},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	t.Logf("Lookup reply: %s", string(rsp.Reply.Data))

	for i, v := range rsp.Users {
		t.Logf("User %d: %+v", i+1, v)
	}
}

func yesterday() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d-1, 21, 0, 0, 0, time.UTC)
}

func TestSearchRecent(t *testing.T) {
	cli := &twitter.Client{
		Authorize: checkAuth(t),
		Log:       logFunc(t),
	}

	ctx := context.Background()
	const query = `from:benjaminwittes "Today on @inlieuoffunshow"`
	rsp, err := tweets.SearchRecent(query, &tweets.SearchOpts{
		MaxResults:  10,
		StartTime:   yesterday(),
		TweetFields: []string{types.Tweet_Entities},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("SearchRecent failed: %v", err)
	}

	if len(rsp.Tweets) == 0 {
		t.Fatal("No matching results")
	}
	for i, tweet := range rsp.Tweets {
		t.Logf("Match %d: %+v", i+1, tweet)
		for j, u := range tweet.Entities.URLs {
			t.Logf("-- URL [%d]: %s", j+1, u.Expanded)
		}
	}
}
