// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

// Package types defines types for the Twitter API v2 object model.
package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

//go:generate go run mkenum/mkenum.go -output generated.go

// DateFormat defines the encoding format for a Date string.
const DateFormat = "2006-01-02T15:04:05.999Z"

// A Date defines the JSON encoding of an ISO 8601 date.
type Date time.Time

// String returns d as rendered by DateFormat.
func (d Date) String() string { return time.Time(d).Format(DateFormat) }

// UnmarshalJSON decodes d from a JSON string value.
func (d *Date) UnmarshalJSON(bits []byte) error {
	var s string
	if err := json.Unmarshal(bits, &s); err != nil {
		return fmt.Errorf("cannot decode %q as a date", string(bits))
	}
	ts, err := time.Parse(DateFormat, s)
	if err != nil {
		return fmt.Errorf("invalid date: %v", err)
	}
	*d = Date(ts)
	return nil
}

// MarshalJSON encodes d as a JSON string value.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// Minutes defines the JSON encoding of a duration in minutes.
type Minutes time.Duration

// UnmarshalJSON decodes d from a JSON integer number of minutes.
func (m *Minutes) UnmarshalJSON(bits []byte) error {
	var min int64
	if err := json.Unmarshal(bits, &min); err != nil {
		return fmt.Errorf("cannot decode %q as an integer", string(bits))
	}
	*m = Minutes(time.Duration(min) * time.Minute)
	return nil
}

// MarshalJSON encodes d as a JSON integer number of minutes.  Time intervals
// smaller than a minute are rounded toward zero.
func (m Minutes) MarshalJSON() ([]byte, error) {
	min := strconv.FormatInt(int64(time.Duration(m)/time.Minute), 10)
	return []byte(min), nil
}

// Milliseconds defines the JSON encoding of a duration in milliseconds.
type Milliseconds time.Duration

// UnmarshalJSON decodes d from a JSON integer number of milliseconds.
func (m *Milliseconds) UnmarshalJSON(bits []byte) error {
	var ms int64
	if err := json.Unmarshal(bits, &ms); err != nil {
		return fmt.Errorf("cannot decode %q as an integer", string(bits))
	}
	*m = Milliseconds(time.Duration(ms) * time.Millisecond)
	return nil
}

// MarshalJSON encodes m as a JSON integer number of milliseconds.
func (m Milliseconds) MarshalJSON() ([]byte, error) {
	ms := strconv.FormatInt(int64(time.Duration(m)/time.Millisecond), 10)
	return []byte(ms), nil

}
