package keeper_test

import (
	gocontext "context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAllProposal(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.GovKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	// check for the proposals with no proposal added should return null.
	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req := types.NewQueryProposalsRequest(0, nil, nil, pageReq)

	proposals, err := queryClient.AllProposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.Equal(t, string(proposals.Proposals), "null")

	// create 2 test proposals
	for i := 0; i < 2; i++ {
		num := strconv.Itoa(i + 1)
		testProposal := types.NewTextProposal("Proposal"+num, "testing proposal "+num)
		_, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
		require.NoError(t, err)
	}

	// Query for proposals after adding 2 proposals to the store.
	// give page limit as 1 and expect NextKey should not to be empty
	proposals, err = queryClient.AllProposals(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, proposals.Proposals)
	require.NotEmpty(t, proposals.Res.NextKey)

	pageReq = &query.PageRequest{
		Key:        proposals.Res.NextKey,
		Limit:      1,
		CountTotal: false,
	}

	req = types.NewQueryProposalsRequest(0, nil, nil, pageReq)

	// query for the next page which is 2nd proposal at present context.
	// and expect NextKey should be empty
	proposals, err = queryClient.AllProposals(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, proposals.Proposals)
	require.Empty(t, proposals.Res)

	pageReq = &query.PageRequest{
		Key:        nil,
		Limit:      2,
		CountTotal: false,
	}

	req = types.NewQueryProposalsRequest(0, nil, nil, pageReq)

	// Query the page with limit 2 and expect NextKey should ne nil
	proposals, err = queryClient.AllProposals(gocontext.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, proposals.Proposals)
	require.Empty(t, proposals.Res)
}
