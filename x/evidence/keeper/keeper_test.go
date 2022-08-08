package keeper_test

import (
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetestutil "github.com/cosmos/cosmos-sdk/x/evidence/testutil"
	"github.com/golang/mock/gomock"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/stretchr/testify/suite"
)

var (
	pubkeys = []cryptotypes.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	}

	valAddresses = []sdk.ValAddress{
		sdk.ValAddress(pubkeys[0].Address()),
		sdk.ValAddress(pubkeys[1].Address()),
		sdk.ValAddress(pubkeys[2].Address()),
	}

	// The default power validators are initialized to have within tests
	initAmt   = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))
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
	return func(ctx sdk.Context, e exported.Evidence) error {
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

	ctx     sdk.Context
	querier sdk.Querier

	evidenceKeeper keeper.Keeper
	bankKeeper     *evidencetestutil.MockBankKeeper
	accountKeeper  *evidencetestutil.MockAccountKeeper
	slashingKeeper *evidencetestutil.MockSlashingKeeper
	stakingKeeper  *evidencetestutil.MockStakingKeeper
	queryClient    types.QueryClient
	encCfg         moduletestutil.TestEncodingConfig
}

func (suite *KeeperTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(evidence.AppModuleBasic{})
	key := sdk.NewKVStoreKey(types.StoreKey)
	tkey := sdk.NewTransientStoreKey("evidence_transient_store")
	testCtx := testutil.DefaultContext(key, tkey)
	suite.ctx = testCtx

	ctrl := gomock.NewController(suite.T())

	stakingKeeper := evidencetestutil.NewMockStakingKeeper(ctrl)
	slashingKeeper := evidencetestutil.NewMockSlashingKeeper(ctrl)
	accountKeeper := evidencetestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := evidencetestutil.NewMockBankKeeper(ctrl)

	evidenceKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		stakingKeeper,
		slashingKeeper,
	)

	suite.stakingKeeper = stakingKeeper
	suite.slashingKeeper = slashingKeeper
	suite.bankKeeper = bankKeeper

	router := types.NewRouter()
	router = router.AddRoute(types.RouteEquivocation, testEquivocationHandler(evidenceKeeper))
	evidenceKeeper.SetRouter(router)
	suite.ctx = testCtx.WithBlockHeader(tmproto.Header{Height: 1})
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(evidence.AppModuleBasic{})

	suite.accountKeeper = accountKeeper

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, evidenceKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
	suite.evidenceKeeper = *evidenceKeeper
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

	res, ok := suite.evidenceKeeper.GetEvidence(ctx, e.Hash())
	suite.True(ok)
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

	res, ok := suite.evidenceKeeper.GetEvidence(ctx, e.Hash())
	suite.True(ok)
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

	suite.Error(suite.evidenceKeeper.SubmitEvidence(ctx, e))

	res, ok := suite.evidenceKeeper.GetEvidence(ctx, e.Hash())
	suite.False(ok)
	suite.Nil(res)
}

func (suite *KeeperTestSuite) TestIterateEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100
	suite.populateEvidence(ctx, numEvidence)

	evidence := suite.evidenceKeeper.GetAllEvidence(ctx)
	suite.Len(evidence, numEvidence)
}

func (suite *KeeperTestSuite) TestGetEvidenceHandler() {
	handler, err := suite.evidenceKeeper.GetEvidenceHandler((&types.Equivocation{}).Route())
	suite.NoError(err)
	suite.NotNil(handler)

	handler, err = suite.evidenceKeeper.GetEvidenceHandler("invalidHandler")
	suite.Error(err)
	suite.Nil(handler)
}
