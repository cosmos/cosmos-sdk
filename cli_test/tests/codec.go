package tests

import (
	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	"github.com/cosmos/cosmos-sdk/simapp"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

var (
	cdc      = codecstd.MakeCodec(simapp.ModuleBasics)
	appCodec = codecstd.NewAppCodec(cdc)
)

func init() {
	authclient.Codec = appCodec
}
