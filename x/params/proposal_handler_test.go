package params_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func validateNoOp(_ interface{}) error { return nil }

type testInput struct {
	ctx    sdk.Context
	cdc    *codec.LegacyAmino
	keeper keeper.Keeper
}

var (
	_ types.ParamSet = (*testParams)(nil)

	keyMaxValidators = "MaxValidators"
	keySlashingRate  = "SlashingRate"
	testSubspace     = "TestSubspace"
)

type testParamsSlashingRate struct {
	DoubleSign uint16 `json:"double_sign,omitempty" yaml:"double_sign,omitempty"`
	Downtime   uint16 `json:"downtime,omitempty" yaml:"downtime,omitempty"`
}

type testParams struct {
	MaxValidators uint16                 `json:"max_validators" yaml:"max_validators"` // maximum number of validators (max uint16 = 65535)
	SlashingRate  testParamsSlashingRate `json:"slashing_rate" yaml:"slashing_rate"`
}

func (tp *testParams) ParamSetPairs() types.ParamSetPairs {
	return types.ParamSetPairs{
		types.NewParamSetPair([]byte(keyMaxValidators), &tp.MaxValidators, validateNoOp),
		types.NewParamSetPair([]byte(keySlashingRate), &tp.SlashingRate, validateNoOp),
	}
}

func testProposal(changes ...proposal.ParamChange) *proposal.ParameterChangeProposal {
	return proposal.NewParameterChangeProposal(
		"Test",
		"description",
		changes,
	)
}

func newTestInput(t *testing.T) testInput {
	cdc := codec.NewLegacyAmino()
	proposal.RegisterLegacyAminoCodec(cdc)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)

	keyParams := sdk.NewKVStoreKey("params")
	tKeyParams := sdk.NewTransientStoreKey("transient_params")

	cms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tKeyParams, sdk.StoreTypeTransient, db)

	err := cms.LoadLatestVersion()
	require.Nil(t, err)

	encCfg := simapp.MakeTestEncodingConfig()
	keeper := keeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, keyParams, tKeyParams)
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	return testInput{ctx, cdc, keeper}
}

func TestProposalHandlerPassed(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		types.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	tp := testProposal(proposal.NewParamChange(testSubspace, keyMaxValidators, "1"))
	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	require.NoError(t, hdlr(input.ctx, tp))

	var param uint16
	ss.Get(input.ctx, []byte(keyMaxValidators), &param)
	require.Equal(t, param, uint16(1))
}

func TestProposalHandlerFailed(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		types.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	tp := testProposal(proposal.NewParamChange(testSubspace, keyMaxValidators, "invalidType"))
	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	require.Error(t, hdlr(input.ctx, tp))

	require.False(t, ss.Has(input.ctx, []byte(keyMaxValidators)))
}

func TestProposalHandlerUpdateOmitempty(t *testing.T) {
	input := newTestInput(t)
	ss := input.keeper.Subspace(testSubspace).WithKeyTable(
		types.NewKeyTable().RegisterParamSet(&testParams{}),
	)

	hdlr := params.NewParamChangeProposalHandler(input.keeper)
	var param testParamsSlashingRate

	tp := testProposal(proposal.NewParamChange(testSubspace, keySlashingRate, `{"downtime": 7}`))
	require.NoError(t, hdlr(input.ctx, tp))

	ss.Get(input.ctx, []byte(keySlashingRate), &param)
	require.Equal(t, testParamsSlashingRate{0, 7}, param)

	tp = testProposal(proposal.NewParamChange(testSubspace, keySlashingRate, `{"double_sign": 10}`))
	require.NoError(t, hdlr(input.ctx, tp))

	ss.Get(input.ctx, []byte(keySlashingRate), &param)
	require.Equal(t, testParamsSlashingRate{10, 7}, param)
}
