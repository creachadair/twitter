// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package query defines a structured builder for search query strings.
package query

import "strings"

// A Query represents a query structure that can be rendered into a query
// string and checked for validity.
type Query interface {
	// Query returns the query in string format.
	String() string

	// Valid reports whether the query is valid, meaning it contains at least
	// one standalone query term. A valid query may contain invalid subqueries.
	Valid() bool
}

// A Builder exports methods to construct query terms. A zero value is ready
// for use.
type Builder struct{}

// New returns a new query builder.
func New() Builder { return Builder{} }

// All matches a conjunction of words, equivalent to And(Word(s) ...).
func (Builder) All(ss ...string) Query { return newAndQuery(words(ss)) }

// Some matches a disjunction of words, equivalent to Or(Word(s) ...).
func (Builder) Some(ss ...string) Query { return newOrQuery(words(ss)) }

// Word converts a string into a keyword term. If s contains spaces it will be
// quoted.
func (Builder) Word(s string) Query { return newWord(s) }

// And matches the conjunction of the specified queries.
func (Builder) And(qs ...Query) Query { return newAndQuery(qs) }

// Or matches the disjunction of the specified queries.
func (Builder) Or(qs ...Query) Query { return newOrQuery(qs) }

// Not matches the negation of the specified query.
func (Builder) Not(q Query) Query { return newNotQuery(q) }

// From matches tweets from the specified user.
func (Builder) From(s string) Query { return solo("from:" + untag("@", s)) }

// To matches tweets that reply to the specified user.
func (Builder) To(s string) Query { return solo("to:" + untag("@", s)) }

// URL matches tweets that contain the specified URL.  The match applies to
// both the plain and expanded URL.
func (Builder) URL(s string) Query { return quoted{tag: "url:", arg: s} }

// Hashtag matches tweets that contain the specified hashtag.
func (Builder) Hashtag(s string) Query { return solo("#" + untag("#", s)) }

// Mention matches tweets that mention the specified username.
func (Builder) Mention(s string) Query { return solo("@" + untag("@", s)) }

// RetweetOf matches retweets of the specified user.
func (Builder) RetweetOf(s string) Query { return solo("retweets_of:" + untag("@", s)) }

// Entity matches tweets containing an entity with the given value.
func (Builder) Entity(s string) Query { return solo("entity:" + s) }

// InThread matches tweets with the specified conversation ID.
func (Builder) InThread(s string) Query { return solo("conversation_id:" + s) }

// IsRetweet matches "natural" retweets (not quoted).
func (Builder) IsRetweet() Query { return nsolo("is:retweet") }

// IsVerified matches tweets whose authors are verified.
func (Builder) IsVerified() Query { return nsolo("is:verified") }

// HasHashtags matches tweets that contain at least one hashtag.
func (Builder) HasHashtags() Query { return nsolo("has:hashtags") }

// HasLinks matches tweets that contain at least one link in its body.
func (Builder) HasLinks() Query { return nsolo("has:links") }

// HasMentions matches tweets that contain at least one mention.
func (Builder) HasMentions() Query { return nsolo("has:mentions") }

// HasMedia matches tweets that contain a recognized media URL.
func (Builder) HasMedia() Query { return nsolo("has:media") }

// HasImages matches tweets that contain a recognized image URL.
func (Builder) HasImages() Query { return nsolo("has:images") }

// HasVideos matches tweets that contain "native" twitter videos.
// This does not include links to video on other sites.
func (Builder) HasVideos() Query { return nsolo("has:videos") }

// Lang matches tweets marked as being in the specified language.  A tweet will
// have at most one language tag assigned.
func (Builder) Lang(s string) Query { return nsolo("lang:" + s) }

type solo string

func (s solo) String() string { return string(s) }
func (solo) Valid() bool      { return true }

type quoted struct {
	tag, arg string
}

func (s quoted) String() string { return s.tag + `"` + s.arg + `"` }
func (quoted) Valid() bool      { return true }

type nsolo string

func (s nsolo) String() string { return string(s) }
func (nsolo) Valid() bool      { return false }

type orQuery []Query

func (q orQuery) String() string { return strings.Join(compile(q), " OR ") }
func (q orQuery) Valid() bool    { return len(q) != 0 && isStandalone(q) }

type andQuery []Query

func (q andQuery) String() string { return strings.Join(compile(q), " ") }
func (q andQuery) Valid() bool    { return len(q) != 0 && isStandalone(q) }

type notQuery struct{ sub Query }

func (q notQuery) String() string {
	switch t := q.sub.(type) {
	case andQuery:
		return newOrQuery(negateAll(t)).String()
	case orQuery:
		return newAndQuery(negateAll(t)).String()
	default:
		return "-" + q.sub.String()
	}
}

func (q notQuery) Valid() bool { return q.sub.Valid() }

func negateAll(qs []Query) []Query {
	neg := make([]Query, len(qs))
	for i, elt := range qs {
		neg[i] = newNotQuery(elt)
	}
	return neg
}

func compile(qs []Query) []string {
	var args []string
	for _, elt := range qs {
		s := elt.String()
		if isCompound(elt) {
			args = append(args, "("+s+")")
		} else {
			args = append(args, s)
		}
	}
	return args
}

func isStandalone(qs []Query) bool {
	for _, elt := range qs {
		if elt.Valid() {
			return true
		}
	}
	return false
}

func isCompound(q Query) bool {
	switch t := q.(type) {
	case andQuery:
		return len(t) > 1
	case orQuery:
		return len(t) > 1
	case notQuery:
		return isCompound(t.sub)
	default:
		return false
	}
}

func untag(tag, s string) string { return strings.TrimPrefix(s, tag) }

func newWord(s string) Query {
	trim := strings.TrimSpace(s)
	if strings.ContainsAny(trim, " \t") {
		return quoted{arg: trim}
	}
	return solo(trim)
}

func words(ss []string) []Query {
	q := make([]Query, len(ss))
	for i, s := range ss {
		q[i] = newWord(s)
	}
	return q
}

func collapse(qs []Query, same func(Query) []Query) []Query {
	var all []Query
	for _, q := range qs {
		if vs := same(q); vs != nil {
			all = append(all, vs...)
		} else {
			all = append(all, q)
		}
	}
	return all
}

func newOrQuery(qs []Query) Query {
	if len(qs) == 1 {
		return qs[0]
	}
	return orQuery(collapse(qs, func(q Query) []Query {
		vs, _ := q.(orQuery)
		return vs
	}))
}

func newAndQuery(qs []Query) Query {
	if len(qs) == 1 {
		return qs[0]
	}
	return andQuery(collapse(qs, func(q Query) []Query {
		vs, _ := q.(andQuery)
		return vs
	}))
}

func newNotQuery(q Query) Query {
	if v, ok := q.(notQuery); ok {
		return v.sub
	}
	return notQuery{q}
}
