// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/stretchr/testify/assert"
)

func TestParseBasicAuth(t *testing.T) {
	tests := []struct {
		id         string
		header     string
		user, pass string
	}{
		{"t1", "", "", ""},
		{"t2", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==", "Aladdin", "open sesame"},
		{"t3", "Basic xyz", "", ""},
	}
	for _, test := range tests {
		user, pass := parseBasicAuth(test.header)
		assert.Equal(t, test.user, user, test.id)
		assert.Equal(t, test.pass, pass, test.id)
	}
}

func basicAuth(c *routing.Context, username, password string) (Identity, error) {
	if username == "Aladdin" && password == "open sesame" {
		return "yes", nil
	}
	return nil, errors.New("no")
}

func TestBasic(t *testing.T) {
	h := Basic(basicAuth, "App")
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req)
	err := h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, "no", err.Error())
	}
	assert.Equal(t, `Basic realm="App"`, res.Header().Get("WWW-Authenticate"))
	assert.Nil(t, c.Get(User))

	req, _ = http.NewRequest("GET", "/users/", nil)
	req.Header.Set("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "", res.Header().Get("WWW-Authenticate"))
	assert.Equal(t, "yes", c.Get(User))
}

func TestParseBearerToken(t *testing.T) {
	tests := []struct {
		id     string
		header string
		token  string
	}{
		{"t1", "", ""},
		{"t2", "Bearer QWxhZGRpbjpvcGVuIHNlc2FtZQ==", "Aladdin:open sesame"},
		{"t3", "Bearer xyz", ""},
	}
	for _, test := range tests {
		token := parseBearerAuth(test.header)
		assert.Equal(t, test.token, token, test.id)
	}
}

func bearerAuth(c *routing.Context, token string) (Identity, error) {
	if token == "Aladdin:open sesame" {
		return "yes", nil
	}
	return nil, errors.New("no")
}

func TestBearer(t *testing.T) {
	h := Bearer(bearerAuth, "App")
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/", nil)
	c := routing.NewContext(res, req)
	err := h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, "no", err.Error())
	}
	assert.Equal(t, `Bearer realm="App"`, res.Header().Get("WWW-Authenticate"))
	assert.Nil(t, c.Get(User))

	req, _ = http.NewRequest("GET", "/users/", nil)
	req.Header.Set("Authorization", "Bearer QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "", res.Header().Get("WWW-Authenticate"))
	assert.Equal(t, "yes", c.Get(User))

	req, _ = http.NewRequest("GET", "/users/", nil)
	req.Header.Set("Authorization", "Bearer QW")
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, "no", err.Error())
	}
	assert.Equal(t, `Bearer realm="App"`, res.Header().Get("WWW-Authenticate"))
	assert.Nil(t, c.Get(User))
}

func TestQuery(t *testing.T) {
	h := Query(bearerAuth, "token")
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	c := routing.NewContext(res, req)
	err := h(c)
	if assert.NotNil(t, err) {
		assert.Equal(t, "no", err.Error())
	}
	assert.Nil(t, c.Get(User))

	req, _ = http.NewRequest("GET", "/users?token=Aladdin:open sesame", nil)
	res = httptest.NewRecorder()
	c = routing.NewContext(res, req)
	err = h(c)
	assert.Nil(t, err)
	assert.Equal(t, "", res.Header().Get("WWW-Authenticate"))
	assert.Equal(t, "yes", c.Get(User))
}

func TestJWT(t *testing.T) {
	secret := "secret-key"
	{
		// valid token
		tokenString, err := NewJWT(jwt.MapClaims{
			"id": "100",
		}, secret)
		assert.Nil(t, err)

		h := JWT(secret)
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		c := routing.NewContext(res, req)
		err = h(c)
		assert.Nil(t, err)
		token := c.Get("JWT")
		if assert.NotNil(t, token) {
			assert.Equal(t, "100", token.(*jwt.Token).Claims.(jwt.MapClaims)["id"])
		}
	}

	{
		// invalid signing method
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["name"] = "Qiang"
		claims["admin"] = true
		bearer, _ := token.SignedString([]byte("secret"))

		h := JWT("secret", JWTOptions{
			SigningMethod: "HS512",
		})
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/", nil)
		req.Header.Set("Authorization", "Bearer "+bearer)
		c := routing.NewContext(res, req)
		err := h(c)
		assert.NotNil(t, err)
	}

	{
		// invalid token
		h := JWT("secret")
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/", nil)
		req.Header.Set("Authorization", "Bearer QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
		c := routing.NewContext(res, req)
		err := h(c)
		assert.NotNil(t, err)
		assert.Equal(t, `Bearer realm="API"`, res.Header().Get("WWW-Authenticate"))
		assert.Nil(t, c.Get("JWT"))
	}

	{
		// invalid token with options
		h := JWT("secret", JWTOptions{
			Realm: "App",
		})
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/", nil)
		req.Header.Set("Authorization", "Bearer QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
		c := routing.NewContext(res, req)
		err := h(c)
		assert.NotNil(t, err)
		assert.Equal(t, `Bearer realm="App"`, res.Header().Get("WWW-Authenticate"))
		assert.Nil(t, c.Get("JWT"))
	}
}
