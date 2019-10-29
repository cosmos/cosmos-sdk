package evidence_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	handler sdk.Handler
	keeper  *evidence.Keeper
}

func (suite *HandlerTestSuite) SetupTest() {
	// create required store keys
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	storeKey := sdk.NewKVStoreKey(evidence.StoreKey)

	// create required keepers
	paramsKeeper := params.NewKeeper(types.TestingCdc, keyParams, tkeyParams, params.DefaultCodespace)
	subspace := paramsKeeper.Subspace(evidence.DefaultParamspace)
	evidenceKeeper := evidence.NewKeeper(types.TestingCdc, storeKey, subspace, evidence.DefaultCodespace)

	// create Evidence router, mount Handlers, and set keeper's router
	router := evidence.NewRouter()
	router = router.AddRoute(types.TestEvidenceRouteEquivocation, types.TestEquivocationHandler(*evidenceKeeper))
	evidenceKeeper.SetRouter(router)

	// create DB, mount stores, and load latest version
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	cms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	suite.Nil(cms.LoadLatestVersion())

	// create initial Context
	ctx := sdk.NewContext(cms, abci.Header{ChainID: "test-chain"}, false, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)

	suite.ctx = ctx
	suite.handler = evidence.NewHandler(*evidenceKeeper)
	suite.keeper = evidenceKeeper
}

func (suite *HandlerTestSuite) TestMsgSubmitEvidence_Valid() {
	pk := ed25519.GenPrivKey()
	sv := types.TestVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}

	sig, err := pk.Sign(sv.SignBytes("test-chain"))
	suite.NoError(err)
	sv.Signature = sig

	s := sdk.AccAddress("test")
	e := types.TestEquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA:      sv,
		VoteB:      sv,
	}

	ctx := suite.ctx.WithIsCheckTx(false)
	msg := evidence.NewMsgSubmitEvidence(e, s)
	res := suite.handler(ctx, msg)
	suite.True(res.IsOK())
	suite.Equal(e.Hash().Bytes(), res.Data)
}

func (suite *HandlerTestSuite) TestMsgSubmitEvidence_Invalid() {
	pk := ed25519.GenPrivKey()
	sv := types.TestVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}

	sig, err := pk.Sign(sv.SignBytes("test-chain"))
	suite.NoError(err)
	sv.Signature = sig

	s := sdk.AccAddress("test")
	e := types.TestEquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA:      sv,
		VoteB:      types.TestVote{Height: 10, Round: 1},
	}

	ctx := suite.ctx.WithIsCheckTx(false)
	msg := evidence.NewMsgSubmitEvidence(e, s)
	res := suite.handler(ctx, msg)
	suite.False(res.IsOK())
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
