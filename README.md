# twitter

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/creachadair/twitter)

This repository provides Go package that implements an experimental client for
the [Twitter API v2][tv2]. It is no longer under active development.

This was an experiment, and is not ready for production use. In particular,
test coverage is not nearly as good as it should be. There are replay tests
using responses captured from the production service, but a lot of invariants
remain unverified.

The documentation could also use more work. All the packages and exported types
have doc comments, but working examples are lacking.  Normal test-based
examples are tricky because there is not (as far as I know) a good test double
for the API.

[tv2]: https://developer.twitter.com/en/docs/twitter-api

## API Index

Here is the current status of v2 API endpoint implementations.

### Edits

- [x] DELETE 2/tweets/:id
- [x] PUT 2/tweets/:id/hidden
- [x] POST 2/users/:id/blocking
- [x] DELETE 2/users/:id/blocking/:other
- [x] POST 2/users/:id/bookmarks
- [x] DELETE 2/users/:id/bookmarks/:tid
- [x] POST 2/users/:id/following
- [x] DELETE 2/users/:id/following/:other
- [x] POST 2/users/:id/likes
- [x] DELETE 2/users/:id/likes/:tid
- [x] POST 2/users/:id/muting
- [x] DELETE 2/users/:id/muting/:other
- [x] POST 2/users/:id/pinned_lists
- [x] DELETE 2/users/:id/pinned_lists/:lid
- [x] POST 2/users/:id/retweets
- [x] DELETE 2/users/:id/retweets/:tid

### Lists

- [x] GET 2/lists
- [x] POST 2/lists
- [x] DELETE 2/lists/:id
- [x] PUT 2/lists/:id
- [x] GET 2/lists/:id/followers
- [x] GET 2/lists/:id/members
- [x] POST 2/lists/:id/members
- [x] DELETE 2/lists/:id/members/:userid
- [x] GET 2/lists/:id/tweets
- [x] GET 2/lists/:id/pinned_lists

### Rules

- [x] GET 2/tweets/search/stream/rules
- [x] POST 2/tweets/search/stream/rules
- [x] POST 2/tweets/search/stream/rules, dry_run=true

### Tweets

- [x] GET 2/tweets
- [x] POST 2/tweets
- [x] GET 2/tweets/:id/liking_users
- [x] GET 2/tweets/:id/quote_tweets
- [ ] GET 2/tweets/count/all (requires academic access)
- [ ] GET 2/tweets/count/recent
- [x] GET 2/tweets/sample/stream
- [ ] GET 2/tweets/search/all (requires academic access)
- [x] GET 2/tweets/search/recent
- [x] GET 2/tweets/search/stream

### Users

- [x] GET 2/users
- [x] GET 2/users/:id/blocking
- [x] GET 2/users/:id/bookmarks
- [x] GET 2/users/:id/followed_lists
- [x] GET 2/users/:id/followers
- [x] GET 2/users/:id/following
- [x] GET 2/users/:id/liked_tweets
- [x] GET 2/users/:id/list_memberships
- [x] GET 2/users/:id/mentions
- [x] GET 2/users/:id/muting
- [x] GET 2/users/:id/owned_lists
- [x] GET 2/users/:id/retweeted_by
- [x] GET 2/users/:id/tweets
- [x] GET 2/users/by
- [ ] GET 2/users/me
