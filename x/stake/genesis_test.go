package stake

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestInitGenesis(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)

	pool := keeper.GetPool(ctx)
	pool.BondedTokens = sdk.NewDec(2)

	params := keeper.GetParams(ctx)
	var delegations []Delegation

	validators := []Validator{
		NewValidator(sdk.ValAddress(keep.Addrs[0]), keep.PKs[0], Description{Moniker: "hoop"}),
		NewValidator(sdk.ValAddress(keep.Addrs[1]), keep.PKs[1], Description{Moniker: "bloop"}),
	}
	genesisState := types.NewGenesisState(pool, params, validators, delegations)
	_, err := InitGenesis(ctx, keeper, genesisState)
	require.Error(t, err)

	// initialize the validators
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = sdk.OneDec()
	validators[0].DelegatorShares = sdk.OneDec()
	validators[1].Status = sdk.Bonded
	validators[1].Tokens = sdk.OneDec()
	validators[1].DelegatorShares = sdk.OneDec()

	genesisState = types.NewGenesisState(pool, params, validators, delegations)
	vals, err := InitGenesis(ctx, keeper, genesisState)
	require.NoError(t, err)

	actualGenesis := WriteGenesis(ctx, keeper)
	require.Equal(t, genesisState.Pool, actualGenesis.Pool)
	require.Equal(t, genesisState.Params, actualGenesis.Params)
	require.Equal(t, genesisState.Bonds, actualGenesis.Bonds)
	require.EqualValues(t, keeper.GetAllValidators(ctx), actualGenesis.Validators)

	// now make sure the validators are bonded and intra-tx counters are correct
	resVal, found := keeper.GetValidator(ctx, sdk.ValAddress(keep.Addrs[0]))
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)
	require.Equal(t, int16(0), resVal.BondIntraTxCounter)

	resVal, found = keeper.GetValidator(ctx, sdk.ValAddress(keep.Addrs[1]))
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)
	require.Equal(t, int16(1), resVal.BondIntraTxCounter)

	abcivals := make([]abci.Validator, len(vals))
	for i, val := range validators {
		abcivals[i] = sdk.ABCIValidator(val)
	}

	require.Equal(t, abcivals, vals)
}

func TestInitGenesisLargeValidatorSet(t *testing.T) {
	size := 200
	require.True(t, size > 100)

	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)

	// Assigning 2 to the first 100 vals, 1 to the rest
	pool := keeper.GetPool(ctx)
	pool.BondedTokens = sdk.NewDec(int64(200 + (size - 100)))

	params := keeper.GetParams(ctx)
	delegations := []Delegation{}
	validators := make([]Validator, size)

	for i := range validators {
		validators[i] = NewValidator(sdk.ValAddress(keep.Addrs[i]), keep.PKs[i], Description{Moniker: fmt.Sprintf("#%d", i)})

		validators[i].Status = sdk.Bonded
		if i < 100 {
			validators[i].Tokens = sdk.NewDec(2)
			validators[i].DelegatorShares = sdk.NewDec(2)
		} else {
			validators[i].Tokens = sdk.OneDec()
			validators[i].DelegatorShares = sdk.OneDec()
		}
	}

	genesisState := types.NewGenesisState(pool, params, validators, delegations)
	vals, err := InitGenesis(ctx, keeper, genesisState)
	require.NoError(t, err)

	abcivals := make([]abci.Validator, 100)
	for i, val := range validators[:100] {
		abcivals[i] = sdk.ABCIValidator(val)
	}

	require.Equal(t, abcivals, vals)
}
