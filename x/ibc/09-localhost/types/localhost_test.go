package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
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
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.aminoCdc = app.Codec()
	suite.cdc = clientexported.NewAnyCodec(app.AppCodec())

	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	suite.store = cachekv.NewStore(mem)
}

func TestLocalhostTestSuite(t *testing.T) {
	suite.Run(t, new(LocalhostTestSuite))
}
