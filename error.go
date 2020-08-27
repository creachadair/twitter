// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package twitter

// Error is the concrete type of errors returned by a Call.
type Error struct {
	Message string // a description of the error
	Status  int    // an HTTP status code, if known
	Err     error  // the underlying error, if any
	Data    []byte // the response data from the server, if any
}

// Error satisfies the error interface.
func (e *Error) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return e.Message + ": " + e.Err.Error()
}

// Unwrap satisfies the wrapping interface for the errors package.
func (e *Error) Unwrap() error { return e.Err }
