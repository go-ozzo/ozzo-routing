// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routing

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockStore struct {
	*store
	data map[string]interface{}
}

func newMockStore() *mockStore {
	return &mockStore{newStore(), make(map[string]interface{})}
}

func (s *mockStore) Add(key string, data interface{}) int {
	for _, handler := range data.([]Handler) {
		handler(nil)
	}
	return s.store.Add(key, data)
}

func TestRouteNew(t *testing.T) {
	router := New()
	group := newRouteGroup("/admin", router, nil)

	r1 := newRoute("/users", group)
	assert.Equal(t, "/admin/users", r1.name, "route.name =")
	assert.Equal(t, "/admin/users", r1.path, "route.path =")
	assert.Equal(t, "/admin/users", r1.template, "route.template =")
	_, exists := router.routes[r1.name]
	assert.True(t, exists, "router.routes[name] is ")

	r2 := newRoute("/users/<id:\\d+>/*", group)
	assert.Equal(t, "/admin/users/<id:\\d+>/*", r2.name, "route.name =")
	assert.Equal(t, "/admin/users/<id:\\d+>/<:.*>", r2.path, "route.path =")
	assert.Equal(t, "/admin/users/<id>/<>", r2.template, "route.template =")
	_, exists = router.routes[r2.name]
	assert.True(t, exists, "router.routes[name] is ")
}

func TestRouteName(t *testing.T) {
	router := New()
	group := newRouteGroup("/admin", router, nil)

	r1 := newRoute("/users", group)
	assert.Equal(t, "/admin/users", r1.name, "route.name =")
	r1.Name("user")
	assert.Equal(t, "user", r1.name, "route.name =")
	_, exists := router.routes[r1.name]
	assert.True(t, exists, "router.routes[name] is ")
}

func TestRouteURL(t *testing.T) {
	router := New()
	group := newRouteGroup("/admin", router, nil)
	r := newRoute("/users/<id:\\d+>/<action>/*", group)
	assert.Equal(t, "/admin/users/123/address/<>", r.URL("id", 123, "action", "address"), "Route.URL@1 =")
	assert.Equal(t, "/admin/users/123/<action>/<>", r.URL("id", 123), "Route.URL@2 =")
	assert.Equal(t, "/admin/users/123//<>", r.URL("id", 123, "action"), "Route.URL@3 =")
	assert.Equal(t, "/admin/users/123/profile/", r.URL("id", 123, "action", "profile", ""), "Route.URL@4 =")
	assert.Equal(t, "/admin/users/123/profile/xyz%2Fabc", r.URL("id", 123, "action", "profile", "", "xyz/abc"), "Route.URL@5 =")
	assert.Equal(t, "/admin/users/123/a%2C%3C%3E%3F%23/<>", r.URL("id", 123, "action", "a,<>?#"), "Route.URL@6 =")
}

func newHandler(tag string, buf *bytes.Buffer) Handler {
	return func(*Context) error {
		fmt.Fprintf(buf, tag)
		return nil
	}
}

func TestRouteAdd(t *testing.T) {
	store := newMockStore()
	router := New()
	router.stores["GET"] = store
	assert.Equal(t, 0, store.count, "router.stores[GET].count =")

	var buf bytes.Buffer

	group := newRouteGroup("/admin", router, []Handler{newHandler("1.", &buf), newHandler("2.", &buf)})
	newRoute("/users", group).Get(newHandler("3.", &buf), newHandler("4.", &buf))
	assert.Equal(t, "1.2.3.4.", buf.String(), "buf@1 =")

	buf.Reset()
	group = newRouteGroup("/admin", router, []Handler{})
	newRoute("/users", group).Get(newHandler("3.", &buf), newHandler("4.", &buf))
	assert.Equal(t, "3.4.", buf.String(), "buf@2 =")

	buf.Reset()
	group = newRouteGroup("/admin", router, []Handler{newHandler("1.", &buf), newHandler("2.", &buf)})
	newRoute("/users", group).Get()
	assert.Equal(t, "1.2.", buf.String(), "buf@3 =")
}

func TestRouteMethods(t *testing.T) {
	router := New()
	for _, method := range Methods {
		store := newMockStore()
		router.stores[method] = store
		assert.Equal(t, 0, store.count, "router.stores["+method+"].count =")
	}
	group := newRouteGroup("/admin", router, nil)

	newRoute("/users", group).Get()
	assert.Equal(t, 1, router.stores["GET"].(*mockStore).count, "router.stores[GET].count =")
	newRoute("/users", group).Post()
	assert.Equal(t, 1, router.stores["POST"].(*mockStore).count, "router.stores[POST].count =")
	newRoute("/users", group).Patch()
	assert.Equal(t, 1, router.stores["PATCH"].(*mockStore).count, "router.stores[PATCH].count =")
	newRoute("/users", group).Put()
	assert.Equal(t, 1, router.stores["PUT"].(*mockStore).count, "router.stores[PUT].count =")
	newRoute("/users", group).Delete()
	assert.Equal(t, 1, router.stores["DELETE"].(*mockStore).count, "router.stores[DELETE].count =")
	newRoute("/users", group).Connect()
	assert.Equal(t, 1, router.stores["CONNECT"].(*mockStore).count, "router.stores[CONNECT].count =")
	newRoute("/users", group).Head()
	assert.Equal(t, 1, router.stores["HEAD"].(*mockStore).count, "router.stores[HEAD].count =")
	newRoute("/users", group).Options()
	assert.Equal(t, 1, router.stores["OPTIONS"].(*mockStore).count, "router.stores[OPTIONS].count =")
	newRoute("/users", group).Trace()
	assert.Equal(t, 1, router.stores["TRACE"].(*mockStore).count, "router.stores[TRACE].count =")

	newRoute("/posts", group).To("GET,POST")
	assert.Equal(t, 2, router.stores["GET"].(*mockStore).count, "router.stores[GET].count =")
	assert.Equal(t, 2, router.stores["POST"].(*mockStore).count, "router.stores[POST].count =")
	assert.Equal(t, 1, router.stores["PUT"].(*mockStore).count, "router.stores[PUT].count =")
}

func TestBuildURLTemplate(t *testing.T) {
	tests := []struct {
		path, expected string
	}{
		{"", ""},
		{"/users", "/users"},
		{"<id>", "<id>"},
		{"<id", "<id"},
		{"/users/<id>", "/users/<id>"},
		{"/users/<id:\\d+>", "/users/<id>"},
		{"/users/<:\\d+>", "/users/<>"},
		{"/users/<id>/xyz", "/users/<id>/xyz"},
		{"/users/<id:\\d+>/xyz", "/users/<id>/xyz"},
		{"/users/<id:\\d+>/<test>", "/users/<id>/<test>"},
		{"/users/<id:\\d+>/<test>/", "/users/<id>/<test>/"},
		{"/users/<id:\\d+><test>", "/users/<id><test>"},
		{"/users/<id:\\d+><test>/", "/users/<id><test>/"},
	}
	for _, test := range tests {
		actual := buildURLTemplate(test.path)
		assert.Equal(t, test.expected, actual, "buildURLTemplate("+test.path+") =")
	}
}
