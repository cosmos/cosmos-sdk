package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
)

func TestEquivocation_Valid(t *testing.T) {
	n, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	e := types.Equivocation{
		Height:           100,
		Time:             n,
		Power:            1000000,
		ConsensusAddress: sdk.ConsAddress("foo"),
	}

	require.Equal(t, e.GetTotalPower(), int64(0))
	require.Equal(t, e.GetValidatorPower(), e.Power)
	require.Equal(t, e.GetTime(), e.Time)
	require.Equal(t, e.GetConsensusAddress(), e.ConsensusAddress)
	require.Equal(t, e.GetHeight(), e.Height)
	require.Equal(t, e.Type(), types.TypeEquivocation)
	require.Equal(t, e.Route(), types.RouteEquivocation)
	require.Equal(t, e.Hash().String(), "808DA679674C9C0599965D02EBC5D4DCFD5E700D03035BBCD2DECCBBF44386F7")
	require.Equal(t, e.String(), "height: 100\ntime: 2006-01-02T15:04:05Z\npower: 1000000\nconsensus_address: cosmosvalcons1vehk7pqt5u4\n")
	require.NoError(t, e.ValidateBasic())
}

func TestEquivocationValidateBasic(t *testing.T) {
	var zeroTime time.Time

	n, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	testCases := []struct {
		name      string
		e         types.Equivocation
		expectErr bool
	}{
		{"valid", types.Equivocation{100, n, 1000000, sdk.ConsAddress("foo")}, false},
		{"invalid time", types.Equivocation{100, zeroTime, 1000000, sdk.ConsAddress("foo")}, true},
		{"invalid height", types.Equivocation{0, n, 1000000, sdk.ConsAddress("foo")}, true},
		{"invalid power", types.Equivocation{100, n, 0, sdk.ConsAddress("foo")}, true},
		{"invalid address", types.Equivocation{100, n, 1000000, nil}, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectErr, tc.e.ValidateBasic() != nil)
		})
	}
}
