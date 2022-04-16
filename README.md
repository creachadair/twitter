# twitter

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/creachadair/twitter)

This repository provides Go package that implements a [Twitter API v2][tv2]
client.

This is a work in progress, and is not ready for production use. In particular,
test coverage is not nearly as good as it should be. There are replay tests
using responses captured from the production service, but a lot of invariants
remain unverified.

The documentation could also use more work. All the packages and exported types
have doc comments, but working examples are lacking.  Normal test-based
examples are tricky because there is not (as far as I know) a good test double
for the API; however, some command-line tools would help.

I plan to improve on all of these, but in the meantime I do not recommend using
this library for serious work. Please feel free to file issues, however.  The
library API is very much subject to change.

[tv2]: https://developer.twitter.com/en/docs/twitter-api

## API Index

Here is the current status of v2 API endpoint implementations.

### Lists

- [x] GET 2/lists/:id/followers
- [x] DELETE 2/lists/:id/members/:userid
- [x] GET 2/lists/:id/members
- [x] POST 2/lists/:id/members
- [x] GET 2/lists/:id/tweets
- [x] DELETE 2/lists/:id
- [x] PUT 2/lists/:id
- [x] GET 2/lists
- [x] POST 2/lists

### Tweets

- [ ] PUT 2/tweets/:id/hidden
- [x] GET 2/tweets/:id/liking_users
- [ ] GET 2/tweets/:id/quote_tweets
- [ ] DELETE 2/tweets/:id
- [ ] GET 2/tweets/count/all
- [ ] GET 2/tweets/count/recent
- [x] GET 2/tweets/sample/stream
- [ ] GET 2/tweets/search/all
- [x] GET 2/tweets/search/recent
- [x] POST 2/tweets/search/stream/rules, dry_run=true
- [x] GET 2/tweets/search/stream/rules
- [x] POST 2/tweets/search/stream/rules
- [x] GET 2/tweets/search/stream
- [x] GET 2/tweets
- [x] POST 2/tweets

### Users

- [ ] DELETE 2/users/:id/bookmarks/:tid
- [ ] GET 2/users/:id/bookmarks
- [ ] POST 2/users/:id/bookmarks
- [x] GET 2/users/:id/followed_lists
- [x] GET 2/users/:id/followers
- [x] GET 2/users/:id/following
- [x] GET 2/users/:id/liked_tweets
- [ ] DELETE 2/users/:id/likes/:tid
- [ ] POST 2/users/:id/likes
- [x] GET 2/users/:id/list_memberships
- [ ] GET 2/users/:id/mentions
- [x] GET 2/users/:id/owned_lists
- [ ] GET 2/users/:id/retweeted_by
- [ ] DELETE 2/users/:id/retweets/:tid
- [ ] GET 2/users/:id/retweets
- [ ] GET 2/users/:id/tweets
- [x] GET 2/users/by
- [x] GET 2/users
