// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fault provides a panic and error handler for the ozzo routing package.
package fault

import (
	"net/http"

	"github.com/go-ozzo/ozzo-routing"
)

// ErrorHandler returns a handler that handles errors returned by the handlers following this one.
// If the error implements routing.HTTPError, the handler will set the HTTP status code accordingly.
// Otherwise the HTTP status is set as http.StatusInternalServerError. The handler will also write the error
// as the response body.
//
// A log function can be provided to log a message whenever an error is handled. If nil, no message will be logged.
//
// An optional error conversion function can also be provided to convert an error into a normalized one
// before sending it to the response.
//
//     import (
//         "log"
//         "github.com/go-ozzo/ozzo-routing"
//         "github.com/go-ozzo/ozzo-routing/fault"
//     )
//
//     r := routing.New()
//     r.Use(fault.ErrorHandler(log.Printf))
//     r.Use(fault.PanicHandler(log.Printf))
func ErrorHandler(logf LogFunc, errorf ...ConvertErrorFunc) routing.Handler {
	return func(c *routing.Context) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		if logf != nil {
			logf("%v", err)
		}

		if len(errorf) > 0 {
			err = errorf[0](c, err)
		}

		writeError(c, err)
		c.Abort()

		return nil
	}
}

// writeError writes the error to the response.
// If the error implements HTTPError, it will set the HTTP status as the result of the StatusCode() call of the error.
// Otherwise, the HTTP status will be set as http.StatusInternalServerError.
func writeError(c *routing.Context, err error) {
	if httpError, ok := err.(routing.HTTPError); ok {
		c.Response.WriteHeader(httpError.StatusCode())
	} else {
		c.Response.WriteHeader(http.StatusInternalServerError)
	}
	c.Write(err)
}
