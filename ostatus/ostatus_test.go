// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package ostatus_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/auth"
	"github.com/creachadair/twitter/ostatus"
	"github.com/creachadair/twitter/types"
)

var (
	testCaseDelay = flag.Duration("pause", 0, "How long to pause between tests (for observation)")
	doVerboseLog  = flag.Bool("verbose-log", false, "Enable verbose client logging")
)

func getOrSkip(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skip("Missing " + key + " in environment; skipping this test")
	}
	return val
}

func pause(t *testing.T) {
	t.Helper()
	if *testCaseDelay > 0 {
		t.Logf("Pausing %v before next test...", *testCaseDelay)
		time.Sleep(*testCaseDelay)
	}
}

func TestUserCall(t *testing.T) {
	cfg := auth.Config{
		APIKey:    getOrSkip(t, "AUTHTEST_API_KEY"),
		APISecret: getOrSkip(t, "AUTHTEST_API_SECRET"),
	}
	userToken := strings.SplitN(getOrSkip(t, "OSTATUSTEST_USER_TOKEN"), ":", 2)
	if len(userToken) != 2 {
		t.Fatal("Invalid user token format; want TOKEN:SECRET [redacted]")
	}

	ctx := context.Background()
	cli := twitter.NewClient(&twitter.ClientOpts{
		Authorize: cfg.Authorizer(userToken[0], userToken[1]),
	})
	if *doVerboseLog {
		cli.Log = func(tag jhttp.LogTag, msg string) {
			t.Logf("DEBUG :: %s | %s", tag, msg)
		}
	}

	// Create a status and then delete it.  Since you cannot post two statuses
	// with the same content too nearby in time, embed a timestamp in the text
	// so that check is unlikely to interfere.
	testMessage := fmt.Sprintf("Test message %d 😃 #funtimes", time.Now().Unix())

	var createdID string
	t.Run("Create", func(t *testing.T) {
		rsp, err := ostatus.Create(testMessage, &ostatus.CreateOpts{
			Optional: types.TweetFields{
				AuthorID:  true,
				CreatedAt: true,
				Language:  true,
				Entities:  true,
			},
		}).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		t.Logf("Created ID %s, author=%s, lang=%q, date=%s, text=%q",
			rsp.Tweet.ID, rsp.Tweet.AuthorID, rsp.Tweet.Language, rsp.Tweet.CreatedAt,
			rsp.Tweet.Text)
		createdID = rsp.Tweet.ID
		for _, m := range rsp.Tweet.Entities.HashTags {
			t.Logf("Hashtag [%d..%d] %q", m.Start, m.End, m.Tag)
		}
	})
	pause(t)

	t.Run("Like", func(t *testing.T) {
		rsp, err := ostatus.Like(createdID, nil).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Like failed: %v", err)
		}
		t.Logf("Liked ID %s, text=%q", rsp.Tweet.ID, rsp.Tweet.Text)
	})
	pause(t)

	t.Run("UnLike", func(t *testing.T) {
		rsp, err := ostatus.UnLike(createdID, nil).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("UnLike failed: %v", err)
		}
		t.Logf("UnLiked ID %s, text=%q", rsp.Tweet.ID, rsp.Tweet.Text)
	})
	pause(t)

	t.Run("Delete", func(t *testing.T) {
		rsp, err := ostatus.Delete(createdID, nil).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		t.Logf("Deleted ID %s, text=%q", rsp.Tweet.ID, rsp.Tweet.Text)
		if rsp.Tweet.Text != testMessage {
			t.Errorf("Unexpected message content:\ngot:  %q\nwant: %q", rsp.Tweet.Text, testMessage)
		}
	})

}
