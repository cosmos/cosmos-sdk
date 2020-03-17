package types

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonStripDefaults(t *testing.T) {
	var x interface{}
	err := json.Unmarshal([]byte(`{
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
	}`), &x)
	assert.NoError(t, err)
	y := JsonStripDefaults(x)
	bz, err := json.Marshal(y)
	assert.NoError(t, err)
	assert.Equal(t, `{"g":1,"h":true,"i":"abc","j":[null,null,null,null,null,1],"k":{"y":1}}`, string(bz))
}
