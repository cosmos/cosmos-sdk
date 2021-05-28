package module

import (
	codec2 "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/core/codec"
	"github.com/cosmos/cosmos-sdk/core/module/app"
	"github.com/cosmos/cosmos-sdk/core/store"
	"github.com/cosmos/cosmos-sdk/x/authn"
)

var _ codec.TypeProvider = Module{}

func (m Module) RegisterTypes(registry codec.TypeRegistry) {
	authn.RegisterTypes(registry)
}

func (m Module) NewAppHandler(cdc codec2.Codec, storeKey store.KVStoreKey) app.Handler {

}
