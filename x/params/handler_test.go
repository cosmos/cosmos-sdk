package params_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
	"github.com/cosmos/cosmos-sdk/x/params"
)

func testProposal(changes ...params.Change) params.ChangeProposal {
	return params.NewChangeProposal(
		"Test",
		"description",
		changes,
	)
}

func TestProposalPassedEndblocker(t *testing.T) {
	mapp, gk, router, sk, addrs, _, _ := gov.GetMockApp(t, 1, gov.GenesisState{}, nil)

	params.RegisterCodec(mapp.Cdc)

	pk := params.NewProposalKeeper(mapp.ParamsKeeper, gk)
	router.AddRoute(params.RouterKey, params.NewProposalHandler(pk))
	space := mapp.ParamsKeeper.Subspace("myspace").WithKeyTable(params.NewKeyTable(
		[]byte("key"), uint64(0),
	))

	tp := testProposal(params.NewChange("myspace", []byte("key"), nil, []byte("\"1\"")))
	resTags := gov.TestProposal(t, mapp, addrs[0], gk, sk, tp)

	require.Equal(t, sdk.MakeTag(tags.ProposalResult, tags.ActionProposalPassed), resTags[1])

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	var param uint64
	space.Get(ctx, []byte("key"), &param)
	require.Equal(t, param, uint64(1))
}

func TestProposalFailedEndblocker(t *testing.T) {
	mapp, gk, router, sk, addrs, _, _ := gov.GetMockApp(t, 1, gov.GenesisState{}, nil)

	params.RegisterCodec(mapp.Cdc)

	pk := params.NewProposalKeeper(mapp.ParamsKeeper, gk)
	router.AddRoute(params.RouterKey, params.NewProposalHandler(pk))
	space := mapp.ParamsKeeper.Subspace("myspace").WithKeyTable(params.NewKeyTable(
		[]byte("key"), uint64(0),
	))

	tp := testProposal(params.NewChange("myspace", []byte("key"), nil, []byte("invalid")))
	resTags := gov.TestProposal(t, mapp, addrs[0], gk, sk, tp)

	require.Equal(t, sdk.MakeTag(tags.ProposalResult, tags.ActionProposalFailed), resTags[1])

	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	ok := space.Has(ctx, []byte("key"))
	require.False(t, ok)
}
