// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package access provides an access logging handler for the ozzo routing package.
package access

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-ozzo/ozzo-routing"
)

// LogFunc logs a message using the given format and optional arguments.
// The usage of format and arguments is similar to that for fmt.Printf().
// LogFunc should be thread safe.
type LogFunc func(format string, a ...interface{})

// Logger returns a handler that logs a message for every request.
// The access log messages contain information including client IPs, time used to serve each request, request line,
// response status and size.
//
//     import (
//         "log"
//         "github.com/go-ozzo/ozzo-routing"
//         "github.com/go-ozzo/ozzo-routing/access"
//     )
//
//     r := routing.New()
//     r.Use(access.Logger(log.Printf))
func Logger(log LogFunc) routing.Handler {
	return func(c *routing.Context) error {
		startTime := time.Now()

		req := c.Request
		rw := &LogResponseWriter{c.Response, http.StatusOK, 0}
		c.Response = rw

		err := c.Next()

		clientIP := getClientIP(req)
		elapsed := float64(time.Now().Sub(startTime).Nanoseconds()) / 1e6
		requestLine := fmt.Sprintf("%s %s %s", req.Method, req.URL.Path, req.Proto)
		log(`[%s] [%.3fms] %s %d %d`, clientIP, elapsed, requestLine, rw.Status, rw.BytesWritten)

		return err
	}
}

// LogResponseWriter wraps http.ResponseWriter in order to capture HTTP status and response length information.
type LogResponseWriter struct {
	http.ResponseWriter
	Status       int
	BytesWritten int64
}

func (r *LogResponseWriter) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.BytesWritten += int64(written)
	return written, err
}

func (r *LogResponseWriter) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
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
