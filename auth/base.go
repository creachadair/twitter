// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package auth supports OAuth request signing and requests to generate
// authorization tokens.
//
// The core type in this package is Config, which carries the application and
// user secrets. Methods of the Config type implement signing of requests and
// handle queries to the API for tokens. At minimum, the APIKey and APISecret
// fields must be populated with the application's credentials.
//
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/creachadair/twitter"
)

/*
Working notes:

To get an access token and secret for a user:

1. POST api.twitter.com/oauth/request_token ? oauth_callback=URL
   This request must be signed with the app's own access token and secret.
   It returns an ephemeral token (ET) and secret (ETS) to use for step 2.
   These are returned as form-encoded query terms in the response body.
   ETS is not used for anything.
   Note that the "oob" callback causes Twitter to issue a PIN for verification.

2. GET api.twitter.com/oauth/authorize ? oauth_token=ET [& force_login=true & screen_name=who]
   This URL is visited by the user. Assuming the user grants access:
   - In the ordinary flow, the site redirects to the app's callback with a verifier (V).
   - In the PIN flow, the site gives the user a PIN (P) to hand-deliver to the app.
   Either V or P is needed for step (3).

3. GET api.twitter.com/oauth/access_toekn ? oauth_token=ET & oauth_consumer_key=APIKEY & oauth_verifier=P
   It returns the durable user token (UT) and secret (UTS).
   These are returned as form-encoded query terms in the response body.
   These must be stored securly and used to sign requests on behalf of the user.

When the app acts on its own behalf (e.g., to request a user token), it uses
its own AccessToken, from the app settings.

When the app acts on the user's behalf, it uses the user's AccessToken, issued
by the server in Step (3).
*/

// Config carries the keys and secrets to generate OAuth 1.0 signatures.
//
// The APIKey and APISecret fields must be populated for all requests.
// The rules for AccessToken and AccessTokenSecret are described in the
// documentation for each query type.
type Config struct {
	APIKey            string `oauth:"oauth_consumer_key"`    // also: Consumer or App Key
	APISecret         string `oauth:"oauth_consumer_secret"` // also: Consumer or App Secret
	AccessToken       string `oauth:"oauth_token"`
	AccessTokenSecret string `oauth:"oauth_token_secret"`

	// If set, use this function to generate a nonce.
	// If unset, a non-cryptographic pseudorandom nonce will be used.
	MakeNonce func() string

	// If set, use this function to generate a timestamp.
	// If unset, use time.Now.
	Timestamp func() time.Time
}

// Authorizer returns a twitter.Authorizer that uses the specified access token
// to sign requests.
func (c Config) Authorizer(token, secret string) twitter.Authorizer {
	uc := c // shallow copy
	uc.AccessToken = token
	uc.AccessTokenSecret = secret
	return uc.Authorize
}

// Authorize attaches an OAuth 1.0 signature to the given request.
//
// This operation requires c.AccessToken and c.AccessTokenSecret to be set.
// To authorize a ticket request, use the application's credentials.
// To authorize a user request, use the user's credentials.
func (c Config) Authorize(req *http.Request) error {
	if c.AccessToken == "" || c.AccessTokenSecret == "" {
		return errors.New("missing access credentials")
	}
	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return fmt.Errorf("invalid query: %v", err)
	}
	sigURL := (&url.URL{
		Scheme:  req.URL.Scheme,
		Host:    req.URL.Host,
		Path:    req.URL.Path,
		RawPath: req.URL.RawPath,
	}).String()

	params := make(Params)
	for key, vals := range q {
		if len(vals) != 0 {
			params[key] = strings.Join(vals, ",")
		}
	}

	// TODO: Maybe parse query terms out of the body?

	authData := c.Sign(req.Method, sigURL, params)
	req.Header.Add("Authorization", authData.Authorization)
	return nil
}

// AuthData carries the result of authorizing a request.
type AuthData struct {
	Params        Params // the annotated request parameters (as signed)
	Signature     string // the HMAC-SHA1 signature
	Authorization string // the Authorization field value
}

// makeAuthParams returns a copy of params with oauth metadata added.
// Any oauth_* parameters are copied to the result, and removed from params.
func (c Config) makeAuthParams(params Params) Params {
	tmp := Params{
		"oauth_version":          "1.0",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_consumer_key":     c.APIKey,
		"oauth_token":            c.AccessToken,
		"oauth_timestamp":        c.makeTimestamp(),
		"oauth_nonce":            c.makeNonce(),
	}
	for key, val := range params {
		if _, ok := tmp[key]; ok {
			delete(params, key)
		}
		tmp[key] = val
	}
	return tmp
}

// signature computes the signature for the specified request parameters.
func (c Config) signature(method, requestURL string, authParams Params) string {
	urlWithoutQuery := strings.SplitN(requestURL, "?", 2)[0]

	base := strings.ToUpper(method) + // e.g., POST
		"&" + url.QueryEscape(urlWithoutQuery) +
		"&" + url.QueryEscape(authParams.Encode())
	// N.B.: Escaping the encoded authParams is intentional and required, to
	// hide the "&" separators from the base string.

	key := url.QueryEscape(c.APISecret) + "&" + url.QueryEscape(c.AccessTokenSecret)
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(base))
	sig := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(sig)
}

// Sign computes an authorization signature for the request parameters.
// The requestURL must not contain any query parameters or fragments.
//
// If params contains parameters that affect the OAuth signature, such as
// "oauth_timestamp" or "oauth_nonce", their values are copied for signing and
// deleted from params. The contents of params are not otherwise modified. The
// parameters as-signed can be recovered from the Params field of the AuthData
// value returned.
func (c Config) Sign(method, requestURL string, params Params) AuthData {
	authParams := c.makeAuthParams(params)
	sig := c.signature(method, requestURL, authParams)

	qfmt := func(key, val string) string { return key + `="` + url.QueryEscape(val) + `"` }
	qesc := func(key string) string { return qfmt(key, authParams[key]) }
	args := []string{
		qesc("oauth_consumer_key"),
		qesc("oauth_token"),
		qesc("oauth_nonce"),
		qesc("oauth_timestamp"),
		qesc("oauth_signature_method"),
		qesc("oauth_version"),
		qfmt("oauth_signature", sig),
	}
	auth := "OAuth " + strings.Join(args, ", ")

	return AuthData{
		Params:        authParams,
		Signature:     sig,
		Authorization: auth,
	}
}

func (c Config) makeNonce() string {
	if c.MakeNonce != nil {
		return c.MakeNonce()
	}
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func (c Config) makeTimestamp() string {
	var now time.Time
	if c.Timestamp != nil {
		now = c.Timestamp()
	} else {
		now = time.Now()
	}
	return strconv.FormatInt(int64(now.Unix()), 10)
}

// Params represent URL query parameters.
type Params map[string]string

// Encode encodes p as a URL query string, not including the "?" prefix.
func (p Params) Encode() string {
	q := make(url.Values)
	for key, val := range p {
		q.Set(key, val)
	}

	// QueryEscape correctly escapes "+" as "%2B", but uses "+" for " ".
	// Since we aren't allowed to use "+' in this context, fix it up after.
	return strings.ReplaceAll(q.Encode(), "+", "%20")
}
