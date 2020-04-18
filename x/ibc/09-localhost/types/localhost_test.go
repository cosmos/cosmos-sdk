package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	height = 4
)

type LocalhostTestSuite struct {
	suite.Suite

	cdc   *codec.Codec
	store *cachekv.Store
}

func (suite *LocalhostTestSuite) SetupTest() {
	suite.cdc = codec.New()
	codec.RegisterCrypto(suite.cdc)
	commitmenttypes.RegisterCodec(suite.cdc)

	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	suite.store = cachekv.NewStore(mem)
}

func TestLocalhostTestSuite(t *testing.T) {
	suite.Run(t, new(LocalhostTestSuite))
}
