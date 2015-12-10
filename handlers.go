// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"net/http"
	"strings"
	"fmt"
	"time"
	"sync"
)

// LogFunc logs a message using the given format and optional arguments.
// The usage of format and arguments is similar to those for fmt.Printf().
type LogFunc func(format string, a ...interface{})

// HTTPHandlerFunc adapts a http.HandlerFunc into a routing.Handler.
func HTTPHandlerFunc(h http.HandlerFunc) Handler {
	return func(c *Context) {
		h(c.Response, c.Request)
	}
}

// HTTPHandler adapts a http.Handler into a routing.Handler.
func HTTPHandler(h http.Handler) Handler {
	return func(c *Context) {
		h.ServeHTTP(c.Response, c.Request)
	}
}

// ErrorHandler returns a handler that handles the error recorded in Context.Error.
//
// If Context.Error is an HTTPError, ErrorHandler will set the response status code
// as the status code specified by Context.Error; Otherwise, ErrorHandler will
// set the status code to be 500 (http.StatusInternalServerError) and log
// the error using the specified LogFunc (if it is not nil).
//
// This handler is usually used as one of the last handlers for a router.
func ErrorHandler(f LogFunc) Handler {
	return func(c *Context) HTTPError {
		if err, ok := c.Error.(HTTPError); ok {
			c.Response.WriteHeader(err.Code())
			return err
		}
		if f != nil {
			f("%v", c.Error)
		}
		c.Response.WriteHeader(http.StatusInternalServerError)
		return NewHTTPError(http.StatusInternalServerError)
	}
}

// NotFoundHandler returns a handler that triggers an HTTPError with the status http.StatusNotFound.
//
// This handler is usually used as one of the last handlers for a router.
func NotFoundHandler() Handler {
	return func() {
		panic(NewHTTPError(http.StatusNotFound))
	}
}

// TrailingSlashRemover returns a handler that removes trailing slashes from the requested URL.
// The handler will redirect the browser to the new URL without trailing slashes.
// The status parameter should be either http.StatusMovedPermanently (301) or http.StatusFound (302).
// If the original URL has no trailing slashes, nothing will be done by this handler.
func TrailingSlashRemover(status int) Handler {
	return func(c *Context) {
		if c.Request.URL.Path != "/" && strings.HasSuffix(c.Request.URL.Path, "/") {
			http.Redirect(c.Response, c.Request, strings.TrimRight(c.Request.URL.Path, "/"), status)
		} else {
			c.Next()
		}
	}
}

// AccessLogger returns a handler that logs an entry for every request.
// The access log entries will be written using the specified writer in the Apache httpd access log format.
func AccessLogger(log LogFunc) Handler {
	var mu sync.Mutex
	return func(c *Context) {
		startTime := time.Now()

		req := c.Request
		rw := &logResponseWriter{c.Response, http.StatusOK, 0}
		c.Response = rw

		c.Next()

		clientIP := getClientIP(req)
		start := startTime.Format("02/Jan/2006 15:04:05 -0700")
		elapsed := float64(time.Now().Sub(startTime).Nanoseconds()) / 1e6
		requestLine := fmt.Sprintf("%s %s %s", req.Method, req.RequestURI, req.Proto)
		mu.Lock()
		defer mu.Unlock()
		log("%s - - [%s] \"%s %d %d\" %.3fms\n", clientIP, start, requestLine, rw.status, rw.bytesWritten, elapsed)
	}
}

func getClientIP(req *http.Request) string {
	ip := req.Header.Get("X-Real-IP")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = req.RemoteAddr
		}
	}
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

type logResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
}

func (r *logResponseWriter) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.bytesWritten += int64(written)
	return written, err
}

func (r *logResponseWriter) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
