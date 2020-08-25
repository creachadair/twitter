// Copyright (C) 2020 Michael J. Fromberger. All Rights Reserved.

package twitter

import "fmt"

// Error is the concrete type of errors returned by a Call.
type Error struct {
	Message string // a description of the error
	Status  int    // an HTTP status code, if known
	Err     error  // the underlying error, if any
	Body    []byte // the response body from the server, if any
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

// Errorf returns an error of concrete type *Error.
func Errorf(data []byte, msg string, args ...interface{}) error {
	var err error
	if len(args) != 0 {
		v, ok := args[len(args)-1].(error)
		if ok {
			err = v
			args = args[:len(args)-1]
		}
	}
	return newErrorf(err, 0, data, msg, args...)
}

func newErrorf(err error, status int, body []byte, msg string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(msg, args...),
		Status:  status,
		Err:     err,
		Body:    body,
	}
}
