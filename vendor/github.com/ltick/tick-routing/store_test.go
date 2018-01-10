// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type storeTestEntry struct {
	key, data string
	params    int
}

func TestStoreAdd(t *testing.T) {
	tests := []struct {
		id       string
		entries  []storeTestEntry
		expected string
	}{
		{
			"all static",
			[]storeTestEntry{
				{"/gopher/bumper.png", "1", 0},
				{"/gopher/bumper192x108.png", "2", 0},
				{"/gopher/doc.png", "3", 0},
				{"/gopher/bumper320x180.png", "4", 0},
				{"/gopher/docpage.png", "5", 0},
				{"/gopher/doc.png", "6", 0},
				{"/gopher/doc", "7", 0},
			},
			`{key: , regex: <nil>, data: <nil>, order: 0, minOrder: 0, pindex: -1, pnames: []}
    {key: /gopher/, regex: <nil>, data: <nil>, order: 1, minOrder: 1, pindex: -1, pnames: []}
        {key: bumper, regex: <nil>, data: <nil>, order: 1, minOrder: 1, pindex: -1, pnames: []}
            {key: .png, regex: <nil>, data: 1, order: 1, minOrder: 1, pindex: -1, pnames: []}
            {key: 192x108.png, regex: <nil>, data: 2, order: 2, minOrder: 2, pindex: -1, pnames: []}
            {key: 320x180.png, regex: <nil>, data: 4, order: 4, minOrder: 4, pindex: -1, pnames: []}
        {key: doc, regex: <nil>, data: 7, order: 7, minOrder: 3, pindex: -1, pnames: []}
            {key: .png, regex: <nil>, data: 3, order: 3, minOrder: 3, pindex: -1, pnames: []}
            {key: page.png, regex: <nil>, data: 5, order: 5, minOrder: 5, pindex: -1, pnames: []}
`,
		},
		{
			"parametric",
			[]storeTestEntry{
				{"/users/<id>", "11", 1},
				{"/users/<id>/profile", "12", 1},
				{"/users/<id>/<accnt:\\d+>/address", "13", 2},
				{"/users/<id>/age", "14", 1},
				{"/users/<id>/<accnt:\\d+>", "15", 2},
			},
			`{key: , regex: <nil>, data: <nil>, order: 0, minOrder: 0, pindex: -1, pnames: []}
    {key: /users/, regex: <nil>, data: <nil>, order: 0, minOrder: 1, pindex: -1, pnames: []}
        {key: <id>, regex: <nil>, data: 11, order: 1, minOrder: 1, pindex: 0, pnames: [id]}
            {key: /, regex: <nil>, data: <nil>, order: 2, minOrder: 2, pindex: 0, pnames: [id]}
                {key: age, regex: <nil>, data: 14, order: 4, minOrder: 4, pindex: 0, pnames: [id]}
                {key: profile, regex: <nil>, data: 12, order: 2, minOrder: 2, pindex: 0, pnames: [id]}
                {key: <accnt:\d+>, regex: ^\d+, data: 15, order: 5, minOrder: 3, pindex: 1, pnames: [id accnt]}
                    {key: /address, regex: <nil>, data: 13, order: 3, minOrder: 3, pindex: 1, pnames: [id accnt]}
`,
		},
		{
			"corner cases",
			[]storeTestEntry{
				{"/users/<id>/test/<name>", "101", 2},
				{"/users/abc/<id>/<name>", "102", 2},
				{"", "103", 0},
			},
			`{key: , regex: <nil>, data: 103, order: 3, minOrder: 0, pindex: -1, pnames: []}
    {key: /users/, regex: <nil>, data: <nil>, order: 0, minOrder: 1, pindex: -1, pnames: []}
        {key: abc/, regex: <nil>, data: <nil>, order: 0, minOrder: 2, pindex: -1, pnames: []}
            {key: <id>, regex: <nil>, data: <nil>, order: 0, minOrder: 2, pindex: 0, pnames: [id]}
                {key: /, regex: <nil>, data: <nil>, order: 0, minOrder: 2, pindex: 0, pnames: [id]}
                    {key: <name>, regex: <nil>, data: 102, order: 2, minOrder: 2, pindex: 1, pnames: [id name]}
        {key: <id>, regex: <nil>, data: <nil>, order: 0, minOrder: 1, pindex: 0, pnames: [id]}
            {key: /test/, regex: <nil>, data: <nil>, order: 0, minOrder: 1, pindex: 0, pnames: [id]}
                {key: <name>, regex: <nil>, data: 101, order: 1, minOrder: 1, pindex: 1, pnames: [id name]}
`,
		},
	}
	for _, test := range tests {
		h := newStore()
		for _, entry := range test.entries {
			n := h.Add(entry.key, entry.data)
			assert.Equal(t, entry.params, n, test.id+" > "+entry.key+" > param count =")
		}
		assert.Equal(t, test.expected, h.String(), test.id+" > store.String() =")
	}
}

func TestStoreGet(t *testing.T) {
	pairs := []struct {
		key, value string
	}{
		{"/gopher/bumper.png", "1"},
		{"/gopher/bumper192x108.png", "2"},
		{"/gopher/doc.png", "3"},
		{"/gopher/bumper320x180.png", "4"},
		{"/gopher/docpage.png", "5"},
		{"/gopher/doc.png", "6"},
		{"/gopher/doc", "7"},
		{"/users/<id>", "8"},
		{"/users/<id>/profile", "9"},
		{"/users/<id>/<accnt:\\d+>/address", "10"},
		{"/users/<id>/age", "11"},
		{"/users/<id>/<accnt:\\d+>", "12"},
		{"/users/<id>/test/<name>", "13"},
		{"/users/abc/<id>/<name>", "14"},
		{"", "15"},
		{"/all/<:.*>", "16"},
	}
	h := newStore()
	maxParams := 0
	for _, pair := range pairs {
		n := h.Add(pair.key, pair.value)
		if n > maxParams {
			maxParams = n
		}
	}
	assert.Equal(t, 2, maxParams, "param count = ")

	tests := []struct {
		key    string
		value  interface{}
		params string
	}{
		{"/gopher/bumper.png", "1", ""},
		{"/gopher/bumper192x108.png", "2", ""},
		{"/gopher/doc.png", "3", ""},
		{"/gopher/bumper320x180.png", "4", ""},
		{"/gopher/docpage.png", "5", ""},
		{"/gopher/doc.png", "3", ""},
		{"/gopher/doc", "7", ""},
		{"/users/abc", "8", "id:abc,"},
		{"/users/abc/profile", "9", "id:abc,"},
		{"/users/abc/123/address", "10", "id:abc,accnt:123,"},
		{"/users/abcd/age", "11", "id:abcd,"},
		{"/users/abc/123", "12", "id:abc,accnt:123,"},
		{"/users/abc/test/123", "13", "id:abc,name:123,"},
		{"/users/abc/xyz/123", "14", "id:xyz,name:123,"},
		{"", "15", ""},
		{"/g", nil, ""},
		{"/all", nil, ""},
		{"/all/", "16", ":,"},
		{"/all/abc", "16", ":abc,"},
		{"/users/abc/xyz", nil, ""},
	}
	pvalues := make([]string, maxParams)
	for _, test := range tests {
		data, pnames := h.Get(test.key, pvalues)
		assert.Equal(t, test.value, data, "store.Get("+test.key+") =")
		params := ""
		if len(pnames) > 0 {
			for i, name := range pnames {
				params += fmt.Sprintf("%v:%v,", name, pvalues[i])
			}
		}
		assert.Equal(t, test.params, params, "store.Get("+test.key+").params =")
	}
}
