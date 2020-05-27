package baseapp_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

func TestValidateBlockParams(t *testing.T) {
	testCases := []struct {
		arg       interface{}
		expectErr bool
	}{
		{nil, true},
		{&abci.BlockParams{}, true},
		{abci.BlockParams{}, true},
		{abci.BlockParams{MaxBytes: -1, MaxGas: -1}, true},
		{abci.BlockParams{MaxBytes: 2000000, MaxGas: -5}, true},
		{abci.BlockParams{MaxBytes: 2000000, MaxGas: 300000}, false},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.expectErr, baseapp.ValidateBlockParams(tc.arg) != nil)
	}
}

func TestValidateEvidenceParams(t *testing.T) {
	testCases := []struct {
		arg       interface{}
		expectErr bool
	}{
		{nil, true},
		{&abci.EvidenceParams{}, true},
		{abci.EvidenceParams{}, true},
		{abci.EvidenceParams{MaxAgeNumBlocks: -1, MaxAgeDuration: 18004000}, true},
		{abci.EvidenceParams{MaxAgeNumBlocks: 360000, MaxAgeDuration: -1}, true},
		{abci.EvidenceParams{MaxAgeNumBlocks: 360000, MaxAgeDuration: 18004000}, false},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.expectErr, baseapp.ValidateEvidenceParams(tc.arg) != nil)
	}
}

func TestValidateValidatorParams(t *testing.T) {
	testCases := []struct {
		arg       interface{}
		expectErr bool
	}{
		{nil, true},
		{&abci.ValidatorParams{}, true},
		{abci.ValidatorParams{}, true},
		{abci.ValidatorParams{PubKeyTypes: []string{}}, true},
		{abci.ValidatorParams{PubKeyTypes: []string{"secp256k1"}}, false},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.expectErr, baseapp.ValidateValidatorParams(tc.arg) != nil)
	}
}
