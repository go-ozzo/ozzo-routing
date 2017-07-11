// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package slash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
)

func TestRemover(t *testing.T) {
	h := Remover(http.StatusMovedPermanently)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req)
	err := h(c)
	assert.Nil(t, err, "return value is nil")
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "/users", res.Header().Get("Location"))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "", res.Header().Get("Location"))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/users", nil)
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "", res.Header().Get("Location"))

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/users/", nil)
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Equal(t, "/users", res.Header().Get("Location"))
}
