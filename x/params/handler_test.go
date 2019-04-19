package params_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

type testInput struct {
	ctx    sdk.Context
	cdc    *codec.Codec
	keeper params.Keeper
}

func testProposal(changes ...params.ParamChange) params.ParameterChangeProposal {
	return params.NewParameterChangeProposal(
		"Test",
		"description",
		changes,
	)
}

func newTestInput(t *testing.T) testInput {
	cdc := codec.New()
	types.RegisterCodec(cdc)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)

	keyParams := sdk.NewKVStoreKey("params")
	tKeyParams := sdk.NewTransientStoreKey("transient_params")

	cms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tKeyParams, sdk.StoreTypeTransient, db)

	err := cms.LoadLatestVersion()
	require.Nil(t, err)

	keeper := params.NewKeeper(cdc, keyParams, tKeyParams, params.DefaultCodespace)
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())

	return testInput{ctx, cdc, keeper}
}

func TestProposalHandlerPassed(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace("testSubspace").WithKeyTable(
		params.NewKeyTable([]byte("testKey"), uint64(0)),
	)

	tp := testProposal(params.NewParamChange("testSubspace", "testKey", nil, []byte("\"1\"")))
	hdlr := params.NewProposalHandler(input.keeper)
	require.NoError(t, hdlr(input.ctx, tp))

	var param uint64
	ss.Get(input.ctx, []byte("testKey"), &param)
	require.Equal(t, param, uint64(1))
}

func TestProposalHandlerFailed(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace("testSubspace").WithKeyTable(
		params.NewKeyTable([]byte("testKey"), uint64(0)),
	)

	tp := testProposal(params.NewParamChange("testSubspace", "testKey", nil, []byte("invalid")))
	hdlr := params.NewProposalHandler(input.keeper)
	require.Error(t, hdlr(input.ctx, tp))

	require.False(t, ss.Has(input.ctx, []byte("testKey")))
}
