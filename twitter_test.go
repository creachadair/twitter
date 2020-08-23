package twitter_test

import (
	"context"
	"os"
	"testing"

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

func TestClientCall(t *testing.T) {
	cli := &twitter.Client{Authorize: checkAuth(t)}

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
	cli := &twitter.Client{Authorize: checkAuth(t)}

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
	cli := &twitter.Client{Authorize: checkAuth(t)}

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
	cli := &twitter.Client{Authorize: checkAuth(t)}

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
