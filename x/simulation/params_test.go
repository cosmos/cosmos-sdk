package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
	content := &testContent{}
	f := func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		return content
	}

	pContent := NewWeightedProposalContent(key, weight, f)

	require.Equal(t, key, pContent.AppParamsKey())
	require.Equal(t, weight, pContent.DefaultWeight())

	ctx := sdk.NewContext(nil, tmproto.Header{}, true, nil)
	require.Equal(t, content, pContent.ContentSimulatorFn()(nil, ctx, nil))
}

type testContent struct {
}

func (t testContent) GetTitle() string       { return "" }
func (t testContent) GetDescription() string { return "" }
func (t testContent) ProposalRoute() string  { return "" }
func (t testContent) ProposalType() string   { return "" }
func (t testContent) ValidateBasic() error   { return nil }
func (t testContent) String() string         { return "" }
