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

	cms     sdk.CommitMultiStore
	ctx     sdk.Context
	querier sdk.Querier
	keeper  *keeper.Keeper
}

func (suite *KeeperTestSuite) SetupTest() {
	// create required store keys
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	storeKey := sdk.NewKVStoreKey(types.StoreKey)

	// create required keepers
	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	subspace := paramsKeeper.Subspace(types.DefaultParamspace)
	evidenceKeeper := keeper.NewKeeper(cdc, storeKey, subspace, types.DefaultCodespace)

	// create Evidence router, mount Handlers, and set keeper's router
	router := types.NewRouter()
	router = router.AddRoute(EvidenceRouteEquivocation, EquivocationHandler(*evidenceKeeper))
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

	suite.cms = cms
	suite.ctx = ctx
	suite.querier = keeper.NewQuerier(*evidenceKeeper)
	suite.keeper = evidenceKeeper
}

func (suite *KeeperTestSuite) populateEvidence(ctx sdk.Context, numEvidence int) []types.Evidence {
	evidence := make([]types.Evidence, numEvidence)

	for i := 0; i < numEvidence; i++ {
		pk := ed25519.GenPrivKey()
		sv := SimpleVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           int64(i),
			Round:            0,
		}

		sig, err := pk.Sign(sv.SignBytes(ctx.ChainID()))
		suite.NoError(err)
		sv.Signature = sig

		evidence[i] = EquivocationEvidence{
			Power:      100,
			TotalPower: 100000,
			PubKey:     pk.PubKey(),
			VoteA:      sv,
			VoteB:      sv,
		}

		suite.Nil(suite.keeper.SubmitEvidence(ctx, evidence[i]))
	}

	return evidence
}

func (suite *KeeperTestSuite) TestSubmitValidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	sv := SimpleVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}

	sig, err := pk.Sign(sv.SignBytes(ctx.ChainID()))
	suite.NoError(err)
	sv.Signature = sig

	e := EquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA:      sv,
		VoteB:      sv,
	}

	suite.Nil(suite.keeper.SubmitEvidence(ctx, e))

	res, ok := suite.keeper.GetEvidence(ctx, e.Hash())
	suite.True(ok)
	suite.Equal(e, res)
}

func (suite *KeeperTestSuite) TestSubmitInvalidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	e := EquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA: SimpleVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           10,
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

func (suite *KeeperTestSuite) TestIterateEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100
	suite.populateEvidence(ctx, numEvidence)

	evidence := suite.keeper.GetAllEvidence(ctx)
	suite.Len(evidence, numEvidence)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
