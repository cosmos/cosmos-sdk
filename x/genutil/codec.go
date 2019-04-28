package genutil

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// generic sealed codec to be used throughout this module
var moduleCdc *codec.Codec

func init() {
	cdc := codec.New()

	// TODO abstract genesis transactions registration back to staking
	// required for genesis transactions
	staking.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	moduleCdc = cdc.Seal()
}
