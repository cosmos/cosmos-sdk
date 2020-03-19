package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCanonicalProtoJSON(t *testing.T) {
	cases := map[string]struct {
		orig      string
		canonical string
	}{
		"bool": {
			orig:      `{"x":true,"y":false}`,
			canonical: `{"x":true}`,
		},
		"num":{
			orig:      `{"x":1,"y":0}`,
			canonical: `{"x":1}`,
		},
		"str":{
			orig:      `{"x":"abc","y":""}`,
			canonical: `{"x":"abc"}`,
		},
		"null":{
			orig:      `{"x":1,"y":null}`,
			canonical: `{"x":1}`,
		},
		"array":{
			orig:      `{"x":[1],"y":[]}`,
			canonical: `{"x":[1]}`,
		},
		"obj":{
			orig:      `{"x":{"a":1,"b":0},"y":{}}`,
			canonical: `{"x":{"a":1}}`,
		},
		"array with defaults":{
			orig:      `[0,false,"",null,[],{}]`,
			canonical: `[0,false,"",null,[],{}]`,
		},
		"unsorted":{
			orig:      `{"z":1,"a":2,"c":3}`,
			canonical: `{"a":2,"c":3,"z":1}`,
		},
		"tx": {
			orig:      `{"base":{"accountNumber":"13","chainId":"test-chain","fee":null}}`,
			canonical: `{"base":{"accountNumber":"13","chainId":"test-chain"}}`,
		},
		"complex": {
			orig: `{
				"a":0,
				"b":false,
				"c":"",
				"d":[],
				"e":{},
				"f":null,
				"g":1,
				"h":true,
				"i":"abc",
				"j":[0,false,"",[],{},1],
				"k":{"x":0,"y":1}
			}`,
			canonical: `{"g":1,"h":true,"i":"abc","j":[0,false,"",[],{},1],"k":{"y":1}}`,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			bz2, err := canonicalProtoJSON([]byte(tc.orig))
			assert.NoError(t, err)
			assert.Equal(t, tc.canonical, string(bz2))
		})
	}
}

