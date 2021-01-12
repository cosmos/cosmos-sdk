package baseapp_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
		{&tmproto.EvidenceParams{}, true},
		{tmproto.EvidenceParams{}, true},
		{tmproto.EvidenceParams{MaxAgeNumBlocks: -1, MaxAgeDuration: 18004000, MaxBytes: 5000000}, true},
		{tmproto.EvidenceParams{MaxAgeNumBlocks: 360000, MaxAgeDuration: -1, MaxBytes: 5000000}, true},
		{tmproto.EvidenceParams{MaxAgeNumBlocks: 360000, MaxAgeDuration: 18004000, MaxBytes: -1}, true},
		{tmproto.EvidenceParams{MaxAgeNumBlocks: 360000, MaxAgeDuration: 18004000, MaxBytes: 5000000}, false},
		{tmproto.EvidenceParams{MaxAgeNumBlocks: 360000, MaxAgeDuration: 18004000, MaxBytes: 0}, false},
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
		{&tmproto.ValidatorParams{}, true},
		{tmproto.ValidatorParams{}, true},
		{tmproto.ValidatorParams{PubKeyTypes: []string{}}, true},
		{tmproto.ValidatorParams{PubKeyTypes: []string{"secp256k1"}}, false},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.expectErr, baseapp.ValidateValidatorParams(tc.arg) != nil)
	}
}
