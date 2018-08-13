// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package content

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ltick/tick-routing"
	"github.com/stretchr/testify/assert"
)

func TestLanguageNegotiator(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users/", nil)
	req.Header.Set("Accept-Language", "ru-RU;q=0.6,ru;q=0.5,zh-CN;q=1.0,zh;q=0.9")

	// test no arguments
	res := httptest.NewRecorder()
	c := routing.NewContext(res, req)
	h := LanguageNegotiator()
	err := h(c)
	assert.Nil(t, err)
	assert.Equal(t, "en-US", c.Get(Language))

	h = LanguageNegotiator("ru-RU", "ru", "zh", "zh-CN")
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "zh-CN", c.Get(Language))

	h = LanguageNegotiator("en", "en-US")
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "en", c.Get(Language))

	req.Header.Set("Accept-Language", "ru-RU;q=0")
	h = LanguageNegotiator("en", "ru-RU")
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "en", c.Get(Language))
}
