package types

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJSONStripDefaults(t *testing.T) {
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
	y := jsonStripDefaults(x)
	bz, err := json.Marshal(y)
	assert.NoError(t, err)
	assert.Equal(t, `{"g":1,"h":true,"i":"abc","j":[0,false,"",[],{},1],"k":{"y":1}}`, string(bz))
}

func TestJSONStripDefaults2(t *testing.T) {
	var x interface{}
	err := json.Unmarshal([]byte(`{"base":{"accountNumber":"13","chainId":"test-chain","fee":null}}`), &x)
	assert.NoError(t, err)
	y := jsonStripDefaults(x)
	bz, err := json.Marshal(y)
	assert.NoError(t, err)
	assert.Equal(t, `{"base":{"accountNumber":"13","chainId":"test-chain"}}`, string(bz))
}
