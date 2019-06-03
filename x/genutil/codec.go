package genutil

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// generic sealed codec to be used throughout this module
var moduleCdc *codec.Codec

// TODO abstract genesis transactions registration back to staking
// required for genesis transactions
func init() {
	moduleCdc = codec.New()
	staking.RegisterCodec(moduleCdc)
	auth.RegisterCodec(moduleCdc)
	sdk.RegisterCodec(moduleCdc)
	codec.RegisterCrypto(moduleCdc)
	moduleCdc.Seal()
}
