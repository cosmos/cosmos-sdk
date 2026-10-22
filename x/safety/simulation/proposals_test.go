package simulation_test

import (
	"math/rand"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/safety/simulation"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

func TestProposalMsgs(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	ctx := sdk.NewContext(nil, cmtproto.Header{}, true, nil)
	accounts := simtypes.RandomAccounts(r, 3)
	gov := sdk.AccAddress(address.Module("gov")).String()
	ac := addresscodec.NewBech32Codec("cosmos")

	msgs := simulation.ProposalMsgs()
	require.Len(t, msgs, 1)
	require.Equal(t, simulation.OpWeightMsgUpdateParams, msgs[0].AppParamsKey())
	require.Equal(t, simulation.DefaultWeightMsgUpdateParams, msgs[0].DefaultWeight())

	var sawOperator, sawNoOperator, sawScope bool
	for range 50 {
		msg, ok := msgs[0].MsgSimulatorFn()(r, ctx, accounts).(*types.MsgUpdateParams)
		require.True(t, ok)
		require.Equal(t, gov, msg.Authority)
		require.Equal(t, gov, msg.Params.Authority, "the authority must stay gov so later proposals still pass")
		require.NoError(t, msg.Validate(ac))
		sawOperator = sawOperator || msg.Params.Operator != ""
		sawNoOperator = sawNoOperator || msg.Params.Operator == ""
		sawScope = sawScope || len(msg.Params.Scope) > 0
	}
	require.True(t, sawOperator && sawNoOperator && sawScope)
}

type moduleAccounts map[string]sdk.AccAddress

func (m moduleAccounts) GetModuleAddress(name string) sdk.AccAddress { return m[name] }

// The simsx factory builds a valid MsgUpdateParams that gov signs and that
// keeps gov as the authority, so later proposals still pass.
func TestMsgUpdateParamsFactory(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	accounts := simtypes.RandomAccounts(r, 3)
	govAddr := sdk.AccAddress(address.Module("gov"))
	gov := govAddr.String()
	ac := addresscodec.NewBech32Codec("cosmos")
	data := simsx.NewChainDataSource(sdk.Context{}, r, moduleAccounts{"gov": govAddr}, nil, ac, accounts...)
	factory := simulation.MsgUpdateParamsFactory()

	var sawOperator, sawNoOperator, sawScope bool
	for range 50 {
		reporter := simsx.NewBasicSimulationReporter()
		signers, msg := factory(sdk.Context{}, data, reporter)
		require.False(t, reporter.IsSkipped())
		require.Empty(t, signers)
		require.Equal(t, gov, msg.Authority)
		require.Equal(t, gov, msg.Params.Authority)
		require.NoError(t, msg.Validate(ac))
		sawOperator = sawOperator || msg.Params.Operator != ""
		sawNoOperator = sawNoOperator || msg.Params.Operator == ""
		sawScope = sawScope || len(msg.Params.Scope) > 0
	}
	require.True(t, sawOperator && sawNoOperator && sawScope)
}
