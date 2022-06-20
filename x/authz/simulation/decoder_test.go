package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/depinject"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestDecodeStore(t *testing.T) {
	var cdc codec.Codec
	depinject.Inject(testutil.AppConfig, &cdc)

	dec := simulation.NewDecodeStore(cdc)

	now := time.Now().UTC()
	e := now.Add(1)
	grant, _ := authz.NewGrant(now, banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewInt64Coin("foo", 123))), &e)
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
