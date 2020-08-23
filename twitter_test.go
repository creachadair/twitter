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
	t.Logf("Reply: %s", string(rsp.Data))
	for key, vals := range rsp.Includes {
		t.Logf("Include type %q:", key)
		for i, val := range vals {
			t.Logf("[%d] %s", i+1, string(val))
		}
	}
}
