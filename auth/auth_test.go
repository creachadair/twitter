// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package auth_test

/*
About the tests in this file:

Most of the tests defined below require production access and credentials.
The tests that require this expect their credentials to be passed in the
environment, and if they are missing the test will be skipped.

The environment variables affecting the tests are as follows:

All tests:
  AUTHTEST_API_KEY    : the application's API key ("consumer key")
  AUTHTEST_API_SECRET : the application's API secret ("consumer secret")

TestRequestFlow:
  AUTHTEST_ACCESS_TOKEN        : the application's access token
  AUTHTEST_ACCESS_TOKEN_SECRET : the application's access token secret

TestAccessGrant:
  AUTHTEST_REQUEST_TOKEN    : the request token from the request flow
  AUTHTEST_REQUEST_VERIFIER : the access verifier (PIN or nonce)

TestUserQuery:
  AUTHTEST_USER_TOKEN : the user's name and access token, in the format
     <username>:<token>:<secret>

To generate user credentials:

1. Run TestRequestFlow and use the URL logged in the test output to manually
   generate a verification PIN.

2. Run TestAccessGrant with the token and verifier from (1).

3. Use the log output from (2) to construct the AUTHTEST_USER_TOKEN string.

TestUserQuery verifies that the user credentials work by requesting information
that the API will not grant without user context.
*/

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/auth"
	"github.com/creachadair/twitter/tweets"
	"github.com/creachadair/twitter/types"
)

func debugClient(t *testing.T) *twitter.Client {
	t.Helper()
	return twitter.NewClient(&twitter.ClientOpts{
		Log: func(tag, msg string) {
			t.Logf("DEBUG :: %s | %s", tag, msg)
		},
	})
}

func getOrSkip(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skip("Missing " + key + " in environment; skipping this test")
	}
	return val
}

func baseConfigOrSkip(t *testing.T) auth.Config {
	t.Helper()
	return auth.Config{
		APIKey:    getOrSkip(t, "AUTHTEST_API_KEY"),
		APISecret: getOrSkip(t, "AUTHTEST_API_SECRET"),
	}
}

func authConfigOrSkip(t *testing.T) auth.Config {
	t.Helper()
	cfg := baseConfigOrSkip(t)
	cfg.AccessToken = getOrSkip(t, "AUTHTEST_ACCESS_TOKEN")
	cfg.AccessTokenSecret = getOrSkip(t, "AUTHTEST_ACCESS_TOKEN_SECRET")
	return cfg
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestRequestFlow(t *testing.T) {
	cfg := authConfigOrSkip(t)
	cli := debugClient(t)
	ctx := context.Background()

	tok, err := cfg.GetRequestToken(auth.UsePIN, nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("GetRequestToken failed: %v", err)
	}

	t.Logf("Request token secret: %s", tok.Secret)
	t.Logf("Auth URL: %s/oauth/authorize?oauth_token=%s", twitter.BaseURL, tok.Key)
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestAccessGrant(t *testing.T) {
	cfg := authConfigOrSkip(t)
	reqToken := getOrSkip(t, "AUTHTEST_REQUEST_TOKEN")
	verifier := getOrSkip(t, "AUTHTEST_REQUEST_VERIFIER")
	cli := debugClient(t)
	ctx := context.Background()

	tok, err := cfg.GetAccessToken(reqToken, verifier, nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
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

	tok, err := cfg.GetBearerToken(nil).Invoke(ctx, cli)
	if err != nil {
		t.Fatalf("GetBearerToken failed: %v", err)
	}
	t.Logf(`Bearer token:
Token:  %q
Secret: %q`, tok.Key, tok.Secret)
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestUserQuery(t *testing.T) {
	cfg := baseConfigOrSkip(t)
	creds := strings.SplitN(getOrSkip(t, "AUTHTEST_USER_TOKEN"), ":", 3)
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

// Test vectors from http://lti.tools/oauth/ to verify the basic computations.
func TestKnownInputs(t *testing.T) {
	// Example inputs
	const (
		requestURL = "http://photos.example.net/photos"
		wantParams = `file=vacation.jpg&oauth_consumer_key=dpf43f3p2l4k3l03&oauth_nonce=kllo9940pd9333jh&` +
			`oauth_signature_method=HMAC-SHA1&oauth_timestamp=1191242096&oauth_token=nnch734d00sl2jdk&` +
			`oauth_version=1.0&size=original`
		wantSig = `tR3+Ty81lMeYAr/Fid0kMTYa/WM=`

		// N.B. This value has been redacted by removing newlines.
		wantAuth = `OAuth oauth_consumer_key="dpf43f3p2l4k3l03", oauth_token="nnch734d00sl2jdk", ` +
			`oauth_nonce="kllo9940pd9333jh", oauth_timestamp="1191242096", oauth_signature_method="HMAC-SHA1", ` +
			`oauth_version="1.0", oauth_signature="tR3%2BTy81lMeYAr%2FFid0kMTYa%2FWM%3D"`
	)

	cfg := auth.Config{
		APIKey:            "dpf43f3p2l4k3l03",
		APISecret:         "kd94hf93k423kf44",
		AccessToken:       "nnch734d00sl2jdk",
		AccessTokenSecret: "pfkkdhi9sl3r4s00",
	}
	params := auth.Params{
		"oauth_nonce":     "kllo9940pd9333jh",
		"oauth_timestamp": "1191242096",
		"size":            "original",
		"file":            "vacation.jpg",
	}
	ad := cfg.Sign("GET", requestURL, params)
	if got := ad.Params.Encode(); got != wantParams {
		t.Errorf("Encoded parameters:\ngot:  %s\nwant: %s", got, wantParams)
	}
	if ad.Signature != wantSig {
		t.Errorf("Signature:\ngot:  %s\nwant: %s", ad.Signature, wantSig)
	}
	if ad.Authorization != wantAuth {
		t.Errorf("Authorization:\ngot:  %s\nwant: %s", ad.Authorization, wantAuth)
	}
}
