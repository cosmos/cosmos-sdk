package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
)

//----------------------------------

type Dummy struct{}

func (d *Dummy) Name() string {
	return "dummy"
}
func (d *Dummy) RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result) {
	return
}
func (d *Dummy) SetOption(storei KVStore, key, value string) (log string) {
	return ""
}
func (d *Dummy) InitChain(store KVStore, vals []*abci.Validator) {
}
func (d *Dummy) BeginBlock(store KVStore, hash []byte, header *abci.Header) {
}
func (d *Dummy) EndBlock(store KVStore, height uint64) (res abci.ResponseEndBlock) {
	return
}

//----------------------------------

func TestPlugin(t *testing.T) {

	plugins := NewPlugins()

	//define the test list
	var testList = []struct {
		testPass func() bool
		errMsg   string
	}{
		{func() bool { return (len(plugins.GetList()) == 0) },
			"plugins object init with a objects"},
		{func() bool { plugins.RegisterPlugin(&Dummy{}); return (len(plugins.GetList()) == 1) },
			"plugin wasn't added to plist after registered"},
		{func() bool { return (plugins.GetByName("dummy").Name() == "dummy") },
			"plugin wasn't retrieved properly with GetByName"},
	}

	//execute the tests
	for _, tl := range testList {
		assert.True(t, tl.testPass(), tl.errMsg)
	}

}
