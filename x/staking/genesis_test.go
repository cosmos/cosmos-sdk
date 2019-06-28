package staking

import (
	"fmt"
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestInitGenesis(t *testing.T) {
	ctx, accKeeper, keeper, supplyKeeper := keep.CreateTestInput(t, false, 1000)

	valTokens := sdk.TokensFromConsensusPower(1)

	params := keeper.GetParams(ctx)
	validators := make([]Validator, 2)
	var delegations []Delegation

	// initialize the validators
	validators[0].OperatorAddress = sdk.ValAddress(keep.Addrs[0])
	validators[0].ConsPubKey = keep.PKs[0]
	validators[0].Description = NewDescription("hoop", "", "", "")
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = valTokens
	validators[0].DelegatorShares = valTokens.ToDec()
	validators[1].OperatorAddress = sdk.ValAddress(keep.Addrs[1])
	validators[1].ConsPubKey = keep.PKs[1]
	validators[1].Description = NewDescription("bloop", "", "", "")
	validators[1].Status = sdk.Bonded
	validators[1].Tokens = valTokens
	validators[1].DelegatorShares = valTokens.ToDec()

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := InitGenesis(ctx, keeper, accKeeper, supplyKeeper, genesisState)

	actualGenesis := ExportGenesis(ctx, keeper)
	require.Equal(t, genesisState.Params, actualGenesis.Params)
	require.Equal(t, genesisState.Delegations, actualGenesis.Delegations)
	require.EqualValues(t, keeper.GetAllValidators(ctx), actualGenesis.Validators)

	// now make sure the validators are bonded and intra-tx counters are correct
	resVal, found := keeper.GetValidator(ctx, sdk.ValAddress(keep.Addrs[0]))
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)

	resVal, found = keeper.GetValidator(ctx, sdk.ValAddress(keep.Addrs[1]))
	require.True(t, found)
	require.Equal(t, sdk.Bonded, resVal.Status)

	abcivals := make([]abci.ValidatorUpdate, len(vals))
	for i, val := range validators {
		abcivals[i] = val.ABCIValidatorUpdate()
	}

	require.Equal(t, abcivals, vals)
}

func TestInitGenesisLargeValidatorSet(t *testing.T) {
	size := 200
	require.True(t, size > 100)

	ctx, accKeeper, keeper, supplyKeeper := keep.CreateTestInput(t, false, 1000)

	params := keeper.GetParams(ctx)
	delegations := []Delegation{}
	validators := make([]Validator, size)

	for i := range validators {
		validators[i] = NewValidator(sdk.ValAddress(keep.Addrs[i]),
			keep.PKs[i], NewDescription(fmt.Sprintf("#%d", i), "", "", ""))

		validators[i].Status = sdk.Bonded

		tokens := sdk.TokensFromConsensusPower(1)
		if i < 100 {
			tokens = sdk.TokensFromConsensusPower(2)
		}
		validators[i].Tokens = tokens
		validators[i].DelegatorShares = tokens.ToDec()
	}

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := InitGenesis(ctx, keeper, accKeeper, supplyKeeper, genesisState)

	abcivals := make([]abci.ValidatorUpdate, 100)
	for i, val := range validators[:100] {
		abcivals[i] = val.ABCIValidatorUpdate()
	}

	require.Equal(t, abcivals, vals)
}

func TestValidateGenesis(t *testing.T) {
	genValidators1 := make([]types.Validator, 1, 5)
	pk := ed25519.GenPrivKey().PubKey()
	genValidators1[0] = types.NewValidator(sdk.ValAddress(pk.Address()), pk, types.NewDescription("", "", "", ""))
	genValidators1[0].Tokens = sdk.OneInt()
	genValidators1[0].DelegatorShares = sdk.OneDec()

	tests := []struct {
		name    string
		mutate  func(*types.GenesisState)
		wantErr bool
	}{
		{"default", func(*types.GenesisState) {}, false},
		// validate genesis validators
		{"duplicate validator", func(data *types.GenesisState) {
			(*data).Validators = genValidators1
			(*data).Validators = append((*data).Validators, genValidators1[0])
		}, true},
		{"no delegator shares", func(data *types.GenesisState) {
			(*data).Validators = genValidators1
			(*data).Validators[0].DelegatorShares = sdk.ZeroDec()
		}, true},
		{"jailed and bonded validator", func(data *types.GenesisState) {
			(*data).Validators = genValidators1
			(*data).Validators[0].Jailed = true
			(*data).Validators[0].Status = sdk.Bonded
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.DefaultGenesisState()
			tt.mutate(&genesisState)
			if tt.wantErr {
				assert.Error(t, ValidateGenesis(genesisState))
			} else {
				assert.NoError(t, ValidateGenesis(genesisState))
			}
		})
	}
}
