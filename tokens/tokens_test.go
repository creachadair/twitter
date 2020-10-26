// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tokens_test

/*
About the tests in this file:

Most of the tests defined below require production access and credentials.
The tests that require this expect their credentials to be passed in the
environment, and if they are missing the test will be skipped.

The environment variables affecting the tests are as follows:

All tests:
  AUTHTEST_API_KEY    : the application's API key ("consumer key")
  AUTHTEST_API_SECRET : the application's API secret ("consumer secret")

TestRequestToken:
  AUTHTEST_ACCESS_TOKEN        : the application's access token
  AUTHTEST_ACCESS_TOKEN_SECRET : the application's access token secret

TestAccessGrant:
  AUTHTEST_REQUEST_TOKEN    : the request token from the request flow
  AUTHTEST_REQUEST_VERIFIER : the access verifier (PIN or nonce)

TestUserQuery:
  AUTHTEST_USER_TOKEN : the user's name and access token, in the format
     <username>:<token>:<secret>

TestInvalidateAccess:
  AUTHTEST_INVALIDATE_TOKEN  : the access token to invalidate
  AUTHTEST_INVALIDATE_SECRET : the access token secret to invalidate

TestInvalidateBearer:
  AUTHTEST_INVALIDATE_BEARER : the bearer token to invalidate

To generate user credentials:

1. Run TestRequestToken and use the URL logged in the test output to manually
   generate a verification PIN.

   Note that a given URL will only work once, whether or not the resulting
   verification is used for anything. Test caching may re-use a previous
   result. To bypass this, add -count=1 to the test run.

2. Run TestAccessGrant with the token and verifier from (1).

3. Use the log output from (2) to construct the AUTHTEST_USER_TOKEN string.

TestUserQuery verifies that the user credentials work by requesting information
that the API will not grant without user context.
*/

import (
	"context"
	"strings"
	"testing"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/jhttp/auth"
	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/internal/otest"
	"github.com/creachadair/twitter/tokens"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
)

func debugClient(t *testing.T) *twitter.Client {
	t.Helper()
	return twitter.NewClient(&jhttp.Client{
		Log: func(tag jhttp.LogTag, msg string) {
			t.Logf("DEBUG :: %s | %s", tag, msg)
		},
	})
}

func baseConfigOrSkip(t *testing.T) auth.Config {
	t.Helper()
	return auth.Config{
		APIKey:    otest.GetOrSkip(t, "AUTHTEST_API_KEY"),
		APISecret: otest.GetOrSkip(t, "AUTHTEST_API_SECRET"),
	}
}

func authConfigOrSkip(t *testing.T) auth.Config {
	t.Helper()
	cfg := baseConfigOrSkip(t)
	cfg.AccessToken = otest.GetOrSkip(t, "AUTHTEST_ACCESS_TOKEN")
	cfg.AccessTokenSecret = otest.GetOrSkip(t, "AUTHTEST_ACCESS_TOKEN_SECRET")
	return cfg
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestRequestToken(t *testing.T) {
	cfg := authConfigOrSkip(t)
	cli := debugClient(t)
	ctx := context.Background()

	tok, err := tokens.GetRequest(cfg, tokens.UsePIN, nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("GetRequest failed: %v", err)
	}

	t.Logf("Request token secret: %s", tok.Secret)
	t.Logf("Auth URL: %s/oauth/authorize?oauth_token=%s", twitter.BaseURL, tok.Key)
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestAccessGrant(t *testing.T) {
	cfg := authConfigOrSkip(t)
	reqToken := otest.GetOrSkip(t, "AUTHTEST_REQUEST_TOKEN")
	verifier := otest.GetOrSkip(t, "AUTHTEST_REQUEST_VERIFIER")
	cli := debugClient(t)
	ctx := context.Background()

	tok, err := tokens.GetAccess(cfg, reqToken, verifier, nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("GetAccess failed: %v", err)
	}
	t.Logf(`Access token:
UserID:   %q
Username: %q
Token:    %q
Secret:   %q`, tok.UserID, tok.Username, tok.Key, tok.Secret)
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestBearerToken(t *testing.T) {
	cfg := baseConfigOrSkip(t)
	cli := debugClient(t)
	ctx := context.Background()

	tok, err := tokens.GetBearer(cfg, nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("GetBearer failed: %v", err)
	}
	t.Logf(`Bearer token:
Token:  %q
Secret: %q`, tok.Key, tok.Secret)
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestUserQuery(t *testing.T) {
	cfg := baseConfigOrSkip(t)
	creds := strings.SplitN(otest.GetOrSkip(t, "AUTHTEST_USER_TOKEN"), ":", 3)
	if len(creds) != 3 {
		t.Fatal("Invalid AUTHTEST_USER_TOKEN; want user:token:secret")
	}
	userName, token, secret := creds[0], creds[1], creds[2]

	cli := debugClient(t)
	cli.Authorize = cfg.Authorizer(token, secret)
	ctx := context.Background()

	// Verify that we can get information that requires a user token.
	// Search to find the user's latest tweets, then request the non-public
	// metrics for that tweet.

	srsp, err := tweets.SearchRecent("from:"+userName, nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("SearchResults failed: %v", err)
	} else if len(srsp.Tweets) == 0 {
		t.Fatal("No results found")
	}
	latest := srsp.Tweets[0]
	t.Logf("Found latest tweet id=%q", latest.ID)

	lrsp, err := tweets.Lookup(latest.ID, &tweets.LookupOpts{
		Optional: []types.Fields{
			types.TweetFields{PublicMetrics: true, NonPublicMetrics: true},
		},
	}).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	for key, val := range lrsp.Tweets[0].PublicMetrics {
		t.Logf("Public     %-15s : %d", key, val)
	}
	for key, val := range lrsp.Tweets[0].NonPublicMetrics {
		t.Logf("Non-public %-15s : %d", key, val)
	}
}

func TestInvalidateAccess(t *testing.T) {
	cfg := baseConfigOrSkip(t)
	token := otest.GetOrSkip(t, "AUTHTEST_INVALIDATE_TOKEN")
	secret := otest.GetOrSkip(t, "AUTHTEST_INVALIDATE_SECRET")

	cli := debugClient(t)
	ctx := context.Background()

	rsp, err := tokens.InvalidateAccess(cfg, token, secret).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("InvalidateAccess failed: %v", err)
	}
	t.Logf("Invalidated access token: %s", rsp)
}

func TestInvalidateBearer(t *testing.T) {
	cfg := authConfigOrSkip(t)
	bearer := otest.GetOrSkip(t, "AUTHTEST_INVALIDATE_BEARER")

	cli := debugClient(t)
	ctx := context.Background()

	rsp, err := tokens.InvalidateBearer(cfg, bearer).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("InvalidateBearer failed: %v", err)
	}
	t.Logf("Invalidated bearer token: %s", rsp)
}
