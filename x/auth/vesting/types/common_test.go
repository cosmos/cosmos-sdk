package types_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	simappcodec "github.com/cosmos/cosmos-sdk/simapp/codec"
)

var (
	app      = simapp.Setup(false)
	appCodec = simappcodec.NewAppCodec(app.Codec())
)
