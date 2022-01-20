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
