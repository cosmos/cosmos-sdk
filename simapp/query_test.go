package simapp

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestQuery(t *testing.T) {
	app := Setup(false)
	ctx := app.NewContext(false, abci.Header{Height: app.LastBlockHeight()})

	chainSupply := supply.Supply{
		Total: sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(10))),
	}

	app.SupplyKeeper.SetSupply(ctx, chainSupply)
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1}})
	app.Commit()

	retr := app.SupplyKeeper.GetSupply(ctx)
	require.Equal(t, &chainSupply, retr, "Get does not work")

	res := app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", supply.StoreKey),
		Data:  supply.SupplyKey,
		Prove: true,
	})

	fmt.Printf("Res: %#v\n", res)

	require.NotNil(t, res.Value, "Query is nil")
}
