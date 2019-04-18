package upgrade

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(Plan{}, "upgrade/Plan", nil)
	cdc.RegisterConcrete(SoftwareUpgradeProposal{}, "upgrade/SoftwareUpgradeProposal", nil)
	cdc.RegisterConcrete(CancelSoftwareUpgradeProposal{}, "upgrade/CancelSoftwareUpgradeProposal", nil)
}

// module codec
var moduleCdc = codec.New()

func init() {
	RegisterCodec(moduleCdc)
}
