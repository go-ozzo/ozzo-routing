// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"testing"
	"net/http"
)

func TestNewHttpError(t *testing.T) {
	e := NewHTTPError(http.StatusNotFound)
	if e.Code() != http.StatusNotFound {
		t.Errorf("Context.Error(http.StatusNotFound).ErrorCode = %v, want %v", e.Code(), http.StatusNotFound)
	}
	if e.Error() != http.StatusText(http.StatusNotFound) {
		t.Errorf("Context.Error(http.StatusNotFound).ErrorMessage %q, want %q", e.Error(), http.StatusText(http.StatusNotFound))
	}

	e = NewHTTPError(http.StatusNotFound, "abc")
	if e.Code() != http.StatusNotFound {
		t.Errorf("Context.Error(http.StatusNotFound).ErrorCode %v, want %v", e.Code(), http.StatusNotFound)
	}
	if e.Error() != "abc" {
		t.Errorf("Context.Error(http.StatusNotFound).ErrorMessage %q, want %q", e.Error(), "abc")
	}
}
