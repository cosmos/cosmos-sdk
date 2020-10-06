package v036

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"

	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	var genesisState GenesisState
	require.NotPanics(t, func() {
		genesisState = Migrate(v034auth.GenesisState{
			CollectedFees: types.Coins{
				{
					Amount: types.NewInt(10),
					Denom:  "stake",
				},
			},
			Params: v034auth.Params{}, // forwarded structure: filling and checking will be testing a no-op
		})
	})
	require.Equal(t, genesisState, GenesisState{Params: v034auth.Params{}})
}
