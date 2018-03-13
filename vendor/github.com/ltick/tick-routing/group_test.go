// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteGroupTo(t *testing.T) {
	router := New(context.Background())
	for _, method := range Methods {
		store := newMockStore()
		router.stores[method] = store
	}
	group := newRouteGroup("/admin", router, nil, nil, nil, nil)

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
	router := New(context.Background())
	for _, method := range Methods {
		store := newMockStore()
		router.stores[method] = store
		assert.Equal(t, 0, store.count, "router.stores["+method+"].count =")
	}
	group := newRouteGroup("/admin", router, nil, nil, nil, nil)

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
	group := newRouteGroup("/admin", New(context.Background()), nil, nil, nil, nil)
	g1 := group.Group("/users", nil, nil)
	assert.Equal(t, "/admin/users", g1.prefix, "g1.prefix =")
	assert.Equal(t, 0, len(g1.groupStartupHandlers), "len(g1.groupStartupHandlers) =")
	assert.Equal(t, 0, len(g1.groupShutdownHandlers), "len(g1.groupShutdownHandlers) =")
	var buf bytes.Buffer
	g2 := group.Group("", []Handler{newHandler("1", &buf), newHandler("2", &buf)}, nil)
	assert.Equal(t, "/admin", g2.prefix, "g2.prefix =")
	assert.Equal(t, 2, len(g2.groupStartupHandlers), "len(g2.groupStartupHandlers) =")

	g3 := group.Group("", nil, []Handler{newHandler("1", &buf), newHandler("2", &buf)})
	assert.Equal(t, "/admin", g3.prefix, "g2.prefix =")
	assert.Equal(t, 2, len(g3.groupShutdownHandlers), "len(g2.groupShutdownHandlers) =")

	group2 := newRouteGroup("/admin", New(context.Background()), []Handler{newHandler("1", &buf), newHandler("2", &buf)}, []Handler{}, []Handler{}, []Handler{})
	g4 := group2.Group("/users", nil, nil)
	assert.Equal(t, "/admin/users", g4.prefix, "g4.prefix =")
	assert.Equal(t, 2, len(g4.groupStartupHandlers), "len(g4.groupStartupHandlers) =")
	g5 := group2.Group("", []Handler{newHandler("3", &buf)}, nil)
	assert.Equal(t, "/admin", g5.prefix, "g5.prefix =")
	assert.Equal(t, 1, len(g5.groupStartupHandlers), "len(g5.groupStartupHandlers) =")
}

func TestRouteGroupAddStartupHandler(t *testing.T) {
	var buf bytes.Buffer
	group := newRouteGroup("/admin", New(context.Background()), nil, nil, nil, nil)
	group.AddStartupHandler(newHandler("1", &buf), newHandler("2", &buf))
	assert.Equal(t, 2, len(group.groupStartupHandlers), "len(group.groupStartupHandlers) =")

	group2 := newRouteGroup("/admin", New(context.Background()), []Handler{newHandler("1", &buf), newHandler("2", &buf)}, []Handler{}, []Handler{}, []Handler{})
	group2.AddStartupHandler(newHandler("3", &buf))
	assert.Equal(t, 3, len(group2.groupStartupHandlers), "len(group2.groupStartupHandlers) =")
}

func TestRouteGroupAddShutdownHandler(t *testing.T) {
	var buf bytes.Buffer
	group := newRouteGroup("/admin", New(context.Background()), nil, nil, nil, nil)
	group.AddShutdownHandler(newHandler("1", &buf), newHandler("2", &buf))
	assert.Equal(t, 2, len(group.groupShutdownHandlers), "len(group.groupShutdownHandlers) =")

	group2 := newRouteGroup("/admin", New(context.Background()), []Handler{}, []Handler{}, []Handler{}, []Handler{newHandler("1", &buf), newHandler("2", &buf)})
	group2.AddShutdownHandler(newHandler("3", &buf))
	assert.Equal(t, 3, len(group2.groupShutdownHandlers), "len(group2.groupShutdownHandlers) =")
}
