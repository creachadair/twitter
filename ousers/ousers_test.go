// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package ousers_test

import (
	"context"
	"testing"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter/internal/otest"
	"github.com/creachadair/twitter/ousers"
)

func TestUserCall(t *testing.T) {
	bearerToken := otest.GetOrSkip(t, "OUSERS_TWITTER_TOKEN")
	cli := otest.NewClient(t, &jhttp.Client{
		Authorize: jhttp.BearerTokenAuthorizer(bearerToken),
	})
	ctx := context.Background()

	// Read a couple of pages to test pagination, but don't pull too many as the
	// app rate limit is only 30/15m at these endpoints.
	const maxPagesToRead = 2

	t.Run("Followers", func(t *testing.T) {
		q := ousers.Followers("jack", &ousers.FollowOpts{
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

	t.Run("Following", func(t *testing.T) {
		q := ousers.Following("jack", &ousers.FollowOpts{
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
