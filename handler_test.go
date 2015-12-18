// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"testing"
	"fmt"
	"net/http/httptest"
	"net/http"
)

type LoggerMock struct {
	message string
}

func (l *LoggerMock) Error(format string, a ...interface{}) {
	l.message = fmt.Sprintf(format, a...)
}

func TestErrorHandler(t *testing.T) {
	res := httptest.NewRecorder()
	c := NewContext(res, nil)
	l := &LoggerMock{}

	c.Error = NewHTTPError(http.StatusNotFound)
	h := ErrorHandler(l.Error).(func(*Context) HTTPError)
	e := h(c)
	if e.Code() != http.StatusNotFound {
		t.Errorf("Expected error status %v, got %v", http.StatusNotFound, e.Code())
	}
	if res.Code != http.StatusNotFound {
		t.Errorf("Expected response status %v, got %v", http.StatusNotFound, res.Code)
	}

	res = httptest.NewRecorder()
	c.Error = "xyz"
	c.Response = res
	e2 := h(c)
	if e2.Code() != http.StatusInternalServerError {
		t.Errorf("Expected error status %v, got %v", http.StatusInternalServerError, e.Code())
	}
	if res.Code != http.StatusInternalServerError {
		t.Errorf("Expected response status %v, got %v", http.StatusInternalServerError, res.Code)
	}
	if l.message != "xyz" {
		t.Errorf("Expected log message %q, got %q", "xyz", l.message)
	}
}

func TestNotFoundHandler(t *testing.T) {
	h := NotFoundHandler().(func())
	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("Expected error not received")
		}
		httpError, ok := err.(HTTPError)
		if !ok || httpError.Code() != http.StatusNotFound {
			t.Errorf("Got an unexpected error")
		}
	}()

	h()
}

func TestTrailingSlashRemover(t *testing.T) {
}

func TestAccessLogger(t *testing.T) {
}
