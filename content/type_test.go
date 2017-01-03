// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package content

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
)

func TestJSONFormatter(t *testing.T) {
	res := httptest.NewRecorder()
	w := &JSONDataWriter{}
	w.SetHeader(res)
	err := w.Write(res, "xyz")
	assert.Nil(t, err)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "\"xyz\"\n", res.Body.String())
}

func TestXMLFormatter(t *testing.T) {
	res := httptest.NewRecorder()
	w := &XMLDataWriter{}
	w.SetHeader(res)
	err := w.Write(res, "xyz")
	assert.Nil(t, err)
	assert.Equal(t, "application/xml; charset=UTF-8", res.Header().Get("Content-Type"))
	assert.Equal(t, "<string>xyz</string>", res.Body.String())
}

func TestHTMLFormatter(t *testing.T) {
	res := httptest.NewRecorder()
	w := &HTMLDataWriter{}
	w.SetHeader(res)
	err := w.Write(res, "xyz")
	assert.Nil(t, err)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header().Get("Content-Type"))
	assert.Equal(t, "xyz", res.Body.String())
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
	assert.Equal(t, "\"xyz\"\n", res.Body.String())

	assert.Panics(t, func() {
		TypeNegotiator("unknown")
	})
}

var (
	v1JSON = "application/json;v=1"
	v2JSON = "application/json;v=2"
)

type JSONDataWriter1 struct {
	JSONDataWriter
}

func (w *JSONDataWriter1) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", v1JSON)
}

type JSONDataWriter2 struct {
	JSONDataWriter
}

func (w *JSONDataWriter2) SetHeader(res http.ResponseWriter) {
	res.Header().Set("Content-Type", v2JSON)
}

func TestTypeNegotiatorWithVersion(t *testing.T) {

	req, _ := http.NewRequest("GET", "/users/", nil)
	req.Header.Set("Accept", "application/xml,"+v1JSON)

	// test no arguments
	res := httptest.NewRecorder()
	c := routing.NewContext(res, req)
	h := TypeNegotiator()
	assert.Nil(t, h(c))
	c.Write("xyz")
	assert.Equal(t, "text/html; charset=UTF-8", res.Header().Get("Content-Type"))
	assert.Equal(t, "xyz", res.Body.String())

	DataWriters[v1JSON] = &JSONDataWriter1{}
	DataWriters[v2JSON] = &JSONDataWriter2{}

	// test format chosen based on Accept
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	h = TypeNegotiator(v2JSON, v1JSON, XML)
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, "application/json;v=1", res.Header().Get("Content-Type"))
	assert.Equal(t, `"xyz"`+"\n", res.Body.String())

	// test default format used when no match
	req.Header.Set("Accept", "application/pdf")
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	assert.Nil(t, h(c))
	assert.Nil(t, c.Write("xyz"))
	assert.Equal(t, v2JSON, res.Header().Get("Content-Type"))
	assert.Equal(t, "\"xyz\"\n", res.Body.String())

	assert.Panics(t, func() {
		TypeNegotiator("unknown")
	})
}
