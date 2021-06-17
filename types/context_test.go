package types_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/tests/mocks"
	"github.com/cosmos/cosmos-sdk/types"
)

type contextTestSuite struct {
	suite.Suite
}

func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(contextTestSuite))
}

func (s *contextTestSuite) defaultContext(key types.StoreKey) types.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, types.StoreTypeIAVL, db)
	s.Require().NoError(cms.LoadLatestVersion())
	ctx := types.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())
	return ctx
}

func (s *contextTestSuite) TestCacheContext() {
	key := types.NewKVStoreKey(s.T().Name() + "_TestCacheContext")
	k1 := []byte("hello")
	v1 := []byte("world")
	k2 := []byte("key")
	v2 := []byte("value")

	ctx := s.defaultContext(key)
	store := ctx.KVStore(key)
	store.Set(k1, v1)
	s.Require().Equal(v1, store.Get(k1))
	s.Require().Nil(store.Get(k2))

	cctx, write := ctx.CacheContext()
	cstore := cctx.KVStore(key)
	s.Require().Equal(v1, cstore.Get(k1))
	s.Require().Nil(cstore.Get(k2))

	cstore.Set(k2, v2)
	s.Require().Equal(v2, cstore.Get(k2))
	s.Require().Nil(store.Get(k2))

	write()

	s.Require().Equal(v2, store.Get(k2))
}

func (s *contextTestSuite) TestLogContext() {
	key := types.NewKVStoreKey(s.T().Name())
	ctx := s.defaultContext(key)
	ctrl := gomock.NewController(s.T())
	s.T().Cleanup(ctrl.Finish)

	logger := mocks.NewMockLogger(ctrl)
	logger.EXPECT().Debug("debug")
	logger.EXPECT().Info("info")
	logger.EXPECT().Error("error")

	ctx = ctx.WithLogger(logger)
	ctx.Logger().Debug("debug")
	ctx.Logger().Info("info")
	ctx.Logger().Error("error")
}

type dummy int64 //nolint:unused

func (d dummy) Clone() interface{} {
	return d
}

// Testing saving/loading sdk type values to/from the context
func (s *contextTestSuite) TestContextWithCustom() {
	var ctx types.Context
	s.Require().True(ctx.IsZero())

	ctrl := gomock.NewController(s.T())
	s.T().Cleanup(ctrl.Finish)

	header := tmproto.Header{}
	height := int64(1)
	chainid := "chainid"
	ischeck := true
	txbytes := []byte("txbytes")
	logger := mocks.NewMockLogger(ctrl)
	voteinfos := []abci.VoteInfo{{}}
	meter := types.NewGasMeter(10000)
	blockGasMeter := types.NewGasMeter(20000)
	minGasPrices := types.DecCoins{types.NewInt64DecCoin("feetoken", 1)}
	headerHash := []byte("headerHash")

	ctx = types.NewContext(nil, header, ischeck, logger)
	s.Require().Equal(header, ctx.BlockHeader())

	ctx = ctx.
		WithBlockHeight(height).
		WithChainID(chainid).
		WithTxBytes(txbytes).
		WithVoteInfos(voteinfos).
		WithGasMeter(meter).
		WithMinGasPrices(minGasPrices).
		WithBlockGasMeter(blockGasMeter).
		WithHeaderHash(headerHash)
	s.Require().Equal(height, ctx.BlockHeight())
	s.Require().Equal(chainid, ctx.ChainID())
	s.Require().Equal(ischeck, ctx.IsCheckTx())
	s.Require().Equal(txbytes, ctx.TxBytes())
	s.Require().Equal(logger, ctx.Logger())
	s.Require().Equal(voteinfos, ctx.VoteInfos())
	s.Require().Equal(meter, ctx.GasMeter())
	s.Require().Equal(minGasPrices, ctx.MinGasPrices())
	s.Require().Equal(blockGasMeter, ctx.BlockGasMeter())
	s.Require().Equal(headerHash, ctx.HeaderHash().Bytes())
	s.Require().False(ctx.WithIsCheckTx(false).IsCheckTx())

	// test IsReCheckTx
	s.Require().False(ctx.IsReCheckTx())
	ctx = ctx.WithIsCheckTx(false)
	ctx = ctx.WithIsReCheckTx(true)
	s.Require().True(ctx.IsCheckTx())
	s.Require().True(ctx.IsReCheckTx())

	// test consensus param
	s.Require().Nil(ctx.ConsensusParams())
	cp := &abci.ConsensusParams{}
	s.Require().Equal(cp, ctx.WithConsensusParams(cp).ConsensusParams())

	// test inner context
	newContext := context.WithValue(ctx.Context(), "key", "value") //nolint:golint,staticcheck
	s.Require().NotEqual(ctx.Context(), ctx.WithContext(newContext).Context())
}

// Testing saving/loading of header fields to/from the context
func (s *contextTestSuite) TestContextHeader() {
	var ctx types.Context

	height := int64(5)
	time := time.Now()
	addr := secp256k1.GenPrivKey().PubKey().Address()
	proposer := types.ConsAddress(addr)

	ctx = types.NewContext(nil, tmproto.Header{}, false, nil)

	ctx = ctx.
		WithBlockHeight(height).
		WithBlockTime(time).
		WithProposer(proposer)
	s.Require().Equal(height, ctx.BlockHeight())
	s.Require().Equal(height, ctx.BlockHeader().Height)
	s.Require().Equal(time.UTC(), ctx.BlockHeader().Time)
	s.Require().Equal(proposer.Bytes(), ctx.BlockHeader().ProposerAddress)
}

func (s *contextTestSuite) TestContextHeaderClone() {
	cases := map[string]struct {
		h tmproto.Header
	}{
		"empty": {
			h: tmproto.Header{},
		},
		"height": {
			h: tmproto.Header{
				Height: 77,
			},
		},
		"time": {
			h: tmproto.Header{
				Time: time.Unix(12345677, 12345),
			},
		},
		"zero time": {
			h: tmproto.Header{
				Time: time.Unix(0, 0),
			},
		},
		"many items": {
			h: tmproto.Header{
				Height:  823,
				Time:    time.Unix(9999999999, 0),
				ChainID: "silly-demo",
			},
		},
		"many items with hash": {
			h: tmproto.Header{
				Height:        823,
				Time:          time.Unix(9999999999, 0),
				ChainID:       "silly-demo",
				AppHash:       []byte{5, 34, 11, 3, 23},
				ConsensusHash: []byte{11, 3, 23, 87, 3, 1},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		s.T().Run(name, func(t *testing.T) {
			ctx := types.NewContext(nil, tc.h, false, nil)
			s.Require().Equal(tc.h.Height, ctx.BlockHeight())
			s.Require().Equal(tc.h.Time.UTC(), ctx.BlockTime())

			// update only changes one field
			var newHeight int64 = 17
			ctx = ctx.WithBlockHeight(newHeight)
			s.Require().Equal(newHeight, ctx.BlockHeight())
			s.Require().Equal(tc.h.Time.UTC(), ctx.BlockTime())
		})
	}
}

func (s *contextTestSuite) TestUnwrapSDKContext() {
	sdkCtx := types.NewContext(nil, tmproto.Header{}, false, nil)
	ctx := types.WrapSDKContext(sdkCtx)
	sdkCtx2 := types.UnwrapSDKContext(ctx)
	s.Require().Equal(sdkCtx, sdkCtx2)

	ctx = context.Background()
	s.Require().Panics(func() { types.UnwrapSDKContext(ctx) })
}
