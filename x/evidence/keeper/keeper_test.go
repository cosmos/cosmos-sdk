package keeper_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetestutil "github.com/cosmos/cosmos-sdk/x/evidence/testutil"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

var (
	pubkeys = []cryptotypes.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	}

	valAddress = sdk.ValAddress(pubkeys[0].Address())
)

func newPubKey(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}

	pubkey := &ed25519.PubKey{Key: pkBytes}

	return pubkey
}

func testEquivocationHandler(_ interface{}) types.Handler {
	return func(ctx context.Context, e exported.Evidence) error {
		if err := e.ValidateBasic(); err != nil {
			return err
		}

		ee, ok := e.(*types.Equivocation)
		if !ok {
			return fmt.Errorf("unexpected evidence type: %T", e)
		}
		if ee.Height%2 == 0 {
			return fmt.Errorf("unexpected even evidence height: %d", ee.Height)
		}

		return nil
	}
}

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context

	evidenceKeeper keeper.Keeper
	bankKeeper     *evidencetestutil.MockBankKeeper
	accountKeeper  *evidencetestutil.MockAccountKeeper
	slashingKeeper *evidencetestutil.MockSlashingKeeper
	stakingKeeper  *evidencetestutil.MockStakingKeeper
	blockInfo      *evidencetestutil.MockCometinfo
	queryClient    types.QueryClient
	encCfg         moduletestutil.TestEncodingConfig
	msgServer      types.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(evidence.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	tkey := storetypes.NewTransientStoreKey("evidence_transient_store")
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, tkey)
	suite.ctx = testCtx.Ctx

	ctrl := gomock.NewController(suite.T())

	stakingKeeper := evidencetestutil.NewMockStakingKeeper(ctrl)
	slashingKeeper := evidencetestutil.NewMockSlashingKeeper(ctrl)
	accountKeeper := evidencetestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := evidencetestutil.NewMockBankKeeper(ctrl)
	suite.blockInfo = &evidencetestutil.MockCometinfo{}

	evidenceKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		stakingKeeper,
		slashingKeeper,
		address.NewBech32Codec("cosmos"),
		&evidencetestutil.MockCometinfo{},
	)

	suite.stakingKeeper = stakingKeeper
	suite.slashingKeeper = slashingKeeper
	suite.bankKeeper = bankKeeper

	router := types.NewRouter()
	router = router.AddRoute(types.RouteEquivocation, testEquivocationHandler(evidenceKeeper))
	evidenceKeeper.SetRouter(router)

	suite.ctx = testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(evidence.AppModuleBasic{})

	suite.accountKeeper = accountKeeper

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(evidenceKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)
	suite.evidenceKeeper = *evidenceKeeper

	suite.Require().Equal(testCtx.Ctx.Logger().With("module", "x/"+types.ModuleName),
		suite.evidenceKeeper.Logger(testCtx.Ctx))

	suite.msgServer = keeper.NewMsgServerImpl(suite.evidenceKeeper)
}

func (suite *KeeperTestSuite) populateEvidence(ctx sdk.Context, numEvidence int) []exported.Evidence {
	evidence := make([]exported.Evidence, numEvidence)

	for i := 0; i < numEvidence; i++ {
		pk := ed25519.GenPrivKey()

		evidence[i] = &types.Equivocation{
			Height:           11,
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()).String(),
		}

		suite.Nil(suite.evidenceKeeper.SubmitEvidence(ctx, evidence[i]))
	}

	return evidence
}

func (suite *KeeperTestSuite) TestSubmitValidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()

	e := &types.Equivocation{
		Height:           1,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()).String(),
	}

	suite.Nil(suite.evidenceKeeper.SubmitEvidence(ctx, e))

	res, err := suite.evidenceKeeper.Evidences.Get(ctx, e.Hash())
	suite.NoError(err)
	suite.Equal(e, res)
}

func (suite *KeeperTestSuite) TestSubmitValidEvidence_Duplicate() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()

	e := &types.Equivocation{
		Height:           1,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()).String(),
	}

	suite.Nil(suite.evidenceKeeper.SubmitEvidence(ctx, e))
	suite.Error(suite.evidenceKeeper.SubmitEvidence(ctx, e))

	res, err := suite.evidenceKeeper.Evidences.Get(ctx, e.Hash())
	suite.NoError(err)
	suite.Equal(e, res)
}

func (suite *KeeperTestSuite) TestSubmitInvalidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	e := &types.Equivocation{
		Height:           0,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()).String(),
	}

	err := suite.evidenceKeeper.SubmitEvidence(ctx, e)
	suite.ErrorIs(err, types.ErrInvalidEvidence)

	res, err := suite.evidenceKeeper.Evidences.Get(ctx, e.Hash())
	suite.ErrorIs(err, collections.ErrNotFound)
	suite.Nil(res)
}

func (suite *KeeperTestSuite) TestIterateEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100
	suite.populateEvidence(ctx, numEvidence)

	var evidences []exported.Evidence
	suite.Require().NoError(suite.evidenceKeeper.Evidences.Walk(ctx, nil, func(key []byte, value exported.Evidence) (stop bool, err error) {
		evidences = append(evidences, value)
		return false, nil
	}))
	suite.Len(evidences, numEvidence)
}

func (suite *KeeperTestSuite) TestGetEvidenceHandler() {
	handler, err := suite.evidenceKeeper.GetEvidenceHandler((&types.Equivocation{}).Route())
	suite.NoError(err)
	suite.NotNil(handler)

	handler, err = suite.evidenceKeeper.GetEvidenceHandler("invalidHandler")
	suite.Error(err)
	suite.Nil(handler)
}
