// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fault provides a panic and error handler for the ozzo routing package.
package fault

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-ozzo/ozzo-routing"
)

type (
	// LogFunc logs a message using the given format and optional arguments.
	// The usage of format and arguments is similar to that for fmt.Printf().
	// LogFunc should be thread safe.
	LogFunc func(format string, a ...interface{})

	// ErrorHandler is called whenever a panic or error is captured by the middleware.
	ErrorHandler func(c *routing.Context, err error, log LogFunc)
)

// Recovery returns a handler that handles panics and errors occurred while servicing an HTTP request.
//
// The handler will recover from panics and render the recovered error or the error returned by a handler.
// If the error is not a routing.HTTPError, it will respond with http.StatusInternalServerError.
// Otherwise, it will use the status code of the HTTPError.
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
func Recovery(log LogFunc, errorHandler ...ErrorHandler) routing.Handler {
	handler := handleError
	if len(errorHandler) > 0 {
		handler = errorHandler[0]
	}
	return func(c *routing.Context) error {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					handler(c, e, log)
				} else {
					handler(c, errors.New(fmt.Sprint(err)), log)
				}
				c.Abort()
			}
		}()

		if err := c.Next(); err != nil {
			handler(c, err, log)
			c.Abort()
		}

		return nil
	}
}

// handleError handles the specified error by rendering it with the context.
func handleError(c *routing.Context, err error, log LogFunc) {
	if log != nil {
		log("%v", err)
	}
	httpError, ok := err.(routing.HTTPError)
	if !ok {
		httpError = routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Response.WriteHeader(httpError.StatusCode())
	c.Write(httpError)
}
