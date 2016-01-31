// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fault

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"bytes"
	"fmt"
	"errors"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
)

func TestHandleError(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req)
	var buf bytes.Buffer
	handleError(c, errors.New("abc"), getLogger(&buf))
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req)
	handleError(c, routing.NewHTTPError(http.StatusNotFound, "xyz"), nil)
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, "xyz", res.Body.String())
	assert.Equal(t, "", buf.String())
}

func TestRecovery(t *testing.T) {
	var buf bytes.Buffer
	h := Recovery(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req, h, handler1, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())
	assert.Equal(t, "abc", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "test", res.Body.String())
	assert.Equal(t, "", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h, handler3, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "xyz", res.Body.String())
	assert.Equal(t, "xyz", buf.String())

	buf.Reset()
	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users/", nil)
	c = routing.NewContext(res, req, h, handler4, handler2)
	assert.Nil(t, c.Next())
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "123", res.Body.String())
	assert.Equal(t, "123", buf.String())
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *routing.Context) error {
	return errors.New("abc")
}

func handler2(c *routing.Context) error {
	c.Write("test")
	return nil
}

func handler3(c *routing.Context) error {
	panic("xyz")
	return nil
}

func handler4(c *routing.Context) error {
	panic(routing.NewHTTPError(http.StatusBadRequest, "123"))
	return nil
}
