// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package olists_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/olists"
)

var (
	doVerboseLog = flag.Bool("verbose-log", false, "Enable verbose client logging")
)

func getOrSkip(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skip("Missing " + key + " in environment; skipping this test")
	}
	return val
}

func TestUserCall(t *testing.T) {
	bearerToken := getOrSkip(t, "OLISTS_TWITTER_TOKEN")
	cli := twitter.NewClient(&jhttp.Client{
		Authorize: jhttp.BearerTokenAuthorizer(bearerToken),
	})
	ctx := context.Background()
	if *doVerboseLog {
		cli.Log = func(tag jhttp.LogTag, msg string) {
			t.Logf("DEBUG :: %s | %s", tag, msg)
		}
	}

	// Read a couple of pages to test pagination, but don't pull too many as the
	// app rate limit is only 30/15m at these endpoints.
	const maxPagesToRead = 2

	t.Run("Members", func(t *testing.T) {
		q := olists.Members("1318922483496591360", &olists.ListOpts{
			PerPage: 10,
		})

		for p := 0; p < maxPagesToRead && q.HasMorePages(); p++ {
			rsp, err := q.Invoke(ctx, cli)
			if err != nil {
				t.Fatalf("Invoke failed: %v", err)
			}
			for _, u := range rsp.Users {
				t.Logf("User id=%s username=%q name=%q", u.ID, u.Username, u.Name)
			}

			t.Logf("Next page token: %q", rsp.NextToken)
		}
	})

	t.Run("Subscribers", func(t *testing.T) {
		q := olists.Subscribers("1318922483496591360", &olists.ListOpts{
			PerPage: 10,
		})

		for p := 0; p < maxPagesToRead && q.HasMorePages(); p++ {
			rsp, err := q.Invoke(ctx, cli)
			if err != nil {
				t.Fatalf("Invoke failed: %v", err)
			}
			for _, u := range rsp.Users {
				t.Logf("User id=%s username=%q name=%q", u.ID, u.Username, u.Name)
			}

			t.Logf("Next page token: %q", rsp.NextToken)
		}
	})
}
