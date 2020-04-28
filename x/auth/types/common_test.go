package types_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
)

var (
	app      = simapp.Setup(false)
	appCodec = std.NewAppCodec(app.Codec())
)
