package types_test

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	require.Equal(t, e.GetConsensusAddress(address.NewBech32Codec("cosmosvalcons")).String(), e.ConsensusAddress)
	require.Equal(t, e.GetHeight(), e.Height)
	require.Equal(t, e.Route(), types.RouteEquivocation)
	require.Equal(t, strings.ToUpper(hex.EncodeToString(e.Hash())), "1E10F9267BEA3A9A4AB5302C2C510CC1AFD7C54E232DA5B2E3360DFAFACF7A76")
	require.Equal(t, "height:100 time:<seconds:1136214245 > power:1000000 consensus_address:\"cosmosvalcons1vehk7h6lta047h6lta047h6lta047h6l8m4r53\" ", e.String())
	require.NoError(t, e.ValidateBasic())

	require.Equal(t, int64(0), e.GetTotalPower())
	require.Equal(t, e.Power, e.GetValidatorPower())
	require.Equal(t, e.Time, e.GetTime())
	require.Equal(t, e.ConsensusAddress, e.GetConsensusAddress(address.NewBech32Codec("cosmosvalcons")).String())
	require.Equal(t, e.Height, e.GetHeight())
	require.Equal(t, types.RouteEquivocation, e.Route())
	require.Equal(t, "1E10F9267BEA3A9A4AB5302C2C510CC1AFD7C54E232DA5B2E3360DFAFACF7A76", strings.ToUpper(hex.EncodeToString(e.Hash())))
	require.Equal(t, "height:100 time:<seconds:1136214245 > power:1000000 consensus_address:\"cosmosvalcons1vehk7h6lta047h6lta047h6lta047h6l8m4r53\" ", e.String())
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
	tmEvidence := NewCometMisbehavior(1, 100, time.Now(), comet.DuplicateVote,
		validator{address: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, power: 100})

	evidence := types.FromABCIEvidence(tmEvidence, address.NewBech32Codec("testcnclcons"))
	consAddr := evidence.GetConsensusAddress(address.NewBech32Codec("testcnclcons"))
	// Check the address is the same after conversion
	require.Equal(t, tmEvidence.Validator().Address(), consAddr.Bytes())
	sdk.GetConfig().SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
}

type Misbehavior struct {
	height           int64
	time             time.Time
	totalVotingPower int64
	validator        validator
	misBehaviorType  comet.MisbehaviorType
}

func NewCometMisbehavior(height, tvp int64, t time.Time, tpe comet.MisbehaviorType, val validator) comet.Evidence {
	return Misbehavior{
		height:           height,
		time:             t,
		totalVotingPower: tvp,
		misBehaviorType:  tpe,
		validator:        val,
	}
}

func (m Misbehavior) Type() comet.MisbehaviorType {
	return m.misBehaviorType
}

func (m Misbehavior) Height() int64 {
	return m.height
}

func (m Misbehavior) Validator() comet.Validator {
	return m.validator
}

func (m Misbehavior) Time() time.Time {
	return m.time
}

func (m Misbehavior) TotalVotingPower() int64 {
	return m.totalVotingPower
}

type validator struct {
	address []byte
	power   int64
}

var _ comet.Validator = (*validator)(nil)

func (v validator) Address() []byte {
	return v.address
}

func (v validator) Power() int64 {
	return v.power
}
