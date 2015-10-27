// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"testing"
	"net/http"
)

func TestContextPanic(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Expected error not found")
		}
		if _, ok := err.(HTTPError); !ok {
			t.Error("Expected HttpError not found")
		}
	}()
	c := NewContext(nil, nil)
	c.Panic(http.StatusNotFound)
}
