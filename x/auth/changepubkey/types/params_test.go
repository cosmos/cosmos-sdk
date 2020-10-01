package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

func TestParamsEqual(t *testing.T) {
	p1 := types.DefaultParams()
	p2 := types.DefaultParams()
	require.Equal(t, p1, p2)

	p1.PubKeyChangeCost += 10
	require.NotEqual(t, p1, p2)
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  types.Params
		wantErr error
	}{
		{"default params", types.DefaultParams(), nil},
		{"invalid pubkey change cost value", types.NewParams(0), fmt.Errorf("invalid pubkey change cost: 0")},
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
