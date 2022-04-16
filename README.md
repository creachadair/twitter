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

The following v2 API endpoints are currently implemented:

| Method | Path                                       |
|--------|--------------------------------------------|
| DELETE | 2/lists/:id                                |
| DELETE | 2/lists/:id/members/:userid                |
| GET    | 2/lists                                    |
| GET    | 2/lists/:id/followers                      |
| GET    | 2/lists/:id/members                        |
| GET    | 2/lists/:id/tweets                         |
| POST   | 2/lists                                    |
| POST   | 2/lists/:id/members                        |
| PUT    | 2/lists/:id                                |
| GET    | 2/tweets                                   |
| GET    | 2/tweets/:id/liking_users                  |
| GET    | 2/tweets/sample/stream                     |
| GET    | 2/tweets/search/recent                     |
| GET    | 2/tweets/search/stream                     |
| GET    | 2/tweets/search/stream/rules               |
| POST   | 2/tweets                                   |
| POST   | 2/tweets/search/stream/rules               |
| POST   | 2/tweets/search/stream/rules, dry_run=true |
| GET    | 2/users                                    |
| GET    | 2/users/:id/followed_lists                 |
| GET    | 2/users/:id/followers                      |
| GET    | 2/users/:id/following                      |
| GET    | 2/users/:id/liked_tweets                   |
| GET    | 2/users/:id/list_memberships               |
| GET    | 2/users/:id/owned_lists                    |
| GET    | 2/users/by                                 |

### Not Yet Implemented

| Method | Path                       |
|--------|----------------------------|
| DELETE | 2/tweets/:id               |
| GET    | 2/users/:id/tweets         |
| GET    | 2/users/:id/mentions       |
| GET    | 2/tweets/search/all        |
| GET    | 2/tweets/count/recent      |
| GET    | 2/tweets/count/all         |
| GET    | 2/users/:id/retweeted_by   |
| GET    | 2/users/:id/retweets       |
| DELETE | 2/users/:id/retweets/:tid  |
| GET    | 2/tweets/:id/quote_tweets  |
| POST   | 2/users/:id/likes          |
| DELETE | 2/users/:id/likes/:tid     |
| GET    | 2/users/:id/bookmarks      |
| POST   | 2/users/:id/bookmarks      |
| DELETE | 2/users/:id/bookmarks/:tid |
| PUT    | 2/tweets/:id/hidden        |
