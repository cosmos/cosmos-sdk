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
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine"
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
	clientID string
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
	suite.clientID = "solomachineclient"
	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, Time: suite.now})
}

func TestSoloMachineTestSuite(t *testing.T) {
	suite.Run(t, new(SoloMachineTestSuite))
}

func (suite *SoloMachineTestSuite) CreateHeader() solomachinetypes.Header {
	newPrivKey := ed25519.GenPrivKey()
	signature, err := suite.privKey.Sign(newPrivKey.PubKey().Bytes())
	suite.Require().NoError(err)

	suite.sequence++
	suite.privKey = newPrivKey

	return solomachinetypes.Header{
		Sequence:  suite.sequence,
		Signature: signature,
		NewPubKey: suite.privKey.PubKey(),
	}
}

func (suite *SoloMachineTestSuite) ConsensusState() solomachinetypes.ConsensusState {
	return solomachinetypes.ConsensusState{
		Sequence: suite.sequence,
		PubKey:   suite.privKey.PubKey(),
	}
}

func (suite *SoloMachineTestSuite) Evidence() solomachinetypes.Evidence {
	dataOne := []byte("DATA ONE")
	dataTwo := []byte("DATA TWO")

	sig, err := suite.privKey.Sign(append(sdk.Uint64ToBigEndian(suite.sequence), dataOne...))
	suite.Require().NoError(err)

	signatureOne := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataOne,
	}

	sig, err = suite.privKey.Sign(append(sdk.Uint64ToBigEndian(suite.sequence), dataTwo...))
	suite.Require().NoError(err)

	signatureTwo := solomachinetypes.SignatureAndData{
		Signature: sig,
		Data:      dataTwo,
	}

	return solomachinetypes.Evidence{
		ClientID:     suite.clientID,
		Sequence:     suite.sequence,
		SignatureOne: signatureOne,
		SignatureTwo: signatureTwo,
	}
}
