// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package twitter_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/rules"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
)

var (
	doManual = flag.Bool("manual", false, "Enable manual tests")
)

func newProdClient(t *testing.T) *twitter.Client {
	t.Helper()
	if !*doManual {
		t.Skip("Skipping test because -manual=false")
	}

	bearerToken := os.Getenv("TWITTER_TOKEN")
	if bearerToken == "" {
		// When talking to production, we need a real credential.
		log.Fatal("No TWITTER_TOKEN found in the environment; cannot run manual tests")
	}
	cli = &twitter.Client{Authorize: twitter.BearerTokenAuthorizer(bearerToken)}
	if *doVerboseLog {
		cli.Log = func(tag, msg string) {
			log.Printf("CLIENT :: %s | %s", tag, msg)
		}
	}
	return cli
}

func TestStream(t *testing.T) {
	cli := newProdClient(t)
	ctx := context.Background()

	req := &twitter.Request{
		Method: "tweets/sample/stream",
		Params: twitter.Params{
			types.TweetFields: []string{
				types.Tweet_AuthorID,
				types.Tweet_Entities,
			},
		},
	}

	const maxResults = 3

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

func TestSearchStream(t *testing.T) {
	cli := newProdClient(t)
	ctx := context.Background()

	r := rules.Adds{{Query: `cat has:images lang:en`}}
	rsp, err := rules.Update(r).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("Updating rules: %v", err)
	}
	id := rsp.Rules[0].ID

	t.Run("Search", func(t *testing.T) {
		const maxResults = 3

		nr := 0
		err := tweets.SearchStream(func(rsp *tweets.Reply) error {
			for _, tw := range rsp.Tweets {
				nr++
				t.Logf("Result %d: id=%s, author=%s, text=%s", nr, tw.ID, tw.AuthorID, tw.Text)
			}
			if nr >= maxResults {
				return twitter.ErrStopStreaming
			}
			return nil
		}, &tweets.StreamOpts{
			TweetFields: []string{types.Tweet_AuthorID},
		}).Invoke(ctx, cli)
		if err != nil {
			t.Errorf("SearchStream failed: %v", err)
		}
	})

	del := rules.Deletes{id}
	if _, err := rules.Update(del).Invoke(ctx, cli); err != nil {
		t.Fatalf("Deleting rules: %v", err)
	}
}
