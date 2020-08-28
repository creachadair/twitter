# twitter

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/creachadair/twitter)
[![Go Report Card](https://goreportcard.com/badge/github.com/creachadair/twitter)](https://goreportcard.com/report/github.com/creachadair/twitter)

This repository provides Go package that implements a [Twitter API v2][tv2]
client.

This is a work in progress, and is not ready for production use. In particular:

- There is very little test coverage, mostly smoke tests.
  - [x] Replay tests with captured API data.
  - [x] Manual tests for streaming APIs.
  - [x] Basic CI actions workflow.
  - [x] Make replay tests work for streaming methods.

- The documentation is still incomplete.
  - [x] Doc comment for package `rules`
  - [ ] Executable examples for package `rules`
  - [x] Doc comment for package `tweets`
  - [ ] Executable examples for package `tweets`
  - [x] Doc comment for package `users`
  - [ ] Executable examples for package `users`

I plan to improve on all of these, but in the meantime I do not recommend using
this library for serious work. Please feel free to file issues, however.  The
library API is very much subject to change.

[tv2]: https://developer.twitter.com/en/docs/twitter-api/early-access
