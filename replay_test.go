// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package twitter_test

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/rules"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
	"github.com/creachadair/twitter/users"
)

var (
	testDataFile = flag.String("testdata", "testdata/test-record", "Path of test data file")
	testMode     = flag.String("mode", "replay", "Test mode (record, replay, run)")
	doVerboseLog = flag.Bool("verbose-log", false, "Enable verbose client logging")
	maxBodyBytes = flag.Int64("max-body-size", 12000,
		"Maximum response body size when recording (0=unlimited)")

	cli *twitter.Client // see TestMain
)

// limitTransport wraps an http.RoundTripper to artificially truncate the
// response body to a fixed limit. We need to do this when recording stream
// methods, because the recorder consumes the whole response body before it
// returns any data to the client.
type limitTransport struct {
	real  http.RoundTripper
	limit int64
}

func (t limitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rsp, err := t.real.RoundTrip(req)
	if err == nil {
		rsp.Body = readCloser{
			Reader: io.LimitReader(rsp.Body, t.limit),
			Closer: rsp.Body,
		}
	}
	return rsp, err
}

type readCloser struct {
	io.Reader
	io.Closer
}

const fakeAuthToken = "this-is-a-fake-auth-token-for-testing"

// This test uses the go-vcr module to replay recorded HTTP interactions,
// captured from the live Twitter API.
//
// For ordinary use, run "go test", which will use the test-data.yaml file
// checked in at the root of the repository.
//
// To record a new file, run "go test -mode=record". Don't forget to check in any
// changes you obtain in this way.
//
// Use the -testdata flag to specify the location of the test data file.
//
// Use -verbose-log to get spammy client debug logging. This is mainly useful
// when you are verifying that the recording worked.
//
// Known deficiencies:
//
// - Each interaction is marked as "played" once it has been used so that it
//   cannot be replayed. This is sensible, but means if you run go test with
//   -count > 1 or multiple -cpu options, it will fail on all runs after the
//   first because it can't find the interactions again.
//
func TestMain(m *testing.M) {
	flag.Parse()

	var mode recorder.Mode
	switch *testMode {
	case "replay":
		mode = recorder.ModeReplaying
	case "record":
		mode = recorder.ModeRecording
	case "run":
		mode = recorder.ModeDisabled
	default:
		log.Fatalf("Unknown recorder mode %q (options: record, replay, run)", *testMode)
	}

	// Recording or replaying require a test data file and a recorder.
	if *testMode != "run" && *testDataFile == "" {
		log.Fatal("You must provide a non-empty -testdata file path")
	}

	// When recording, we need to limit the size of response bodies from the
	// server so that streaming methods do not stall the recorder.
	//
	// You want as small a limit as possible so as not to blow up the test data
	// size, but if it's too small the client will fail spuriously. The
	// practical solution is empiricism: Run production queries with a trial
	// limit and adjust the limit till they all pass.

	var rt http.RoundTripper = http.DefaultTransport
	if *testMode == "record" && *maxBodyBytes > 0 {
		rt = limitTransport{real: rt, limit: *maxBodyBytes}
		log.Printf("Limiting response bodies to %d bytes", *maxBodyBytes)
	}

	rec, err := recorder.NewAsMode(*testDataFile, mode, rt)
	if err != nil {
		log.Fatalf("Opening recorder %q: %v", *testDataFile, err)
	}

	// Running or recording require a production credential.
	// Replaying requires a fake credential.
	var auth twitter.Authorizer
	switch *testMode {
	case "run", "record":
		bearerToken := os.Getenv("TWITTER_TOKEN")
		if bearerToken == "" {
			// When talking to production, we need a real credential.
			log.Fatalf("No TWITTER_TOKEN found in the environment; cannot %s tests", *testMode)
		}
		auth = twitter.BearerTokenAuthorizer(bearerToken)
	default:
		auth = twitter.BearerTokenAuthorizer(fakeAuthToken)
	}

	// Filter Authorization headers when recording to swap the real token with
	// the fake, so we don't check in production credentials with testdata.
	if *testMode == "record" {
		rec.AddFilter(func(in *cassette.Interaction) error {
			// This relies on the fact that Values promises not to return a copy.
			auth := in.Request.Headers.Values("Authorization")
			for i, v := range auth {
				if strings.HasPrefix(v, "Bearer ") {
					auth[i] = "Bearer " + fakeAuthToken
					return nil
				}
			}
			log.Printf("WARNING: No Authorization found in request")
			return nil
		})
	}

	cli = &twitter.Client{
		HTTPClient: &http.Client{Transport: rec},
		Authorize:  auth,
	}
	if *doVerboseLog {
		log.Printf("Enabled verbose client logging")
		cli.Log = func(tag, msg string) {
			log.Printf("CLIENT :: %s | %s", tag, msg)
		}
	}
	os.Exit(func() int {
		if rec != nil {
			defer func() {
				if err := rec.Stop(); err != nil {
					log.Fatalf("Stopping recorder: %v", err)
				}
			}()
		}
		log.Print("Running tests...")
		return m.Run() // run the actual tests
	}())
}

// Verify that the direct call plumbing works.
func TestClientCall(t *testing.T) {
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
				types.Expand_PinnedTweetID,
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
	ctx := context.Background()
	rsp, err := tweets.Lookup("1297524288245895168", &tweets.LookupOpts{
		TweetFields: []string{
			types.Tweet_CreatedAt,
			types.Tweet_Entities,
			types.Tweet_AuthorID,
		},
		Expansions: []string{types.Expand_MentionUsername},
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
	ctx := context.Background()
	rsp, err := users.LookupByName("kanyewest", &users.LookupOpts{
		More: []string{"jack", "Popehat"},
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

func TestSearchPages(t *testing.T) {
	ctx := context.Background()

	const maxResults = 25
	for _, test := range []struct {
		label, query string
	}{
		{"OnePage", `from:creachadair has:mentions`},
		{"MultiPage", `from:benjaminwittes -has:images`},
	} {
		t.Run(test.label, func(t *testing.T) {
			q := tweets.SearchRecent(test.query, nil)

			nr := 0
			for q.HasMorePages() {
				rsp, err := q.Invoke(ctx, cli)
				if err != nil {
					t.Fatalf("SearchRecent failed: %v", err)
				}
				for _, tw := range rsp.Tweets {
					nr++
					t.Logf("Tweet %d: id=%s, text=%q", nr, tw.ID, tw.Text)
				}

				qpage := q.PageToken()
				t.Logf("Next page token: %q", qpage)

				if mpage := rsp.Meta.NextToken; mpage != qpage {
					t.Errorf("Query page token: got %q, want %q", qpage, mpage)
				}

				if nr > maxResults {
					t.Logf("Done: Got %d (> %d) results", nr, maxResults)
					return
				}
			}
			t.Log("Done: No more pages")
		})
	}
}

func TestSearchRecent(t *testing.T) {
	ctx := context.Background()

	// N.B. Don't set timestamps in the search options. Twitter only provides
	// about a week of data, so fixing a static timestamp will break recording.
	// But moving time will break playback, which matches on time.
	//
	// TODO: See about writing a matcher to ignore the time fields.

	const query = `from:benjaminwittes "Today on @inlieuoffunshow"`
	rsp, err := tweets.SearchRecent(query, &tweets.SearchOpts{
		MaxResults:  10,
		TweetFields: []string{types.Tweet_AuthorID, types.Tweet_Entities},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("SearchRecent failed: %v", err)
	}
	if rsp.Meta != nil {
		t.Logf("Response metadata: count=%d, oldest=%s, newest=%s",
			rsp.Meta.ResultCount, rsp.Meta.OldestID, rsp.Meta.NewestID)
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

func TestRules(t *testing.T) {
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
		r := rules.Adds{
			{Query: `cat has:images lang:en`, Tag: testRuleTag},
		}
		rsp, err := rules.Update(r).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		} else if len(rsp.Rules) != 1 {
			t.Errorf("Update: got %d rules, want 1", len(rsp.Rules))
		} else {
			testRuleID = rsp.Rules[0].ID
			t.Logf("Update assigned rule ID %q", testRuleID)
		}
		logResponse(t, rsp)
	})

	t.Run("Get", func(t *testing.T) {
		rsp, err := rules.Get(testRuleID).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		} else if len(rsp.Rules) != 1 {
			t.Errorf("Get: got %d rules, want 1", len(rsp.Rules))
		} else if r := rsp.Rules[0]; r.Tag != testRuleTag {
			t.Errorf("Rule %q tag: got %q, want %q", r.ID, r.Tag, testRuleTag)
		}
		logResponse(t, rsp)
	})

	t.Run("GetAll", func(t *testing.T) {
		rsp, err := rules.Get().Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		for _, r := range rsp.Rules {
			if r.ID == testRuleID && r.Tag == testRuleTag {
				t.Logf("Found rule ID %q with tag %q", r.ID, r.Tag)
				return
			}
		}
		t.Fatalf("Did not find expected rule ID %q", testRuleID)
	})

	// Requesting a non-existent rule ID should return an empty result.
	t.Run("GetMissing", func(t *testing.T) {
		const badID = "12345678"
		rsp, err := rules.Get(badID).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Get(%q) failed: %v", badID, err)
		} else if len(rsp.Rules) != 0 {
			t.Errorf("Get(%q): got %d rules, want 0", badID, len(rsp.Rules))
		}
		logResponse(t, rsp)
	})

	t.Run("Validate", func(t *testing.T) {
		del := rules.Deletes{testRuleID}
		rsp, err := rules.Validate(del).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		logResponse(t, rsp)
	})

	t.Run("Update", func(t *testing.T) {
		del := rules.Deletes{testRuleID}
		rsp, err := rules.Update(del).Invoke(ctx, cli)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		logResponse(t, rsp)
	})
}

func TestCallRaw(t *testing.T) {
	ctx := context.Background()

	save := cli.BaseURL
	defer func() { cli.BaseURL = save }()
	cli.BaseURL = "https://api.twitter.com/1.1"

	req := &twitter.Request{
		Method: "statuses/show.json",
		Params: twitter.Params{
			"id":         []string{"1297524288245895168"},
			"tweet_mode": []string{"extended"},
		},
	}
	data, err := cli.CallRaw(ctx, req)
	if err != nil {
		t.Fatalf("CallRaw failed: %v", err)
	}
	t.Logf("Received %d response bytes", len(data))

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("Decoding response: %v", err)
	}
	for key, val := range obj {
		switch s := val.(type) {
		case string:
			t.Logf("%-15s : %q", key, s)
		case float64:
			t.Logf("%-15s : %d", key, int64(s))
		}
	}
}
