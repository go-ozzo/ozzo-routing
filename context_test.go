// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"errors"
)

func TestContext_Panic(t *testing.T) {
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

type DataResponse struct {
	*httptest.ResponseRecorder
}

func (r *DataResponse) WriteData(data interface{}) error {
	if data == nil {
		return errors.New("cannot be nil")
	}
	s, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.Write(s)
	return nil
}

func TestContext_Write(t *testing.T) {
	res := &DataResponse{httptest.NewRecorder()}
	c := NewContext(res, nil)
	c.Write(100)
	if result := res.Body.String(); result != "100" {
		t.Errorf("Write(100) = %q, expected %q", result, "100")
	}

	res.Body.Reset()
	c.Write("abc")
	if result := res.Body.String(); result != `"abc"` {
		t.Errorf("Write(`abc`) = %q, expected %q", result, "abc")
	}

	defer func() {
		if e := recover(); e == nil {
			t.Errorf("Expected panic not occured")
		}
	}()
	res.Body.Reset()
	c.Write(nil)
}

func TestContext_Write2(t *testing.T) {
	res := httptest.NewRecorder()
	c := NewContext(res, nil)
	c.Write(100)
	if result := res.Body.String(); result != "100" {
		t.Errorf("Write(100) = %q, expected %q", result, "100")
	}

	res.Body.Reset()
	c.Write("abc")
	if result := res.Body.String(); result != "abc" {
		t.Errorf("Write(`abc`) = %q, expected %q", result, "abc")
	}

	res.Body.Reset()
	c.Write([]byte("abc"))
	if result := res.Body.String(); result != "abc" {
		t.Errorf("Write(`abc`) = %q, expected %q", result, "abc")
	}
}
