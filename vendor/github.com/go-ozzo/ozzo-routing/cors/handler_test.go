// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
)

func TestBuildAllowMap(t *testing.T) {
	m := buildAllowMap("", false)
	assert.Equal(t, 0, len(m))

	m = buildAllowMap("", true)
	assert.Equal(t, 0, len(m))

	m = buildAllowMap("GET , put", false)
	assert.Equal(t, 2, len(m))
	assert.True(t, m["GET"])
	assert.True(t, m["PUT"])
	assert.False(t, m["put"])

	m = buildAllowMap("GET , put", true)
	assert.Equal(t, 2, len(m))
	assert.True(t, m["GET"])
	assert.False(t, m["PUT"])
	assert.True(t, m["put"])
}

func TestOptionsInit(t *testing.T) {
	opts := &Options{
		AllowHeaders: "Accept, Accept-Language",
		AllowMethods: "PATCH, PUT",
		AllowOrigins: "https://example.com",
	}
	opts.init()
	assert.Equal(t, 2, len(opts.allowHeaderMap))
	assert.Equal(t, 2, len(opts.allowMethodMap))
	assert.Equal(t, 1, len(opts.allowOriginMap))
}

func TestOptionsIsOriginAllowed(t *testing.T) {
	tests := []struct {
		id      string
		allowed string
		origin  string
		result  bool
	}{
		{"t1", "*", "http://example.com", true},
		{"t2", "null", "http://example.com", false},
		{"t3", "http://foo.com", "http://example.com", false},
		{"t4", "http://example.com", "http://example.com", true},
	}

	for _, test := range tests {
		opts := &Options{AllowOrigins: test.allowed}
		opts.init()
		assert.Equal(t, test.result, opts.isOriginAllowed(test.origin), test.id)
	}
}

func TestOptionsSetOriginHeaders(t *testing.T) {
	headers := http.Header{}
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowCredentials: false,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "https://example.com", headers.Get(headerAllowOrigin))
	assert.Equal(t, "", headers.Get(headerAllowCredentials))

	headers = http.Header{}
	opts = &Options{
		AllowOrigins:     "*",
		AllowCredentials: false,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "*", headers.Get(headerAllowOrigin))
	assert.Equal(t, "", headers.Get(headerAllowCredentials))

	headers = http.Header{}
	opts = &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowCredentials: true,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "https://example.com", headers.Get(headerAllowOrigin))
	assert.Equal(t, "true", headers.Get(headerAllowCredentials))

	headers = http.Header{}
	opts = &Options{
		AllowOrigins:     "*",
		AllowCredentials: true,
	}
	opts.setOriginHeader("https://example.com", headers)
	assert.Equal(t, "https://example.com", headers.Get(headerAllowOrigin))
	assert.Equal(t, "true", headers.Get(headerAllowCredentials))
}

func TestOptionsSetActualHeaders(t *testing.T) {
	headers := http.Header{}
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowCredentials: false,
		ExposeHeaders:    "X-Ping, X-Pong",
	}
	opts.init()
	opts.setActualHeaders("https://example.com", headers)
	assert.Equal(t, "https://example.com", headers.Get(headerAllowOrigin))
	assert.Equal(t, "X-Ping, X-Pong", headers.Get(headerExposeHeaders))

	opts.ExposeHeaders = ""
	headers = http.Header{}
	opts.setActualHeaders("https://example.com", headers)
	assert.Equal(t, "https://example.com", headers.Get(headerAllowOrigin))
	assert.Equal(t, "", headers.Get(headerExposeHeaders))

	headers = http.Header{}
	opts.setActualHeaders("https://bar.com", headers)
	assert.Equal(t, "", headers.Get(headerAllowOrigin))
}

func TestOptionsIsPreflightAllowed(t *testing.T) {
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowMethods:     "PUT, PATCH",
		AllowCredentials: false,
		ExposeHeaders:    "X-Ping, X-Pong",
	}
	opts.init()
	allowed, headers := opts.isPreflightAllowed("https://foo.com", "PUT", "")
	assert.True(t, allowed)
	assert.Equal(t, "", headers)

	opts = &Options{
		AllowOrigins: "https://example.com, https://foo.com",
		AllowMethods: "PUT, PATCH",
	}
	opts.init()
	allowed, headers = opts.isPreflightAllowed("https://foo.com", "DELETE", "")
	assert.False(t, allowed)
	assert.Equal(t, "", headers)

	opts = &Options{
		AllowOrigins: "https://example.com, https://foo.com",
		AllowMethods: "PUT, PATCH",
		AllowHeaders: "X-Ping, X-Pong",
	}
	opts.init()
	allowed, headers = opts.isPreflightAllowed("https://foo.com", "PUT", "X-Unknown")
	assert.False(t, allowed)
	assert.Equal(t, "", headers)
}

func TestOptionsSetPreflightHeaders(t *testing.T) {
	headers := http.Header{}
	opts := &Options{
		AllowOrigins:     "https://example.com, https://foo.com",
		AllowMethods:     "PUT, PATCH",
		AllowHeaders:     "X-Ping, X-Pong",
		AllowCredentials: false,
		ExposeHeaders:    "X-Ping, X-Pong",
		MaxAge:           time.Duration(100) * time.Second,
	}
	opts.init()
	opts.setPreflightHeaders("https://bar.com", "PUT", "", headers)
	assert.Zero(t, len(headers))

	headers = http.Header{}
	opts.setPreflightHeaders("https://foo.com", "PUT", "X-Pong", headers)
	assert.Equal(t, "https://foo.com", headers.Get(headerAllowOrigin))
	assert.Equal(t, "PUT, PATCH", headers.Get(headerAllowMethods))
	assert.Equal(t, "100", headers.Get(headerMaxAge))
	assert.Equal(t, "X-Pong", headers.Get(headerAllowHeaders))

	headers = http.Header{}
	opts = &Options{
		AllowOrigins: "*",
		AllowMethods: "*",
		AllowHeaders: "*",
	}
	opts.init()
	opts.setPreflightHeaders("https://bar.com", "PUT", "X-Pong", headers)
	assert.Equal(t, "*", headers.Get(headerAllowOrigin))
	assert.Equal(t, "PUT", headers.Get(headerAllowMethods))
	assert.Equal(t, "X-Pong", headers.Get(headerAllowHeaders))
}

func TestHandlers(t *testing.T) {
	h := Handler(Options{
		AllowOrigins: "https://example.com, https://foo.com",
		AllowMethods: "PUT, PATCH",
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/users/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	c := routing.NewContext(res, req)
	assert.Nil(t, h(c))
	assert.Equal(t, "https://example.com", res.Header().Get(headerAllowOrigin))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/users/", nil)
	req.Header.Set("Origin", "https://example.com")
	c = routing.NewContext(res, req)
	assert.Nil(t, h(c))
	assert.Equal(t, "https://example.com", res.Header().Get(headerAllowOrigin))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/users/", nil)
	c = routing.NewContext(res, req)
	assert.Nil(t, h(c))
	assert.Equal(t, "", res.Header().Get(headerAllowOrigin))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/users/", nil)
	req.Header.Set("Origin", "https://example.com")
	c = routing.NewContext(res, req)
	assert.Nil(t, h(c))
	assert.Equal(t, "", res.Header().Get(headerAllowOrigin))
}
