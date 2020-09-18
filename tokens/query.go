// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package tokens

// See https://developer.twitter.com/en/docs/api-reference-index#platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/creachadair/jhttp"
	"github.com/creachadair/jhttp/auth"
	"github.com/creachadair/twitter"
)

func clientWithAuth(cli *twitter.Client, auth jhttp.Authorizer) *twitter.Client {
	cp := *cli // shallow copy
	cp.Authorize = auth
	return &cp
}

// UsePIN is used as the callback in an authorization ticket request to request
// "out-of-band" or PIN based verification.
const UsePIN = "oob"

// GetRequest constructs a query to obtain an authorization request ticket for
// the specified callback URL. Pass UsePIN for the callback to use PIN based
// verification.
//
// This query requires c.AccessToken and c.AccessTokenSecret to be set to the
// application's own credentials.
//
// API: oauth/request_token
func GetRequest(c auth.Config, callback string, opts *RequestOpts) RequestQuery {
	req := &jhttp.Request{
		Method:     "oauth/request_token",
		HTTPMethod: "POST",
		Params:     jhttp.Params{"oauth_callback": []string{callback}},
	}
	opts.addRequestParams(req)
	return RequestQuery{Request: req, authorize: c.Authorize}
}

// A RequestQuery is a query for an authorization ticket.
type RequestQuery struct {
	*jhttp.Request
	authorize jhttp.Authorizer
}

// Invoke issues the query to the given client and returns the request Token.
func (q RequestQuery) Invoke(ctx context.Context, cli *twitter.Client) (Token, error) {
	data, err := clientWithAuth(cli, q.authorize).CallRaw(ctx, q.Request)
	if err != nil {
		return Token{}, err
	}
	tok, err := url.ParseQuery(string(data))
	if err != nil {
		return Token{}, &jhttp.Error{Message: "parsing response", Err: err}
	}
	return Token{
		Key:    tok.Get("oauth_token"),
		Secret: tok.Get("oauth_token_secret"),
	}, nil
}

// RequestOpts provides optional values for a ticket-granting request.
// A nil *RequestOpts provides empty values for all fields.
type RequestOpts struct {
	AccessType string // access override; "read" or "write"
}

func (o *RequestOpts) addRequestParams(req *jhttp.Request) {
	if o != nil && o.AccessType != "" {
		req.Params.Set("x_auth_access_type", o.AccessType)
	}
}

// GetAccess constructs a query to obtain an access token from the given
// request token and verifier.
//
// This query does not require c.AccessToken or c.AccessTokenSecret.
//
// API: oauth/access_token
func GetAccess(c auth.Config, reqToken, verifier string, opts *AccessOpts) AccessQuery {
	req := &jhttp.Request{
		Method:     "oauth/access_token",
		HTTPMethod: "POST",
		Params: jhttp.Params{
			"oauth_token":    []string{reqToken},
			"oauth_verifier": []string{verifier},
		},
	}
	return AccessQuery{Request: req}
}

// An AccessQuery is a query for an access token.
type AccessQuery struct {
	*jhttp.Request
}

// Invoke issues the query and returns the access Token.
func (a AccessQuery) Invoke(ctx context.Context, cli *twitter.Client) (AccessToken, error) {
	data, err := cli.CallRaw(ctx, a.Request)
	if err != nil {
		return AccessToken{}, err
	}
	tok, err := url.ParseQuery(string(data))
	if err != nil {
		return AccessToken{}, &jhttp.Error{Message: "parsing response", Err: err}
	}
	return AccessToken{
		Token: Token{
			Key:    tok.Get("oauth_token"),
			Secret: tok.Get("oauth_token_secret"),
		},
		UserID:   tok.Get("user_id"),
		Username: tok.Get("screen_name"),
	}, nil
}

// AccessOpts provides optional values for an access-token request.
// A nil *AccessOpts provides empty values for all fields.
type AccessOpts struct{}

// A Token carries a token key and its corresponding secret.
type Token struct {
	Key    string
	Secret string
}

// An AccessToken is a Token with optional user identification data.
type AccessToken struct {
	Token
	UserID   string
	Username string
}

// GetBearer constructs a query to obtain an OAuth2 bearer token.
//
// Bearer token requests are authenticated using c.APIKey and c.APISecret.
// This query does not require c.AccessToken or c.AccessTokenSecret.
//
// API: oauth2/token
func GetBearer(c auth.Config, opts *BearerOpts) BearerQuery {
	req := &jhttp.Request{
		Method:     "oauth2/token",
		HTTPMethod: "POST",
		Params: jhttp.Params{
			"grant_type": []string{"client_credentials"},
			// This is the only grant type currently supported, but the parameter
			// is required to be set.
			//
			// See https://developer.twitter.com/en/docs/authentication/api-reference/token
		},
	}
	return BearerQuery{Request: req, user: c.APIKey, password: c.APISecret}
}

// A BearerQuery is a query for an OAuth 2 bearer token.
type BearerQuery struct {
	*jhttp.Request
	user, password string
}

// BearerOpts provides optional values for a bearer-token request.
// A nil *BearerOpts provides empty values for all fields.
type BearerOpts struct{}

// Invoke issues the query and returns the bearer token.
func (q BearerQuery) Invoke(ctx context.Context, cli *twitter.Client) (Token, error) {
	data, err := clientWithAuth(cli, func(hreq *http.Request) error {
		hreq.SetBasicAuth(url.QueryEscape(q.user), url.QueryEscape(q.password))
		return nil
	}).CallRaw(ctx, q.Request)
	if err != nil {
		return Token{}, err
	}

	var wrapper struct {
		Type  string `json:"token_type"`
		Token string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return Token{}, &jhttp.Error{Message: "decoding token", Err: err}
	}
	return Token{
		Key:    wrapper.Type,
		Secret: wrapper.Token,
	}, nil
}

// InvalidateAccess constructs a query to invalidate an access token.
// This query does not use c.AccessToken or c.AccessTokenSecret.
//
// API: oauth/invalidate_token
func InvalidateAccess(c auth.Config, token, secret string) InvalidateQuery {
	return InvalidateQuery{
		Request: &jhttp.Request{
			Method:     "1.1/oauth/invalidate_token",
			HTTPMethod: "POST",
		},
		authorize: c.Authorizer(token, secret),
	}
}

// InvalidateBearer constructs a query to invalidate a bearer token.
//
// This query requires c.AccessToken and c.AccessTokenSecret to be the access
// token of the owner of the bearer token.
//
// API: oauth2/invalidate_token
func InvalidateBearer(c auth.Config, bearerToken string) InvalidateQuery {
	return InvalidateQuery{
		Request: &jhttp.Request{
			Method:     "oauth2/invalidate_token",
			HTTPMethod: "POST",

			// For reasons I don't fully understand, this query does not work with
			// the token stored in the URL parameters; the server reports a 403.
			// I thought it was maybe escaping, but unescaping and over-escaping
			// didn't seem to help. The body works, though. 🤷

			Data:        []byte("access_token=" + bearerToken),
			ContentType: "application/x-www-form-urlencoded",
		},
		authorize: c.Authorize,
	}
}

// InvalidateQuery is a query for a token invalidation request.
type InvalidateQuery struct {
	*jhttp.Request
	authorize jhttp.Authorizer
}

// Invoke issues the query and returns the invalidated token.
func (q InvalidateQuery) Invoke(ctx context.Context, cli *twitter.Client) (string, error) {
	data, err := clientWithAuth(cli, q.authorize).CallRaw(ctx, q.Request)
	if err != nil {
		return "", err
	}
	var tok struct {
		Token string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &tok); err != nil {
		return "", &jhttp.Error{Message: "decoding token", Err: err}
	}
	return tok.Token, nil
}
