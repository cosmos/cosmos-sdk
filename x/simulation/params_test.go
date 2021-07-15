package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestParamChange(t *testing.T) {
	subspace, key := "theSubspace", "key"
	f := func(r *rand.Rand) string {
		return "theResult"
	}

	pChange := NewSimParamChange(subspace, key, f)

	require.Equal(t, subspace, pChange.Subspace())
	require.Equal(t, key, pChange.Key())
	require.Equal(t, f(nil), pChange.SimValue()(nil))
	require.Equal(t, fmt.Sprintf("%s/%s", subspace, key), pChange.ComposedKey())
}

func TestNewWeightedProposalContent(t *testing.T) {
	key := "theKey"
	weight := 1
	msgs := []sdk.Msg{&testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}}
	f := func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) []sdk.Msg {
		return msgs
	}

	pContent := NewWeightedProposalMessageSim(key, weight, f)

	require.Equal(t, key, pContent.AppParamsKey())
	require.Equal(t, weight, pContent.DefaultWeight())

	ctx := sdk.NewContext(nil, tmproto.Header{}, true, nil)
	require.Equal(t, msgs, pContent.ProposalSimulatorFn()(nil, ctx, nil))
}
