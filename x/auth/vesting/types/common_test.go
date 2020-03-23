package types_test

import (
	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	"github.com/cosmos/cosmos-sdk/simapp"
)

var (
	app      = simapp.Setup(false)
	appCodec = codecstd.NewAppCodec(app.Codec())
)
