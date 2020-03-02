package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/stretchr/testify/require"
)

func TestIncrementProposalNumber(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	tp := TestProposal
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	proposal6, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.ProposalID)
}
