package types

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	validators = []Validator{
		NewValidator(valAddr1, pk1, Description{}),
		NewValidator(valAddr2, pk2, Description{}),
		NewValidator(valAddr3, pk3, Description{}),
	}
	header = abci.Header{
		ChainID: "hello",
		Height:  5,
	}
)

func TestHistoricalInfo(t *testing.T) {
	hi := NewHistoricalInfo(header, validators)
	require.True(t, sort.IsSorted(Validators(hi.ValSet)), "Validators are not sorted")

	var value []byte
	require.NotPanics(t, func() {
		value = MustMarshalHistoricalInfo(ModuleCdc, hi)
	})

	require.NotNil(t, value, "Marshalled HistoricalInfo is nil")

	recv, err := UnmarshalHistoricalInfo(ModuleCdc, value)
	require.Nil(t, err, "Unmarshalling HistoricalInfo failed")
	require.Equal(t, hi, recv, "Unmarshalled HistoricalInfo is different from original")
	require.True(t, sort.IsSorted(Validators(hi.ValSet)), "Validators are not sorted")
}

func TestValidateBasic(t *testing.T) {
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
		ValSet: validators,
	}
	err = ValidateBasic(hi)
	require.Error(t, err, "ValidateBasic passed on unsorted ValSet")

	hi = NewHistoricalInfo(header, validators)
	err = ValidateBasic(hi)
	require.NoError(t, err, "ValidateBasic failed on valid HistoricalInfo")
}
