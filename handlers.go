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
	"os"
	"path/filepath"
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
	return func(c *Context) {
		if err, ok := c.Error.(HTTPError); ok {
			c.Response.WriteHeader(err.Code())
			c.Write(err)
			return
		}
		if f != nil {
			f("%v", c.Error)
		}
		c.Response.WriteHeader(http.StatusInternalServerError)
		c.Write(NewHTTPError(http.StatusInternalServerError))
	}
}

// NotFoundHandler returns a handler that triggers an HTTPError with the status http.StatusNotFound.
//
// This handler is usually used as one of the last handlers for a router.
func NotFoundHandler() Handler {
	return func(*Context) {
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

// AccessLogger returns a handler that logs a message for every request.
// The access log messages contain information including client IPs, time used to serve each request, request line,
// response status and size.
func AccessLogger(log LogFunc) Handler {
	var mu sync.Mutex
	return func(c *Context) {
		startTime := time.Now()

		req := c.Request
		rw := &logResponseWriter{c.Response, http.StatusOK, 0}
		c.Response = rw

		c.Next()

		clientIP := getClientIP(req)
		elapsed := float64(time.Now().Sub(startTime).Nanoseconds()) / 1e6
		requestLine := fmt.Sprintf("%s %s %s", req.Method, req.RequestURI, req.Proto)
		mu.Lock()
		defer mu.Unlock()
		log(`[%s] [%.3fms] %s %d %d`, clientIP, elapsed, requestLine, rw.status, rw.bytesWritten)
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

// StaticOptions defines the possible options for StaticFolder handler.
type StaticOptions struct {
	// The prefix in the URL path that should not be considered as part of the file path.
	// For example, if the URL path is "/foo/bar/index.html" and the Prefix is set as "/foo",
	// then the file "/bar/index.html" would be served.
	Prefix    string
	// The file (e.g. index.html) to be served when the current request corresponds to a directory.
	// It is defaulted to "index.html". If the file does not exist, the handler will pass the control
	// to the next available handler.
	IndexFile string
	// A function that checks if the requested file path is allowed. If allowed, the function
	// may do additional work such as setting Expires HTTP header.
	// Note that if the requested file path is not allowed, the function should decide whether to
	// call Context.Next() to pass the control to the next available handler.
	Allow     func(*Context, string) bool
}

// Static returns a handler that serves the files under the specified folder as response content.
// For example, if root is "static" and the handler is handling the URL path "/app/index.html",
// then the content of the file "<working dir>/static/app/index.html" may be served as the response.
func Static(root string, opts ...StaticOptions) Handler {
	if !filepath.IsAbs(root) {
		root = filepath.Join(RootPath, root)
	}
	options := StaticOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}
	if options.IndexFile == "" {
		options.IndexFile = "index.html"
	}

	// limit the files to be served within the specified folder
	dir := http.Dir(root)

	return func(c *Context) {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if options.Prefix != "" {
			if !strings.HasPrefix(path, options.Prefix) {
				c.Next()
				return
			}
			path = path[len(options.Prefix):]
			if path != "" && path[0] != '/' {
				c.Next()
				return
			}
		}
		if options.Allow != nil && !options.Allow(c, path) {
			return
		}

		var (
			file http.File
			fstat os.FileInfo
			err error
		)

		if file, err = dir.Open(path); err != nil {
			c.Next()
			return
		}
		defer file.Close()

		if fstat, err = file.Stat(); err != nil {
			c.Next()
			return
		}

		// if it's a directory, try the index file
		if fstat.IsDir() {
			path = filepath.Join(path, options.IndexFile)
			if file, err = dir.Open(path); err != nil {
				c.Next()
				return
			}
			defer file.Close()

			if fstat, err = file.Stat(); err != nil || fstat.IsDir() {
				c.Next()
				return
			}
		}

		http.ServeContent(c.Response, c.Request, path, fstat.ModTime(), file)
	}
}

// StaticFile returns a handler that serves the content of the specified file as the response.
// If the specified file does not exist, the handler will pass the control to the next available handler.
func StaticFile(path string) Handler {
	if !filepath.IsAbs(path) {
		path = filepath.Join(RootPath, path)
	}
	return func(c *Context) {
		if file, err := os.Open(path); err == nil {
			if fs, err2 := file.Stat(); err2 == nil {
				http.ServeContent(c.Response, c.Request, path, fs.ModTime(), file)
				return
			}
		}
		c.Next()
	}
}
