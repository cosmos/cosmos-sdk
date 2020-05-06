package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
)

const (
	height = 4
)

type LocalhostTestSuite struct {
	suite.Suite

	aminoCdc *codec.Codec
	cdc      codec.Marshaler
	store    *cachekv.Store
}

func (suite *LocalhostTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.aminoCdc = app.Codec()
	suite.cdc = app.AppCodec()

	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	suite.store = cachekv.NewStore(mem)
}

func TestLocalhostTestSuite(t *testing.T) {
	suite.Run(t, new(LocalhostTestSuite))
}
