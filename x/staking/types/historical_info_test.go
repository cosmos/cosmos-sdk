package types_test

import (
	"math/rand"
	"sort"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var header = cmtproto.Header{
	ChainID: "hello",
	Height:  5,
}

func createValidators(t *testing.T) []types.Validator {
	t.Helper()
	return []types.Validator{
		newValidator(t, valAddr1, pk1),
		newValidator(t, valAddr2, pk2),
		newValidator(t, valAddr3, pk3),
	}
}

func TestValidateBasic(t *testing.T) {
	validators := createValidators(t)
	hi := types.HistoricalInfo{
		Header: header,
	}
	ac := address.NewBech32Codec("cosmosvaloper")
	err := types.ValidateBasic(hi, ac)
	require.Error(t, err, "ValidateBasic passed on nil ValSet")

	// Ensure validators are not sorted
	for sort.IsSorted(types.Validators{Validators: validators, ValidatorCodec: ac}) {
		rand.Shuffle(len(validators), func(i, j int) {
			validators[i], validators[j] = validators[j], validators[i]
		})
	}
	hi = types.HistoricalInfo{
		Header: header,
		Valset: validators,
	}
	err = types.ValidateBasic(hi, ac)
	require.Error(t, err, "ValidateBasic passed on unsorted ValSet")

	hi = types.NewHistoricalInfo(header, types.Validators{Validators: validators, ValidatorCodec: ac}, sdk.DefaultPowerReduction)
	err = types.ValidateBasic(hi, ac)
	require.NoError(t, err, "ValidateBasic failed on valid HistoricalInfo")
}
