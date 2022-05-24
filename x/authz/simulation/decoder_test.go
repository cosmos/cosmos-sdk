package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/cosmos-sdk/simapp"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
	"github.com/Stride-Labs/cosmos-sdk/types/kv"
	"github.com/Stride-Labs/cosmos-sdk/x/authz"
	"github.com/Stride-Labs/cosmos-sdk/x/authz/keeper"
	"github.com/Stride-Labs/cosmos-sdk/x/authz/simulation"
	banktypes "github.com/Stride-Labs/cosmos-sdk/x/bank/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	grant, _ := authz.NewGrant(banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("foo", 123))), time.Now().UTC())
	grantBz, err := cdc.Marshal(&grant)
	require.NoError(t, err)
	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: []byte(keeper.GrantKey), Value: grantBz},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectErr   bool
		expectedLog string
	}{
		{"Grant", false, fmt.Sprintf("%v\n%v", grant, grant)},
		{"other", true, ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
