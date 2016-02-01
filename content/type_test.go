// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package content

import (
	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	res := httptest.NewRecorder()
	w := JSONFormatter(res)
	bytes, err := w("xyz")
	assert.Nil(t, err)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "\"xyz\"", string(bytes))
}

func TestXMLFormatter(t *testing.T) {
	res := httptest.NewRecorder()
	w := XMLFormatter(res)
	bytes, err := w("xyz")
	assert.Nil(t, err)
	assert.Equal(t, "application/xml; charset=UTF-8", res.Header().Get("Content-Type"))
	assert.Equal(t, "<string>xyz</string>", string(bytes))
}

func TestHTMLFormatter(t *testing.T) {
	res := httptest.NewRecorder()
	w := HTMLFormatter(res)
	bytes, err := w("xyz")
	assert.Nil(t, err)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header().Get("Content-Type"))
	assert.Equal(t, "xyz", string(bytes))
}

func TestTypeNegotiator(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users/", nil)
	req.Header.Set("Accept", "application/xml")

	// test no arguments
	res := httptest.NewRecorder()
	c := routing.NewContext(res, req)
	h := TypeNegotiator()
	assert.Nil(t, h(c))
	c.Write("xyz")
	assert.Equal(t, "text/html; charset=UTF-8", res.Header().Get("Content-Type"))
	assert.Equal(t, "xyz", res.Body.String())

	// test format chosen based on Accept
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	h = TypeNegotiator(JSON, XML)
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, "application/xml; charset=UTF-8", res.Header().Get("Content-Type"))
	assert.Equal(t, "<string>xyz</string>", res.Body.String())

	// test default format used when no match
	req.Header.Set("Accept", "application/pdf")
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "\"xyz\"", res.Body.String())

	assert.Panics(t, func() {
		TypeNegotiator("unknown")
	})
}
