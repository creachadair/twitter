// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package otest carries some shared setup code for unit tests.
package otest

import (
	"flag"
	"os"
	"testing"

	"github.com/creachadair/twitter"
	"github.com/creachadair/twitter/jape"
)

var (
	doVerboseLog = flag.Bool("verbose-log", false, "Enable verbose client logging")
)

// GetOrSkip returns the value of the specified environment variable, marking t
// as skipped if the value is unset or empty.
func GetOrSkip(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skip("Missing " + key + " in environment; skipping this test")
	}
	return val
}

// NewClient is a wrapper for twitter.NewClient that sets up debug logging if
// enabled by the -verbose-log flag.
func NewClient(t *testing.T, opts *jape.Client) *twitter.Client {
	cli := twitter.NewClient(opts)
	if *doVerboseLog {
		cli.Log = func(tag jape.LogTag, msg string) {
			t.Logf("DEBUG :: %s | %s", tag, msg)
		}
	}
	return cli
}
