package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

type KeeperTestSuite struct {
	suite.Suite

	cms    sdk.CommitMultiStore
	ctx    sdk.Context
	keeper *keeper.Keeper
}

func (suite *KeeperTestSuite) SetupTest() {
	// create required store keys
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	storeKey := sdk.NewKVStoreKey(types.StoreKey)

	// create required keepers
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	subspace := paramsKeeper.Subspace(types.DefaultParamspace)
	keeper := keeper.NewKeeper(cdc, storeKey, subspace, types.DefaultCodespace)

	// create Evidence router, mount Handlers, and set keeper's router
	router := types.NewRouter()
	router = router.AddRoute(EvidenceRouteEquivocation, EquivocationHandler(*keeper))
	keeper.SetRouter(router)

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

	suite.cms = cms
	suite.ctx = ctx
	suite.keeper = keeper
}

// func (suite *KeeperTestSuite) TestSubmitValidEvidence(t *testing.T) {

// }

func (suite *KeeperTestSuite) TestSubmitInvalidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	e := EquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA: SimpleVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           11,
			Round:            0,
		},
		VoteB: SimpleVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           11,
			Round:            0,
		},
	}

	suite.Error(suite.keeper.SubmitEvidence(ctx, e))

	res, ok := suite.keeper.GetEvidence(ctx, e.Hash())
	suite.False(ok)
	suite.Nil(res)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
