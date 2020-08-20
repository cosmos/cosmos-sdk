package types_test

import (
	"fmt"
	pt "github.com/gogo/protobuf/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestParamsEqual(t *testing.T) {
	p1 := types.DefaultParams()
	p2 := types.DefaultParams()
	require.Equal(t, p1, p2)

	p1.TxSigLimit.Value += 10
	require.NotEqual(t, p1, p2)
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  types.Params
		wantErr error
	}{
		{"default params", types.DefaultParams(), nil},
		{
			"invalid tx signature limit",
			types.NewParams(types.DefaultMaxMemoCharacters, pt.UInt64Value{Value: 0},
				types.DefaultTxSizeCostPerByte, types.DefaultSigVerifyCostED25519, types.DefaultSigVerifyCostSecp256k1),
			fmt.Errorf("invalid tx signature limit: 0")},
		{
			"invalid ED25519 signature verification cost",
			types.NewParams(types.DefaultMaxMemoCharacters, types.DefaultTxSigLimit,
				types.DefaultTxSizeCostPerByte, pt.UInt64Value{Value: 0}, types.DefaultSigVerifyCostSecp256k1),
			fmt.Errorf("invalid ED25519 signature verification cost: 0")},
		{
			"invalid SECK256k1 signature verification cost",
			types.NewParams(types.DefaultMaxMemoCharacters, types.DefaultTxSigLimit,
				types.DefaultTxSizeCostPerByte, types.DefaultSigVerifyCostED25519, pt.UInt64Value{Value: 0}),
			fmt.Errorf("invalid SECK256k1 signature verification cost: 0")},
		{
			"invalid max memo characters",
			types.NewParams(pt.UInt64Value{Value: 0}, types.DefaultTxSigLimit, types.DefaultTxSizeCostPerByte,
				types.DefaultSigVerifyCostED25519, types.DefaultSigVerifyCostSecp256k1),
			fmt.Errorf("invalid max memo characters: 0")},
		{
			"invalid tx size cost per byte",
			types.NewParams(types.DefaultMaxMemoCharacters, types.DefaultTxSigLimit, pt.UInt64Value{Value: 0},
				types.DefaultSigVerifyCostED25519, types.DefaultSigVerifyCostSecp256k1),
			fmt.Errorf("invalid tx size cost per byte: 0")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.Validate()
			if tt.wantErr == nil {
				require.NoError(t, got)
				return
			}
			require.Equal(t, tt.wantErr, got)
		})
	}
}
