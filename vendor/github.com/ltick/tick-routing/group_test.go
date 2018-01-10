// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteGroupTo(t *testing.T) {
	router := New()
	for _, method := range Methods {
		store := newMockStore()
		router.stores[method] = store
	}
	group := newRouteGroup("/admin", router, nil)

	group.Any("/users")
	for _, method := range Methods {
		assert.Equal(t, 1, router.stores[method].(*mockStore).count, "router.stores["+method+"].count@1 =")
	}

	group.To("GET", "/articles")
	assert.Equal(t, 2, router.stores["GET"].(*mockStore).count, "router.stores[GET].count@2 =")
	assert.Equal(t, 1, router.stores["POST"].(*mockStore).count, "router.stores[POST].count@2 =")

	group.To("GET,POST", "/comments")
	assert.Equal(t, 3, router.stores["GET"].(*mockStore).count, "router.stores[GET].count@3 =")
	assert.Equal(t, 2, router.stores["POST"].(*mockStore).count, "router.stores[POST].count@3 =")
}

func TestRouteGroupMethods(t *testing.T) {
	router := New()
	for _, method := range Methods {
		store := newMockStore()
		router.stores[method] = store
		assert.Equal(t, 0, store.count, "router.stores["+method+"].count =")
	}
	group := newRouteGroup("/admin", router, nil)

	group.Get("/users")
	assert.Equal(t, 1, router.stores["GET"].(*mockStore).count, "router.stores[GET].count =")
	group.Post("/users")
	assert.Equal(t, 1, router.stores["POST"].(*mockStore).count, "router.stores[POST].count =")
	group.Patch("/users")
	assert.Equal(t, 1, router.stores["PATCH"].(*mockStore).count, "router.stores[PATCH].count =")
	group.Put("/users")
	assert.Equal(t, 1, router.stores["PUT"].(*mockStore).count, "router.stores[PUT].count =")
	group.Delete("/users")
	assert.Equal(t, 1, router.stores["DELETE"].(*mockStore).count, "router.stores[DELETE].count =")
	group.Connect("/users")
	assert.Equal(t, 1, router.stores["CONNECT"].(*mockStore).count, "router.stores[CONNECT].count =")
	group.Head("/users")
	assert.Equal(t, 1, router.stores["HEAD"].(*mockStore).count, "router.stores[HEAD].count =")
	group.Options("/users")
	assert.Equal(t, 1, router.stores["OPTIONS"].(*mockStore).count, "router.stores[OPTIONS].count =")
	group.Trace("/users")
	assert.Equal(t, 1, router.stores["TRACE"].(*mockStore).count, "router.stores[TRACE].count =")
}

func TestRouteGroupGroup(t *testing.T) {
	group := newRouteGroup("/admin", New(), nil)
	g1 := group.Group("/users")
	assert.Equal(t, "/admin/users", g1.prefix, "g1.prefix =")
	assert.Equal(t, 0, len(g1.handlers), "len(g1.handlers) =")
	var buf bytes.Buffer
	g2 := group.Group("", newHandler("1", &buf), newHandler("2", &buf))
	assert.Equal(t, "/admin", g2.prefix, "g2.prefix =")
	assert.Equal(t, 2, len(g2.handlers), "len(g2.handlers) =")

	group2 := newRouteGroup("/admin", New(), []Handler{newHandler("1", &buf), newHandler("2", &buf)})
	g3 := group2.Group("/users")
	assert.Equal(t, "/admin/users", g3.prefix, "g3.prefix =")
	assert.Equal(t, 2, len(g3.handlers), "len(g3.handlers) =")
	g4 := group2.Group("", newHandler("3", &buf))
	assert.Equal(t, "/admin", g4.prefix, "g4.prefix =")
	assert.Equal(t, 1, len(g4.handlers), "len(g4.handlers) =")
}

func TestRouteGroupUse(t *testing.T) {
	var buf bytes.Buffer
	group := newRouteGroup("/admin", New(), nil)
	group.Use(newHandler("1", &buf), newHandler("2", &buf))
	assert.Equal(t, 2, len(group.handlers), "len(group.handlers) =")

	group2 := newRouteGroup("/admin", New(), []Handler{newHandler("1", &buf), newHandler("2", &buf)})
	group2.Use(newHandler("3", &buf))
	assert.Equal(t, 3, len(group2.handlers), "len(group2.handlers) =")
}
