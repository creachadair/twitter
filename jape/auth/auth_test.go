// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package auth_test

import (
	"testing"

	"github.com/creachadair/twitter/jape/auth"
)

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
