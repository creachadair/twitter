// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package query_test

import (
	"fmt"
	"testing"

	"github.com/creachadair/twitter/query"
)

func Example() {
	b := query.New()
	q := b.And(
		b.Or(
			b.All("red", "green", "blue"),
			b.Some("black", "white"),
		),
		b.HasImages(),
		b.Not(b.IsRetweet()),
	)

	fmt.Printf("Valid: %v\nQuery: %s\n", q.Valid(), q.String())
	// Output:
	// Valid: true
	// Query: ((red green blue) OR black OR white) has:images -is:retweet
}

func TestValidQueries(t *testing.T) {
	var b query.Builder

	tests := []struct {
		input query.Query
		want  string
	}{
		{b.Word("cat"), "cat"},
		{b.Word("you are here"), `"you are here"`},
		{b.Not(b.Word("you are gone")), `-"you are gone"`},

		{b.Or(
			b.Word("cat"),
			b.Word("dog"),
		),
			"cat OR dog"},
		{b.Some("cat", "dog"), "cat OR dog"},

		{b.Or(
			b.All("cat", "dog"),
			b.All("sheep", "goat"),
		),
			"(cat dog) OR (sheep goat)"},

		{b.And(
			b.Word("cat"),
			b.Word("dog"),
		), "cat dog"},
		{b.All("cat", "dog"), "cat dog"},

		{b.And(
			b.HasImages(),
			b.Some("cat", "dog"),
		),
			"has:images (cat OR dog)"},

		// Nested logic is flattened.
		{b.Not(b.Not(b.All("x", "y"))), "x y"},

		{b.And(
			b.Word("a"),
			b.All("b", "c"),
			b.And(b.Hashtag("x")),
		),
			"a b c #x"},

		{b.Or(
			b.Word("a"),
			b.Or(b.Word("b")),
			b.Or(
				b.Mention("foo"),
				b.Mention("bar"),
			),
		),
			"a OR b OR @foo OR @bar"},

		// Negated quantities get DeMorganized.
		{b.Not(b.All("cat", "dog")), "-cat OR -dog"},
		{b.Not(b.Some("cat", "dog")), "-cat -dog"},

		{b.Not(b.And(
			b.All("sheep", "goat", "pig"),
			b.Not(b.Word("horny")), // interior negation is correctly undone
			b.Mention("jack"),
		)),
			"-sheep OR -goat OR -pig OR horny OR -@jack"},

		{b.Not(b.Or(
			b.And(b.Word("six"), b.Word("strapping stars")),
			b.Not(b.And(b.HasLinks(), b.InThread("122"))),
		)),
			`(-six OR -"strapping stars") has:links conversation_id:122`},
	}
	for _, test := range tests {
		if !test.input.Valid() {
			t.Errorf("Query: %+v is invalid", test.input)
		}
		got := test.input.String()
		if got != test.want {
			t.Errorf("Query: %+v\ngot:  %s\nwant: %s", test.input, got, test.want)
		}
	}
}

// Verify that queries that do not contain any standalone terms are correctly
// reported as invalid.
func TestInvalidQueries(t *testing.T) {
	var b query.Builder
	tests := []query.Query{
		b.HasMentions(),
		b.And(
			b.IsVerified(),
			b.Not(b.HasMedia()),
		),
		b.Not(b.Or(
			b.Lang("en"),
			b.IsRetweet(),
			b.Not(b.HasVideos()),
		)),
		b.And(), // empty
		b.Or(),  // empty
	}

	for _, test := range tests {
		if test.Valid() {
			t.Errorf("Query: %s\n-- unexpectedly valid", test)
		}
	}
}
