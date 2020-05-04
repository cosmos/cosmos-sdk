package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	height = 4
)

type LocalhostTestSuite struct {
	suite.Suite

	aminoCdc *codec.Codec
	cdc      clientexported.Codec
	store    *cachekv.Store
}

func (suite *LocalhostTestSuite) SetupTest() {
	suite.aminoCdc = codec.New()
	codec.RegisterCrypto(suite.aminoCdc)
	commitmenttypes.RegisterCodec(suite.aminoCdc)
	suite.cdc = std.NewAppCodec(suite.aminoCdc)

	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	suite.store = cachekv.NewStore(mem)
}

func TestLocalhostTestSuite(t *testing.T) {
	suite.Run(t, new(LocalhostTestSuite))
}
