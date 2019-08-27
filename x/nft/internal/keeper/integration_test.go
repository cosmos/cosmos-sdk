package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
)

// nolint: deadcode unused
var (
	denom     = "test-denom"
	denom2    = "test-denom2"
	denom3    = "test-denom3"
	id        = "1"
	id2       = "2"
	id3       = "3"
	address   = types.CreateTestAddrs(1)[0]
	address2  = types.CreateTestAddrs(2)[1]
	address3  = types.CreateTestAddrs(3)[2]
	tokenURI  = "https://google.com/token-1.json"
	tokenURI2 = "https://google.com/token-2.json"
)

func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})

	return app, ctx
}
