// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-ozzo/ozzo-routing/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// JWTOptions.Leeway allows clock skew on exp, nbf and iat; without it there is none.
func TestJWTLeeway(t *testing.T) {
	now := time.Now().Unix()
	tokens := map[string]string{
		"expired 30s ago":         mustNewJWT(t, jwt.MapClaims{"exp": now - 30}),
		"valid from 30s from now": mustNewJWT(t, jwt.MapClaims{"nbf": now + 30}),
		"issued 30s from now":     mustNewJWT(t, jwt.MapClaims{"iat": now + 30}),
	}
	expired := mustNewJWT(t, jwt.MapClaims{"exp": now - 120})
	for _, test := range []struct {
		id     string
		leeway time.Duration
		token  string
		ok     bool
	}{
		{"expired, no leeway", 0, tokens["expired 30s ago"], false},
		{"expired, within leeway", time.Minute, tokens["expired 30s ago"], true},
		{"expired, beyond leeway", time.Minute, expired, false},
		{"nbf, no leeway", 0, tokens["valid from 30s from now"], false},
		{"nbf, within leeway", time.Minute, tokens["valid from 30s from now"], true},
		{"iat, no leeway", 0, tokens["issued 30s from now"], false},
		{"iat, within leeway", time.Minute, tokens["issued 30s from now"], true},
	} {
		h := JWT(leewaySecret, JWTOptions{Leeway: test.leeway})
		req, _ := http.NewRequest("GET", "/users/", nil)
		req.Header.Set("Authorization", "Bearer "+test.token)
		err := h(routing.NewContext(httptest.NewRecorder(), req))
		if test.ok {
			assert.Nil(t, err, test.id)
		} else {
			assert.NotNil(t, err, test.id)
		}
	}
}

const leewaySecret = "secret-key"

// mustNewJWT signs claims with leewaySecret.
func mustNewJWT(t *testing.T, claims jwt.MapClaims) string {
	s, err := NewJWT(claims, leewaySecret)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// unsignedToken builds a token with the given header and claims JSON and no valid signature.
func unsignedToken(header, claims, signature string) string {
	enc := base64.RawURLEncoding
	return enc.EncodeToString([]byte(header)) + "." + enc.EncodeToString([]byte(claims)) + "." + signature
}

// TestJWTValidation pins which tokens auth.JWT accepts. These are the rules of the
// golang-jwt v3 era (exp, nbf and iat are all checked; the signing method is pinned),
// and they must hold whichever JWT library the handler uses.
func TestJWTValidation(t *testing.T) {
	const secret = "secret-key"
	now := time.Now().Unix()
	sign := func(claims jwt.MapClaims, key string, method ...jwt.SigningMethod) string {
		s, err := NewJWT(claims, key, method...)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	valid := sign(jwt.MapClaims{"id": "100"}, secret)
	dot := strings.LastIndex(valid, ".")

	tests := []struct {
		id    string
		token string
		ok    bool
	}{
		{"no time claims", valid, true},
		{"exp in the future", sign(jwt.MapClaims{"exp": now + 3600}, secret), true},
		{"exp in the past", sign(jwt.MapClaims{"exp": now - 3600}, secret), false},
		{"exp not a number", sign(jwt.MapClaims{"exp": "tomorrow"}, secret), false},
		{"exp zero means unset", sign(jwt.MapClaims{"exp": 0}, secret), true},
		{"nbf in the past", sign(jwt.MapClaims{"nbf": now - 3600}, secret), true},
		{"nbf in the future", sign(jwt.MapClaims{"nbf": now + 3600}, secret), false},
		{"iat in the past", sign(jwt.MapClaims{"iat": now - 3600}, secret), true},
		{"iat in the future", sign(jwt.MapClaims{"iat": now + 3600}, secret), false},
		{"wrong key", sign(jwt.MapClaims{"id": "100"}, "other-key"), false},
		{"other HMAC method", sign(jwt.MapClaims{"id": "100"}, secret, jwt.SigningMethodHS512), false},
		{"alg none", unsignedToken(`{"alg":"none","typ":"JWT"}`, `{"id":"100"}`, ""), false},
		{"tampered claims", unsignedToken(`{"alg":"HS256","typ":"JWT"}`, `{"id":"1"}`, valid[dot+1:]), false},
		{"signature cut", valid[:dot+1], false},
		{"garbage segments", "a.b.c", false},
		{"two segments", "a.b", false},
		{"empty", "", false},
	}
	for _, test := range tests {
		h := JWT(secret)
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/", nil)
		req.Header.Set("Authorization", "Bearer "+test.token)
		c := routing.NewContext(res, req)
		err := h(c)
		if test.ok {
			assert.Nil(t, err, test.id)
			assert.NotNil(t, c.Get("JWT"), test.id)
			continue
		}
		if assert.NotNil(t, err, test.id) {
			assert.Equal(t, http.StatusUnauthorized, err.(routing.HTTPError).StatusCode(), test.id)
		}
		assert.Equal(t, `Bearer realm="API"`, res.Header().Get("WWW-Authenticate"), test.id)
		assert.Nil(t, c.Get("JWT"), test.id)
	}
}
