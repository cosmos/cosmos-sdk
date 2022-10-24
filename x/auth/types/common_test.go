package types_test

import (
	"github.com/pointnetwork/cosmos-point-sdk/simapp"
)

var (
	ecdc                  = simapp.MakeTestEncodingConfig()
	appCodec, legacyAmino = ecdc.Codec, ecdc.Amino
)
