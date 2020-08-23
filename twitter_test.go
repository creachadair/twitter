package twitter_test

import (
	"context"
	"os"
	"testing"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/types"
)

func TestClient(t *testing.T) {
	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		t.Skip("No BEARER_TOKEN found in the environment; test cannot run")
	}

	cli := &twitter.Client{
		Authorize: twitter.BearerTokenAuthorizer(bearerToken),
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
