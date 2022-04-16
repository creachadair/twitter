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

- [ ] `DELETE 2/tweets/:id`
- [ ] `DELETE 2/users/:id/bookmarks/:tid`
- [ ] `DELETE 2/users/:id/likes/:tid`
- [ ] `DELETE 2/users/:id/retweets/:tid`
- [ ] `GET    2/tweets/:id/quote_tweets`
- [ ] `GET    2/tweets/count/all`
- [ ] `GET    2/tweets/count/recent`
- [ ] `GET    2/tweets/search/all`
- [ ] `GET    2/users/:id/bookmarks`
- [ ] `GET    2/users/:id/mentions`
- [ ] `GET    2/users/:id/retweeted_by`
- [ ] `GET    2/users/:id/retweets`
- [ ] `GET    2/users/:id/tweets`
- [ ] `POST   2/users/:id/bookmarks`
- [ ] `POST   2/users/:id/likes`
- [ ] `PUT    2/tweets/:id/hidden`
- [x] `DELETE 2/lists/:id/members/:userid`
- [x] `DELETE 2/lists/:id`
- [x] `GET    2/lists/:id/followers`
- [x] `GET    2/lists/:id/members`
- [x] `GET    2/lists/:id/tweets`
- [x] `GET    2/lists`
- [x] `GET    2/tweets/:id/liking_users`
- [x] `GET    2/tweets/sample/stream`
- [x] `GET    2/tweets/search/recent`
- [x] `GET    2/tweets/search/stream/rules`
- [x] `GET    2/tweets/search/stream`
- [x] `GET    2/tweets`
- [x] `GET    2/users/:id/followed_lists`
- [x] `GET    2/users/:id/followers`
- [x] `GET    2/users/:id/following`
- [x] `GET    2/users/:id/liked_tweets`
- [x] `GET    2/users/:id/list_memberships`
- [x] `GET    2/users/:id/owned_lists`
- [x] `GET    2/users/by`
- [x] `GET    2/users`
- [x] `POST   2/lists/:id/members`
- [x] `POST   2/lists`
- [x] `POST   2/tweets/search/stream/rules, dry_run=true`
- [x] `POST   2/tweets/search/stream/rules`
- [x] `POST   2/tweets`
- [x] `PUT    2/lists/:id`
