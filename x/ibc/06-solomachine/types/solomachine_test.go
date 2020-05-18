package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcsmtypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine"
)

const (
	zero = 0
)

type SoloMachineTestSuite struct {
	suite.Suite

	ctx      sdk.Context
	aminoCdc *codec.Codec
	cdc      codec.Marshaler
	privKey  crypto.PrivKey
	sequence uint64
	now      time.Time
}

func (suite *SoloMachineTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.aminoCdc = app.Codec()
	suite.cdc = app.AppCodec()

	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.privKey = ed25519.GenPrivKey()

	suite.sequence = 1
	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, Time: suite.now})
}

func TestSoloMachineTestSuite(t *testing.T) {
	suite.Run(t, new(SoloMachineTestSuite))
}

func (suite *SoloMachineTestSuite) CreateHeader() ibcsmtypes.Header {
	newPrivKey := ed25519.GenPrivKey()
	signature, err := suite.privKey.Sign(newPrivKey.PubKey().Bytes())
	suite.Require().NoError(err)

	suite.sequence++
	suite.privKey = newPrivKey

	return ibcsmtypes.Header{
		Sequence:  suite.sequence,
		Signature: signature,
		NewPubKey: suite.privKey.PubKey(),
	}
}
