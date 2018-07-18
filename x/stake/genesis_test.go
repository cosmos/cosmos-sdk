package stake

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestInitGenesis(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)

	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewRat(2)

	params := keeper.GetParams(ctx)
	var delegations []Delegation

	validators := []Validator{
		NewValidator(keep.Addrs[0], keep.PKs[0], Description{Moniker: "hoop"}),
		NewValidator(keep.Addrs[1], keep.PKs[1], Description{Moniker: "bloop"}),
	}
	genesisState := types.NewGenesisState(pool, params, validators, delegations)
	err := InitGenesis(ctx, keeper, genesisState)
	require.Error(t, err)

	// initialize the validators
	validators[0].Tokens = sdk.OneRat()
	validators[0].DelegatorShares = sdk.OneRat()
	validators[1].Tokens = sdk.OneRat()
	validators[1].DelegatorShares = sdk.OneRat()

	genesisState = types.NewGenesisState(pool, params, validators, delegations)
	err = InitGenesis(ctx, keeper, genesisState)
	require.NoError(t, err)

	// now make sure the validators are bonded
	resVal, found := keeper.GetValidator(ctx, keep.Addrs[0])
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)

	resVal, found = keeper.GetValidator(ctx, keep.Addrs[1])
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)
}
