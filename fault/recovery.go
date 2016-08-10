// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fault provides a panic and error handler for the ozzo routing package.
package fault

import "github.com/go-ozzo/ozzo-routing"

type (
	// LogFunc logs a message using the given format and optional arguments.
	// The usage of format and arguments is similar to that for fmt.Printf().
	// LogFunc should be thread safe.
	LogFunc func(format string, a ...interface{})

	// HandleErrorFunc is called whenever a panic or error is captured by the middleware.
	HandleErrorFunc func(c *routing.Context, err error, log LogFunc)
)

// Recovery returns a handler that handles both panics and errors occurred while servicing an HTTP request.
// Recovery can be considered as a combination of ErrorHandler and PanicHandler.
//
// The handler will recover from panics and render the recovered error or the error returned by a handler.
// If the error implements routing.HTTPError, the handler will set the HTTP status code accordingly.
// Otherwise the HTTP status is set as http.StatusInternalServerError. The handler will also write the error
// as the response body.
//
// A log function can be provided to log a message whenever an error is handled. If nil, no message will be logged.
//
//     import (
//         "log"
//         "github.com/go-ozzo/ozzo-routing"
//         "github.com/go-ozzo/ozzo-routing/fault"
//     )
//
//     r := routing.New()
//     r.Use(fault.Recovery(log.Printf))
func Recovery(logf LogFunc) routing.Handler {
	handlePanic := PanicHandler(logf)
	return func(c *routing.Context) error {
		if err := handlePanic(c); err != nil {
			if logf != nil {
				logf("%v", err)
			}
			writeError(c, err)
			c.Abort()
		}
		return nil
	}
}
