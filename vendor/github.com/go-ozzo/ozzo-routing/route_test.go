// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

	r1 := group.newRoute("GET", "/users").Get()
	assert.Equal(t, "", r1.name, "route.name =")
	assert.Equal(t, "/users", r1.path, "route.path =")
	assert.Equal(t, 1, len(router.Routes()))

	r2 := group.newRoute("GET", "/users/<id:\\d+>/*").Post()
	assert.Equal(t, "", r2.name, "route.name =")
	assert.Equal(t, "/users/<id:\\d+>/*", r2.path, "route.path =")
	assert.Equal(t, "/admin/users/<id>/", r2.template, "route.template =")
	assert.Equal(t, 2, len(router.Routes()))
}

func TestRouteName(t *testing.T) {
	router := New()
	group := newRouteGroup("/admin", router, nil)

	r1 := group.newRoute("GET", "/users")
	assert.Equal(t, "", r1.name, "route.name =")
	r1.Name("user")
	assert.Equal(t, "user", r1.name, "route.name =")
	_, exists := router.namedRoutes[r1.name]
	assert.True(t, exists)
}

func TestRouteURL(t *testing.T) {
	router := New()
	group := newRouteGroup("/admin", router, nil)
	r := group.newRoute("GET", "/users/<id:\\d+>/<action>/*")
	assert.Equal(t, "/admin/users/123/address/", r.URL("id", 123, "action", "address"))
	assert.Equal(t, "/admin/users/123/<action>/", r.URL("id", 123))
	assert.Equal(t, "/admin/users/123//", r.URL("id", 123, "action"))
	assert.Equal(t, "/admin/users/123/profile/", r.URL("id", 123, "action", "profile", ""))
	assert.Equal(t, "/admin/users/123/profile/", r.URL("id", 123, "action", "profile", "", "xyz/abc"))
	assert.Equal(t, "/admin/users/123/a%2C%3C%3E%3F%23/", r.URL("id", 123, "action", "a,<>?#"))
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
	group.newRoute("GET", "/users").Get(newHandler("3.", &buf), newHandler("4.", &buf))
	assert.Equal(t, "1.2.3.4.", buf.String(), "buf@1 =")

	buf.Reset()
	group = newRouteGroup("/admin", router, []Handler{})
	group.newRoute("GET", "/users").Get(newHandler("3.", &buf), newHandler("4.", &buf))
	assert.Equal(t, "3.4.", buf.String(), "buf@2 =")

	buf.Reset()
	group = newRouteGroup("/admin", router, []Handler{newHandler("1.", &buf), newHandler("2.", &buf)})
	group.newRoute("GET", "/users").Get()
	assert.Equal(t, "1.2.", buf.String(), "buf@3 =")
}

func TestRouteTag(t *testing.T) {
	router := New()
	router.Get("/posts").Tag("posts")
	router.Any("/users").Tag("users")
	router.To("PUT,PATCH", "/comments").Tag("comments")
	router.Get("/orders").Tag("GET orders").Post().Tag("POST orders")
	routes := router.Routes()
	for _, route := range routes {
		if !assert.True(t, len(route.Tags()) > 0, route.method+" "+route.path+" should have a tag") {
			continue
		}
		tag := route.Tags()[0].(string)
		switch route.path {
		case "/posts":
			assert.Equal(t, "posts", tag)
		case "/users":
			assert.Equal(t, "users", tag)
		case "/comments":
			assert.Equal(t, "comments", tag)
		case "/orders":
			if route.method == "GET" {
				assert.Equal(t, "GET orders", tag)
			} else {
				assert.Equal(t, "POST orders", tag)
			}
		}
	}
}

func TestRouteMethods(t *testing.T) {
	router := New()
	for _, method := range Methods {
		store := newMockStore()
		router.stores[method] = store
		assert.Equal(t, 0, store.count, "router.stores["+method+"].count =")
	}
	group := newRouteGroup("/admin", router, nil)

	group.newRoute("GET", "/users").Get()
	assert.Equal(t, 1, router.stores["GET"].(*mockStore).count, "router.stores[GET].count =")
	group.newRoute("GET", "/users").Post()
	assert.Equal(t, 1, router.stores["POST"].(*mockStore).count, "router.stores[POST].count =")
	group.newRoute("GET", "/users").Patch()
	assert.Equal(t, 1, router.stores["PATCH"].(*mockStore).count, "router.stores[PATCH].count =")
	group.newRoute("GET", "/users").Put()
	assert.Equal(t, 1, router.stores["PUT"].(*mockStore).count, "router.stores[PUT].count =")
	group.newRoute("GET", "/users").Delete()
	assert.Equal(t, 1, router.stores["DELETE"].(*mockStore).count, "router.stores[DELETE].count =")
	group.newRoute("GET", "/users").Connect()
	assert.Equal(t, 1, router.stores["CONNECT"].(*mockStore).count, "router.stores[CONNECT].count =")
	group.newRoute("GET", "/users").Head()
	assert.Equal(t, 1, router.stores["HEAD"].(*mockStore).count, "router.stores[HEAD].count =")
	group.newRoute("GET", "/users").Options()
	assert.Equal(t, 1, router.stores["OPTIONS"].(*mockStore).count, "router.stores[OPTIONS].count =")
	group.newRoute("GET", "/users").Trace()
	assert.Equal(t, 1, router.stores["TRACE"].(*mockStore).count, "router.stores[TRACE].count =")

	group.newRoute("GET", "/posts").To("GET,POST")
	assert.Equal(t, 2, router.stores["GET"].(*mockStore).count, "router.stores[GET].count =")
	assert.Equal(t, 2, router.stores["POST"].(*mockStore).count, "router.stores[POST].count =")
	assert.Equal(t, 1, router.stores["PUT"].(*mockStore).count, "router.stores[PUT].count =")

	group.newRoute("GET", "/posts").To("POST")
	assert.Equal(t, 2, router.stores["GET"].(*mockStore).count, "router.stores[GET].count =")
	assert.Equal(t, 3, router.stores["POST"].(*mockStore).count, "router.stores[POST].count =")
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

func TestRouteString(t *testing.T) {
	router := New()
	router.Get("/users/<id>")
	router.To("GET,POST", "/users/<id>/profile")
	group := router.Group("/admin")
	group.Post("/users")
	s := ""
	for _, route := range router.Routes() {
		s += fmt.Sprintln(route)
	}

	assert.Equal(t, `GET /users/<id>
GET /users/<id>/profile
POST /users/<id>/profile
POST /admin/users
`, s)
}
