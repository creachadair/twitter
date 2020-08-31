package auth_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/auth"
)

func getOrSkip(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skip("Missing " + key + " in environment; skipping this test")
	}
	return val
}

// This is a manual test that requires production credentials.
// Skip the test if they are not set in the environment.
func TestTwitter(t *testing.T) {
	cfg := auth.Config{
		APIKey:            getOrSkip(t, "TWITTER_API_KEY"),
		APISecret:         getOrSkip(t, "TWITTER_API_SECRET"),
		AccessToken:       getOrSkip(t, "TWITTER_ACCESS_TOKEN"),
		AccessTokenSecret: getOrSkip(t, "TWITTER_ACCESS_TOKEN_SECRET"),
	}
	cli := &twitter.Client{
		Authorize: func(req *http.Request) error {
			err := cfg.Authorize(req)
			if err == nil {
				t.Logf("Authorized request:\n%s", req.Header.Get("Authorization"))
			}
			return err
		},
		Log: func(tag, msg string) {
			t.Logf("DEBUG :: %s | %s", tag, msg)
		},
	}

	data, err := cli.CallRaw(context.Background(), &twitter.Request{
		Method: "oauth/request_token",
		Params: twitter.Params{"oauth_callback": []string{"oob"}},
	})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	q, err := url.ParseQuery(string(data))
	if err != nil {
		t.Fatalf("Parsing query: %v", err)
	}
	t.Logf("Auth URL: %s/oauth/authorize?oauth_token=%s", twitter.BaseURL, q.Get("oauth_token"))
	t.Logf("Auth secret: %s", q.Get("oauth_token_secret"))
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
