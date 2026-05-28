package staking_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestValidateGenesis(t *testing.T) {
	genValidators1 := make([]types.Validator, 1, 5)
	pk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	newPkAny, err := codectypes.NewAnyWithValue(newPk)
	assert.NoError(t, err)
	livePk := ed25519.GenPrivKey().PubKey()
	livePkAny, err := codectypes.NewAnyWithValue(livePk)
	assert.NoError(t, err)

	genValidators1[0] = testutil.NewValidator(t, sdk.ValAddress(pk.Address()), pk)
	genValidators1[0].Tokens = math.OneInt()
	genValidators1[0].DelegatorShares = math.LegacyOneDec()
	valAddr := genValidators1[0].OperatorAddress
	oldConsAddr := sdk.ConsAddress(pk.Address()).String()
	maturity := time.Now().UTC()

	genValidators2 := append([]types.Validator(nil), genValidators1...)
	liveVal := testutil.NewValidator(t, sdk.ValAddress(livePk.Address()), livePk)
	liveVal.Tokens = math.OneInt()
	liveVal.DelegatorShares = math.LegacyOneDec()
	genValidators2 = append(genValidators2, liveVal)
	cloneValidators := func(vals []types.Validator) []types.Validator {
		return append([]types.Validator(nil), vals...)
	}

	tests := []struct {
		name    string
		mutate  func(*types.GenesisState)
		wantErr bool
	}{
		{"default", func(*types.GenesisState) {}, false},
		// validate genesis validators
		{"duplicate validator", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.Validators = append(data.Validators, genValidators1[0])
		}, true},
		{"no delegator shares", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.Validators[0].DelegatorShares = math.LegacyZeroDec()
		}, true},
		{"jailed and bonded validator", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.Validators[0].Jailed = true
			data.Validators[0].Status = types.Bonded
		}, true},
		{"valid consensus key rotation state", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.ConsensusKeyRotationHistory = []types.ConsensusKeyRotationHistory{{
				ValidatorAddress:    valAddr,
				OldConsensusAddress: oldConsAddr,
				MaturityTime:        maturity,
			}}
			data.PendingConsensusKeyRotations = []types.PendingConsensusKeyRotation{{
				ValidatorAddress: valAddr,
				NewPubkey:        newPkAny,
				ApplyHeight:      10,
			}}
		}, false},
		{"pending consensus key rotation without history", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.PendingConsensusKeyRotations = []types.PendingConsensusKeyRotation{{
				ValidatorAddress: valAddr,
				NewPubkey:        newPkAny,
				ApplyHeight:      10,
			}}
		}, true},
		{"duplicate consensus key rotation history", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.ConsensusKeyRotationHistory = []types.ConsensusKeyRotationHistory{
				{
					ValidatorAddress:    valAddr,
					OldConsensusAddress: oldConsAddr,
					MaturityTime:        maturity,
				},
				{
					ValidatorAddress:    valAddr,
					OldConsensusAddress: oldConsAddr,
					MaturityTime:        maturity,
				},
			}
		}, true},
		{"negative pending consensus key rotation apply height", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.ConsensusKeyRotationHistory = []types.ConsensusKeyRotationHistory{{
				ValidatorAddress:    valAddr,
				OldConsensusAddress: oldConsAddr,
				MaturityTime:        maturity,
			}}
			data.PendingConsensusKeyRotations = []types.PendingConsensusKeyRotation{{
				ValidatorAddress: valAddr,
				NewPubkey:        newPkAny,
				ApplyHeight:      -1,
			}}
		}, true},
		{"pending consensus key rotation for unknown validator", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			unknownValAddr := sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address()).String()
			data.ConsensusKeyRotationHistory = []types.ConsensusKeyRotationHistory{{
				ValidatorAddress:    unknownValAddr,
				OldConsensusAddress: oldConsAddr,
				MaturityTime:        maturity,
			}}
			data.PendingConsensusKeyRotations = []types.PendingConsensusKeyRotation{{
				ValidatorAddress: unknownValAddr,
				NewPubkey:        newPkAny,
				ApplyHeight:      10,
			}}
		}, true},
		{"pending consensus key rotation to live validator key", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators2)
			data.ConsensusKeyRotationHistory = []types.ConsensusKeyRotationHistory{{
				ValidatorAddress:    valAddr,
				OldConsensusAddress: oldConsAddr,
				MaturityTime:        maturity,
			}}
			data.PendingConsensusKeyRotations = []types.PendingConsensusKeyRotation{{
				ValidatorAddress: valAddr,
				NewPubkey:        livePkAny,
				ApplyHeight:      10,
			}}
		}, true},
		{"pending consensus key rotation to history key", func(data *types.GenesisState) {
			data.Validators = cloneValidators(genValidators1)
			data.ConsensusKeyRotationHistory = []types.ConsensusKeyRotationHistory{{
				ValidatorAddress:    valAddr,
				OldConsensusAddress: sdk.ConsAddress(newPk.Address()).String(),
				MaturityTime:        maturity,
			}}
			data.PendingConsensusKeyRotations = []types.PendingConsensusKeyRotation{{
				ValidatorAddress: valAddr,
				NewPubkey:        newPkAny,
				ApplyHeight:      10,
			}}
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.DefaultGenesisState()
			tt.mutate(genesisState)

			if tt.wantErr {
				assert.Error(t, staking.ValidateGenesis(genesisState))
			} else {
				assert.NoError(t, staking.ValidateGenesis(genesisState))
			}
		})
	}
}
