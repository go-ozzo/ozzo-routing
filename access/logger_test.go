// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package access

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	h := Logger(getLogger(&buf))

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
	c := routing.NewContext(res, req, h, handler1)
	assert.NotNil(t, c.Next())
	assert.Contains(t, buf.String(), "GET /users")
}

func TestLogResponseWriter(t *testing.T) {
	res := httptest.NewRecorder()
	w := &LogResponseWriter{res, 0, 0}
	w.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, http.StatusBadRequest, w.Status)
	n, _ := w.Write([]byte("test"))
	assert.Equal(t, 4, n)
	assert.Equal(t, int64(4), w.BytesWritten)
	assert.Equal(t, "test", res.Body.String())
}

func TestGetClientIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users/", nil)
	req.Header.Set("X-Real-IP", "192.168.100.1")
	req.Header.Set("X-Forwarded-For", "192.168.100.2")
	req.RemoteAddr = "192.168.100.3"

	assert.Equal(t, "192.168.100.1", getClientIP(req))
	req.Header.Del("X-Real-IP")
	assert.Equal(t, "192.168.100.2", getClientIP(req))
	req.Header.Del("X-Forwarded-For")
	assert.Equal(t, "192.168.100.3", getClientIP(req))

	req.RemoteAddr = "192.168.100.3:8080"
	assert.Equal(t, "192.168.100.3", getClientIP(req))
}

func getLogger(buf *bytes.Buffer) LogFunc {
	return func(format string, a ...interface{}) {
		fmt.Fprintf(buf, format, a...)
	}
}

func handler1(c *routing.Context) error {
	return errors.New("abc")
}
