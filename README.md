# twitter

http://godoc.org/github.com/creachadair/twitter

[![Go Report Card](https://goreportcard.com/badge/github.com/creachadair/twitter)](https://goreportcard.com/report/github.com/creachadair/twitter)

This repository provides Go package that implements a [Twitter API v2][tv2]
client.

This is a work in progress, and is not ready for production use. In particular:

- ~Not All the API endpoints are supported yet.~
  - [x] tweets
  - [x] tweets/sample/stream
  - [x] tweets/search/recent
      - [x] pagination
  - [x] tweets/search/stream
  - [x] GET tweets/search/stream/rules
  - [x] POST tweets/search/stream/rules
  - [x] users
  - [x] users/by

- There is very little test coverage (only a few manual smoke tests).
- The documentation is still incomplete.
  - [x] Doc comment for package `rules`
  - [ ] Executable examples for package `rules`
  - [x] Doc comment for package `tweets`
  - [ ] Executable examples for package `tweets`
  - [ ] Doc comment for package `types`
  - [ ] Executable examples for package `types`
  - [x] Doc comment for package `users`
  - [ ] Executable examples for package `users`

I plan to improve on all of these, but in the meantime I do not recommend using
this library for serious work. Please feel free to file issues, however.  The
library API is very much subject to change.

[tv2]: https://developer.twitter.com/en/docs/twitter-api/early-access
