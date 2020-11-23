package types_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var header = tmproto.Header{
	ChainID: "hello",
	Height:  5,
}

func createValidators(t *testing.T) []stakingtypes.Validator {
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
	err := types.ValidateBasic(hi)
	require.Error(t, err, "ValidateBasic passed on nil ValSet")

	// Ensure validators are not sorted
	for sort.IsSorted(stakingtypes.Validators(validators)) {
		rand.Shuffle(len(validators), func(i, j int) {
			it := validators[i]
			validators[i] = validators[j]
			validators[j] = it
		})
	}
	hi = types.HistoricalInfo{
		Header: header,
		Valset: validators,
	}
	err = types.ValidateBasic(hi)
	require.Error(t, err, "ValidateBasic passed on unsorted ValSet")

	hi = types.NewHistoricalInfo(header, validators)
	err = types.ValidateBasic(hi)
	require.NoError(t, err, "ValidateBasic failed on valid HistoricalInfo")
}
