package types

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var header = tmproto.Header{
	ChainID: "hello",
	Height:  5,
}

func createValidators(t *testing.T) []Validator {
	return []Validator{
		newValidator(t, valAddr1, pk1),
		newValidator(t, valAddr2, pk2),
		newValidator(t, valAddr3, pk3),
	}
}

func TestHistoricalInfo(t *testing.T) {
	validators := createValidators(t)
	hi := NewHistoricalInfo(header, validators)
	require.True(t, sort.IsSorted(Validators(hi.Valset)), "Validators are not sorted")

	var value []byte
	require.NotPanics(t, func() {
		value = ModuleCdc.MustMarshalBinaryBare(&hi)
	})
	require.NotNil(t, value, "Marshalled HistoricalInfo is nil")

	recv, err := UnmarshalHistoricalInfo(ModuleCdc, value)
	require.Nil(t, err, "Unmarshalling HistoricalInfo failed")
	require.Equal(t, hi.Header, recv.Header)
	for i := range hi.Valset {
		require.True(t, hi.Valset[i].Equal(&recv.Valset[i]))
	}
	require.True(t, sort.IsSorted(Validators(hi.Valset)), "Validators are not sorted")
}

func TestValidateBasic(t *testing.T) {
	validators := createValidators(t)
	hi := HistoricalInfo{
		Header: header,
	}
	err := ValidateBasic(hi)
	require.Error(t, err, "ValidateBasic passed on nil ValSet")

	// Ensure validators are not sorted
	for sort.IsSorted(Validators(validators)) {
		rand.Shuffle(len(validators), func(i, j int) {
			it := validators[i]
			validators[i] = validators[j]
			validators[j] = it
		})
	}
	hi = HistoricalInfo{
		Header: header,
		Valset: validators,
	}
	err = ValidateBasic(hi)
	require.Error(t, err, "ValidateBasic passed on unsorted ValSet")

	hi = NewHistoricalInfo(header, validators)
	err = ValidateBasic(hi)
	require.NoError(t, err, "ValidateBasic failed on valid HistoricalInfo")
}
