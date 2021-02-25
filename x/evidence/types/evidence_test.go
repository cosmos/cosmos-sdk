package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func TestEquivocation_Valid(t *testing.T) {
	n, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	addr := sdk.ConsAddress("foo_________________")

	e := types.Equivocation{
		Height:           100,
		Time:             n,
		Power:            1000000,
		ConsensusAddress: addr.String(),
	}

	require.Equal(t, e.GetTotalPower(), int64(0))
	require.Equal(t, e.GetValidatorPower(), e.Power)
	require.Equal(t, e.GetTime(), e.Time)
	require.Equal(t, e.GetConsensusAddress().String(), e.ConsensusAddress)
	require.Equal(t, e.GetHeight(), e.Height)
	require.Equal(t, e.Type(), types.TypeEquivocation)
	require.Equal(t, e.Route(), types.RouteEquivocation)
	require.Equal(t, e.Hash().String(), "1E10F9267BEA3A9A4AB5302C2C510CC1AFD7C54E232DA5B2E3360DFAFACF7A76")
	require.Equal(t, e.String(), "height: 100\ntime: 2006-01-02T15:04:05Z\npower: 1000000\nconsensus_address: cosmosvalcons1vehk7h6lta047h6lta047h6lta047h6l8m4r53\n")
	require.NoError(t, e.ValidateBasic())
}

func TestEquivocationValidateBasic(t *testing.T) {
	var zeroTime time.Time
	addr := sdk.ConsAddress("foo_________________")

	n, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	testCases := []struct {
		name      string
		e         types.Equivocation
		expectErr bool
	}{
		{"valid", types.Equivocation{100, n, 1000000, addr.String()}, false},
		{"invalid time", types.Equivocation{100, zeroTime, 1000000, addr.String()}, true},
		{"invalid height", types.Equivocation{0, n, 1000000, addr.String()}, true},
		{"invalid power", types.Equivocation{100, n, 0, addr.String()}, true},
		{"invalid address", types.Equivocation{100, n, 1000000, ""}, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expectErr, tc.e.ValidateBasic() != nil)
		})
	}
}

func TestEvidenceAddressConversion(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForConsensusNode("testcnclcons", "testcnclconspub")
	tmEvidence := abci.Evidence{
		Type: abci.EvidenceType_DUPLICATE_VOTE,
		Validator: abci.Validator{
			Address: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Power:   100,
		},
		Height:           1,
		Time:             time.Now(),
		TotalVotingPower: 100,
	}
	evidence := types.FromABCIEvidence(tmEvidence).(*types.Equivocation)
	consAddr := evidence.GetConsensusAddress()
	// Check the address is the same after conversion
	require.Equal(t, tmEvidence.Validator.Address, consAddr.Bytes())
	sdk.GetConfig().SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
}
