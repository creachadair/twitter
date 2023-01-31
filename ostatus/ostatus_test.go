// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package ostatus_test

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/internal/otest"
	"github.com/creachadair/twitter/jape"
	"github.com/creachadair/twitter/jape/auth"
	"github.com/creachadair/twitter/ostatus"
	"github.com/creachadair/twitter/types"
)

var (
	testCaseDelay = flag.Duration("pause", 0, "How long to pause between tests (for observation)")
)

func pause(t *testing.T) {
	t.Helper()
	if *testCaseDelay > 0 {
		t.Logf("Pausing %v before next test...", *testCaseDelay)
		time.Sleep(*testCaseDelay)
	}
}

func newTestClient(t *testing.T) (context.Context, string, *twitter.Client) {
	t.Helper()
	cfg := auth.Config{
		APIKey:    otest.GetOrSkip(t, "AUTHTEST_API_KEY"),
		APISecret: otest.GetOrSkip(t, "AUTHTEST_API_SECRET"),
	}
	userToken := strings.SplitN(otest.GetOrSkip(t, "OSTATUSTEST_USER_TOKEN"), ":", 2)
	if len(userToken) != 2 {
		t.Fatal("Invalid user token format; want TOKEN:SECRET [redacted]")
	}

	userID := strings.SplitN(userToken[0], "-", 2)[0]
	ctx := context.Background()
	cli := otest.NewClient(t, &jape.Client{
		Authorize: cfg.Authorizer(userToken[0], userToken[1]),
	})
	return ctx, userID, cli
}

func TestTimelines(t *testing.T) {
	ctx, userID, cli := newTestClient(t)

	t.Run("UserTimeline", func(t *testing.T) {
		rsp, err := ostatus.UserTimeline(userID, &ostatus.TimelineOpts{
			ByID: true,
		}).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("UserTimeline failed: %v", err)
		}
		t.Logf("Found %d tweets in user timeline", len(rsp.Tweets))
		for i, s := range rsp.Tweets {
			t.Logf(" - Tweet %d: id=%s text=%q", i+1, s.ID, s.Text)
		}
	})

	t.Run("HomeTimeline", func(t *testing.T) {
		rsp, err := ostatus.HomeTimeline(userID, &ostatus.TimelineOpts{
			ByID: true,
		}).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("HomeTimeline failed: %v", err)
		}
		t.Logf("Found %d tweets in home timeline", len(rsp.Tweets))
		for i, s := range rsp.Tweets {
			t.Logf(" - Tweet %d: id=%s text=%q", i+1, s.ID, s.Text)
		}
	})

	t.Run("MentionsTimeline", func(t *testing.T) {
		rsp, err := ostatus.MentionsTimeline(userID, &ostatus.TimelineOpts{
			ByID: true,
		}).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("HomeTimeline failed: %v", err)
		}
		t.Logf("Found %d tweets in mentions timeline", len(rsp.Tweets))
		for i, s := range rsp.Tweets {
			t.Logf(" - Tweet %d: id=%s text=%q", i+1, s.ID, s.Text)
		}
	})
}

func TestUserCall(t *testing.T) {
	ctx, _, cli := newTestClient(t)

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
		c := rsp.Tweets[0]
		t.Logf("Created ID %s, author=%s, lang=%q, date=%s, text=%q",
			c.ID, c.AuthorID, c.Language, c.CreatedAt, c.Text)
		createdID = c.ID
		for _, m := range c.Entities.HashTags {
			t.Logf("Hashtag [%d..%d] %q", m.Start, m.End, m.Tag)
		}
	})
	pause(t)

	t.Run("Like", func(t *testing.T) {
		rsp, err := ostatus.Like(createdID, nil).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Like failed: %v", err)
		}
		t.Logf("Liked ID %s, text=%q", rsp.Tweets[0].ID, rsp.Tweets[0].Text)
	})
	pause(t)

	t.Run("Unlike", func(t *testing.T) {
		rsp, err := ostatus.Unlike(createdID, nil).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Unlike failed: %v", err)
		}
		c := rsp.Tweets[0]
		t.Logf("Unliked ID %s, text=%q", c.ID, c.Text)
	})
	pause(t)

	t.Run("Delete", func(t *testing.T) {
		rsp, err := ostatus.Delete(createdID, nil).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		c := rsp.Tweets[0]
		t.Logf("Deleted ID %s, text=%q", c.ID, c.Text)
		if rsp.Tweets[0].Text != testMessage {
			t.Errorf("Unexpected message content:\ngot:  %q\nwant: %q", rsp.Tweets[0].Text, testMessage)
		}
	})

}
