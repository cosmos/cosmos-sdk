package types

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestHistoricalInfo(t *testing.T) {
	validators := []Validator{
		NewValidator(valAddr1, pk1, Description{}),
		NewValidator(valAddr2, pk2, Description{}),
		NewValidator(valAddr3, pk3, Description{}),
	}
	header := abci.Header{
		ChainID: "hello",
		Height:  5,
	}
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
