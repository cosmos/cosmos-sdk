package types_test

import (
	"github.com/KiraCore/cosmos-sdk/simapp"
)

var (
	app         = simapp.Setup(false)
	appCodec, _ = simapp.MakeCodecs()
)
